# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Forsaken-Mail is a self-hosted disposable/temporary email service. Users receive emails at random or custom addresses on a configured domain and view them in real-time via a web UI. Optional DingTalk webhook notifications are supported.

The active codebase is **Go backend + React frontend**. Legacy Node.js code (app.js, bin/www, modules/, routes/, public/) still exists in the repo but is superseded by the Go rewrite.

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

**Tests:** No test suite exists for either backend or frontend.

## Architecture

### Data Flow

```
External mail --(port 25)--> internal/smtp (go-smtp)
  -> enmime parse -> mail.Router.Handle()
    -> SQLite (mail.Store)
    -> WebSocket Hub broadcast to clients on that shortId
    -> DingTalk webhook (async)

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
- **`auth/`** — OAuth2 (GitHub/Google), AES-GCM session cookies, email whitelist middleware
- **`api/`** — HTTP handlers on stdlib `http.NewServeMux`; `router.go` defines all routes
- **`audit/`** — SQLite audit log CRUD
- **`webhook/`** — DingTalk webhook sender
- **`logger/`** — slog with lumberjack log rotation

### Frontend (`web/src/`)

Stack: React 19 + React Router 7 + Tailwind 4 + DaisyUI 5 + Vite 6.

- **`App.jsx`** — Router + AuthContext provider; routes: `/login`, `/`, `/admin`
- **`hooks/useWebSocket.js`** — WebSocket lifecycle, mail state, exponential backoff reconnect (1s–30s), localStorage shortId history
- **`lib/api.js`** — fetch wrapper with `credentials: 'same-origin'`, auto-redirect on 401
- **`pages/`** — LoginPage (OAuth buttons), MainPage (mailbox UI), AdminPage (audit/settings/status tabs)
- **`components/`** — MailboxAddress, MailList, MailDetail (DOMPurify-sanitized HTML), MailHistory, HelpModal, Navbar

### SPA Serving

`cmd/server/main.go` `spaHandler` function: requests to `/api/`, `/auth/`, `/ws` go to the API mux; existing static files in `embed/` are served directly; everything else falls through to `index.html` for client-side routing.

### WebSocket Protocol (JSON)

- Client→Server: `{"type":"request_shortid"}` or `{"type":"set_shortid","short_id":"..."}`
- Server→Client: `{"type":"shortid","short_id":"..."}`, `{"type":"mail","data":{...}}`, `{"type":"error","message":"..."}`

### Configuration

All config is environment-variable based. See `.env.example`. Two tiers:
- **A-class** (env vars, immutable at runtime): PORT, OAUTH_*, SESSION_SECRET, DB_PATH, MAILIN_*
- **B-class** (SQLite `settings` table, mutable via `PUT /api/admin/settings`): mail_host, site_title, allowed_emails, keyword_blacklist, retention settings

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
4. `docker compose build && docker compose up -d`

The server runs via Docker Compose (ports 25 for SMTP, 3081 for HTTP). No Go toolchain needed on the dev machine — all builds happen in Docker.

## Important Notes

- **No local Go**: The dev Mac does not have Go installed. All Go builds must happen via Docker or on the server.
- **QQ SMTP test**: The `/api/test-email` endpoint sends test emails via QQ SMTP. Credentials are passed per-request from the frontend, not stored server-side. Never commit `.env` files or auth codes.
- **Private data**: `.env` contains OAuth secrets, session keys, and webhook tokens. It is gitignored. Never expose these in commits or logs.
