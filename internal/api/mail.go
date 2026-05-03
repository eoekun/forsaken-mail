package api

import (
	"net/http"

	"forsaken-mail/internal/i18n"
)

// handleMails responds to GET /api/mails?shortId=xxx with the mail list.
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
