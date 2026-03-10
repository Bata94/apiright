package core

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

type Session struct {
	ID        string
	Data      map[string]any
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (s *Session) Get(key string) any {
	return s.Data[key]
}

func (s *Session) Set(key string, value any) {
	s.Data[key] = value
}

func (s *Session) Delete(key string) {
	delete(s.Data, key)
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type SessionStore interface {
	Get(id string) (*Session, error)
	Set(session *Session) error
	Delete(id string) error
}

type MemorySessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemorySessionStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}

	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *MemorySessionStore) Set(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *MemorySessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}

func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
