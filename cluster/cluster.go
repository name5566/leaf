package cluster

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/network"
	"math"
	"time"
	"reflect"
	"net"
)

var (
	server  *network.TCPServer
	clients []*network.TCPClient
	agents = map[string]*Agent{}
)

func Init() {
	if conf.ListenAddr != "" {
		server = new(network.TCPServer)
		server.Addr = conf.ListenAddr
		server.MaxConnNum = int(math.MaxInt32)
		server.PendingWriteNum = conf.PendingWriteNum
		server.LenMsgLen = 4
		server.MaxMsgLen = math.MaxUint32
		server.NewAgent = newAgent

		server.Start()
	}

	for _, addr := range conf.ConnAddrs {
		client := new(network.TCPClient)
		client.Addr = addr
		client.ConnectInterval = 3 * time.Second
		client.PendingWriteNum = conf.PendingWriteNum
		client.LenMsgLen = 4
		client.MaxMsgLen = math.MaxUint32
		client.NewAgent = newAgent

		client.Start()
		clients = append(clients, client)
	}
}

func GetAgent(serverName string) *Agent {
	agent, ok := agents[serverName]
	if ok {
		return agent
	} else {
		return nil
	}
}

func Destroy() {
	if server != nil {
		server.Close()
	}

	for _, client := range clients {
		client.Close()
	}
}

type Agent struct {
	ServerName	string
	conn       	*network.TCPConn
	userData 	interface{}
}

func newAgent(conn *network.TCPConn) network.Agent {
	a := new(Agent)
	a.conn = conn

	msg := &S2S_NotifyServerName{ServerName:conf.ServerName}
	a.WriteMsg(msg)
	return a
}

func (a *Agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		if Processor != nil {
			msg, err := Processor.Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			err = Processor.Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *Agent) OnClose() {}

func (a *Agent) WriteMsg(msg interface{}) {
	if Processor != nil {
		data, err := Processor.Marshal(msg)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		}
	}
}

func (a *Agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Agent) Close() {
	a.conn.Close()
}

func (a *Agent) Destroy() {
	a.conn.Destroy()
}

func (a *Agent) UserData() interface{} {
	return a.userData
}

func (a *Agent) SetUserData(data interface{}) {
	a.userData = data
}
