package configure

type Bridge struct {
	// connect portal address
	Portal string
	// 'portal' validates the token to determine if it is legitimate
	Token string
	// memory pool
	Pool Pool
}

func LoadBridge(filename string) (cnfs []Bridge, e error) {
	var tmp []Bridge
	e = loadObject(filename, &tmp)
	if e != nil {
		return
	}
	cnfs = tmp
	return
}
