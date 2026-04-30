package api

import (
	"net/http"
	"strconv"

	"forsaken-mail/internal/auth"
)

// handleAuditLogs responds to GET /api/admin/audit-logs with paginated audit logs.
func (rt *Router) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	event := r.URL.Query().Get("event")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := rt.auditStore.Query(event, offset, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query audit logs")
		return
	}

	// Convert to JSON-friendly format.
	jsonLogs := make([]auditLogJSON, 0, len(logs))
	for _, l := range logs {
		jsonLogs = append(jsonLogs, auditLogJSON{
			ID:        l.ID,
			Event:     l.Event,
			Email:     l.Email,
			Detail:    l.Detail,
			IP:        l.IP,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"logs":  jsonLogs,
		"total": total,
	})
}

// auditLogJSON is the JSON representation of an audit log entry.
type auditLogJSON struct {
	ID        int64  `json:"id"`
	Event     string `json:"event"`
	Email     string `json:"email"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
	CreatedAt string `json:"created_at"`
}

// handleGetSettings responds to GET /api/admin/settings with all settings.
func (rt *Router) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	settings, err := rt.settings.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// handleUpdateSettings responds to PUT /api/admin/settings by updating settings.
func (rt *Router) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var kvs map[string]string
	if err := readJSON(r, &kvs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := rt.settings.SetAll(kvs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	// Record audit event.
	email := auth.GetEmail(r)
	ip := r.RemoteAddr
	if err := rt.auditStore.Record("CONFIG_CHANGED", email, "{}", ip); err != nil {
		writeError(w, http.StatusInternalServerError, "settings updated but audit log failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleStatus responds to GET /api/admin/status with system status information.
func (rt *Router) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mailCount, err := rt.mailStore.Count()
	if err != nil {
		mailCount = -1
	}

	mailHost, _ := rt.settings.Get("mail_host")

	writeJSON(w, http.StatusOK, map[string]any{
		"uptime":     rt.startTime.Format("2006-01-02T15:04:05Z"),
		"mail_count": mailCount,
		"ws_clients": rt.hub.ClientCount(),
		"db_path":    rt.cfg.DBPath,
		"mail_host":  mailHost,
	})
}
