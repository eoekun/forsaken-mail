package api

import (
	"net/http"

	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/service"
)

// webhookTestRequest is the expected body for POST /api/webhook/test.
type webhookTestRequest struct {
	Service string `json:"service"`
	Config  string `json:"config"`
	Message string `json:"message"`
	// Legacy fields for backward compatibility
	Token string `json:"token"`
}

// handleWebhookTest responds to POST /api/webhook/test by sending a test message.
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

	result, err := rt.webhookService.SendTest(service.WebhookTestRequest{
		Service: req.Service,
		Config:  req.Config,
		Message: req.Message,
		Token:   req.Token,
	}, lang)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "webhook_request_failed", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
