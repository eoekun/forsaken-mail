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

// Send sends a notification about a received email.
// It reads the service type and config from the settings store at call time.
// If webhook is not enabled or misconfigured, it silently skips.
func (s *Sender) Send(from, to, subject, text string, codes []string) {
	enabled, _ := s.settings.Get("webhook_enabled")
	if enabled != "1" {
		// Fallback: check legacy dingtalk setting for backward compatibility
		legacyToken, _ := s.settings.Get("dingtalk_webhook_token")
		if strings.TrimSpace(legacyToken) == "" {
			return
		}
		// Legacy mode: send via DingTalk
		s.sendLegacyDingTalk(from, to, subject, text, codes, legacyToken)
		return
	}

	service, _ := s.settings.Get("webhook_service")
	configStr, _ := s.settings.Get("webhook_config")
	messageTemplate, _ := s.settings.Get("webhook_message")

	var config map[string]string
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			slog.Error("failed to parse webhook_config", "error", err)
			return
		}
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
	token := config["token"]
	if token == "" {
		token = config["url"]
	}
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

	token := config["token"]
	chatID := config["chat_id"]
	if token == "" || chatID == "" {
		slog.Warn("telegram webhook misconfigured: missing token or chat_id")
		return
	}

	svc, err := telegram.New(token)
	if err != nil {
		slog.Error("failed to create telegram service", "error", err)
		return
	}

	// Note: Telegram custom API endpoint requires overriding tgbotapi.APIEndpoint global.
	// For network-level proxy, use HTTP_PROXY env var or a reverse proxy.

	var chatIDInt int64
	if err := json.Unmarshal([]byte(chatID), &chatIDInt); err != nil {
		// Try parsing as plain number
		chatIDInt = 0
		for _, c := range chatID {
			if c >= '0' && c <= '9' {
				chatIDInt = chatIDInt*10 + int64(c-'0')
			}
		}
	}
	if chatIDInt == 0 {
		slog.Error("invalid telegram chat_id", "chat_id", chatID)
		return
	}

	svc.AddReceivers(chatIDInt)

	n := notify.New()
	n.UseServices(svc)

	body := BuildPlainText(template, from, to, subject, text, codes)
	if err := n.Send(context.Background(), "New Mail", body); err != nil {
		slog.Error("telegram webhook send failed", "error", err)
	}
}

// sendSlack sends via notify Slack service.
func (s *Sender) sendSlack(from, to, subject, text string, codes []string, template string, config map[string]string) {

	token := config["token"]
	channel := config["channel"]
	if token == "" || channel == "" {
		slog.Warn("slack webhook misconfigured: missing token or channel")
		return
	}

	svc := slack.New(token)
	svc.AddReceivers(channel)

	n := notify.New()
	n.UseServices(svc)

	body := BuildPlainText(template, from, to, subject, text, codes)
	if err := n.Send(context.Background(), "New Mail", body); err != nil {
		slog.Error("slack webhook send failed", "error", err)
	}
}

// sendLegacyDingTalk handles backward compatibility with old dingtalk_webhook_token setting.
func (s *Sender) sendLegacyDingTalk(from, to, subject, text string, codes []string, token string) {
	messageTemplate, _ := s.settings.Get("dingtalk_webhook_message")
	body := BuildMailMarkdown(messageTemplate, from, to, subject, text, codes)
	result, err := SendDingTalkMarkdown(token, "New Mail", body)
	if err != nil {
		slog.Error("DingTalk webhook request failed", "error", err)
		return
	}
	if !result.OK {
		slog.Error("DingTalk webhook returned non-ok result", "status_code", result.StatusCode, "message", result.Message)
	}
}

// SendTest sends a test message with the given config.
// lang is used for translating user-facing messages.
func (s *Sender) SendTest(service, configStr, message, lang string) (*Result, error) {
	var config map[string]string
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return &Result{OK: false, Message: "Invalid webhook config JSON."}, nil
		}
	}

	text := strings.TrimSpace(message)
	if text == "" {
		text = "Forsaken-Mail test message."
	}

	switch strings.ToLower(strings.TrimSpace(service)) {
	case "dingtalk":
		token := config["token"]
		if token == "" {
			token = config["url"]
		}
		if token == "" {
			return &Result{OK: false, Message: i18n.T(lang, "webhook_token_empty")}, nil
		}
		return SendDingTalkText(token, text)

	case "telegram":
		token := config["token"]
		chatID := config["chat_id"]
		if token == "" || chatID == "" {
			return &Result{OK: false, Message: "Missing token or chat_id."}, nil
		}
		svc, err := telegram.New(token)
		if err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		var chatIDInt int64
		for _, c := range chatID {
			if c >= '0' && c <= '9' {
				chatIDInt = chatIDInt*10 + int64(c-'0')
			}
		}
		if chatIDInt == 0 {
			return &Result{OK: false, Message: "Invalid chat_id."}, nil
		}
		svc.AddReceivers(chatIDInt)
		n := notify.New()
		n.UseServices(svc)
		if err := n.Send(context.Background(), "Test", text); err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		return &Result{OK: true, Message: "ok"}, nil

	case "slack":
		token := config["token"]
		channel := config["channel"]
		if token == "" || channel == "" {
			return &Result{OK: false, Message: "Missing token or channel."}, nil
		}
		svc := slack.New(token)
		svc.AddReceivers(channel)
		n := notify.New()
		n.UseServices(svc)
		if err := n.Send(context.Background(), "Test", text); err != nil {
			return &Result{OK: false, Message: err.Error()}, nil
		}
		return &Result{OK: true, Message: "ok"}, nil

	default:
		return &Result{OK: false, Message: "Unknown service: " + service}, nil
	}
}
