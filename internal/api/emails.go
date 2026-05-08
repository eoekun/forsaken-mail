package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/mail"
)

// toEmailResponse converts a mail.Mail to an emailResponse.
func toEmailResponse(m *mail.Mail) emailResponse {
	return emailResponse{
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
	}
}

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

// handleListEmailsByShortID returns all emails for /api/emails/{shortId}.
func (rt *Router) handleListEmailsByShortID(w http.ResponseWriter, r *http.Request) {
	lang := i18n.LangFromRequest(r)
	shortID := r.PathValue("shortId")
	if shortID == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "shortid_required"))
		return
	}

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
		resp.Emails = append(resp.Emails, toEmailResponse(&m))
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleDeleteAllEmailsByShortID deletes all emails for /api/emails/{shortId}.
func (rt *Router) handleDeleteAllEmailsByShortID(w http.ResponseWriter, r *http.Request) {
	shortID := r.PathValue("shortId")
	if shortID == "" {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "shortid_required"))
		return
	}
	rt.handleDeleteAllEmails(w, r, shortID)
}

// handleGetEmailByPath returns a single email for /api/emails/{shortId}/{mailId}.
func (rt *Router) handleGetEmailByPath(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("mailId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "invalid_mail_id"))
		return
	}
	rt.handleGetEmail(w, r, id)
}

// handleDeleteEmailByPath deletes a single email for /api/emails/{shortId}/{mailId}.
func (rt *Router) handleDeleteEmailByPath(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("mailId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "invalid_mail_id"))
		return
	}
	rt.handleDeleteEmail(w, r, id)
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

	writeJSON(w, http.StatusOK, toEmailResponse(m))
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
	lang := i18n.LangFromRequest(r)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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
