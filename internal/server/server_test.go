package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aloysb/auth-session/internal/session"
)

// MockSessionService is a mock implementation of session.ISessionService
type MockSessionService struct {
	CreateSessionFunc     func(token string, userID string) (*session.Session, error)
	GenerateTokenFunc     func() string
	ValidateSessionFunc   func(token string) (*session.Session, error)
	InvalidateSessionFunc func(token string) error
}

func (m *MockSessionService) CreateSession(token string, userID string) (*session.Session, error) {
	return m.CreateSessionFunc(token, userID)
}

func (m *MockSessionService) GenerateToken() string {
	return m.GenerateTokenFunc()
}

func (m *MockSessionService) ValidateSession(token string) (*session.Session, error) {
	return m.ValidateSessionFunc(token)
}

func (m *MockSessionService) InvalidateSession(token string) error {
	return nil
}

// MockBasicAuthService is a mock implementation of auth.BasicAuthService
type MockBasicAuthService struct {
	SignInFunc func(email string, password string) error
	SignUpFunc func(email string, password string) error
}

func (m *MockBasicAuthService) SignIn(email string, password string) error {
	return m.SignInFunc(email, password)
}

func (m *MockBasicAuthService) SignUp(email string, password string) error {
	return m.SignUpFunc(email, password)
}

func TestLoginHandler_Success(t *testing.T) {
	mockSessionService := &MockSessionService{
		GenerateTokenFunc: func() string {
			return "mockToken"
		},
		CreateSessionFunc: func(token string, userID string) (*session.Session, error) {
			return &session.Session{UserId: userID}, nil
		},
	}

	basicAuthService := &MockBasicAuthService{
		SignInFunc: func(email string, password string) error {
			return nil
		},
		SignUpFunc: func(email string, password string) error {
			return nil
		},
	}

	srv := New(mockSessionService, basicAuthService)

	req, err := http.NewRequest("POST", "/login", bytes.NewBufferString("email=valid@email.com&password=validPassword"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.loginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response SessionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Token != "mockToken" || response.Session.UserId != "valid@email.com" {
		t.Errorf("unexpected response: got %+v", response)
	}
}

func TestLoginHandler_MissingUserID(t *testing.T) {
	mockSessionService := &MockSessionService{}
	basicAuthService := &MockBasicAuthService{}

	srv := New(mockSessionService, basicAuthService)

	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.loginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestValidateSessionHandler_Success(t *testing.T) {
	mockSessionService := &MockSessionService{
		ValidateSessionFunc: func(token string) (*session.Session, error) {
			return &session.Session{UserId: "testUser"}, nil
		},
	}
	basicAuthService := &MockBasicAuthService{}

	srv := New(mockSessionService, basicAuthService)

	req, err := http.NewRequest("POST", "/validate", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: session.COOKIE_NAME, Value: "mockToken"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.validateSessionHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != "testUser" {
		t.Errorf("unexpected response body: got %v want %v", rr.Body.String(), "testUser")
	}
}

func TestValidateSessionHandler_NoCookie(t *testing.T) {
	mockSessionService := &MockSessionService{}
	basicAuthService := &MockBasicAuthService{}

	srv := New(mockSessionService, basicAuthService)

	req, err := http.NewRequest("POST", "/validate", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.validateSessionHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestLogoutUserHandler_Success(t *testing.T) {
	mockSessionService := &MockSessionService{
		InvalidateSessionFunc: func(token string) error {
			return nil
		},
	}
	basicAuthService := &MockBasicAuthService{}

	srv := New(mockSessionService, basicAuthService)

	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: session.COOKIE_NAME, Value: "mockToken"})

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.logoutHandler)
	handler.ServeHTTP(rec, req)

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
