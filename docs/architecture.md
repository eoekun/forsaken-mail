# Architecture Notes

This document captures the enforced boundaries after the current backend and frontend refactor.

## Backend

- HTTP routes are split by behavior:
  - page-oriented auth flows keep redirect semantics (`/auth/*`, `/login`)
  - API and WebSocket endpoints return structured errors, including `401 {"error":"unauthorized"}` for unauthenticated requests
- `GET` endpoints are read-only. Administrative or repair work must use explicit mutation routes.
- Static asset serving is a runtime concern of the Go server. The server serves files from `embed/`, which is treated as generated output.
- Backend code is layered as:
  - `handler`: request parsing, auth, status codes, JSON encoding
  - `service`: config read/write, status queries, webhook testing, incoming-mail orchestration
  - `store`: SQLite-backed mail, audit, and raw settings persistence
- Incoming mail flow now reads runtime settings through `settings.Service`, then fans out through dedicated collaborators:
  - persistence via `mail.Store`
  - audit via `audit.Store`
  - live mailbox delivery via `ws.Hub`
  - webhook notification via `webhook.Sender`

## Frontend

- `web/dist/` is the frontend bundler output.
- `embed/` is a server staging directory populated from `web/dist/` during `npm run build`.
- Frontend code is layered as:
  - connection layer: `useWebSocketConnection` owns socket lifecycle and reconnect behavior
  - state layer: `useMailboxState` owns mailbox, unread, selection, and recent-mail transitions through a reducer
  - data access and side effects: `web/src/lib/*Api.js`, mailbox storage helpers, and notification helpers
  - page layer: page components compose hooks and render UI without owning persistence or reconnect logic

## Configuration

- Startup configuration comes from environment variables and is required to boot the process.
- Runtime settings live in the database-backed settings store and are editable via admin APIs through a typed `settings.Values` model.
- Startup config establishes boot-time defaults. Runtime settings own mutable behavior after startup.
- Runtime settings updates are centrally normalized and validated before persistence.
- Mutable settings that affect live behavior, such as mailbox blacklist rules, are synchronized to long-lived components when settings change.

## API Contracts

- API auth failures return `401` JSON responses instead of HTML redirects.
- Query parameters are for reads; state-changing operations use explicit `POST`, `PUT`, or `DELETE` routes.
- Compatibility routes should be registered explicitly by method and path instead of relying on a catch-all prefix handler.
- `GET /api/config` returns a typed site bootstrap payload for the SPA, including `hosts` and a normalized keyword blacklist array.
- `GET /api/admin/settings` returns typed settings values rather than an arbitrary string map.
- `PUT /api/admin/settings` accepts only the controlled settings field set; unknown keys and invalid values are rejected.
