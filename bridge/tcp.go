package bridge

type TCPAddr string

func (tcp TCPAddr) Network() string {
	return `tcp`
}

func (tcp TCPAddr) String() string {
	return string(tcp)
}
