package network

type Server interface {
	Start()
	Close()
}
