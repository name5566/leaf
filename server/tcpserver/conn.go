package tcpserver

import (
	"net"
)

type Conn struct {
	baseConn net.Conn
}

func (conn *Conn) Run() {
	for {

	}
}
