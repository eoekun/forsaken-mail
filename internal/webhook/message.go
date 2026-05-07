package webhook

import (
	"strings"
	"time"
)

const (
	maxPreviewLength = 200
	maxMessageLength = 1800
)

// BuildMailMarkdown constructs a DingTalk markdown-formatted notification.
// If codes are non-empty, highlights verification codes instead of email body.
func BuildMailMarkdown(template, from, to, subject, text string, codes []string) string {
	title := strings.TrimSpace(template)
	if title == "" {
		title = "📬 New Mail"
	}

	fromSafe := sanitizeSingleLine(from, "unknown")
	toSafe := sanitizeSingleLine(to, "unknown")
	subjectSafe := sanitizeSingleLine(subject, "(no subject)")

	var sb strings.Builder
	sb.WriteString("### " + title + "\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("**From:** " + fromSafe + "\n\n")
	sb.WriteString("**To:** " + toSafe + "\n\n")
	sb.WriteString("**Subject:** " + subjectSafe + "\n\n")

	if len(codes) > 0 {
		sb.WriteString("**Verification Codes:**\n\n")
		for _, code := range codes {
			sb.WriteString("> **`" + code + "`**\n\n")
		}
	} else {
		preview := buildTextPreview(text)
		if preview != "" {
			sb.WriteString("**Preview:**\n\n")
			sb.WriteString("> " + preview + "\n")
		}
	}

	message := sb.String()
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength] + "..."
	}
	return message
}

// BuildPlainText constructs a plain-text notification for services that don't
// support markdown (Telegram, Slack, etc. via notify).
func BuildPlainText(template, from, to, subject, text string, codes []string) string {
	title := strings.TrimSpace(template)
	if title == "" {
		title = "New Mail"
	}

	fromSafe := sanitizeSingleLine(from, "unknown")
	toSafe := sanitizeSingleLine(to, "unknown")
	subjectSafe := sanitizeSingleLine(subject, "(no subject)")
	date := time.Now().Format("2006-01-02 15:04:05")

	lines := []string{
		title,
		"",
		"From: " + fromSafe,
		"To: " + toSafe,
		"Subject: " + subjectSafe,
		"Date: " + date,
	}

	if len(codes) > 0 {
		lines = append(lines, "", "Verification Codes:")
		for _, code := range codes {
			lines = append(lines, "  > "+code)
		}
	} else {
		preview := buildTextPreview(text)
		if preview != "" {
			lines = append(lines, "", "Preview: "+preview)
		}
	}

	message := strings.Join(lines, "\n")
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength] + "..."
	}
	return message
}

// sanitizeSingleLine collapses whitespace and returns fallback if empty.
func sanitizeSingleLine(value, fallback string) string {
	normalized := strings.Join(strings.Fields(value), " ")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return fallback
	}
	return normalized
}

// buildTextPreview returns a truncated preview of the email text body.
func buildTextPreview(text string) string {
	normalized := strings.Join(strings.Fields(text), " ")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return ""
	}
	if len(normalized) <= maxPreviewLength {
		return normalized
	}
	return normalized[:maxPreviewLength] + "..."
}
