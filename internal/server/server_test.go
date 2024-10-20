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

func TestLoginHandler_Success(t *testing.T) {
	mockSessionService := &MockSessionService{
		GenerateTokenFunc: func() string {
			return "mockToken"
		},
		CreateSessionFunc: func(token string, userID string) (*session.Session, error) {
			return &session.Session{UserId: userID}, nil
		},
	}

	srv := New(mockSessionService)

	req, err := http.NewRequest("POST", "/login", bytes.NewBufferString("user_id=testUser"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.loginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response SessionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Token != "mockToken" || response.Session.UserId != "testUser" {
		t.Errorf("unexpected response: got %+v", response)
	}
}

func TestLoginHandler_MissingUserID(t *testing.T) {
	mockSessionService := &MockSessionService{}

	srv := New(mockSessionService)

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

	srv := New(mockSessionService)

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

	srv := New(mockSessionService)

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

	srv := New(mockSessionService)

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
