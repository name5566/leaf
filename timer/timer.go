package timer

import (
	"errors"
	"time"
)

// one dispatcher per goroutine (goroutine not safe)
type Dispatcher struct {
	ChanTimer chan *Timer
}

func NewDispatcher(l int) *Dispatcher {
	disp := new(Dispatcher)
	disp.ChanTimer = make(chan *Timer, l)
	return disp
}

// Timer
type Timer struct {
	t  *time.Timer
	cb func()
}

func (t *Timer) Stop() {
	t.t.Stop()
	t.cb = nil
}

func (t *Timer) Cb() {
	if t.cb != nil {
		t.cb()
		t.cb = nil
	}
}

func (disp *Dispatcher) AfterFunc(d time.Duration, cb func()) *Timer {
	t := new(Timer)
	t.cb = cb
	t.t = time.AfterFunc(d, func() {
		disp.ChanTimer <- t
	})
	return t
}

// Cron
type Cron struct {
	t *Timer
}

func (c *Cron) Stop() {
	c.t.Stop()
}

func (disp *Dispatcher) CronFunc(expr string, _cb func()) (*Cron, error) {
	cronExpr, err := NewCronExpr(expr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return nil, errors.New("next time not found")
	}

	cron := new(Cron)

	// callback
	var cb func()
	cb = func() {
		defer _cb()

		now := time.Now()
		nextTime := cronExpr.Next(now)
		if nextTime.IsZero() {
			return
		}
		cron.t = disp.AfterFunc(nextTime.Sub(now), cb)
	}

	cron.t = disp.AfterFunc(nextTime.Sub(now), cb)
	return cron, nil
}
