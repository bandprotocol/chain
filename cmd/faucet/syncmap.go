package main

import "sync"

type SyncMap struct {
	sync.RWMutex
	container map[string]*LimitStatus
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		RWMutex:   sync.RWMutex{},
		container: make(map[string]*LimitStatus),
	}
}

func (m *SyncMap) Load(key string) (*LimitStatus, bool) {
	m.RLock()
	defer m.RUnlock()
	res, ok := m.container[key]
	return res, ok
}

func (m *SyncMap) Store(key string, value *LimitStatus) {
	m.Lock()
	defer m.Unlock()
	m.container[key] = value
}

func (m *SyncMap) Remove(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.container, key)
}
