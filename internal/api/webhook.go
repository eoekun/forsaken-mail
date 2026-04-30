package api

import "net/http"

// webhookTestRequest is the expected body for POST /api/webhook/test.
type webhookTestRequest struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// handleWebhookTest responds to POST /api/webhook/test by sending a test DingTalk message.
func (rt *Router) handleWebhookTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req webhookTestRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := rt.webhook.SendTest(req.Token, req.Message)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "Webhook request failed: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
