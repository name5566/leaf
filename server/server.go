package server

type Server interface {
	Start()
	Close()
}
