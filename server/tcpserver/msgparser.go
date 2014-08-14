package tcpserver

import (
	"net"
)

type MsgParser interface {
	Parse(net.Conn) (interface{}, interface{}, error)
}
