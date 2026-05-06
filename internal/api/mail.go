package api

import (
	"net/http"
	"strconv"

	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/mail"
)

// handleMails responds to GET /api/mails?shortId=xxx with the mail list.
// Pass ?reextract=true to re-run code extraction on existing mails.
func (rt *Router) handleMails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	shortID := r.URL.Query().Get("shortId")
	if shortID == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "shortid_required"))
		return
	}

	// Re-extract codes/links for existing mails if requested.
	if r.URL.Query().Get("reextract") == "true" {
		rt.mailStore.Reextract(shortID)
	}

	mails, err := rt.mailStore.ListByShortID(shortID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	// Ensure JSON array, not null.
	if mails == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	writeJSON(w, http.StatusOK, mails)
}

// handleRecentMails responds to GET /api/mails/recent with the latest mails across all mailboxes.
func (rt *Router) handleRecentMails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	mails, err := rt.mailStore.ListRecent(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	if mails == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	writeJSON(w, http.StatusOK, mails)
}

// handleAllMails responds to GET /api/mails/all with paginated mails across all mailboxes.
// Supports ?page=1&pageSize=20&short_id=xxx&from=xxx&q=keyword for filtering.
func (rt *Router) handleAllMails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	shortID := r.URL.Query().Get("short_id")
	from := r.URL.Query().Get("from")
	query := r.URL.Query().Get("q")

	mails, total, err := rt.mailStore.ListAll(page, pageSize, shortID, from, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	if mails == nil {
		mails = []mail.Mail{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"emails":   mails,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// handleMailFilters responds to GET /api/mails/filters with distinct senders and recipients.
func (rt *Router) handleMailFilters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	senders, err := rt.mailStore.ListDistinctSenders(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}
	recipients, err := rt.mailStore.ListDistinctRecipients(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "list_mails_failed"))
		return
	}

	if senders == nil {
		senders = []string{}
	}
	if recipients == nil {
		recipients = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"senders":    senders,
		"recipients": recipients,
	})
}
