package gateimpl

import (
	"github.com/name5566/leaf/demo/server/echo"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

type Agent struct {
	conn   *network.TCPConn
	parser *network.MsgParser
}

func (a *Agent) Run() {
	for {
		data, err := a.parser.Read(a.conn)
		if err != nil {
			log.Debug("Network error: %v", err)
			break
		}

		echo.R.Call0("echo", a.conn, a.parser, data)
	}
}

func (a *Agent) OnClose() {

}
