package main

import (
	"github.com/name5566/leaf"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network"
	"time"
)

// config
const (
	serverAddr = "127.0.0.1:8000"
	clientNum  = 1
)

// module
type Module struct {
	client *network.TCPClient
}

func (m *Module) OnInit() {
	// client
	m.client = &network.TCPClient{
		Addr:    serverAddr,
		ConnNum: clientNum,
	}

	// msg parser
	parser := network.NewMsgParser()

	// agent allocator
	m.client.NewAgent = func(conn *network.TCPConn) network.Agent {
		a := &Agent{
			conn:   conn,
			parser: parser,
		}

		// msg "hi"
		parser.Write(conn, []byte("hi"))

		return a
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
	conn   *network.TCPConn
	parser *network.MsgParser
}

func (a *Agent) Run() {
	for {
		// read a msg
		data, err := a.parser.Read(a.conn)
		if err != nil {
			log.Debug("Network error: %v", err)
			break
		}

		log.Debug("msg: %s", data)

		// echo the msg
		a.parser.Write(a.conn, data)

		time.Sleep(time.Second)
	}
}

func (a *Agent) OnClose() {

}

// main
func main() {
	leaf.Run(new(Module))
}
