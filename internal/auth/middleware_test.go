package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func sessionCookie(t *testing.T, sm *SessionManager, expiresAt time.Time) *http.Cookie {
	t.Helper()

	value, err := sm.Encrypt(&SessionData{
		Email:     "user@example.com",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	return &http.Cookie{
		Name:  "session",
		Value: value,
	}
}

func TestRequirePageAuth_RedirectsUnauthenticated(t *testing.T) {
	mw := NewMiddleware(NewSessionManager("secret", false), nil)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	mw.RequirePageAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}
}

func TestRequireAPIAuth_ReturnsJSON401Unauthenticated(t *testing.T) {
	mw := NewMiddleware(NewSessionManager("secret", false), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/mails", nil)
	rec := httptest.NewRecorder()

	mw.RequireAPIAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "unauthorized" {
		t.Fatalf("expected unauthorized error, got %q", body["error"])
	}
}

func TestRequireAPIAuth_ClearsExpiredSession(t *testing.T) {
	sm := NewSessionManager("secret", false)
	mw := NewMiddleware(sm, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/mails", nil)
	req.AddCookie(sessionCookie(t, sm, time.Now().Add(-time.Minute)))
	rec := httptest.NewRecorder()

	mw.RequireAPIAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not run")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "session=") {
		t.Fatalf("expected session clearing cookie, got %q", rec.Header().Get("Set-Cookie"))
	}
}
