package network

type Agent interface {
	Run()
	OnClose()
}
