package session

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

var dbFile = "test_sessions.db"
var Db *sql.DB

func setupService() *SessionService {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_sessions_")
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Create a temporary file for the SQLite database
	dbFile = filepath.Join(tmpDir, "test_sessions.db")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Drop the sessions table if it exists
	_, err = db.Exec(`DROP TABLE IF EXISTS sessions`)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Create the sessions table
	_, err = db.Exec(`
        CREATE TABLE sessions (
          id TEXT PRIMARY KEY,
          user_id TEXT NOT NULL,
          expires_at TIMESTAMP NOT NULL
       );
   `)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	Db = db

	return New(db)
}

func teardownTestDB() {
	os.Remove(dbFile) // Remove the database file after tests
}

func TestCreateSession(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	userID := "user123"
	token := s.GenerateToken()
	session, err := s.CreateSession(token, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.UserId != userID {
		t.Errorf("expected user ID %s, got %s", userID, session.UserId)
	}

	if time.Now().After(session.ExpiresAt) {
		t.Errorf("expected expiration time to be in the future, got %v", session.ExpiresAt)
	}
}

func TestValidateSession_Valid(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	userID := "user123"
	token := s.GenerateToken()
	session, err := s.CreateSession(token, userID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	validatedSession, err := s.ValidateSession(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedSession.Id != session.Id {
		t.Errorf("expected session ID %s, got %s", session.Id, validatedSession.Id)
	}
}

func TestValidateSession_Expired(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	// Create a session that is already expired
	userID := "user123"
	token := s.GenerateToken()
	session := &Session{
		UserId:    userID,
		Id:        generateSessionIdFromToken(token),
		ExpiresAt: time.Now().Add(-time.Hour), // Set expiration time to 1 hour ago
	}

	_, err := Db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", session.Id, session.UserId, session.ExpiresAt)
	if err != nil {
		t.Fatalf("failed to insert expired session: %v", err)
	}

	_, err = s.ValidateSession(token)
	if err != ErrExpiredSession {
		t.Errorf("expected ErrExpiredSession, got %v", err)
	}
}

func TestInvalidateSession(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	userID := "user123"
	token := s.GenerateToken()
	session, err := s.CreateSession(token, userID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = s.InvalidateSession(session.Id)
	if err != nil {
		t.Fatalf("expected no error when invalidating session, got %v", err)
	}

	// Check if the session still exists
	var count int
	row := Db.QueryRow("SELECT COUNT(*) FROM sessions WHERE id = $1", session.Id)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query database: %v", err)
	}

	if count != 0 {
		t.Errorf("expected session to be deleted, but found %d records", count)
	}
}
