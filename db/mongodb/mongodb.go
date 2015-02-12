package mongodb

import (
	"container/heap"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
)

// session
type Session struct {
	*mgo.Session
	ref   int
	index int
}

// session heap
type SessionHeap []*Session

func (h SessionHeap) Len() int {
	return len(h)
}

func (h SessionHeap) Less(i, j int) bool {
	return h[i].ref < h[j].ref
}

func (h SessionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *SessionHeap) Push(s interface{}) {
	s.(*Session).index = len(*h)
	*h = append(*h, s.(*Session))
}

func (h *SessionHeap) Pop() interface{} {
	l := len(*h)
	s := (*h)[l-1]
	s.index = -1
	*h = (*h)[:l-1]
	return s
}

// dial context
type DialContext struct {
	sessions SessionHeap
}

type DialInfo struct {
	Url        string
	SessionNum int
}

func Dial(info DialInfo) (*DialContext, error) {
	if info.SessionNum <= 0 {
		info.SessionNum = 100
		log.Release("invalid SessionNum, reset to %v", info.SessionNum)
	}

	s, err := mgo.Dial(info.Url)
	if err != nil {
		return nil, err
	}

	c := new(DialContext)
	c.sessions = make(SessionHeap, info.SessionNum)
	c.sessions[0] = &Session{s, 0, 0}
	for i := 1; i < info.SessionNum; i++ {
		c.sessions[i] = &Session{s.New(), 0, i}
	}

	heap.Init(&c.sessions)

	return c, nil
}

func (c *DialContext) Close() {
	for _, s := range c.sessions {
		s.Close()
	}
}

func (c *DialContext) Ref() *Session {
	s := c.sessions[0]
	s.ref++
	heap.Fix(&c.sessions, 0)
	s.Refresh()
	return s
}

func (c *DialContext) UnRef(s *Session) {
	s.ref--
	heap.Fix(&c.sessions, s.index)
}
