package network

import (
	_ "github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"net"
	"net/http"
	_ "sync"
	"time"
)

type WSServer struct {
	Addr        string
	HTTPTimeout time.Duration
}

type WSHandler struct {
	*WSServer
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Release("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        &WSHandler{server},
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go func() {
		err := httpServer.Serve(ln)
		if err != nil {
			log.Error("http serve error: %v", err)
		}
	}()
}

func (server *WSServer) Close() {

}
