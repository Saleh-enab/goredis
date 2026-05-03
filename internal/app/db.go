package app

import "sync"

var Data = &kvStore{m: make(map[string]string)}

type kvStore struct {
	mu sync.RWMutex
	m  map[string]string
}

func (s *kvStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *kvStore) Set(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = val
}
