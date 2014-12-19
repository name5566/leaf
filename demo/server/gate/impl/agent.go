package gateimpl

import (
	"github.com/name5566/leaf/demo/server/echo"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

type Agent struct {
	conn *network.TCPConn
}

func newAgent(conn *network.TCPConn) network.Agent {
	return &Agent{conn}
}

func (a *Agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("Network error: %v", err)
			break
		}

		// dispatch the msg
		echo.R.AsynCall0("echo", a.conn, data)
	}
}

func (a *Agent) OnClose() {}
