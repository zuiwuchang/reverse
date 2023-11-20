package bridge

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"log/slog"

	"github.com/powerpuffpenguin/vnet/reverse"
	"github.com/zuiwuchang/reverse/configure"
	"github.com/zuiwuchang/reverse/pool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var ErrClosed = errors.New(`bridge already closed`)

type Bridge struct {
	l      Listener
	server http.Server
	token  string
}

func New(opts *configure.Bridge) (b *Bridge, e error) {
	l := reverse.Listen(TCPAddr(opts.Portal))

	b = &Bridge{
		l:     Listener{Listener: l},
		token: opts.Token,
	}
	var http2Server http2.Server
	mux := http.NewServeMux()
	mux.HandleFunc(`/video/live`, b.handler)
	b.server.Handler = h2c.NewHandler(mux, &http2Server)
	e = http2.ConfigureServer(&b.server, &http2Server)
	if e != nil {
		return
	}
	slog.Info(`net bridge`,
		`portal`, opts.Portal,
	)
	return
}

func (b *Bridge) Serve() error {
	return b.server.Serve(b.l)
}
func (bridge *Bridge) handler(w http.ResponseWriter, r *http.Request) {
	if bridge.token != `` {
		ok := false
		for _, val := range r.Header.Values(`Authorization`) {
			if strings.HasPrefix(val, `Bearer `) && val[7:] == bridge.token {
				ok = true
				break
			}
		}
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	if r.ProtoMajor < 2 || r.Method != http.MethodPost || r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target := r.Header.Get(`target`)
	b, e := base64.RawURLEncoding.DecodeString(target)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set(`Content-Type`, `text/plain; charset=utf-8`)
		w.Write([]byte(e.Error()))
		return
	}
	rawURL := string(b)
	u, e := url.ParseRequestURI(rawURL)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set(`Content-Type`, `text/plain; charset=utf-8`)
		w.Write([]byte(e.Error()))
		return
	} else if u.Scheme != `tcp` {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set(`Content-Type`, `text/plain; charset=utf-8`)
		w.Write([]byte(`not support to scheme: ` + u.Scheme))
		return
	}

	ctx := r.Context()
	var dialer net.Dialer
	c, e := dialer.DialContext(ctx, `tcp`, u.Host)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set(`Content-Type`, `text/plain; charset=utf-8`)
		w.Write([]byte(e.Error()))
		return
	}
	defer c.Close()

	slog.Info(`bridge forword`, `to`, rawURL)
	forwordH2C(w, r, c)

	// waiting to read unread data in the network
	time.Sleep(time.Second)
}
func forwordH2C(w http.ResponseWriter, r *http.Request, c net.Conn) {
	w.WriteHeader(http.StatusCreated)
	f := w.(http.Flusher)
	f.Flush()

	newForword(w, r, c).Serve()

	// done := make(chan int, 2)
	// go func() {
	// 	var (
	// 		b      = pool.GetBytes()
	// 		n      int
	// 		er, ew error
	// 	)
	// 	for {
	// 		n, er = c.Read(b)
	// 		if n > 0 {
	// 			_, ew = w.Write(b[:n])
	// 			if ew != nil {
	// 				break
	// 			}
	// 			f.Flush()
	// 		}
	// 		if er != nil {
	// 			break
	// 		}
	// 	}
	// 	pool.PutBytes(b)
	// 	done <- 1
	// }()

	// go func() {
	// 	var (
	// 		b      = pool.GetBytes()
	// 		n      int
	// 		er, ew error
	// 	)
	// 	for {
	// 		n, er = r.Body.Read(b)
	// 		if n > 0 {
	// 			_, ew = c.Write(b[:n])
	// 			if ew != nil {
	// 				break
	// 			}
	// 		}
	// 		if er != nil {
	// 			break
	// 		}
	// 	}
	// 	pool.PutBytes(b)
	// 	done <- 2
	// }()
	// // wait any error
	// <-done
}

type _Forword struct {
	w      http.ResponseWriter
	r      *http.Request
	c      net.Conn
	done   chan struct{}
	closed uint32

	chr chan []byte
	chw chan []byte
}

func newForword(w http.ResponseWriter, r *http.Request, c net.Conn) *_Forword {
	return &_Forword{
		w:    w,
		r:    r,
		c:    c,
		done: make(chan struct{}),
		chr:  make(chan []byte, 32),
		chw:  make(chan []byte, 32),
	}
}
func (f *_Forword) Close() {
	if f.closed == 0 && atomic.CompareAndSwapUint32(&f.closed, 0, 1) {
		close(f.done)
	}
}
func (f *_Forword) Serve() {
	go f.read(f.chr, f.c)
	go f.read(f.chw, f.r.Body)

	go f.write(f.w, f.chr)
	go f.write(f.c, f.chw)

	<-f.done
}
func (f *_Forword) read(w chan []byte, r io.Reader) {
	for {
		b := pool.GetBytes()
		n, e := r.Read(b)
		if e != nil {
			break
		}
		if n > 0 {
			select {
			case <-f.done:
				return
			case w <- b[:n]:
			}
		}
	}
	f.Close()
}
func (f *_Forword) write(w io.Writer, r chan []byte) {
	flusher, ok := w.(http.Flusher)
	if ok {
		var bs [][]byte
		var merge int
		for {
			select {
			case <-f.done:
				return
			case b := <-r:
				bs = append(bs, b)
			}
		MERGE:
			for merge < 1024*1024 {
				select {
				case <-f.done:
					for _, b := range bs {
						pool.PutBytes(b)
					}
					return
				case b := <-r:
					bs = append(bs, b)
					merge += len(b)
				default:
					break MERGE
				}
			}

			for _, b := range bs {
				_, e := w.Write(b)
				if e != nil {
					f.Close()
					for _, b := range bs {
						pool.PutBytes(b)
					}
					return
				}
			}
			flusher.Flush()

			for _, b := range bs {
				pool.PutBytes(b)
			}
		}
	} else {
		for {
			select {
			case <-f.done:
				return
			case b := <-r:
				_, e := w.Write(b)
				pool.PutBytes(b)
				if e != nil {
					f.Close()
					return
				}
			}
		}
	}
}
