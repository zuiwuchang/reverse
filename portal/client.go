package portal

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/powerpuffpenguin/vnet/reverse"
	"golang.org/x/net/http2"
)

type Client struct {
	send    *uint64
	recv    *uint64
	expired *int64

	client *http.Client
}

func newClient(dialer *reverse.Dialer,
	send, recv *uint64,
	expired *int64,
	maxseconds int64,
) *Client {
	return &Client{
		send:    send,
		recv:    recv,
		expired: expired,
		client: &http.Client{
			Transport: &http2.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					c, e := dialer.DialContext(ctx, network, addr)
					if e != nil {
						e0 := ctx.Err()
						if e0 != nil {
							return nil, e
						}

						c, e0 = dialer.DialContext(ctx, network, addr)
						if e0 != nil {
							return nil, e
						}

						e = nil
					}

					if expired != nil && maxseconds > 0 {
						atomic.StoreInt64(expired, time.Now().Unix()+maxseconds)
					}
					if send != nil || recv != nil {
						if send != nil {
							atomic.StoreUint64(send, 0)
						}
						if recv != nil {
							atomic.StoreUint64(recv, 0)
						}
						c = &Conn{
							Conn: c,
							send: send,
							recv: recv,
						}
					}
					return c, e
				},
				AllowHTTP: true,
			},
		},
	}
}

type Conn struct {
	net.Conn
	send *uint64
	recv *uint64
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 && c.recv != nil {
		atomic.AddUint64(c.recv, uint64(n))
	}
	return
}
func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 && c.send != nil {
		atomic.AddUint64(c.send, uint64(n))
	}
	return
}
