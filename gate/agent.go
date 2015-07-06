package gate

type Agent interface {
	WriteMsg(msg interface{})
	Close()
	UserData() interface{}
	SetUserData(data interface{})
}
