package tcpserver

import (
	"net"
)

type Agent struct {
	conn       net.Conn
	parser     MsgParser
	dispatcher *MsgDispatcher
}

func (agent *Agent) Run() {
	for {
		id, msg, err := agent.parser.Parse(agent.conn)
		if err != nil {
			break
		}

		handler := agent.dispatcher.Handler(id)
		if handler == nil {
			break
		}
		handler(agent, msg)
	}
}
