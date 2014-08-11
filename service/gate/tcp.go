package gate

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type ConnSet map[net.Conn]struct{}

type TcpGate struct {
	ln         net.Listener
	maxConnNum int
	conns      ConnSet
	mutexConns sync.Mutex
	agentMgr   AgentMgr
}

func NewTcpGate(laddr string, maxConnNum int, agentMgr AgentMgr) (*TcpGate, error) {
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}

	tcpGate := new(TcpGate)
	tcpGate.ln = ln
	tcpGate.maxConnNum = maxConnNum
	tcpGate.conns = make(ConnSet)
	tcpGate.agentMgr = agentMgr
	return tcpGate, nil
}

func (tcpGate *TcpGate) Start() {
	go func() {
		for {
			// accept conn
			conn, err := tcpGate.ln.Accept()
			if err != nil {
				if err.Error() == "use of closed network connection" {
					log.Release("tcp gate closed")
					return
				} else {
					log.Error("accept error: %v", err)
					continue
				}
			}

			// conns
			tcpGate.mutexConns.Lock()
			if len(tcpGate.conns) >= tcpGate.maxConnNum {
				tcpGate.mutexConns.Unlock()
				conn.Close()
				log.Error("too many connections")
				continue
			}
			tcpGate.conns[conn] = struct{}{}
			tcpGate.mutexConns.Unlock()

			// handle conn
			go tcpGate.handleConn(conn)
		}
	}()
}

func (tcpGate *TcpGate) Close() {
	tcpGate.ln.Close()

	tcpGate.mutexConns.Lock()
	for conn, _ := range tcpGate.conns {
		conn.Close()
	}
	tcpGate.conns = make(ConnSet)
	tcpGate.mutexConns.Unlock()
}

func (tcpGate *TcpGate) handleConn(conn net.Conn) {
	agent := tcpGate.agentMgr.NewAgent()
	agent.Main(conn)

	conn.Close()

	tcpGate.mutexConns.Lock()
	delete(tcpGate.conns, conn)
	tcpGate.mutexConns.Unlock()
}
