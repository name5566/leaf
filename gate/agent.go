package gate

type Agent interface {
	WriteMsg(msg interface{})
	Close()
}
