package timer

import (
	"time"
)

type Dispatcher struct {
	ChanCb chan func()
}

func NewDispatcher(l int) *Dispatcher {
	disp := new(Dispatcher)
	disp.ChanCb = make(chan func(), l)
	return disp
}

func (disp *Dispatcher) AfterFunc(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, func() {
		disp.ChanCb <- f
	})
}
