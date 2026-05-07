# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tmail is a self-hosted disposable/temporary email service. Users receive emails at random or custom addresses on a configured domain and view them in real-time via a web UI. Optional webhook notifications via DingTalk, Telegram, or Slack.

The codebase is **Go backend + React frontend**.

## Build & Run Commands

**Go backend** (requires `CGO_ENABLED=1` for SQLite):
```bash
go build ./cmd/server          # build binary
go run ./cmd/server            # run directly
```

**Frontend** (in `web/`):
```bash
cd web && npm install          # install deps
cd web && npm run dev          # dev server with proxy to localhost:3000
cd web && npm run build        # production build -> ../embed/
```

**Docker:**
```bash
docker compose build           # 3-stage build (frontend -> Go -> Alpine)
docker compose up              # run with .env config
```

**Tests** (Go only, no frontend tests):
```bash
# Run all tests (requires CGO_ENABLED=1 for SQLite)
docker run --rm -v $(pwd):/src -w /src golang:1.24-alpine sh -c 'apk add --no-cache gcc musl-dev > /dev/null 2>&1 && go test ./...'

# Run a single test
docker run --rm -v $(pwd):/src -w /src golang:1.24-alpine sh -c 'apk add --no-cache gcc musl-dev > /dev/null 2>&1 && go test ./internal/mail/ -run TestListAll -v'
```

Test files: `internal/mail/store_test.go`, `internal/mail/blacklist_test.go`, `internal/api/helpers_test.go`.

## Architecture

Go module: `forsaken-mail`. Key dependencies: `coder/websocket`, `emersion/go-smtp`, `jhillyerd/enmime`, `mattn/go-sqlite3`, `golang.org/x/oauth2`, `nikoksr/notify`.

### Data Flow

```
External mail --(port 25)--> internal/smtp (go-smtp)
  -> enmime parse -> mail.Router.Handle()
    -> SQLite (mail.Store)
    -> WebSocket Hub broadcast to clients on that shortId
    -> Webhook notification (async, DingTalk/Telegram/Slack via nikoksr/notify)

Browser <--(WebSocket /ws)--> internal/ws (Hub)
  -> client subscribes to a shortId
  -> hub broadcasts incoming mail to matching clients

Browser <--(HTTP /api/*)--> internal/api (http.NewServeMux)
  -> serves React SPA from embed/ (Go embed)
  -> API routes behind OAuth session middleware
```

### Backend (`internal/`)

- **`config/`** — env var parsing into Config struct (A-class: immutable at startup)
- **`settings/`** — SQLite-backed runtime settings (B-class: mutable via Admin API, seeded from env on first run)
- **`smtp/`** — go-smtp Backend/Session implementation + per-IP rate limiter
- **`mail/`** — SQLite mail CRUD, router (SMTP->store->WS->webhook), periodic cleanup goroutine
- **`ws/`** — WebSocket hub: manages shortId->client mappings, broadcasts mail
- **`auth/`** — two modes: OAuth2 (GitHub/Google) or local (username/password via `AUTH_MODE`); AES-GCM session cookies; `login_whitelist` restricts OAuth logins (only checked in OAuth callback, not middleware, since local auth stores username not email)
- **`api/`** — HTTP handlers on stdlib `http.NewServeMux` with Go 1.22+ route patterns (`GET /api/mails/{id}`, `PUT /api/mails/{id}/read`, `GET /auth/{provider}/login`); `router.go` defines all routes and applies security headers (CSP, X-Frame-Options)
- **`audit/`** — SQLite audit log CRUD
- **`webhook/`** — Multi-platform webhook notifications via `nikoksr/notify` (Telegram, Slack) + custom DingTalk markdown sender. Config: `webhook_enabled`, `webhook_service`, `webhook_config` (JSON), `webhook_message`.
- **`logger/`** — slog with lumberjack log rotation
- **`i18n/`** — server-side translations (en/zh) for API error messages; resolved from `Accept-Language` header

### Frontend (`web/src/`)

Stack: React 19 + React Router 7 + Tailwind 4 + DaisyUI 5 + Vite 6. i18n via i18next (en/zh locales in `locales/`).

- **`App.jsx`** — Router + AuthContext provider; routes: `/login`, `/`, `/recent`, `/admin`
- **`hooks/useWebSocketConnection.js`** — WebSocket lifecycle, reconnect with exponential backoff (1s–30s)
- **`hooks/useWebSocket.js`** — mailbox tab state, mail operations, localStorage persistence, recent mails
- **`lib/api.js`** — fetch wrapper with `credentials: 'same-origin'`, auto-redirect on 401
- **`pages/`** — LoginPage, MainPage (mailbox UI), RecentMailsPage (all mails with search/filter), AdminPage (audit/settings/status tabs)
- **`components/`** — MailboxTabs, MailList, MailListItem, MailDetail (DOMPurify-sanitized HTML), SettingsTab, StatusTab, AuditLogTab, HelpModal, Navbar, Toast

### SPA Serving

`cmd/server/main.go` `spaHandler` function: requests to `/api/`, `/auth/`, `/ws` go to the API mux; existing static files in `embed/` are served directly; everything else falls through to `index.html` for client-side routing.

### WebSocket Protocol (JSON)

- Client→Server: `{"type":"request_shortid"}`, `{"type":"subscribe","short_id":"..."}`
- Server→Client: `{"type":"shortid","short_id":"..."}`, `{"type":"mail","data":{...}}`, `{"type":"error","message":"..."}`
- Frontend uses `subscribe` for both initial and additional mailboxes; `set_shortid` exists server-side but is unused by the frontend

### Configuration

All config is environment-variable based. See `.env.example`. Two tiers:
- **A-class** (env vars, immutable at runtime): PORT, AUTH_MODE (`oauth`|`local`), OAUTH_*, ADMIN_USERNAME/PASSWORD (for local mode), SESSION_SECRET, COOKIE_SECURE, DB_PATH, MAILIN_*
- **B-class** (SQLite `settings` table, mutable via `PUT /api/admin/settings`): mail_host (comma-separated for multi-domain), site_title, login_whitelist, keyword_blacklist, webhook_* (enabled/service/config/message), audit_mail_received, retention settings (0 = permanent/disabled)

### Key Domain Concepts

- **Multi-domain**: `mail_host` accepts comma-separated domains (e.g. `example.com,test.com`). SMTP accepts all listed domains. Frontend shows a domain selector when multiple hosts are configured.
- **keyword_blacklist**: Controls mailbox **creation**, not mail content. Prevents users from creating shortIds containing blacklisted keywords (e.g. `admin`, `root`). Enforced client-side (from `/api/config`) and server-side in WebSocket `subscribe`/`set_shortid` handlers. Does NOT drop incoming mail to blacklisted addresses.

## Dev Workflow

For local development, run the Go backend and Vite dev server separately:
1. `go run ./cmd/server` (serves API on :3000)
2. `cd web && npm run dev` (serves frontend on :5173, proxies API/WS/auth to :3000)

The Vite dev proxy is configured in `web/vite.config.js` to forward `/api`, `/ws`, and `/auth` paths.

## Deployment

Production server: `ubuntu@130.162.245.79:/home/ubuntu/forsaken-mail`

Deploy steps:
1. Commit and push to `origin/refactor/go-react`
2. SSH to server: `ssh ubuntu@130.162.245.79`
3. `cd /home/ubuntu/forsaken-mail && git pull origin refactor/go-react`
4. `mkdir -p data && sudo chown -R 100:101 data` (first deploy or if data/ was recreated; container `appuser` is UID 100)
5. `docker compose build && docker compose up -d`

The server runs via Docker Compose (ports 25 for SMTP, 3081 for HTTP). No Go toolchain needed on the dev machine — all builds happen in Docker.

## Important Notes

- **No local Go**: The dev Mac does not have Go installed. All Go builds must happen via Docker or on the server.
- **QQ SMTP test**: The `/api/test-email` endpoint sends test emails via QQ SMTP. Credentials are passed per-request from the frontend, not stored server-side. Never commit `.env` files or auth codes.
- **Private data**: `.env` contains OAuth secrets, session keys, and webhook tokens. It is gitignored. Never expose these in commits or logs.
