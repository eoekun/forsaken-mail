package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	dingtalkHost        = "oapi.dingtalk.com"
	dingtalkDefaultPath = "/robot/send"
	httpTimeout         = 10 * time.Second
)

var urlPattern = regexp.MustCompile(`^https?://`)

// httpClient is reused across requests for connection pooling.
// It uses http.DefaultTransport which respects HTTP_PROXY/HTTPS_PROXY env vars.
var httpClient = &http.Client{
	Timeout:   httpTimeout,
	Transport: http.DefaultTransport,
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

// dingtalkTextRequest is the JSON payload for DingTalk text message.
type dingtalkTextRequest struct {
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

// webhookTarget represents a parsed DingTalk webhook endpoint.
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

// SendDingTalkMarkdown sends a markdown message to DingTalk.
func SendDingTalkMarkdown(tokenOrURL, title, text string) (*Result, error) {
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

	return postDingtalk(target, payload)
}

// SendDingTalkText sends a plain text message to DingTalk.
func SendDingTalkText(tokenOrURL, text string) (*Result, error) {
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

	return postDingtalk(target, payload)
}

func postDingtalk(target *webhookTarget, payload []byte) (*Result, error) {
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

