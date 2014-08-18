package util

import (
	"sync"
)

type Map struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func (m *Map) init() {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
}

func (m *Map) Get(key interface{}) interface{} {
	m.RLock()
	defer m.RUnlock()

	if m.m == nil {
		return nil
	} else {
		return m.m[key]
	}
}

func (m *Map) Set(key interface{}, value interface{}) {
	m.Lock()
	defer m.Unlock()

	m.init()
	m.m[key] = value
}

func (m *Map) Del(key interface{}) {
	m.Lock()
	defer m.Unlock()

	m.init()
	delete(m.m, key)
}
