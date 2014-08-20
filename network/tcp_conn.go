package network

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
}

func NewTCPConn(conn net.Conn, pendingWriteNum int) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)

	go func() {
		for b := range tcpConn.writeChan {
			_, err := conn.Write(b)
			if err != nil {
				tcpConn.Close()
				break
			}
		}
	}()

	return tcpConn
}

func (tcpConn *TCPConn) doClose() {
	tcpConn.conn.Close()
	close(tcpConn.writeChan)
	tcpConn.closeFlag = true
}

func (tcpConn *TCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	if tcpConn.closeFlag {
		return
	}

	tcpConn.doClose()
}

// b must not be modified by other goroutines
func (tcpConn *TCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	if tcpConn.closeFlag {
		return
	}
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doClose()
		return
	}

	tcpConn.writeChan <- b
}

func (tcpConn *TCPConn) CopyAndWrite(b []byte) {
	cb := make([]byte, len(b))
	copy(cb, b)
	tcpConn.Write(cb)
}

func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}
