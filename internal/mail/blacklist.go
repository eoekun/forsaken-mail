package mail

import "strings"

// isBlacklisted checks whether shortID contains any keyword from the
// comma-separated blacklist (case-insensitive). Returns false if the
// blacklist is empty.
func IsBlacklisted(shortID, blacklistCSV string) bool {
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

// isDomainAllowed checks whether domain is in the comma-separated
// mailHost list (case-insensitive). Always returns true if mailHostCSV
// is empty.
func IsDomainAllowed(domain, mailHostCSV string) bool {
	if mailHostCSV == "" {
		return true
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	for _, d := range strings.Split(mailHostCSV, ",") {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == domain {
			return true
		}
	}
	return false
}

// parseDomains splits a comma-separated domain list into a clean slice.
func ParseDomains(mailHostCSV string) []string {
	var domains []string
	for _, d := range strings.Split(mailHostCSV, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			domains = append(domains, d)
		}
	}
	return domains
}
