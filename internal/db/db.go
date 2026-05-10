package db

import "sync"

var Data = &kvStore{M: make(map[string]string)}

type kvStore struct {
	mu sync.RWMutex
	M  map[string]string
}

func (s *kvStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.M[key]
	return v, ok
}

func (s *kvStore) Set(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.M[key] = val
}
