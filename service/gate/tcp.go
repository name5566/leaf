package gate

import (
	"errors"
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type TcpGateCfg struct {
	addr       string
	maxConnNum int
	agentMgr   AgentMgr
}

type ConnSet map[net.Conn]struct{}

type TcpGate struct {
	ln         net.Listener
	maxConnNum int
	conns      ConnSet
	mutexConns sync.Mutex
	agentMgr   AgentMgr
	wg         sync.WaitGroup
	running    bool
}

func NewTcpGate(cfg *TcpGateCfg) (*TcpGate, error) {
	if cfg == nil {
		return nil, errors.New("cfg must not be nil")
	}
	if cfg.agentMgr == nil {
		return nil, errors.New("agentMgr must not be nil")
	}

	ln, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return nil, err
	}

	tcpGate := new(TcpGate)
	tcpGate.ln = ln
	tcpGate.maxConnNum = cfg.maxConnNum
	tcpGate.conns = make(ConnSet)
	tcpGate.agentMgr = cfg.agentMgr
	tcpGate.running = true
	return tcpGate, nil
}

func (tcpGate *TcpGate) Start() {
	go func() {
		for {
			// accept conn
			conn, err := tcpGate.ln.Accept()
			if err != nil {
				if !tcpGate.running {
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
			tcpGate.wg.Add(1)
			go tcpGate.handleConn(conn)
		}
	}()
}

func (tcpGate *TcpGate) Close() {
	tcpGate.running = false
	tcpGate.ln.Close()

	tcpGate.mutexConns.Lock()
	for conn, _ := range tcpGate.conns {
		conn.Close()
	}
	tcpGate.conns = make(ConnSet)
	tcpGate.mutexConns.Unlock()

	tcpGate.wg.Wait()
}

func (tcpGate *TcpGate) handleConn(conn net.Conn) {
	agent := tcpGate.agentMgr.NewAgent()
	agent.Main(conn)

	conn.Close()
	tcpGate.mutexConns.Lock()
	delete(tcpGate.conns, conn)
	tcpGate.mutexConns.Unlock()

	tcpGate.wg.Done()
}
