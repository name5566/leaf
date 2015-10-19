package network

import (
	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"net"
	"sync"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	closeFlag bool
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)

	go func() {
		for b := range wsConn.writeChan {
			if b == nil {
				break
			}

			err := conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
			}
		}

		conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()
	close(wsConn.writeChan)
	wsConn.closeFlag = true
}

func (wsConn *WSConn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doDestroy()
}

func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doWrite(nil)
	wsConn.closeFlag = true
}

func (wsConn *WSConn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Debug("close conn: channel full")
		wsConn.doDestroy()
		return
	}

	wsConn.writeChan <- b
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// goroutine not safe
func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	_, b, err := wsConn.conn.ReadMessage()
	return b, err
}

// b must not be modified by other goroutines
func (wsConn *WSConn) WriteMsg(b []byte) {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag || b == nil {
		return
	}

	wsConn.doWrite(b)
}
