package configure

type Pool struct {
	// Multiple memory pools with the same tag will be merged
	Tag string
	// Memory block size
	Cap int
	// How many blocks to cache
	Len int
}
