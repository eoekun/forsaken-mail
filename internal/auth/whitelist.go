package auth

import "strings"

// IsEmailAllowed checks whether the given email is in the comma-separated
// allowedEmails list. If allowedEmails is empty, all emails are allowed.
func IsEmailAllowed(email, allowedEmails string) bool {
	if allowedEmails == "" {
		return true
	}
	for _, e := range strings.Split(allowedEmails, ",") {
		if strings.TrimSpace(e) == email {
			return true
		}
	}
	return false
}
