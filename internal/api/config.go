package api

import (
	"fmt"
	"net"
	"net/http"

	"forsaken-mail/internal/auth"
)

// handleConfig responds to GET /api/config with site configuration.
func (rt *Router) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	host, _ := rt.settings.Get("mail_host")
	siteTitle, _ := rt.settings.Get("site_title")

	resp := map[string]any{
		"host":       host,
		"site_title": siteTitle,
	}

	// Include user email if authenticated.
	email := auth.GetEmail(r)
	if email != "" {
		resp["email"] = email
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleDomainTest responds to GET /api/domain-test?domain=xxx with DNS MX lookup results.
func (rt *Router) handleDomainTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeError(w, http.StatusBadRequest, "domain parameter is required")
		return
	}

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     false,
			"domain": domain,
			"error":  fmt.Sprintf("MX lookup failed: %v", err),
		})
		return
	}

	type mxEntry struct {
		Host string   `json:"host"`
		Pref uint16   `json:"pref"`
		A    []string `json:"a,omitempty"`
		AAAA []string `json:"aaaa,omitempty"`
	}

	var mxList []mxEntry
	for _, mx := range mxRecords {
		entry := mxEntry{
			Host: mx.Host,
			Pref: mx.Pref,
		}

		// Resolve A/AAAA records for each MX host.
		ips, err := net.LookupIP(mx.Host)
		if err == nil {
			for _, ip := range ips {
				if ip4 := ip.To4(); ip4 != nil {
					entry.A = append(entry.A, ip4.String())
				} else {
					entry.AAAA = append(entry.AAAA, ip.String())
				}
			}
		}

		mxList = append(mxList, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"domain": domain,
		"mx":     mxList,
	})
}
