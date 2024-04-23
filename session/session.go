package session

import (
	"errors"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("session key not found")
	ErrSaveFailed  = errors.New("session save failed")
)

type Session interface {
	Get(key string) (any, error)
	Set(key string, val any) error
	Delete(key string) error
	ID() string
	Modified() bool
}

type Builder func(store Store, id string) Session

func DefaultBuilder(store Store, id string) Session {
	return &session{
		Id:   id,
		Data: make(map[string]any),
	}
}

type session struct {
	Id       string
	Data     map[string]any
	mu       sync.RWMutex
	modified bool
}

func (s *session) Get(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (s *session) Set(key string, val any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data[key] = val
	s.modified = true
	return nil
}

func (s *session) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Data, key)
	s.modified = true
	return nil
}

func (s *session) ID() string {
	return s.Id
}

func (s *session) Modified() bool {
	return s.modified
}
