Forsaken-Mail
==============

A self-hosted disposable/temporary email service built with Go and React.

[中文文档](./README.zh-CN.md)

## Features

- Receive emails at random or custom addresses on your domain
- Real-time mail delivery via WebSocket
- OAuth2 (GitHub/Google) or local username/password authentication
- Email whitelist for access control
- DingTalk webhook notifications
- Admin panel with audit logs, runtime settings, and system status
- i18n support (English / Chinese)
- SQLite storage with automatic cleanup

## Quick Start

### DNS Setup

To receive emails at `*@subdomain.domain.com`, add two DNS records:

- **MX record**: `subdomain.domain.com  MX  10  mxsubdomain.domain.com`
- **A record**: `mxsubdomain.domain.com  A  <your-server-ip>`

Verify with an [SMTP tester](http://mxtoolbox.com/diagnostic.aspx).

### Docker Compose (Recommended)

```bash
cp .env.example .env   # edit with your values
docker compose up -d --build
```

The service exposes:
- **Port 25** — SMTP (for receiving emails)
- **Port 3000** — Web UI (configurable via `PORT` env var)

Open `http://localhost:3000` in your browser.

### Environment Variables

See `.env.example` for the full list. Required variables:

| Variable | Description |
|---|---|
| `MAIL_HOST` | Email domain shown in UI (e.g. `mail.example.com`) |
| `SESSION_SECRET` | Session encryption key (`openssl rand -hex 32`) |
| `AUTH_MODE` | `oauth` (default) or `local` |
| `OAUTH_CLIENT_ID` / `OAUTH_CLIENT_SECRET` | Required when `AUTH_MODE=oauth` |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | Required when `AUTH_MODE=local` (password min 8 chars) |

Optional: `DINGTALK_WEBHOOK_TOKEN`, `KEYWORD_BLACKLIST`, `SITE_TITLE`, `LOG_LEVEL`, `COOKIE_SECURE`.

### Data Persistence

SQLite database and data are stored in `./data/` (mounted as a Docker volume). First deploy:

```bash
mkdir -p data && sudo chown -R 100:101 data
```

Container runs as `appuser` (UID 100).

## Development

Requires Node.js 24+ for the frontend. The Go backend builds via Docker (no local Go needed).

```bash
# Terminal 1: Go backend (API on :3000)
go run ./cmd/server

# Terminal 2: Vite dev server (frontend on :5173, proxies API/WS/auth to :3000)
cd web && npm install && npm run dev
```

Build frontend for production:

```bash
cd web && npm run build
```

Vite now builds into `web/dist/`, then the build script syncs the generated assets into the server-facing `embed/` directory.

## Architecture

```
External mail ──(SMTP :25)──► go-smtp server
  → enmime parse → mail.Router
    → SQLite store
    → WebSocket broadcast to subscribed clients
    → DingTalk webhook (async)

Browser ◄──(WebSocket /ws)──► ws.Hub
  → client subscribes to a shortId
  → hub broadcasts incoming mail

Browser ◄──(HTTP /api/*)──► http.NewServeMux
  → REST API + React SPA (Go embed)
  → OAuth/local session middleware
```

**Backend** (`internal/`): config, smtp, mail (store/router/cleanup), ws, auth (OAuth2 + local), api, settings, audit, webhook, i18n, logger.

**Frontend** (`web/src/`): React 19 + React Router 7 + Tailwind 4 + DaisyUI 5 + Vite 6. i18n via i18next.

## License

GPL-2.0
