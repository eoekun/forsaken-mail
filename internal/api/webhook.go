package api

import (
	"net/http"

	"forsaken-mail/internal/i18n"
)

// webhookTestRequest is the expected body for POST /api/webhook/test.
type webhookTestRequest struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// handleWebhookTest responds to POST /api/webhook/test by sending a test DingTalk message.
func (rt *Router) handleWebhookTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	var req webhookTestRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "invalid_request_body"))
		return
	}

	result, err := rt.webhook.SendTest(req.Token, req.Message, lang)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "webhook_request_failed", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
