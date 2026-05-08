package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/config"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
	"forsaken-mail/internal/webhook"
	"forsaken-mail/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

type routerTestEnv struct {
	db       *sql.DB
	handler  http.Handler
	sessions *auth.SessionManager
	store    *mail.Store
}

func newRouterTestEnv(t *testing.T) *routerTestEnv {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	settingsStore := settings.NewStore(db)
	if err := settingsStore.Init(); err != nil {
		t.Fatal(err)
	}
	if err := settingsStore.Set("mail_host", "mail.test"); err != nil {
		t.Fatal(err)
	}

	auditStore := audit.NewStore(db)
	if err := auditStore.Init(); err != nil {
		t.Fatal(err)
	}

	mailStore := mail.NewStore(db)
	if err := mailStore.Init(); err != nil {
		t.Fatal(err)
	}

	sessions := auth.NewSessionManager("secret", false)
	authMW := auth.NewMiddleware(sessions, settingsStore)
	hub := ws.NewHub(nil, "mail.test")
	router := NewRouter(
		&config.Config{
			AuthMode:      "local",
			AdminUsername: "admin",
			AdminPassword: "password123",
			SessionSecret: "secret",
			MailHost:      "mail.test",
			DBPath:        ":memory:",
		},
		sessions,
		authMW,
		mailStore,
		settingsStore,
		auditStore,
		hub,
		webhook.NewSender(settingsStore),
		auth.NewLocalAuth("admin", "password123"),
	)

	return &routerTestEnv{
		db:       db,
		handler:  router.Handler(),
		sessions: sessions,
		store:    mailStore,
	}
}

func (e *routerTestEnv) newRequest(t *testing.T, method, path string, body []byte, withCSRF bool) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Accept-Language", "en")

	sessionValue, err := e.sessions.Encrypt(&auth.SessionData{
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionValue})

	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if withCSRF {
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf-test"})
		req.Header.Set("X-CSRF-Token", "csrf-test")
	}

	return req
}

func (e *routerTestEnv) insertStaleMail(t *testing.T, shortID string) int64 {
	t.Helper()

	result, err := e.db.Exec(
		`INSERT INTO mails (short_id, from_addr, to_addr, subject, text_body, html_body, raw_size, extracted_codes, extracted_links, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		shortID,
		"sender@example.com",
		shortID+"@mail.test",
		"Your verification code",
		"Use code 123456 to continue.",
		"",
		128,
		"[]",
		"[]",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func TestHandler_APIAuthReturnsJSON401(t *testing.T) {
	env := newRouterTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/mails?shortId=alice", nil)
	rec := httptest.NewRecorder()

	env.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected json content type, got %q", ct)
	}
}

func TestHandler_DomainTestRouteUsesGETContract(t *testing.T) {
	env := newRouterTestEnv(t)

	getReq := env.newRequest(t, http.MethodGet, "/api/domain-test", nil, false)
	getRec := httptest.NewRecorder()
	env.handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusBadRequest {
		t.Fatalf("expected GET to reach handler and return 400, got %d", getRec.Code)
	}

	postReq := env.newRequest(t, http.MethodPost, "/api/domain-test", nil, true)
	postRec := httptest.NewRecorder()
	env.handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected POST to be rejected by route contract, got %d", postRec.Code)
	}
}

func TestHandler_GetMailsDoesNotReextract(t *testing.T) {
	env := newRouterTestEnv(t)
	id := env.insertStaleMail(t, "alice")

	req := env.newRequest(t, http.MethodGet, "/api/mails?shortId=alice&reextract=true", nil, false)
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	m, err := env.store.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ExtractedCodes) != 0 {
		t.Fatalf("expected GET path to remain side-effect free, got extracted codes %v", m.ExtractedCodes)
	}
}

func TestHandler_ReextractEndpointMutatesExplicitly(t *testing.T) {
	env := newRouterTestEnv(t)
	id := env.insertStaleMail(t, "alice")

	req := env.newRequest(t, http.MethodPost, "/api/mails/reextract", []byte(`{"short_id":"alice"}`), true)
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if int(resp["updated"].(float64)) != 1 {
		t.Fatalf("expected updated=1, got %v", resp["updated"])
	}

	m, err := env.store.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.ExtractedCodes) == 0 || m.ExtractedCodes[0] != "123456" {
		t.Fatalf("expected extracted code to be backfilled, got %v", m.ExtractedCodes)
	}
}

func TestHandler_EmailsDeleteByShortID(t *testing.T) {
	env := newRouterTestEnv(t)
	if err := env.store.Save(&mail.Mail{
		ShortID:  "alice",
		FromAddr: "sender@example.com",
		ToAddr:   "alice@mail.test",
		Subject:  "hello",
		TextBody: "body",
	}); err != nil {
		t.Fatal(err)
	}

	req := env.newRequest(t, http.MethodDelete, "/api/emails/alice", nil, true)
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	count, err := env.store.CountByShortID("alice")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected mailbox to be deleted, got %d mails", count)
	}
}
