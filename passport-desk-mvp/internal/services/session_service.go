package services

import (
	"context"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// Auto-lock after 5 minutes of inactivity as per TZ
	inactivityTimeout = 5 * time.Minute
)

type SessionService struct {
	Mu              sync.RWMutex
	CurrentOperator *models.Operator
	IsLocked        bool
	LastActivity    time.Time
	EncryptionKey   []byte
	Ctx             context.Context
}

// NewSessionService creates a new SessionService
func NewSessionService(currentOperator *models.Operator) *SessionService {
	return &SessionService{
		Mu:              sync.RWMutex{},
		CurrentOperator: currentOperator,
		IsLocked:        true,
		LastActivity:    time.Now(),
		EncryptionKey:   []byte(""),
	}
}

// IsAuthenticated returns whether a user is authenticated
func (s *SessionService) IsAuthenticated() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.CurrentOperator != nil && !s.IsLocked
}

// InactivityChecker monitors for inactivity and locks the app
func (s *SessionService) InactivityChecker() {
	logger.Info("Inactivity checker started")
	if s.Ctx == nil {
		logger.Error("InactivityChecker started with nil context")
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.Mu.Lock()
			if !s.IsLocked && s.CurrentOperator != nil {
				// Don't auto-lock if no operator is loaded yet to avoid race conditions during startup
				if time.Since(s.LastActivity) > inactivityTimeout {
					s.IsLocked = true
					// Emit event to frontend
					if s.Ctx != nil {
						runtime.EventsEmit(s.Ctx, "session-locked")
					}
				}
			}
			s.Mu.Unlock()
		case <-s.Ctx.Done():
			return
		}
	}
}

// GetCurrentOperator returns the currently logged in operator
func (s *SessionService) GetCurrentOperator() *models.Operator {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.CurrentOperator
}

// UpdateActivity updates the last activity timestamp
func (s *SessionService) UpdateActivity() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.LastActivity = time.Now()
}
