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

// httpClient is reused across requests for connection pooling.
var httpClient = &http.Client{Timeout: httpTimeout}

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
// codes are the extracted verification codes; if non-empty, the notification
// highlights codes instead of the email body.
func (s *Sender) Send(from, to, subject, text string, codes []string) {
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

	body := buildMailMarkdown(messageTemplate, from, to, subject, text, codes)
	result, err := postDingtalkMarkdown(token, "New Mail", body)
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

// dingtalkTextRequest is the JSON payload for DingTalk text message.
type dingtalkTextRequest struct {
	MsgType string              `json:"msgtype"`
	Text    dingtalkTextContent `json:"text"`
}

type dingtalkTextContent struct {
	Content string `json:"content"`
}

// dingtalkMarkdownRequest is the JSON payload for DingTalk markdown message.
type dingtalkMarkdownRequest struct {
	MsgType  string                 `json:"msgtype"`
	Markdown dingtalkMarkdownContent `json:"markdown"`
}

type dingtalkMarkdownContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
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

	payload, err := json.Marshal(dingtalkTextRequest{
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

	resp, err := httpClient.Do(req)
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

func postDingtalkMarkdown(tokenOrURL, title, text string) (*Result, error) {
	target := buildWebhookTarget(tokenOrURL)
	if target == nil {
		return &Result{OK: false, Message: "Webhook token/url is empty or invalid."}, nil
	}

	payload, err := json.Marshal(dingtalkMarkdownRequest{
		MsgType:  "markdown",
		Markdown: dingtalkMarkdownContent{Title: title, Text: text},
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

	resp, err := httpClient.Do(req)
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

// buildMailMarkdown constructs a markdown-formatted notification message.
// If codes are provided, highlights verification codes instead of email body.
func buildMailMarkdown(template, from, to, subject, text string, codes []string) string {
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
