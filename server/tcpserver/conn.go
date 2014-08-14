package tcpserver

import (
	"net"
)

type Conn struct {
	baseConn net.Conn
}

func (conn *Conn) Run(reader MsgReader, dispatcher *MsgDispatcher) {
	for {
		// read
		id, msg, err := reader.Read(conn.baseConn)
		if err != nil {
			break
		}

		// dispatcher
		handler := dispatcher.Handler(id)
		if handler == nil {
			break
		}
		handler(conn, msg)
	}
}
