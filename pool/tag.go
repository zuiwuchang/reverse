package pool

import "log/slog"

var pools = make(map[string]*Pool)

func Get(tag string, cap, len int) *Pool {
	found, ok := pools[tag]
	if !ok {
		if cap == 0 {
			cap = 1024 * 128
		} else if cap < 0 {
			cap = 1024 * 32
		}
		if len == 0 {
			len = 8 * 10
		} else if len < 0 {
			len = 0
		}
		found = New(cap, len)
		pools[tag] = found
		slog.Info(`new pool`,
			`tag`, tag,
			`cap`, cap,
			`len`, len,
		)
	}
	return found
}
