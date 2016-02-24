package gate

type Agent interface {
	WriteMsg(msg interface{})
	Close()
	Destroy()
	UserData() interface{}
	SetUserData(data interface{})
}
