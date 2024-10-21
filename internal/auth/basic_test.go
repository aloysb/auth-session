package auth

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

var dbFile = "test_auth.db"
var Db *sql.DB

func setupService() BasicAuthService {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_sessions_")
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Create a temporary file for the SQLite database
	dbFile = filepath.Join(tmpDir, "test_sessions.db")

	// Create a temporary directory
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Create a temporary file for the SQLite database
	dbFile = filepath.Join(tmpDir, "test_auth.db")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	// Drop the user table if it exists
	_, err = db.Exec(`DROP TABLE IF EXISTS users`)
	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	_, err = db.Exec(`
        CREATE TABLE users (
          id SERIAL PRIMARY KEY,
          email TEXT NOT NULL,
          password TEXT NOT NULL,
          salt TEXT NOT NULL
       );
   `)

	if err != nil {
		log.Fatalf("failed to set up test table: %s", err)
	}

	Db = db

	return *New(db)
}

func teardownTestDB() {
	os.Remove(dbFile) // Remove the database file after tests
}

func TestSignUp_Valid(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignUp("test@user.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSignUp_AlreadyExists(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignUp("test@user.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = s.SignUp("test@user.com", "testpassword")
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestSignUp_InvalidEmail(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignUp("invalidemail", "testpassword")
	if err != ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}

	err = s.SignUp("valid@email.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSignIn_Valid(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignUp("test@user.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = s.SignIn("test@user.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSignIn_InvalidCredentials(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignUp("test@user.com", "testpassword")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = s.SignIn("test@user.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestSignIn_NonexistentUser(t *testing.T) {
	s := setupService()
	defer teardownTestDB()

	err := s.SignIn("nonexistent@user.com", "wrongpassword")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
