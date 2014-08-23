package network

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	Addr              string
	ReconnectInterval time.Duration
	PendingWriteNum   int
	NewAgent          func(*TCPConn) Agent
	wg                sync.WaitGroup
	tcpConn           *TCPConn
	disp              Dispatcher
}

func (client *TCPClient) Start() {
	client.init()
	go client.run()
}

func (client *TCPClient) init() {
	if client.ReconnectInterval == 0 {
		client.ReconnectInterval = 3 * time.Second
		log.Release("invalid ReconnectInterval, reset to %v", client.ReconnectInterval)
	}
	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", client.PendingWriteNum)
	}
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	var conn net.Conn
	var err error
	for conn == nil {
		conn, err = net.Dial("tcp", client.Addr)
		if err != nil {
			time.Sleep(client.ReconnectInterval)
			log.Release("connect to %v error: %v", client.Addr, err)
			continue
		}
	}

	client.wg.Add(1)
	client.tcpConn = NewTCPConn(conn, client.PendingWriteNum)
}

func (client *TCPClient) run() {
	agent := client.NewAgent(client.tcpConn)

	for {
		id, msg, err := agent.Read()
		if err != nil {
			break
		}

		handler := client.disp.Handler(id)
		if handler == nil {
			break
		}
		handler(agent, msg)
	}

	agent.OnClose()
	client.wg.Done()
}

func (client *TCPClient) Close() {
	client.tcpConn.Close()
	client.wg.Wait()
}

func (client *TCPClient) RegHandler(id interface{}, handler Handler) {
	client.disp.RegHandler(id, handler)
}
