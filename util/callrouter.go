package util

import (
	"errors"
	"fmt"
)

type CallRouter struct {
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	mapFunc  map[string]interface{}
	chanCall chan *CallInfo
}

type CallInfo struct {
	funcName string
	args     []interface{}
	// nil
	// chan interface{}
	// chan []interface{}
	chanRet interface{}
}

// new
func NewCallRouter(l int) *CallRouter {
	r := new(CallRouter)
	r.mapFunc = make(map[string]interface{})
	r.chanCall = make(chan *CallInfo, l)

	return r
}

// call (goroutine safe)
func (r *CallRouter) Call0(funcName string, args ...interface{}) {
	r.chanCall <- &CallInfo{
		funcName: funcName,
		args:     args,
		chanRet:  nil,
	}
}

func (r *CallRouter) Call1(funcName string, args ...interface{}) chan interface{} {
	chanRet := make(chan interface{}, 1)

	r.chanCall <- &CallInfo{
		funcName: funcName,
		args:     args,
		chanRet:  chanRet,
	}

	return chanRet
}

func (r *CallRouter) CallN(funcName string, args ...interface{}) chan []interface{} {
	chanRet := make(chan []interface{}, 1)

	r.chanCall <- &CallInfo{
		funcName: funcName,
		args:     args,
		chanRet:  chanRet,
	}

	return chanRet
}

// define (goroutine not safe)
func (r *CallRouter) Def(funcName string, f interface{}) {
	switch f.(type) {
	case func([]interface{}):
	case func([]interface{}) interface{}:
	case func([]interface{}) []interface{}:
	default:
		panic(fmt.Sprintf("%v: definition of function is invalid in CallRouter", funcName))
	}

	if _, ok := r.mapFunc[funcName]; ok {
		panic(fmt.Sprintf("%v: function redefined in CallRouter", funcName))
	}

	r.mapFunc[funcName] = f
}

// router
func (r *CallRouter) Chan() chan *CallInfo {
	return r.chanCall
}

func (r *CallRouter) Route(ci *CallInfo) error {
	// function
	f := r.mapFunc[ci.funcName]
	if f == nil {
		return errors.New(fmt.Sprintf("function %v not defined", ci.funcName))
	}

	switch ci.chanRet.(type) {
	case nil:
		// for Call0
		if _, ok := f.(func([]interface{})); !ok {
			return errors.New(fmt.Sprintf("function %v mismatch Call0", ci.funcName))
		}

		f.(func([]interface{}))(ci.args)
	case chan interface{}:
		// for Call1
		if _, ok := f.(func([]interface{}) interface{}); !ok {
			return errors.New(fmt.Sprintf("function %v mismatch Call1", ci.funcName))
		}

		ci.chanRet.(chan interface{}) <- f.(func([]interface{}) interface{})(ci.args)
	case chan []interface{}:
		// for CallN
		if _, ok := f.(func([]interface{}) []interface{}); !ok {
			return errors.New(fmt.Sprintf("function %v mismatch CallN", ci.funcName))
		}

		ci.chanRet.(chan []interface{}) <- f.(func([]interface{}) []interface{})(ci.args)
	}

	return nil
}
