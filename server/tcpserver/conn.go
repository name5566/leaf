package tcpserver

type Conn interface {
	Read() (
		// id
		interface{},
		// msg
		interface{},
		// err
		error,
	)
}
