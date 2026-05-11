package db

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
)

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

func (s *kvStore) Delete(keys []string) int {
	var n int

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		_, ok := s.M[key]

		if ok {
			delete(s.M, key)
			n++
		}
	}

	return n
}

func (s *kvStore) Exists(keys []string) int {
	var n int

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		_, ok := s.M[key]

		if ok {
			n++
		}
	}

	return n
}

func (s *kvStore) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []string
	for key := range s.M {
		matched, err := filepath.Match(pattern, key)

		if err != nil {
			slog.Error(fmt.Sprintf("error matching keys: (pattern: %s), (key: %s)", pattern, key), "err", err)
			continue
		}

		if matched {
			matches = append(matches, key)
		}

	}

	return matches
}
