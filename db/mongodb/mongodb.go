package mongodb

import (
	"container/heap"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

// one DialContext per goroutine (goroutine not safe)
type DialContext struct {
	sessions SessionHeap
}

// goroutine safe
func Dial(url string, sessionNum int) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
		log.Release("invalid sessionNum, reset to %v", sessionNum)
	}

	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	c := new(DialContext)

	// sessions
	c.sessions = make(SessionHeap, sessionNum)
	c.sessions[0] = &Session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
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

func (c *DialContext) EnsureCounter(db string, collection string, id string) error {
	s := c.Ref()
	defer c.UnRef(s)

	err := s.DB(db).C(collection).Insert(bson.M{
		"_id": id,
		"seq": 0,
	})
	if mgo.IsDup(err) {
		return nil
	} else {
		return err
	}
}

func (c *DialContext) NextSeq(db string, collection string, id string) (int, error) {
	s := c.Ref()
	defer c.UnRef(s)

	var res struct {
		Seq int
	}
	_, err := s.DB(db).C(collection).FindId(id).Apply(mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		ReturnNew: true,
	}, &res)

	return res.Seq, err
}

func (c *DialContext) EnsureIndex(db string, collection string, key []string) error {
	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).EnsureIndex(mgo.Index{
		Key:    key,
		Unique: true,
		Sparse: true,
	})
}
