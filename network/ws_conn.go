package network

import (
	"github.com/gorilla/websocket"
	_ "github.com/name5566/leaf/log"
	_ "sync"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
}
