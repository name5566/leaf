package tcpserver

import (
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type Conn struct {
	sync.Mutex
	baseConn  net.Conn
	writeChan chan []byte
	closeFlag bool
}

func NewConn(baseConn net.Conn, pendingWriteNum int) *Conn {
	conn := new(Conn)
	conn.baseConn = baseConn
	conn.writeChan = make(chan []byte, pendingWriteNum)

	go func() {
		for b := range conn.writeChan {
			_, err := baseConn.Write(b)
			if err != nil {
				conn.Close()
				break
			}
		}
	}()

	return conn
}

func (conn *Conn) doClose() {
	conn.baseConn.Close()
	close(conn.writeChan)
	conn.closeFlag = true
}

func (conn *Conn) Close() {
	conn.Lock()
	defer conn.Unlock()

	if conn.closeFlag {
		return
	}

	conn.doClose()
}

func (conn *Conn) Write(b []byte) {
	conn.Lock()
	defer conn.Unlock()

	if conn.closeFlag {
		return
	}
	if len(conn.writeChan) == cap(conn.writeChan) {
		log.Debug("close conn: channel full")
		conn.doClose()
		return
	}

	conn.writeChan <- b
}

func (conn *Conn) Read(b []byte) (int, error) {
	return conn.baseConn.Read(b)
}
