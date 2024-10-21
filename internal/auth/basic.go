package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"

	"github.com/aloysb/auth-session/internal/utils"
	"golang.org/x/crypto/argon2"
)

const (
	timeCost    = 1         // Time cost parameter
	memoryCost  = 64 * 1024 // Memory cost parameter (64 MB)
	parallelism = 4         // Number of parallel threads
	keyLength   = 32        // Length of the key
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrEmptyPassword      = errors.New("empty password")
)

type IBasicAuthService interface {
	SignIn(email, password string) error
	SignUp(email, password string) error
}

type BasicAuthService struct {
	db *sql.DB
}

type User struct {
	Id       string
	Email    string
	Password []byte
	Salt     []byte
}

func New(db *sql.DB) *BasicAuthService {
	return &BasicAuthService{db}
}

func (b *BasicAuthService) SignUp(email, password string) error {
	// Check if the email already exists
	row := b.db.QueryRow("SELECT id FROM users WHERE email = $1", email)
	var id sql.NullString
	err := row.Scan(&id)

	if err == nil {
		return ErrUserAlreadyExists
	}

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("could not query user: %w", err)
	}

	if validEmail(email) == false {
		return ErrInvalidEmail
	}

	if password == "" {
		return ErrEmptyPassword
	}

	// Hash the password
	salt := generateSalt()
	hashedPassword := hashPassword(password, salt)

	// Save the new user to the database
	if _, err := b.db.Exec("INSERT INTO users (email, password, salt) VALUES ($1, $2, $3)", email, hashedPassword, salt); err != nil {
		return fmt.Errorf("could not insert user: %w", err)
	}

	return nil
}

func (b *BasicAuthService) SignIn(email, password string) error {
	var storedPassword, storedSalt string
	row := b.db.QueryRow("SELECT password, salt FROM users WHERE email = $1", email)

	err := row.Scan(&storedPassword, &storedSalt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrUserNotFound
		default:
			return fmt.Errorf("could not query user: %w", err)
		}
	}
	if !comparePasswords(password, storedSalt, storedPassword) {
		return ErrInvalidCredentials
	}
	return nil
}

func generateSalt() []byte {
	str := utils.GenerateRandomString()
	return []byte(str)
}

// HashPassword hashes the password with a salt
func hashPassword(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)
}

// VerifyPassword checks if the provided password matches the stored hash
func comparePasswords(password, storedSalt, storedHash string) bool {
	hash := hashPassword(password, []byte(storedSalt))
	return string(hash) == storedHash
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
