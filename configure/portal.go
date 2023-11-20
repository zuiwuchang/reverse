package configure

type Portal struct {
	// listen addr, 'bridge' will connect this addr
	Addr string

	// 'bridge' needs to pass in this token to verify that it is legitimate
	Token string

	// The reused transmission channel will no longer transmit data for new connections after the specified number of seconds
	MaxSeconds int64
	// The reused transmission channel will no longer transmit data for new connections after specifying MB data.
	MaxMB uint64

	// Data forwarding target
	Forwards []Forward
}
type Forward struct {
	// listen addr, this is the data source
	From string

	// 'bridge' will connect this address and serve as the data forwarding target
	To string
}

func LoadPortal(filename string) (cnfs []Portal, e error) {
	var tmp []Portal
	e = loadObject(filename, &tmp)
	if e != nil {
		return
	}
	cnfs = tmp
	return
}
