package gate

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

// you must implement this interface
type TcpGateExtension interface {
}

type TcpGate struct {
	ext        TcpGateExtension
	ln         net.Listener
	maxConnNum int
	conns      map[net.Conn]*ConnContext
	lConns     sync.RWMutex
}

type ConnContext struct {
}

func NewTcpGate(ext TcpGateExtension, laddr string, maxConnNum int) (*TcpGate, error) {
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}

	tcpGate := new(TcpGate)
	tcpGate.ext = ext
	tcpGate.ln = ln
	tcpGate.maxConnNum = maxConnNum
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
			tcpGate.lConns.Lock()
			if len(tcpGate.conns) >= tcpGate.maxConnNum {
				tcpGate.lConns.Unlock()
				conn.Close()
				log.Error("too many connections (%v)", tcpGate.maxConnNum)
				continue
			}
			tcpGate.conns[conn] = new(ConnContext)
			tcpGate.lConns.Unlock()

			// handle conn
			go tcpGate.handleConn(conn)
		}
	}()
}

func (tcpGate *TcpGate) Close() {
	tcpGate.ln.Close()

	tcpGate.lConns.Lock()
	for conn, _ := range tcpGate.conns {
		conn.Close()
	}
	tcpGate.conns = make(map[net.Conn]*ConnContext)
	tcpGate.lConns.Unlock()
}

func (tcpGate *TcpGate) handleConn(conn net.Conn) {

}
