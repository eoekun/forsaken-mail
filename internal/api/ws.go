package api

import "net/http"

// handleWS delegates to the WebSocket hub for upgrade handling.
func (rt *Router) handleWS(w http.ResponseWriter, r *http.Request) {
	rt.hub.HandleWS(w, r)
}
