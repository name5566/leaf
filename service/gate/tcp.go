package gate

import (
	"net"
)

const (
	name    = "tcpgate"
	defAddr = ":8080"
)

type TcpGate struct {
	Addr string
	ln   net.Listener
}

func NewTcpGate() (*TcpGate, error) {

}

func (tcpGate *TcpGate) Name() string {
	return name
}

func (tcpGate *TcpGate) Start() {
	if tcpGate.Addr == "" {
		tcpGate.Addr = defAddr
	}

	ln, err := net.Listen("tcp", tcpGate.Addr)
	if err != nil {
		panic(err)
	}
	tcpGate.ln = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			if err.Error() == "use of closed network connection" {
				break
			} else {
				continue
			}
		}
		go handleConn(conn)
	}
}

func (tcpGate *TcpGate) Close() {
	tcpGate.ln.Close()
}

func handleConn(conn net.Conn) {

}
