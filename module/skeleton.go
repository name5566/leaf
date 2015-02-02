package module

import (
	"github.com/name5566/leaf/chanrpc"
	"github.com/name5566/leaf/go"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"time"
)

type Skeleton struct {
	Name               string
	ChanRPCServerLen   int
	GoLen              int
	TimerDispatcherLen int
	ChanRPCServer      *chanrpc.Server
	g                  *g.Go
	timerDispatcher    *timer.Dispatcher
}

func (s *Skeleton) InitSkeleton() {
	if s.Name == "" {
		s.Name = "Unnamed"
	}

	if s.ChanRPCServerLen <= 0 {
		s.ChanRPCServerLen = 10000
		log.Release("invalid ChanRPCServerLen, reset to %v", s.ChanRPCServerLen)
	}
	s.ChanRPCServer = chanrpc.NewServer(s.ChanRPCServerLen)

	if s.GoLen <= 0 {
		s.GoLen = 10000
		log.Release("invalid GoLen, reset to %v", s.GoLen)
	}
	s.g = g.New(s.GoLen)

	if s.TimerDispatcherLen <= 0 {
		s.TimerDispatcherLen = 10000
		log.Release("invalid TimerDispatcherLen, reset to %v", s.TimerDispatcherLen)
	}
	s.timerDispatcher = timer.NewDispatcher(s.TimerDispatcherLen)
}

func (s *Skeleton) RunSkeleton(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			s.ChanRPCServer.Close()
			s.g.Close()
			return
		case ci := <-s.ChanRPCServer.ChanCall:
			err := s.ChanRPCServer.Exec(ci)
			if err != nil {
				log.Error("%v module chanrpc error: %v", s.Name, err)
			}
		case cb := <-s.g.ChanCb:
			s.g.Cb(cb)
		case t := <-s.timerDispatcher.ChanTimer:
			t.Cb()
		}
	}
}

func (s *Skeleton) AfterFunc(d time.Duration, cb func()) *timer.Timer {
	return s.timerDispatcher.AfterFunc(d, cb)
}

func (s *Skeleton) Go(f func(), cb func()) {
	s.g.Go(f, cb)
}

func (s *Skeleton) RegisterChanRPC(id interface{}, f interface{}) {
	s.ChanRPCServer.Register(id, f)
}
