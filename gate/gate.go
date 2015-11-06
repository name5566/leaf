package gate

import (
	"github.com/golang/protobuf/proto"
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"github.com/name5566/leaf/network/json"
	"github.com/name5566/leaf/network/protobuf"
	"reflect"
	"time"
)

type Gate struct {
	MaxConnNum        int
	PendingWriteNum   int
	MaxMsgLen         uint32
	JSONProcessor     *json.Processor
	ProtobufProcessor *protobuf.Processor
	AgentChanRPC      *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool
}

func (gate *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
			}
			return a
		}
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go("NewAgent", a)
			}
			return a
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

type agent struct {
	conn     network.Conn
	gate     *Gate
	userData interface{}
}

func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		if a.gate.JSONProcessor != nil {
			// json
			msg, err := a.gate.JSONProcessor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal json error: %v", err)
				break
			}
			err = a.gate.JSONProcessor.Route(msg, a)
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
			err = a.gate.ProtobufProcessor.Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *agent) OnClose() {
	if a.gate.AgentChanRPC != nil {
		err := a.gate.AgentChanRPC.Open(0).Call0("CloseAgent", a)
		if err != nil {
			log.Error("chanrpc error: %v", err)
		}
	}
}

func (a *agent) WriteMsg(msg interface{}) {
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

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) UserData() interface{} {
	return a.userData
}

func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
