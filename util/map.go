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

func (m *Map) Len() int {
	m.RLock()
	defer m.RUnlock()

	if m.m == nil {
		return 0
	} else {
		return len(m.m)
	}
}

func (m *Map) RLockRange(f func(interface{}, interface{})) {
	m.RLock()
	defer m.RUnlock()

	if m.m == nil {
		return
	}
	for k, v := range m.m {
		f(k, v)
	}
}

func (m *Map) LockRange(f func(interface{}, interface{}, map[interface{}]interface{})) {
	m.Lock()
	defer m.Unlock()

	m.init()
	for k, v := range m.m {
		f(k, v, m.m)
	}
}
