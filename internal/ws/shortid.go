package ws

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// Pools for generating readable short IDs.
var firstNamePool = []string{
	"alex", "mike", "tom", "jack", "leo", "sam", "eric", "lucas", "liam", "noah",
	"emma", "olivia", "sophia", "mia", "ava", "lily", "grace", "ella", "zoe", "nina",
}

var tagPool = []string{
	"mail", "inbox", "user", "note", "cloud", "river", "forest", "stone", "ocean", "field",
	"sun", "moon", "star", "leaf", "bird", "fox", "wolf", "lake", "hill", "wind",
}

var shortIDRegex = regexp.MustCompile(`^[a-z0-9._\-+]{1,64}$`)

// generateShortID produces a cryptographically random human-readable short ID
// of the form firstName + tag + 3-digit number (e.g. "alexsun789"). It validates
// against the blacklist and retries up to 20 times before falling back to a
// timestamp-based ID.
func (h *Hub) generateShortID() string {
	for i := 0; i < 20; i++ {
		name := firstNamePool[secureRandIntn(len(firstNamePool))]
		tag := tagPool[secureRandIntn(len(tagPool))]
		suffix := secureRandIntn(900) + 100
		candidate := fmt.Sprintf("%s%s%d", name, tag, suffix)

		if !shortIDRegex.MatchString(candidate) {
			continue
		}
		if h.isBlacklisted(candidate) {
			continue
		}
		if _, taken := h.clients[candidate]; taken {
			continue
		}
		return candidate
	}

	// Extremely rare fallback path.
	return fmt.Sprintf("mailuser%d", time.Now().UnixMilli())
}

// isBlacklisted checks whether id contains any blacklisted keyword as a
// substring (case-insensitive).
func (h *Hub) isBlacklisted(id string) bool {
	lower := strings.ToLower(id)
	h.blacklistMu.RLock()
	blacklist := append([]string(nil), h.blacklist...)
	h.blacklistMu.RUnlock()
	for _, kw := range blacklist {
		if kw == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// normalizeShortID trims and lowercases the input, returning it only if it
// passes the regex validation. Returns empty string on failure.
func normalizeShortID(id string) string {
	normalized := strings.TrimSpace(strings.ToLower(id))
	if !shortIDRegex.MatchString(normalized) {
		return ""
	}
	return normalized
}

// secureRandIntn returns a cryptographically random int in [0, n).
func secureRandIntn(n int) int {
	nBig := big.NewInt(int64(n))
	result, err := rand.Int(rand.Reader, nBig)
	if err != nil {
		// Fallback: should never happen with crypto/rand
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		return int(binary.BigEndian.Uint64(b) % uint64(n))
	}
	return int(result.Int64())
}
