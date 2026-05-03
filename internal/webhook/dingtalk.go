package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/settings"
)

const (
	dingtalkHost        = "oapi.dingtalk.com"
	dingtalkDefaultPath = "/robot/send"
	maxPreviewLength    = 200
	maxMessageLength    = 1800
	httpTimeout         = 10 * time.Second
)

var urlPattern = regexp.MustCompile(`^https?://`)

// Result represents the outcome of a DingTalk webhook call.
type Result struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
}

// Sender sends DingTalk webhook notifications.
type Sender struct {
	settings *settings.Store
}

// NewSender creates a new DingTalk webhook Sender.
func NewSender(settings *settings.Store) *Sender {
	return &Sender{settings: settings}
}

// Send sends a notification about a received email.
// It reads the token and message template from the settings store.
// If the token is empty, it silently skips.
func (s *Sender) Send(from, to, subject, text string) {
	token, err := s.settings.Get("dingtalk_webhook_token")
	if err != nil {
		slog.Error("failed to get dingtalk_webhook_token", "error", err)
		return
	}
	if strings.TrimSpace(token) == "" {
		return
	}

	messageTemplate, err := s.settings.Get("dingtalk_webhook_message")
	if err != nil {
		slog.Error("failed to get dingtalk_webhook_message", "error", err)
		return
	}

	body := buildMailMessage(messageTemplate, from, to, subject, text)
	result, err := postDingtalkText(token, body)
	if err != nil {
		slog.Error("DingTalk webhook request failed", "error", err)
		return
	}
	if !result.OK {
		slog.Error("DingTalk webhook returned non-ok result", "status_code", result.StatusCode, "message", result.Message)
	}
}

// SendTest sends a test message with the given token.
// lang is used for translating user-facing messages (e.g. "en" or "zh").
func (s *Sender) SendTest(token, message, lang string) (*Result, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return &Result{OK: false, Message: i18n.T(lang, "webhook_token_empty")}, nil
	}

	text := strings.TrimSpace(message)
	if text == "" {
		text = "Forsaken-Mail test message."
	}

	return postDingtalkText(token, text)
}

// buildWebhookTarget parses a token or full URL into the target endpoint.
type webhookTarget struct {
	Hostname string
	Path     string
}

func buildWebhookTarget(tokenOrURL string) *webhookTarget {
	normalized := strings.TrimSpace(tokenOrURL)
	if normalized == "" {
		return nil
	}

	if urlPattern.MatchString(normalized) {
		parsed, err := url.Parse(normalized)
		if err != nil || parsed.Hostname() == "" {
			return nil
		}
		if parsed.Scheme != "https" {
			return nil
		}
		path := parsed.Path
		if parsed.RawQuery != "" {
			path += "?" + parsed.RawQuery
		}
		return &webhookTarget{
			Hostname: parsed.Hostname(),
			Path:     path,
		}
	}

	return &webhookTarget{
		Hostname: dingtalkHost,
		Path:     dingtalkDefaultPath + "?access_token=" + url.QueryEscape(normalized),
	}
}

// dingtalkRequest is the JSON payload for DingTalk robot API.
type dingtalkRequest struct {
	MsgType string              `json:"msgtype"`
	Text    dingtalkTextContent `json:"text"`
}

type dingtalkTextContent struct {
	Content string `json:"content"`
}

// dingtalkResponse is the JSON response from DingTalk robot API.
type dingtalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func postDingtalkText(tokenOrURL, text string) (*Result, error) {
	target := buildWebhookTarget(tokenOrURL)
	if target == nil {
		return &Result{OK: false, Message: "Webhook token/url is empty or invalid."}, nil
	}

	payload, err := json.Marshal(dingtalkRequest{
		MsgType: "text",
		Text:    dingtalkTextContent{Content: text},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("https://%s%s", target.Hostname, target.Path)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var dingResp dingtalkResponse
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return &Result{
			OK:         false,
			Message:    "Failed to parse DingTalk response.",
			StatusCode: resp.StatusCode,
		}, nil
	}

	success := dingResp.ErrCode == 0
	msg := "ok"
	if !success {
		msg = fmt.Sprintf("DingTalk returned errcode=%d, errmsg=%s", dingResp.ErrCode, dingResp.ErrMsg)
	}

	return &Result{
		OK:         success,
		Message:    msg,
		StatusCode: resp.StatusCode,
	}, nil
}

// buildMailMessage constructs a formatted notification message.
func buildMailMessage(template, from, to, subject, text string) string {
	title := strings.TrimSpace(template)
	if title == "" {
		title = "Forsaken-Mail: new email received."
	}

	date := time.Now().Format(time.RFC3339)
	preview := buildTextPreview(text)

	lines := []string{
		title,
		"From: " + sanitizeSingleLine(from, "unknown"),
		"To: " + sanitizeSingleLine(to, "unknown"),
		"Subject: " + sanitizeSingleLine(subject, "(no subject)"),
		"Date: " + date,
	}
	if preview != "" {
		lines = append(lines, "Preview: "+preview)
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
