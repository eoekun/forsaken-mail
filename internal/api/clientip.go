package api

import (
	"net"
	"net/http"
	"strings"
)

// clientIP extracts the real client IP from the request, preferring
// X-Forwarded-For (set by reverse proxies like nginx) over RemoteAddr.
// For X-Forwarded-For with multiple hops, only the first (leftmost) IP is used.
func clientIP(r *http.Request) string {
	// X-Forwarded-For: client, proxy1, proxy2
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}

	// X-Real-IP (common with nginx proxy_set_header X-Real-IP $remote_addr)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr (ip:port format)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
