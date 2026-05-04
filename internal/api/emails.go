package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"forsaken-mail/internal/i18n"
)

// emailResponse is the JSON shape returned by the email API.
type emailResponse struct {
	ID             int64    `json:"id"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Subject        string   `json:"subject"`
	TextBody       string   `json:"text_body"`
	HTMLBody       string   `json:"html_body"`
	IsRead         bool     `json:"is_read"`
	ExtractedCodes []string `json:"extracted_codes"`
	ExtractedLinks []string `json:"extracted_links"`
	CreatedAt      string   `json:"created_at"`
}

// emailListResponse wraps a list of emails with metadata.
type emailListResponse struct {
	Emails  []emailResponse `json:"emails"`
	Total   int             `json:"total"`
	ShortID string          `json:"short_id"`
}

// handleEmails dispatches /api/emails/{shortId} and /api/emails/{shortId}/{mailId}.
func (rt *Router) handleEmails(w http.ResponseWriter, r *http.Request) {
	lang := i18n.LangFromRequest(r)

	// Trim prefix and split path.
	path := strings.TrimPrefix(r.URL.Path, "/api/emails/")
	parts := strings.SplitN(path, "/", 2)
	shortID := parts[0]
	if shortID == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "shortid_required"))
		return
	}

	mailID := ""
	if len(parts) > 1 {
		mailID = parts[1]
	}

	if mailID == "" {
		// /api/emails/{shortId}
		switch r.Method {
		case http.MethodGet:
			rt.handleListEmails(w, r, shortID)
		case http.MethodDelete:
			rt.handleDeleteAllEmails(w, r, shortID)
		default:
			writeError(w, http.StatusMethodNotAllowed, i18n.T(lang, "method_not_allowed"))
		}
	} else {
		// /api/emails/{shortId}/{mailId}
		id, err := strconv.ParseInt(mailID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, i18n.T(lang, "invalid_mail_id"))
			return
		}
		switch r.Method {
		case http.MethodGet:
			rt.handleGetEmail(w, r, id)
		case http.MethodDelete:
			rt.handleDeleteEmail(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, i18n.T(lang, "method_not_allowed"))
		}
	}
}

// handleListEmails returns all emails for a short ID.
func (rt *Router) handleListEmails(w http.ResponseWriter, r *http.Request, shortID string) {
	lang := i18n.LangFromRequest(r)

	mails, err := rt.mailStore.ListByShortID(shortID, 1000)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	total, err := rt.mailStore.CountByShortID(shortID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	resp := emailListResponse{
		Emails:  make([]emailResponse, 0),
		Total:   total,
		ShortID: shortID,
	}
	for _, m := range mails {
		resp.Emails = append(resp.Emails, emailResponse{
			ID:             m.ID,
			From:           m.FromAddr,
			To:             m.ToAddr,
			Subject:        m.Subject,
			TextBody:       m.TextBody,
			HTMLBody:       m.HTMLBody,
			IsRead:         m.IsRead,
			ExtractedCodes: m.ExtractedCodes,
			ExtractedLinks: m.ExtractedLinks,
			CreatedAt:      m.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleGetEmail returns a single email by ID.
func (rt *Router) handleGetEmail(w http.ResponseWriter, r *http.Request, id int64) {
	lang := i18n.LangFromRequest(r)

	m, err := rt.mailStore.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, i18n.T(lang, "mail_not_found"))
			return
		}
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	writeJSON(w, http.StatusOK, emailResponse{
		ID:             m.ID,
		From:           m.FromAddr,
		To:             m.ToAddr,
		Subject:        m.Subject,
		TextBody:       m.TextBody,
		HTMLBody:       m.HTMLBody,
		IsRead:         m.IsRead,
		ExtractedCodes: m.ExtractedCodes,
		ExtractedLinks: m.ExtractedLinks,
		CreatedAt:      m.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// handleDeleteAllEmails deletes all emails for a short ID.
func (rt *Router) handleDeleteAllEmails(w http.ResponseWriter, r *http.Request, shortID string) {
	lang := i18n.LangFromRequest(r)

	if err := rt.mailStore.DeleteByShortID(shortID); err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleDeleteEmail deletes a single email by ID.
func (rt *Router) handleDeleteEmail(w http.ResponseWriter, r *http.Request, id int64) {
	lang := i18n.LangFromRequest(r)

	if err := rt.mailStore.DeleteByID(id); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, i18n.T(lang, "mail_not_found"))
			return
		}
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleMailRead handles PUT /api/mails/{id}/read to mark a mail as read.
func (rt *Router) handleMailRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	// Extract ID from path: /api/mails/{id}/read
	path := strings.TrimPrefix(r.URL.Path, "/api/mails/")
	path = strings.TrimSuffix(path, "/read")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "invalid_mail_id"))
		return
	}

	if err := rt.mailStore.MarkAsRead(id); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, i18n.T(lang, "mail_not_found"))
			return
		}
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
