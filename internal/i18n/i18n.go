package i18n

import (
	"fmt"
	"net/http"
	"strings"
)

// translations maps language -> key -> translated text.
var translations = map[string]map[string]string{
	"en": {
		"method_not_allowed":          "method not allowed",
		"provider_required":           "provider is required",
		"unsupported_provider":        "unsupported provider",
		"internal_server_error":       "internal server error",
		"missing_state_cookie":        "missing or expired state cookie",
		"state_mismatch":              "state mismatch",
		"code_required":               "code parameter is required",
		"oauth_auth_failed":           "OAuth authentication failed",
		"get_email_failed":            "Failed to get user email",
		"shortid_required":            "shortId parameter is required",
		"list_mails_failed":           "failed to list mails",
		"query_audit_failed":          "failed to query audit logs",
		"get_settings_failed":         "failed to get settings",
		"invalid_request_body":        "invalid request body",
		"update_settings_failed":      "failed to update settings",
		"audit_log_failed":            "settings updated but audit log failed",
		"domain_required":             "domain parameter is required",
		"sender_auth_required":        "sender_email and auth_code are required",
		"mail_host_not_configured":    "mail_host not configured",
		"forbidden":                   "forbidden",
		"invalid_message_format":      "invalid message format",
		"invalid_short_id":            "invalid short id",
		"shortid_in_blacklist":        "short id in blacklist",
		"unknown_message_type":        "unknown message type",
		"smtp_connect_failed":         "Failed to connect to QQ SMTP: %v",
		"tls_handshake_failed":        "TLS handshake failed: %v",
		"smtp_client_error":           "SMTP client error: %v",
		"smtp_auth_failed":            "SMTP auth failed (check email and auth code): %v",
		"mail_from_failed":            "MAIL FROM failed: %v",
		"rcpt_to_failed":              "RCPT TO failed: %v",
		"data_command_failed":         "DATA command failed: %v",
		"write_message_failed":        "Failed to write message: %v",
		"close_message_failed":        "Failed to close message: %v",
		"test_email_sent":             "Test email sent from %s to %s via QQ SMTP",
		"webhook_token_empty":         "Webhook token/url is empty or invalid.",
		"webhook_parse_failed":        "Failed to parse DingTalk response.",
		"webhook_request_failed":      "Webhook request failed: %v",
		"mail_not_found":              "mail not found",
		"invalid_mail_id":             "invalid mail ID",
		"login_failed":                "invalid username or password",
		"invalid_request":             "invalid request",
	},
	"zh": {
		"method_not_allowed":          "不允许的请求方法",
		"provider_required":           "需要提供 provider 参数",
		"unsupported_provider":        "不支持的提供商",
		"internal_server_error":       "服务器内部错误",
		"missing_state_cookie":        "state cookie 缺失或已过期",
		"state_mismatch":              "state 不匹配",
		"code_required":               "需要提供 code 参数",
		"oauth_auth_failed":           "OAuth 认证失败",
		"get_email_failed":            "获取用户邮箱失败",
		"shortid_required":            "需要提供 shortId 参数",
		"list_mails_failed":           "获取邮件列表失败",
		"query_audit_failed":          "查询审计日志失败",
		"get_settings_failed":         "获取设置失败",
		"invalid_request_body":        "请求体无效",
		"update_settings_failed":      "更新设置失败",
		"audit_log_failed":            "设置已更新但审计日志记录失败",
		"domain_required":             "需要提供 domain 参数",
		"sender_auth_required":        "需要提供发件邮箱和授权码",
		"mail_host_not_configured":    "mail_host 未配置",
		"forbidden":                   "无权访问",
		"invalid_message_format":      "消息格式无效",
		"invalid_short_id":            "无效的短 ID",
		"shortid_in_blacklist":        "短 ID 包含黑名单关键字",
		"unknown_message_type":        "未知的消息类型",
		"smtp_connect_failed":         "连接 QQ SMTP 失败：%v",
		"tls_handshake_failed":        "TLS 握手失败：%v",
		"smtp_client_error":           "SMTP 客户端错误：%v",
		"smtp_auth_failed":            "SMTP 认证失败（请检查邮箱和授权码）：%v",
		"mail_from_failed":            "MAIL FROM 失败：%v",
		"rcpt_to_failed":              "RCPT TO 失败：%v",
		"data_command_failed":         "DATA 命令失败：%v",
		"write_message_failed":        "写入邮件失败：%v",
		"close_message_failed":        "关闭邮件失败：%v",
		"test_email_sent":             "测试邮件已通过 QQ SMTP 从 %s 发送至 %s",
		"webhook_token_empty":         "Webhook Token/URL 为空或无效。",
		"webhook_parse_failed":        "解析钉钉响应失败。",
		"webhook_request_failed":      "Webhook 请求失败：%v",
		"mail_not_found":              "邮件未找到",
		"invalid_mail_id":             "无效的邮件 ID",
		"login_failed":                "用户名或密码错误",
		"invalid_request":             "无效的请求",
	},
}

// NormalizeLang extracts a two-letter language code from an Accept-Language value.
// Defaults to "en" if the language is not supported.
func NormalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if len(lang) >= 2 {
		lang = lang[:2]
	}
	if _, ok := translations[lang]; ok {
		return lang
	}
	return "en"
}

// LangFromRequest extracts the preferred language from the Accept-Language header.
func LangFromRequest(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	if accept == "" {
		return "en"
	}
	// Take the first language tag (before comma or semicolon).
	for _, part := range strings.FieldsFunc(accept, func(r rune) bool {
		return r == ',' || r == ';'
	}) {
		normalized := NormalizeLang(strings.TrimSpace(part))
		if normalized != "en" {
			return normalized
		}
	}
	// If the first part normalized to "en", still return it.
	return NormalizeLang(accept)
}

// T returns the translated string for the given key and language.
// Falls back to English, then to the key itself.
func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	if m, ok := translations["en"]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	return key
}

// Tfmt returns the translated string formatted with fmt.Sprintf.
func Tfmt(lang, key string, args ...any) string {
	return fmt.Sprintf(T(lang, key), args...)
}
