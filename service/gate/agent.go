package gate

import (
	"net"
)

type AgentMgr interface {
	NewAgent() Agent
}

type Agent interface {
	Main(net.Conn)
}
