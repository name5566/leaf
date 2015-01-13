package util

import (
	"errors"
	"fmt"
)

type CallRouter struct {
	// id -> handler
	//
	// handler:
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	mapFunc  map[interface{}]interface{}
	chanCall chan *CallInfo
}

type CallInfo struct {
	id   interface{}
	args []interface{}
	// nil
	// chan interface{}
	// chan []interface{}
	chanRet interface{}
}

// new
func NewCallRouter(l int) *CallRouter {
	r := new(CallRouter)
	r.mapFunc = make(map[interface{}]interface{})
	r.chanCall = make(chan *CallInfo, l)

	return r
}

// asyn call (goroutine safe)
func (r *CallRouter) AsynCall0(id interface{}, args ...interface{}) {
	r.chanCall <- &CallInfo{
		id:      id,
		args:    args,
		chanRet: nil,
	}
}

func (r *CallRouter) AsynCall1(id interface{}, chanRet chan interface{}, args ...interface{}) {
	r.chanCall <- &CallInfo{
		id:      id,
		args:    args,
		chanRet: chanRet,
	}
}

func (r *CallRouter) AsynCallN(id interface{}, chanRet chan []interface{}, args ...interface{}) {
	r.chanCall <- &CallInfo{
		id:      id,
		args:    args,
		chanRet: chanRet,
	}
}

// sync call (goroutine safe)
func (r *CallRouter) Call1(id interface{}, args ...interface{}) interface{} {
	chanRet := make(chan interface{}, 1)

	r.chanCall <- &CallInfo{
		id:      id,
		args:    args,
		chanRet: chanRet,
	}

	return <-chanRet
}

func (r *CallRouter) CallN(id interface{}, args ...interface{}) []interface{} {
	chanRet := make(chan []interface{}, 1)

	r.chanCall <- &CallInfo{
		id:      id,
		args:    args,
		chanRet: chanRet,
	}

	return <-chanRet
}

// define (goroutine not safe)
func (r *CallRouter) Def(id interface{}, f interface{}) {
	switch f.(type) {
	case func([]interface{}):
	case func([]interface{}) interface{}:
	case func([]interface{}) []interface{}:
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid in CallRouter", id))
	}

	if _, ok := r.mapFunc[id]; ok {
		panic(fmt.Sprintf("function id %v: function redefined in CallRouter", id))
	}

	r.mapFunc[id] = f
}

// route (goroutine not safe)
func (r *CallRouter) Chan() chan *CallInfo {
	return r.chanCall
}

func (r *CallRouter) Route(ci *CallInfo) error {
	// function
	f := r.mapFunc[ci.id]
	if f == nil {
		return errors.New(fmt.Sprintf("function id %v: function not defined", ci.id))
	}

	switch ci.chanRet.(type) {
	case nil:
		// AsynCall0
		if _, ok := f.(func([]interface{})); !ok {
			return errors.New(fmt.Sprintf("function id %v: function mismatch AsynCall0", ci.id))
		}

		f.(func([]interface{}))(ci.args)
	case chan interface{}:
		// Call1 or AsynCall1
		if _, ok := f.(func([]interface{}) interface{}); !ok {
			return errors.New(fmt.Sprintf("function id %v: function mismatch Call1 or AsynCall1", ci.id))
		}

		ci.chanRet.(chan interface{}) <- f.(func([]interface{}) interface{})(ci.args)
	case chan []interface{}:
		// CallN or AsynCallN
		if _, ok := f.(func([]interface{}) []interface{}); !ok {
			return errors.New(fmt.Sprintf("function id %v: function mismatch CallN or AsynCallN", ci.id))
		}

		ci.chanRet.(chan []interface{}) <- f.(func([]interface{}) []interface{})(ci.args)
	}

	return nil
}

func (r *CallRouter) RouteAll() {
	for len(r.chanCall) > 0 {
		r.Route(<-r.chanCall)
	}
}
