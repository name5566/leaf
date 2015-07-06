package gate

import (
	"github.com/golang/protobuf/proto"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"github.com/name5566/leaf/network/json"
	"github.com/name5566/leaf/network/protobuf"
	"reflect"
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
	AgentChanRPC      *chanrpc.Server
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

		if gate.AgentChanRPC != nil {
			gate.AgentChanRPC.Go("NewAgent", a)
		}

		return a
	}

	server.Start()
	<-closeSig
	server.Close()
}

func (gate *TCPGate) OnDestroy() {}

type TCPAgent struct {
	conn     *network.TCPConn
	gate     *TCPGate
	userData interface{}
}

func (a *TCPAgent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message error: %v", err)
			break
		}

		if a.gate.JSONProcessor != nil {
			// json
			msg, err := a.gate.JSONProcessor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal json error: %v", err)
				break
			}
			err = a.gate.JSONProcessor.Route(msg, Agent(a))
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		} else if a.gate.ProtobufProcessor != nil {
			// protobuf
			msg, err := a.gate.ProtobufProcessor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal protobuf error: %v", err)
				break
			}
			err = a.gate.ProtobufProcessor.Route(msg, Agent(a))
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *TCPAgent) OnClose() {
	if a.gate.AgentChanRPC != nil {
		err := a.gate.AgentChanRPC.Open(0).Call0("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
}

func (a *TCPAgent) WriteMsg(msg interface{}) {
	if a.gate.JSONProcessor != nil {
		// json
		data, err := a.gate.JSONProcessor.Marshal(msg)
		if err != nil {
			log.Error("marshal json %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		a.conn.WriteMsg(data)
	} else if a.gate.ProtobufProcessor != nil {
		// protobuf
		id, data, err := a.gate.ProtobufProcessor.Marshal(msg.(proto.Message))
		if err != nil {
			log.Error("marshal protobuf %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		a.conn.WriteMsg(id, data)
	}
}

func (a *TCPAgent) Close() {
	a.conn.Close()
}

func (a *TCPAgent) UserData() interface{} {
	return a.userData
}

func (a *TCPAgent) SetUserData(data interface{}) {
	a.userData = data
}
