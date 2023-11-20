package pool

import "sync"

var bytes sync.Pool

func init() {
	bytes.New = func() any {
		return make([]byte, 1024*32)
	}
}
func GetBytes() []byte {
	return bytes.Get().([]byte)
}
func PutBytes(b []byte) {
	bytes.Put(b)
}
