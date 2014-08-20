package network

type Agent interface {
	Read() (
		// id
		interface{},
		// msg
		interface{},
		// err
		error,
	)

	OnClose()
}
