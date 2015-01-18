package chanrpc

import (
	"errors"
	"fmt"
)

// one server per goroutine (goroutine not safe)
// one client per goroutine (goroutine not safe)
type Server struct {
	// id -> function
	//
	// function:
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	functions map[interface{}]interface{}
	ChanCall  chan *CallInfo
}

type CallInfo struct {
	f       interface{}
	args    []interface{}
	chanRet chan *RetInfo
	cb      interface{}
}

type RetInfo struct {
	// nil
	// interface{}
	// []interface{}
	ret interface{}
	err error
	// callback:
	// func(err error)
	// func(ret interface{}, err error)
	// func(ret []interface{}, err error)
	cb interface{}
}

type Client struct {
	s               *Server
	chanSyncRet     chan *RetInfo
	ChanAsynRet     chan *RetInfo
	pendingAsynCall int
}

func NewServer(l int) *Server {
	s := new(Server)
	s.functions = make(map[interface{}]interface{})
	s.ChanCall = make(chan *CallInfo, l)
	return s
}

// you must call the function before calling Open function
func (s *Server) Register(id interface{}, f interface{}) {
	switch f.(type) {
	case func([]interface{}):
	case func([]interface{}) interface{}:
	case func([]interface{}) []interface{}:
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid", id))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = f
}

func (s *Server) ret(ci *CallInfo, ri *RetInfo) (err error) {
	if ci.chanRet == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ri.cb = ci.cb
	ci.chanRet <- ri
	return
}

func (s *Server) Exec(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			s.ret(ci, &RetInfo{err: err})
		}
	}()

	// execute
	switch ci.f.(type) {
	case func([]interface{}):
		ci.f.(func([]interface{}))(ci.args)
		return s.ret(ci, &RetInfo{})
	case func([]interface{}) interface{}:
		ret := ci.f.(func([]interface{}) interface{})(ci.args)
		return s.ret(ci, &RetInfo{ret: ret})
	case func([]interface{}) []interface{}:
		ret := ci.f.(func([]interface{}) []interface{})(ci.args)
		return s.ret(ci, &RetInfo{ret: ret})
	}

	panic("bug")
}

func (s *Server) Close() {
	close(s.ChanCall)

	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			err: errors.New("chanrpc server closed"),
		})
	}
}

// goroutine safe
func (s *Server) Open(l int) *Client {
	c := new(Client)
	c.s = s
	c.chanSyncRet = make(chan *RetInfo, 1)
	c.ChanAsynRet = make(chan *RetInfo, l)
	return c
}

func (c *Client) call(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	c.s.ChanCall <- ci
	return
}

func (c *Client) f(id interface{}, n int) (f interface{}, err error) {
	f = c.s.functions[id]
	if f == nil {
		err = errors.New(fmt.Sprintf("function id %v: function not registered", id))
		return
	}

	var ok bool
	switch n {
	case 0:
		_, ok = f.(func([]interface{}))
	case 1:
		_, ok = f.(func([]interface{}) interface{})
	case 2:
		_, ok = f.(func([]interface{}) []interface{})
	default:
		panic("bug")
	}

	if !ok {
		err = errors.New(fmt.Sprintf("function id %v: return type mismatch", id))
	}
	return
}

func (c *Client) Call0(id interface{}, args ...interface{}) error {
	f, err := c.f(id, 0)
	if err != nil {
		return err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return err
	}

	ri := <-c.chanSyncRet
	return ri.err
}

func (c *Client) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	f, err := c.f(id, 1)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet
	return ri.ret, ri.err
}

func (c *Client) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	f, err := c.f(id, 2)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet
	return ri.ret.([]interface{}), ri.err
}

func (c *Client) asynCall(id interface{}, args []interface{}, cb interface{}, n int) error {
	f, err := c.f(id, n)
	if err != nil {
		return err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.ChanAsynRet,
		cb:      cb,
	})
	if err != nil {
		return err
	}

	c.pendingAsynCall++
	return nil
}

func (c *Client) AsynCall(id interface{}, _args ...interface{}) {
	if len(_args) < 1 {
		panic("callback function not found")
	}

	// args
	var args []interface{}
	if len(_args) > 1 {
		args = _args[:len(_args)-1]
	}

	// cb
	cb := _args[len(_args)-1]
	switch cb.(type) {
	case func(error):
		err := c.asynCall(id, args, cb, 0)
		if err != nil {
			cb.(func(error))(err)
		}
	case func(interface{}, error):
		err := c.asynCall(id, args, cb, 1)
		if err != nil {
			cb.(func(interface{}, error))(nil, err)
		}
	case func([]interface{}, error):
		err := c.asynCall(id, args, cb, 2)
		if err != nil {
			cb.(func([]interface{}, error))(nil, err)
		}
	default:
		panic("definition of callback function is invalid")
	}
}

func (c *Client) Go(id interface{}, args ...interface{}) {
	f := c.s.functions[id]
	if f == nil {
		return
	}

	c.call(&CallInfo{
		f:    f,
		args: args,
	})
}

func (c *Client) Cb(ri *RetInfo) {
	switch ri.cb.(type) {
	case func(error):
		ri.cb.(func(error))(ri.err)
	case func(interface{}, error):
		ri.cb.(func(interface{}, error))(ri.ret, ri.err)
	case func([]interface{}, error):
		ri.cb.(func([]interface{}, error))(ri.ret.([]interface{}), ri.err)
	default:
		panic("bug")
	}

	c.pendingAsynCall--
}

func (c *Client) Close() {
	for c.pendingAsynCall > 0 {
		c.Cb(<-c.ChanAsynRet)
	}
}
