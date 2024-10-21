package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aloysb/auth-session/internal/auth"
	"github.com/aloysb/auth-session/internal/session"
)

// Response struct to encapsulate session and token
type SessionResponse struct {
	Session session.Session `json:"session"` // The session data
	Token   string          `json:"token"`   // The session token
}

type Server struct {
	sessionService session.ISessionService
	authService    auth.IBasicAuthService
}

func New(sessionService session.ISessionService, authService auth.IBasicAuthService) *Server {
	return &Server{
		sessionService: sessionService,
		authService:    authService,
	}
}

// StartServer initializes the HTTP server and routes
func (s *Server) Start(port int) {
	http.HandleFunc("POST /login", s.loginHandler)
	http.HandleFunc("POST /logout", s.logoutHandler)
	http.HandleFunc("POST /authenticate", s.validateSessionHandler)
	http.HandleFunc("POST /signup", s.signupHandler)

	fmt.Printf("Server is running on port: %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// loginHandler handles user login and creates a session
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	err := s.authService.SignIn(email, password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		case auth.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	token := s.sessionService.GenerateToken()
	session, err := s.sessionService.CreateSession(token, email)

	// Create the response struct
	response := SessionResponse{
		Session: *session,
		Token:   token,
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize the session struct to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Unable to serialize session", http.StatusInternalServerError)
		return
	}

	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	err := s.authService.SignUp(email, password)

	if err != nil {
		fmt.Println(err)
		switch err {
		case auth.ErrInvalidEmail:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case auth.ErrUserAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
			return
		case auth.ErrEmptyPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s.loginHandler(w, r)
}

// validateSessionHandler checks if the session is valid
func (s *Server) validateSessionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(session.COOKIE_NAME)

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	ses, err := s.sessionService.ValidateSession(cookie.Value)

	if err != nil {
		switch {
		case errors.Is(err, session.ErrInvalidSession):
			http.Error(w, "Invalid session", http.StatusUnauthorized)
		case errors.Is(err, session.ErrExpiredSession):
			http.Error(w, "Expired session", http.StatusUnauthorized)
		default:
			http.Error(w, "Error validating session", http.StatusInternalServerError)
		}

	}

	w.Write([]byte(ses.UserId))
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(session.COOKIE_NAME)

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	s.sessionService.InvalidateSession(cookie.Value)
}
