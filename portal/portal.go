package portal

import (
	"errors"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/powerpuffpenguin/vnet/reverse"
	"github.com/zuiwuchang/reverse/configure"
)

var ErrClosed = errors.New(`portal already closed`)

type Portal struct {
	addr     string
	dialer   *reverse.Dialer
	closed   uint32
	done     chan struct{}
	forwords []*Forword
	ch       chan chan *Client

	maxsize    uint64
	maxseconds int64
}

func New(opts *configure.Portal) (p *Portal, e error) {
	count := len(opts.Forwards)
	if count == 0 {
		e = errors.New(`len(forwards) = 0`)
		return
	}

	l, e := net.Listen(`tcp`, opts.Addr)
	if e != nil {
		return
	}
	slog.Info(`portal listen`,
		`addr`, opts.Addr,
	)

	done := make(chan struct{})
	forwords := make([]*Forword, 0, count)
	ch := make(chan chan *Client)
	var forword *Forword
	for i := 0; i < count; i++ {
		forword, e = newForword(&opts.Forwards[i], opts.Token, done, ch)
		if e != nil {
			close(done)
			l.Close()
			for _, forword := range forwords {
				forword.l.Close()
			}
			return
		}
		slog.Info(`portal forword`,
			`listen`, opts.Forwards[i].From,
			`bridge`, opts.Forwards[i].To,
		)
		forwords = append(forwords, forword)
	}

	p = &Portal{
		addr:     opts.Addr,
		dialer:   reverse.NewDialer(l),
		done:     done,
		forwords: forwords,
		ch:       ch,

		maxsize:    opts.MaxMB * 1024 * 1024,
		maxseconds: opts.MaxSeconds,
	}
	return
}
func (p *Portal) Close() error {
	if p.closed == 0 && atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		close(p.done)
		p.dialer.Close()
		for _, forword := range p.forwords {
			forword.l.Close()
		}
		return nil
	}
	return ErrClosed
}
func (p *Portal) neednew(client *Client) bool {
	if client.expired != nil {
		expired := atomic.LoadInt64(client.expired)
		if time.Now().Unix() >= expired {
			return true
		}
	}

	if p.maxsize == 0 {
		return false
	}
	max := atomic.LoadUint64(client.recv)
	min := atomic.LoadUint64(client.send)
	if max < min {
		max = min
	}
	return max >= p.maxsize
}
func (p *Portal) Serve() {
	go p.dialer.Serve()
	for _, f := range p.forwords {
		go f.Serve()
	}
	var (
		ch     chan *Client
		client *Client
	)
	for {
		select {
		case <-p.done:
			return
		case ch = <-p.ch:
		}
		if client == nil || p.neednew(client) {
			var (
				send, recv *uint64
				at         *int64
			)
			if p.maxsize != 0 {
				send = new(uint64)
				recv = new(uint64)
			}
			if p.maxseconds > 0 {
				at = new(int64)
			}
			client = newClient(p.dialer, send, recv, at, p.maxseconds)
			slog.Info(`new client`,
				`Portal`, p.addr,
				`MaxMB`, p.maxsize/1024/1024,
				`MaxSeconds`, p.maxseconds,
			)
		}
		select {
		case <-p.done:
			return
		case ch <- client:
		}
	}
}
