package tcpserver

import (
	"net"
)

type MsgReader interface {
	Read(net.Conn) (
		// id
		interface{},
		// msg
		interface{},
		// err
		error,
	)
}
