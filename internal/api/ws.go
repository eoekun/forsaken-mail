package api

import (
	"net/http"

	"forsaken-mail/internal/i18n"
)

// handleWS delegates to the WebSocket hub for upgrade handling.
func (rt *Router) handleWS(w http.ResponseWriter, r *http.Request) {
	lang := i18n.LangFromRequest(r)
	rt.hub.HandleWS(w, r, lang)
}
