package tcpserver

import (
	"github.com/name5566/leaf/log"
	"sync"
)

type Dispatcher struct {
	sync.RWMutex
	// id -> handler
	handlers map[interface{}]Handler
}

type Handler func(conn Conn, msg interface{})

func (disp *Dispatcher) RegHandler(id interface{}, handler Handler) {
	disp.Lock()
	defer disp.Unlock()

	if disp.handlers == nil {
		disp.handlers = make(map[interface{}]Handler)
	}
	if disp.handlers[id] != nil {
		// TODO: file and line
		log.Error("handler %v already registered", id)
		return
	}
	disp.handlers[id] = handler
}

func (disp *Dispatcher) Handler(id interface{}) Handler {
	disp.RLock()
	defer disp.RUnlock()

	if disp.handlers == nil {
		log.Debug("handler %v not found", id)
		return nil
	}
	handler := disp.handlers[id]
	if handler == nil {
		log.Debug("handler %v not found", id)
		return nil
	}
	return handler
}
