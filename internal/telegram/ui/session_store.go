package ui

import (
	"sync"
)

type SessionStore interface {
	Get(chatID int64) (Session, bool)
	Put(chatID int64, s Session)
	Delete(chatID int64)
}

type sessionStore struct {
	guard    *sync.Mutex
	sessions map[int64]Session
}

func NewSessionStore() SessionStore {
	return &sessionStore{
		guard:    &sync.Mutex{},
		sessions: make(map[int64]Session),
	}
}

func (s *sessionStore) Get(chatID int64) (Session, bool) {
	s.guard.Lock()
	defer s.guard.Unlock()

	session, ok := s.sessions[chatID]
	return session, ok
}

func (s *sessionStore) Put(chatID int64, session Session) {
	s.guard.Lock()
	defer s.guard.Unlock()

	s.sessions[chatID] = session
}

func (s *sessionStore) Delete(chatID int64) {
	s.guard.Lock()
	defer s.guard.Unlock()

	delete(s.sessions, chatID)
}
