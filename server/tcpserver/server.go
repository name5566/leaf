package tcpserver

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type Server struct {
	Addr         string
	MaxConnNum   int
	NewMsgReader func() MsgReader
	ln           net.Listener
	conns        ConnSet
	mutexConns   sync.Mutex
	wg           sync.WaitGroup
	closeFlag    bool
	disp         Dispatcher
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
	if server.NewMsgReader == nil {
		log.Fatal("NewMsgReader must not be nil")
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

func (server *Server) handle(baseConn net.Conn) {
	conn := Conn{baseConn}
	conn.Run(server.NewMsgReader(), &server.disp)

	baseConn.Close()
	server.mutexConns.Lock()
	delete(server.conns, baseConn)
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

func (server *Server) RegHandler(id interface{}, handler Handler) {
	server.disp.RegHandler(id, handler)
}
