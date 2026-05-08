package mail

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/settings"
	"forsaken-mail/internal/ws"
)

// WebhookSender is the interface for sending webhook notifications.
// Implementations live in the webhook package.
type WebhookSender interface {
	Send(from, to, subject, text string, codes []string)
}

// Broadcaster pushes mailbox updates to websocket clients.
type Broadcaster interface {
	SendTo(shortID string, data any)
}

// Processor orchestrates incoming mail persistence and fan-out.
type Processor struct {
	mailStore  *Store
	hub        Broadcaster
	settings   *settings.Service
	auditStore *audit.Store
	webhook    WebhookSender
}

// NewProcessor creates a new Processor with the given dependencies.
func NewProcessor(mailStore *Store, hub Broadcaster, settings *settings.Service, auditStore *audit.Store, webhook WebhookSender) *Processor {
	return &Processor{
		mailStore:  mailStore,
		hub:        hub,
		settings:   settings,
		auditStore: auditStore,
		webhook:    webhook,
	}
}

// Handle processes an incoming mail: saves it, pushes via WebSocket, records an
// audit event, and sends a webhook notification for each valid recipient.
// The senderIP parameter is the remote SMTP client IP for audit logging.
func (r *Processor) Handle(from string, toList []string, subject, textBody, htmlBody string, rawSize int64, senderIP string) {
	values, err := r.settings.Load()
	if err != nil {
		slog.Error("failed to load runtime mail settings", "error", err)
		return
	}
	mailHost := strings.ToLower(strings.TrimSpace(values.MailHost))

	for _, addr := range toList {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		shortID, domain := extractShortID(addr)
		if shortID == "" || domain == "" {
			slog.Warn("invalid recipient address", "addr", addr)
			continue
		}
		if !IsDomainAllowed(domain, mailHost) {
			slog.Warn("recipient domain does not match mail_host", "addr", addr, "mail_host", mailHost)
			continue
		}

		// Save to database.
		m := &Mail{
			ShortID:  shortID,
			FromAddr: from,
			ToAddr:   addr,
			Subject:  subject,
			TextBody: textBody,
			HTMLBody: htmlBody,
			RawSize:  rawSize,
		}
		if err := r.mailStore.Save(m); err != nil {
			slog.Error("failed to save mail", "short_id", shortID, "error", err)
			continue
		}

		// Push to WebSocket clients watching this shortID.
		r.hub.SendTo(shortID, ws.MailData{
			ID:             m.ID,
			From:           from,
			To:             addr,
			Subject:        subject,
			HTML:           htmlBody,
			IsRead:         false,
			ExtractedCodes: m.ExtractedCodes,
			ExtractedLinks: m.ExtractedLinks,
			CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		})

		// Record audit event (if enabled).
		if values.AuditMailReceived {
			detailMap := map[string]any{
				"from":    from,
				"to":      addr,
				"subject": subject,
				"codes":   m.ExtractedCodes,
			}
			detailBytes, _ := json.Marshal(detailMap)
			if err := r.auditStore.Record("MAIL_RECEIVED", addr, string(detailBytes), senderIP); err != nil {
				slog.Error("failed to record audit event", "error", err)
			}
		}

		// Send webhook notification asynchronously.
		if r.webhook != nil {
			go r.webhook.Send(from, addr, subject, textBody, m.ExtractedCodes)
		}
	}
}

// extractShortID parses an email address into the local part (shortID) and
// domain. Returns empty strings if the address is malformed.
func extractShortID(addr string) (shortID, domain string) {
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return "", ""
	}
	shortID = strings.TrimSpace(parts[0])
	domain = strings.ToLower(strings.TrimSpace(parts[1]))
	if shortID == "" || domain == "" {
		return "", ""
	}
	return shortID, domain
}
