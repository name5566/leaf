package gateimpl

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

// config
const (
	serverAddr = ":8000"
	clientNum  = 100
)

// module
type Module struct {
	server *network.TCPServer
}

func (m *Module) OnInit() {
	// server
	m.server = &network.TCPServer{
		Addr:       serverAddr,
		MaxConnNum: clientNum,
	}

	// msg parser
	parser := network.NewMsgParser()

	// agent allocator
	m.server.NewAgent = func(conn *network.TCPConn) network.Agent {
		return &Agent{
			conn:   conn,
			parser: parser,
		}
	}

	log.Release("Gate module addr %v", m.server.Addr)
}

func (m *Module) OnDestroy() {
	m.server.Close()
	log.Release("Destroy the Gate module")
}

func (m *Module) Run(closeSig chan bool) {
	m.server.Start()
}
