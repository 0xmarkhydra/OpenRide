package auth

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrChallengeNotFound = errors.New("otp challenge not found")
	ErrOTPInvalid        = errors.New("otp invalid")
	ErrOTPExpired        = errors.New("otp expired")
	ErrOTPConsumed       = errors.New("otp already consumed")
	ErrTooManyAttempts   = errors.New("too many otp attempts")
	ErrSessionNotFound   = errors.New("refresh session not found")
	ErrSessionExpired    = errors.New("refresh session expired")
	ErrSessionRevoked    = errors.New("refresh session revoked")
)

type Store interface {
	CreateChallenge(Challenge) error
	VerifyChallenge(id string, role Role, codeHash string, now time.Time, maxAttempts int) (Challenge, error)
	CreateRefreshSession(RefreshSession) error
	GetRefreshByHash(tokenHash string) (RefreshSession, error)
	RevokeRefresh(id string, at time.Time) error
}

type MemoryStore struct {
	mu         sync.Mutex
	challenges map[string]Challenge
	sessions   map[string]RefreshSession
	byHash     map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		challenges: make(map[string]Challenge),
		sessions:   make(map[string]RefreshSession),
		byHash:     make(map[string]string),
	}
}

func (s *MemoryStore) CreateChallenge(ch Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.challenges[ch.ID] = ch
	return nil
}

func (s *MemoryStore) VerifyChallenge(id string, role Role, codeHash string, now time.Time, maxAttempts int) (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.challenges[id]
	if !ok || ch.Role != role {
		return Challenge{}, ErrChallengeNotFound
	}
	if ch.ConsumedAt != nil {
		return Challenge{}, ErrOTPConsumed
	}
	if !ch.ExpiresAt.After(now) {
		return Challenge{}, ErrOTPExpired
	}
	if ch.Attempts >= maxAttempts {
		return Challenge{}, ErrTooManyAttempts
	}
	ch.Attempts++
	if ch.CodeHash != codeHash {
		s.challenges[id] = ch
		if ch.Attempts >= maxAttempts {
			return Challenge{}, ErrTooManyAttempts
		}
		return Challenge{}, ErrOTPInvalid
	}
	consumed := now
	ch.ConsumedAt = &consumed
	s.challenges[id] = ch
	return ch, nil
}

func (s *MemoryStore) CreateRefreshSession(session RefreshSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	s.byHash[session.TokenHash] = session.ID
	return nil
}

func (s *MemoryStore) GetRefreshByHash(tokenHash string) (RefreshSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byHash[tokenHash]
	if !ok {
		return RefreshSession{}, ErrSessionNotFound
	}
	session, ok := s.sessions[id]
	if !ok {
		return RefreshSession{}, ErrSessionNotFound
	}
	return session, nil
}

func (s *MemoryStore) RevokeRefresh(id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}
	if session.RevokedAt == nil {
		session.RevokedAt = &at
		s.sessions[id] = session
	}
	return nil
}
