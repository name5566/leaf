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
	Agent             Agent
	conn              net.Conn
	wg                sync.WaitGroup
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
	if client.Agent == nil {
		log.Fatal("Agent must not be nil")
	}

	for client.conn == nil {
		conn, err := net.Dial("tcp", client.Addr)
		if err != nil {
			time.Sleep(client.ReconnectInterval)
			log.Release("connect to %v error: %v", client.Addr, err)
			continue
		}
		client.conn = conn
	}

	tcpConn := NewTCPConn(conn, server.PendingWriteNum)
	agent := server.NewAgent(tcpConn)
	go func() {
		server.handle(agent)

		// cleanup
		tcpConn.Close()
		server.mutexConns.Lock()
		delete(server.conns, conn)
		server.mutexConns.Unlock()

		server.wg.Done()
	}()
}

func (client *TCPClient) run() {

}

func (client *TCPClient) Close() {

}

func (client *TCPClient) RegHandler(id interface{}, handler Handler) {

}
