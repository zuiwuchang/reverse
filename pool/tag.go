package pool

import "log/slog"

var pools = make(map[string]*Pool)

func Get(tag string, cap, len int) *Pool {
	found, ok := pools[tag]
	if !ok {
		if cap < 1 {
			cap = 1024 * 128
			len = 8 * 10
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
