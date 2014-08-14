package tcpserver

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type Server struct {
	Addr         string
	MaxConnNum   int
	NewMsgParser func() MsgParser
	ln           net.Listener
	conns        ConnSet
	mutexConns   sync.Mutex
	wg           sync.WaitGroup
	closeFlag    bool
	dispatcher   MsgDispatcher
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
		go server.handle(conn)
	}
}

func (server *Server) handle(conn net.Conn) {
	agent := Agent{
		conn,
		server.NewMsgParser(),
		&server.dispatcher,
	}
	agent.Run()

	conn.Close()
	server.mutexConns.Lock()
	delete(server.conns, conn)
	server.mutexConns.Unlock()

	server.wg.Done()
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

func (server *Server) HandleFunc(id interface{}, handler Handler) {
	server.dispatcher.HandleFunc(id, handler)
}
