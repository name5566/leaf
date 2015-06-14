package gate

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"github.com/name5566/leaf/network/json"
	"github.com/name5566/leaf/network/protobuf"
)

type TCPGate struct {
	Addr              string
	MaxConnNum        int
	PendingWriteNum   int
	LenMsgLen         int
	MinMsgLen         uint32
	MaxMsgLen         uint32
	LittleEndian      bool
	JSONProcessor     *json.Processor
	ProtobufProcessor *protobuf.Processor
}

func (gate *TCPGate) Run(closeSig chan bool) {
	server := new(network.TCPServer)
	server.Addr = gate.Addr
	server.MaxConnNum = gate.MaxConnNum
	server.PendingWriteNum = gate.PendingWriteNum
	server.LenMsgLen = gate.LenMsgLen
	server.MinMsgLen = gate.MinMsgLen
	server.MaxMsgLen = gate.MaxMsgLen
	server.LittleEndian = gate.LittleEndian
	server.NewAgent = func(conn *network.TCPConn) network.Agent {
		a := new(TCPAgent)
		a.conn = conn
		a.gate = gate
		return a
	}

	server.Start()
	<-closeSig
	server.Close()
}

func (gate *TCPGate) OnDestroy() {}

type TCPAgent struct {
	conn *network.TCPConn
	gate *TCPGate
}

func (a *TCPAgent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read msg error: %v", err)
			break
		}

		// json
		if a.gate.JSONProcessor != nil {
			msg, err := a.gate.JSONProcessor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal json error: %v", err)
				break
			}
			err = a.gate.JSONProcessor.Route(msg, a)
			if err != nil {
				log.Debug("route msg error: %v", err)
				break
			}
		}

		// protobuf
		if a.gate.ProtobufProcessor != nil {
			msg, err := a.gate.ProtobufProcessor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal protobuf error: %v", err)
				break
			}
			err = a.gate.ProtobufProcessor.Route(msg, a)
			if err != nil {
				log.Debug("route msg error: %v", err)
				break
			}
		}
	}
}

func (a *TCPAgent) OnClose() {

}
