package gateimpl

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

type Module struct {
	server *network.TCPServer
}

func (m *Module) OnInit() {
	m.server = &network.TCPServer{
		Addr:            ":8000",
		MaxConnNum:      10000,
		PendingWriteNum: 100,
		NewAgent:        newAgent,
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
