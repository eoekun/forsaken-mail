package mail

import "strings"

// isBlacklisted checks whether shortID contains any keyword from the
// comma-separated blacklist (case-insensitive). Returns false if the
// blacklist is empty.
func isBlacklisted(shortID, blacklistCSV string) bool {
	if blacklistCSV == "" {
		return false
	}
	lower := strings.ToLower(shortID)
	for _, kw := range strings.Split(blacklistCSV, ",") {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
