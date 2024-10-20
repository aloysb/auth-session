package session

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aloysb/auth-session/internal/utils"
)

// Constant defining how long a session is valid
const sessionExpiresIn = 24 * time.Hour

const COOKIE_NAME = "auth_session"

// Error constants for session handling
var (
	ErrExpiredSession = errors.New("expired session")
	ErrInvalidSession = errors.New("invalid session")
)

// Session struct to represent session data
type Session struct {
	UserId    string    `json:"user_id"`    // ID of the user who owns the session
	Id        string    `json:"id"`         // Unique ID of the session
	ExpiresAt time.Time `json:"expires_at"` // Timestamp when the session expires
}

type ISessionService interface {
	CreateSession(token, userId string) (*Session, error)
	ValidateSession(token string) (*Session, error)
	GenerateToken() string
	InvalidateSession(token string) error
}

type SessionService struct {
	db *sql.DB
}

func New(db *sql.DB) *SessionService {
	return &SessionService{
		db: db,
	}
}

// ValidateSession checks if a session is valid and refreshes it if it is close to expiring.
func (s *SessionService) ValidateSession(token string) (*Session, error) {
	// Generate a session ID from the token using SHA-256
	sessionId := generateSessionIdFromToken(token)

	// Query the database to find the session
	row := s.db.QueryRow("SELECT id, user_id, expires_at FROM sessions WHERE id = $1", sessionId)

	var session Session
	err := row.Scan(&session.Id, &session.UserId, &session.ExpiresAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrInvalidSession
		default:
			return nil, fmt.Errorf("could not query session: %w", err)
		}
	}

	// Check if the session is expired
	if time.Now().After(session.ExpiresAt) {
		err := s.InvalidateSession(session.Id) // Invalidate the expired session
		if err != nil {
			fmt.Errorf("could not invalidate session: %w", err)
		}
		return nil, ErrExpiredSession
	}

	// Refresh the session if it's more than halfway to expiration
	if time.Now().After(session.ExpiresAt.Add(-sessionExpiresIn / 2)) {
		session.ExpiresAt = time.Now().Add(sessionExpiresIn)
		_, err := s.db.Exec("UPDATE sessions SET expires_at = $1 WHERE id = $2", session.ExpiresAt, session.Id)
		if err != nil {
			return nil, fmt.Errorf("could not refresh session expiration: %w", err)
		}
	}

	return &session, nil
}

// CreateSession generates a new session and saves it to the database
func (s *SessionService) CreateSession(token, userId string) (*Session, error) {
	// Generate a random session ID
	sessionId := generateSessionIdFromToken(token)

	// Create a new session with an expiration time
	session := &Session{
		UserId:    userId,
		Id:        sessionId,
		ExpiresAt: time.Now().Add(sessionExpiresIn),
	}

	// Save the session to the database
	_, err := s.db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", session.Id, session.UserId, session.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("could not create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GenerateToken() string {
	return utils.GenerateRandomString()
}

// invalidateSession removes a session from the database by ID
func (s *SessionService) InvalidateSession(sessionId string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = $1", sessionId)
	if err != nil {
		return fmt.Errorf("could not invalidate session: %w", err)
	}
	return nil
}

func generateSessionIdFromToken(token string) string {
	h := sha256.New()
	_, err := h.Write([]byte(token))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
