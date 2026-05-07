package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
	"github.com/nikoksr/notify/service/slack"

	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/settings"
)

// Result represents the outcome of a webhook call.
type Result struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
}

// Sender sends webhook notifications to configured services.
type Sender struct {
	settings *settings.Store
}

// NewSender creates a new webhook Sender.
func NewSender(settings *settings.Store) *Sender {
	return &Sender{settings: settings}
}

// parseConfig parses the JSON config string into a map.
func parseConfig(configStr string) (map[string]string, error) {
	var config map[string]string
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

// parseChatID parses a chat ID from JSON number or plain string.
func parseChatID(chatID string) int64 {
	var n int64
	if err := json.Unmarshal([]byte(chatID), &n); err == nil {
		return n
	}
	for _, c := range chatID {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}

// newTelegramService creates a Telegram notify service from config.
// Returns the service, chat ID, and any error.
func newTelegramService(config map[string]string) (*telegram.Telegram, int64, error) {
	token := config["token"]
	chatID := config["chat_id"]
	if token == "" || chatID == "" {
		return nil, 0, nil
	}

	svc, err := telegram.New(token)
	if err != nil {
		return nil, 0, err
	}

	chatIDInt := parseChatID(chatID)
	if chatIDInt == 0 {
		return nil, 0, nil
	}

	svc.AddReceivers(chatIDInt)
	return svc, chatIDInt, nil
}

// newSlackService creates a Slack notify service from config.
// Returns the service, channel, and any error.
func newSlackService(config map[string]string) (*slack.Slack, string, error) {
	token := config["token"]
	channel := config["channel"]
	if token == "" || channel == "" {
		return nil, "", nil
	}

	svc := slack.New(token)
	svc.AddReceivers(channel)
	return svc, channel, nil
}

// dingtalkToken extracts the DingTalk token from config.
func dingtalkToken(config map[string]string) string {
	if t := config["token"]; t != "" {
		return t
	}
	return config["url"]
}

// notifySend sends a message via a notify.Notifier service.
func notifySend(svc notify.Notifier, subject, body string) error {
	n := notify.New()
	n.UseServices(svc)
	return n.Send(context.Background(), subject, body)
}

// Send sends a notification about a received email.
func (s *Sender) Send(from, to, subject, text string, codes []string) {
	enabled, _ := s.settings.Get("webhook_enabled")
	if enabled != "1" {
		return
	}

	service, _ := s.settings.Get("webhook_service")
	configStr, _ := s.settings.Get("webhook_config")
	messageTemplate, _ := s.settings.Get("webhook_message")

	config, err := parseConfig(configStr)
	if err != nil {
		slog.Error("failed to parse webhook_config", "error", err)
		return
	}

	switch strings.ToLower(strings.TrimSpace(service)) {
	case "dingtalk":
		s.sendDingTalk(from, to, subject, text, codes, messageTemplate, config)
	case "telegram":
		s.sendTelegram(from, to, subject, text, codes, messageTemplate, config)
	case "slack":
		s.sendSlack(from, to, subject, text, codes, messageTemplate, config)
	default:
		slog.Warn("unknown webhook service", "service", service)
	}
}

// sendDingTalk sends via custom DingTalk implementation (supports markdown).
func (s *Sender) sendDingTalk(from, to, subject, text string, codes []string, template string, config map[string]string) {
	token := dingtalkToken(config)
	if token == "" {
		return
	}

	body := BuildMailMarkdown(template, from, to, subject, text, codes)
	result, err := SendDingTalkMarkdown(token, "New Mail", body)
	if err != nil {
		slog.Error("DingTalk webhook request failed", "error", err)
		return
	}
	if !result.OK {
		slog.Error("DingTalk webhook returned non-ok result", "status_code", result.StatusCode, "message", result.Message)
	}
}

// sendTelegram sends via notify Telegram service.
func (s *Sender) sendTelegram(from, to, subject, text string, codes []string, template string, config map[string]string) {
	svc, _, err := newTelegramService(config)
	if err != nil {
		slog.Error("failed to create telegram service", "error", err)
		return
	}
	if svc == nil {
		slog.Warn("telegram webhook misconfigured: missing token or chat_id")
		return
	}

	body := BuildPlainText(template, from, to, subject, text, codes)
	if err := notifySend(svc, "New Mail", body); err != nil {
		slog.Error("telegram webhook send failed", "error", err)
	}
}

// sendSlack sends via notify Slack service.
func (s *Sender) sendSlack(from, to, subject, text string, codes []string, template string, config map[string]string) {
	svc, _, err := newSlackService(config)
	if err != nil {
		slog.Error("failed to create slack service", "error", err)
		return
	}
	if svc == nil {
		slog.Warn("slack webhook misconfigured: missing token or channel")
		return
	}

	body := BuildPlainText(template, from, to, subject, text, codes)
	if err := notifySend(svc, "New Mail", body); err != nil {
		slog.Error("slack webhook send failed", "error", err)
	}
}

// SendTest sends a test message with the given config.
func (s *Sender) SendTest(service, configStr, message, lang string) (*Result, error) {
	config, err := parseConfig(configStr)
	if err != nil {
		return &Result{OK: false, Message: "Invalid webhook config JSON."}, nil
	}

	text := strings.TrimSpace(message)
	if text == "" {
		text = "Tmail test message."
	}

	switch strings.ToLower(strings.TrimSpace(service)) {
	case "dingtalk":
		token := dingtalkToken(config)
		if token == "" {
			return &Result{OK: false, Message: i18n.T(lang, "webhook_token_empty")}, nil
		}
		return SendDingTalkText(token, text)

	case "telegram":
		svc, _, err := newTelegramService(config)
		if err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		if svc == nil {
			return &Result{OK: false, Message: "Missing token or chat_id."}, nil
		}
		if err := notifySend(svc, "Test", text); err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		return &Result{OK: true, Message: "ok"}, nil

	case "slack":
		svc, _, err := newSlackService(config)
		if err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		if svc == nil {
			return &Result{OK: false, Message: "Missing token or channel."}, nil
		}
		if err := notifySend(svc, "Test", text); err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		return &Result{OK: true, Message: "ok"}, nil

	default:
		return &Result{OK: false, Message: "Unknown service: " + service}, nil
	}
}
