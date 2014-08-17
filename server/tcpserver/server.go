package tcpserver

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type Server struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(*Conn) Agent
	ln              net.Listener
	conns           ConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
	closeFlag       bool
	disp            Dispatcher
}

type ConnSet map[net.Conn]struct{}

func (server *Server) Start() {
	server.init()
	go server.run()
}

func (server *Server) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.ln = ln
	server.conns = make(ConnSet)
	server.closeFlag = false
}

func (server *Server) run() {
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if server.closeFlag {
				return
			} else {
				log.Error("accept error: %v", err)
				continue
			}
		}

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Debug("too many connections")
			continue
		}
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wg.Add(1)

		// net conn wrapper
		c := NewConn(conn, server.PendingWriteNum)
		agent := server.NewAgent(c)
		go func() {
			server.handle(agent)

			// cleanup
			c.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()

			server.wg.Done()
		}()
	}
}

func (server *Server) handle(agent Agent) {
	for {
		id, msg, err := agent.Read()
		if err != nil {
			break
		}

		handler := server.disp.Handler(id)
		if handler == nil {
			break
		}
		handler(agent, msg)
	}

	agent.OnClose()
}

func (server *Server) Close() {
	server.closeFlag = true
	server.ln.Close()

	server.mutexConns.Lock()
	for conn, _ := range server.conns {
		conn.Close()
	}
	server.conns = make(ConnSet)
	server.mutexConns.Unlock()

	server.wg.Wait()
}

func (server *Server) RegHandler(id interface{}, handler Handler) {
	server.disp.RegHandler(id, handler)
}
