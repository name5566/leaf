package main

import (
	"github.com/name5566/leaf"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"time"
)

// module
type Module struct {
	client *network.TCPClient
}

func (m *Module) OnInit() {
	m.client = &network.TCPClient{
		Addr:            "127.0.0.1:8000",
		ConnNum:         1,
		ConnectInterval: time.Second,
		PendingWriteNum: 100,
		NewAgent:        newAgent,
	}
}

func (m *Module) OnDestroy() {
	m.client.Close()
}

func (m *Module) Run(closeSig chan bool) {
	m.client.Start()
}

// agent
type Agent struct {
	conn *network.TCPConn
}

func newAgent(conn *network.TCPConn) network.Agent {
	a := new(Agent)
	a.conn = conn
	conn.WriteMsg([]byte("My name is Leaf"))
	return a
}

func (a *Agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("Network error: %v", err)
			break
		}

		log.Debug("Echo: %s", data)

		a.conn.WriteMsg(data)
		time.Sleep(time.Second)
	}
}

func (a *Agent) OnClose() {

}

// main
func main() {
	leaf.Run(new(Module))
}
