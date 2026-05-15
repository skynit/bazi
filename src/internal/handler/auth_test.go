package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

// MockUserStore is an in-memory implementation of UserStore for testing.
type MockUserStore struct {
	mu    sync.RWMutex
	users map[string]*model.User // keyed by username
	idSeq uint
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{users: make(map[string]*model.User)}
}

func (m *MockUserStore) Create(user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Username]; exists {
		return &duplicateError{user.Username}
	}

	m.idSeq++
	user.ID = m.idSeq
	m.users[user.Username] = user
	return nil
}

func (m *MockUserStore) FindByUsername(username string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[username]
	if !ok {
		return nil, &notFoundError{username}
	}
	return user, nil
}

func (m *MockUserStore) FindByID(id uint) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, &notFoundError{field: "id"}
}

type duplicateError struct{ username string }

func (e *duplicateError) Error() string { return "duplicate: " + e.username }

type notFoundError struct{ field string }

func (e *notFoundError) Error() string { return "not found: " + e.field }

// ---- test helpers ----

func setupTestRouter(store *MockUserStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &AuthHandler{Store: store}

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/me", middleware.AuthMiddleware(), h.Me)
	}
	return r
}

func jsonBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return strings.NewReader(string(b))
}

// ---- tests ----

func TestRegister(t *testing.T) {
	router := setupTestRouter(NewMockUserStore())

	body := jsonBody(t, model.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify token is present
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in response, got: %v", resp)
	}

	// Verify user object
	userMap, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected user object in response")
	}
	if userMap["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", userMap["username"])
	}
	if userMap["email"] != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %v", userMap["email"])
	}
}

func TestRegisterDuplicate(t *testing.T) {
	store := NewMockUserStore()
	router := setupTestRouter(store)

	// First registration
	body1 := jsonBody(t, model.RegisterRequest{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "secret123",
	})

	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", body1)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d: %s", w1.Code, w1.Body.String())
	}

	// Duplicate registration
	body2 := jsonBody(t, model.RegisterRequest{
		Username: "bob",
		Email:    "bob2@example.com",
		Password: "secret456",
	})

	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", body2)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestLogin(t *testing.T) {
	store := NewMockUserStore()
	router := setupTestRouter(store)

	// Register first
	regBody := jsonBody(t, model.RegisterRequest{
		Username: "charlie",
		Email:    "charlie@example.com",
		Password: "mypassword",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regW.Code, regW.Body.String())
	}

	// Login
	loginBody := jsonBody(t, model.LoginRequest{
		Username: "charlie",
		Password: "mypassword",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", loginW.Code, loginW.Body.String())
	}

	var resp model.LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	store := NewMockUserStore()
	router := setupTestRouter(store)

	// Register
	regBody := jsonBody(t, model.RegisterRequest{
		Username: "dave",
		Email:    "dave@example.com",
		Password: "rightpassword",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regW.Code, regW.Body.String())
	}

	// Login with wrong password
	loginBody := jsonBody(t, model.LoginRequest{
		Username: "dave",
		Password: "wrongpassword",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", loginW.Code, loginW.Body.String())
	}
}

func TestMeUnauthenticated(t *testing.T) {
	store := NewMockUserStore()
	router := setupTestRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMeAuthenticated(t *testing.T) {
	store := NewMockUserStore()
	router := setupTestRouter(store)

	// Register user
	regBody := jsonBody(t, model.RegisterRequest{
		Username: "eve",
		Email:    "eve@example.com",
		Password: "secret123",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regW.Code, regW.Body.String())
	}

	// Extract token from register response
	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	token := regResp["token"].(string)

	// Call /me with token
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var userResp model.RegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &userResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if userResp.Username != "eve" {
		t.Fatalf("expected username eve, got %s", userResp.Username)
	}
	if userResp.Email != "eve@example.com" {
		t.Fatalf("expected email eve@example.com, got %s", userResp.Email)
	}
}
