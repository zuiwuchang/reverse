package portal

import (
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/zuiwuchang/reverse/configure"
)

type Forword struct {
	l     net.Listener
	done  <-chan struct{}
	ch    chan<- chan *Client
	recv  chan *Client
	to    string
	token string
}

func newForword(opts *configure.Forward, token string, done <-chan struct{}, ch chan<- chan *Client) (f *Forword, e error) {
	l, e := net.Listen(`tcp`, opts.From)
	if e != nil {
		return
	}
	f = &Forword{
		l:     l,
		done:  done,
		ch:    ch,
		recv:  make(chan *Client),
		to:    opts.To,
		token: token,
	}
	return
}
func (f *Forword) Serve() {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		c, e := f.l.Accept()
		if e != nil {
			select {
			// already close
			case <-f.done:
				return
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				slog.Warn(`Accept error`,
					`error`, e,
					`retrying in`, tempDelay,
				)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		go f.connect(c)
	}
}
func (f *Forword) connect(c net.Conn) {
	defer c.Close()
	select {
	case <-f.done:
		return
	case f.ch <- f.recv:
	}

	select {
	case <-f.done:
		return
	case client := <-f.recv:
		resp, e := f.connectH2C(client, c)
		if e != nil {
			slog.Warn(`connect h2c fail`,
				`error`, e,
			)
			return
		}
		io.Copy(c, resp.Body)
		time.Sleep(time.Second)
		c.Close()
		resp.Body.Close()
	}
}
func (f *Forword) connectH2C(client *Client, c net.Conn) (resp *http.Response, e error) {
	req, e := http.NewRequest(http.MethodPost, `http://127.0.0.1/video/live`, c)
	if e != nil {
		return
	}
	req.Header.Set(`target`, base64.RawURLEncoding.EncodeToString([]byte(f.to)))
	if f.token != `` {
		req.Header.Set(`Authorization`, `Bearer `+f.token)
	}
	resp, e = client.client.Do(req)
	if e != nil {
		return
	}
	if resp.StatusCode != http.StatusCreated {
		var b []byte
		if resp.Body != nil {
			b, _ = io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
		}
		if len(b) == 0 {
			e = errors.New(resp.Status)
		} else {
			e = errors.New(resp.Status + ` ` + string(b))
		}
		return
	} else if resp.Body == nil {
		e = errors.New(`body nil`)
		return
	}
	return
}
