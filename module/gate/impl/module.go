package gateimpl

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
)

type Module struct {
	server       *network.TCPServer
	Addr         string
	MaxConnNum   int
	LenMsgLen    int // 1 or 2 or 4
	LenMsgID     int // 1 or 2 or 4
	MaxMsgLen    uint32
	LittleEndian bool
}

func (m *Module) OnInit() {
	// LenMsgLen
	switch m.LenMsgLen {
	case 1:
	case 2:
	case 4:
	default:
		m.LenMsgLen = 2
		log.Release("invalid LenMsgLen, reset to %v", m.LenMsgLen)
	}

	// LenMsgID
	switch m.LenMsgID {
	case 1:
	case 2:
	case 4:
	default:
		m.LenMsgID = 2
		log.Release("invalid LenMsgID, reset to %v", m.LenMsgID)
	}

	// MaxMsgLen
	if m.MaxMsgLen == 0 {
		m.MaxMsgLen = 1024
		log.Release("invalid MaxMsgLen, reset to %v", m.MaxMsgLen)
	}

	// AgentConf
	conf := &AgentConf{
		lenMsgLen:    m.LenMsgLen,
		lenMsgID:     m.LenMsgID,
		maxMsgLen:    m.MaxMsgLen,
		littleEndian: m.LittleEndian,
	}

	// server
	m.server = &network.TCPServer{
		Addr:            m.Addr,
		MaxConnNum:      m.MaxConnNum,
		PendingWriteNum: 200,
		NewAgent: func(conn *network.TCPConn) network.Agent {
			return &Agent{
				conn: conn,
				conf: conf,
			}
		},
	}

	log.Release("Gate module addr %v", m.Addr)
}

func (m *Module) OnDestroy() {
	m.server.Close()
	log.Release("Destroy the Gate module")
}

func (m *Module) Run(closeSig chan bool) {
	m.server.Start()
}
