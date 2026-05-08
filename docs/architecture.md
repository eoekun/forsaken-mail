# Architecture Notes

This document captures the enforced boundaries after the phase 1 cleanup and the target layering the codebase is moving toward.

## Backend

- HTTP routes are split by behavior:
  - page-oriented auth flows keep redirect semantics (`/auth/*`, `/login`)
  - API and WebSocket endpoints return structured errors, including `401 {"error":"unauthorized"}` for unauthenticated requests
- `GET` endpoints are read-only. Administrative or repair work must use explicit mutation routes.
- Static asset serving is a runtime concern of the Go server. The server serves files from `embed/`, which is treated as generated output.

Target layering for the next refactor step:

- `handler`: HTTP parsing, auth, response formatting
- `service`: business orchestration
- `store`: persistence and low-level data access

## Frontend

- `web/dist/` is the frontend bundler output.
- `embed/` is a server staging directory populated from `web/dist/` during `npm run build`.
- Pages should consume stable data access and state APIs rather than attaching persistence or subscription side effects directly in page components.

Target layering for the next refactor step:

- connection layer: WebSocket lifecycle and reconnect behavior
- state layer: mailbox, unread, selection, and recent-mail state transitions
- page layer: rendering and interaction orchestration only

## Configuration

- Startup configuration comes from environment variables and is required to boot the process.
- Runtime settings live in the database-backed settings store and are editable via admin APIs.
- When the same concept exists in both places, startup config establishes boot-time defaults and runtime settings own mutable behavior after startup.

## API Contracts

- API auth failures return `401` JSON responses instead of HTML redirects.
- Query parameters are for reads; state-changing operations use explicit `POST`, `PUT`, or `DELETE` routes.
- Compatibility routes should be registered explicitly by method and path instead of relying on a catch-all prefix handler.
