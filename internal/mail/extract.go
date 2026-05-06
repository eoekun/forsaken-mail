package mail

import (
	"regexp"
	"strings"
)

var codePatterns = []*regexp.Regexp{
	// English: "verification code", "security code", "auth code", etc.
	regexp.MustCompile(`(?i)(?:verification|verify|confirm|security|auth(?:entication)?)\s*(?:code|pin|number|token)[\s:]*(\d{4,8})`),
	// English: "your code is 123456", "the code: 123456"
	regexp.MustCompile(`(?i)(?:your|the)\s+(?:code|pin|otp|token)\s+(?:is|:)\s*(\d{4,8})`),
	// English: "OTP: 123456", "OTP is 123456"
	regexp.MustCompile(`(?i)\botp[\s:]+(\d{4,8})`),
	// English: "enter 123456", "use code 123456", "input code 123456"
	regexp.MustCompile(`(?i)(?:enter|use|input|submit)\s+(?:(?:the\s+)?code\s+)?(\d{4,8})`),
	// English: "code: 123456", "code：123456"
	regexp.MustCompile(`(?i)code[\s:：]+(\d{4,8})`),
	// English: "pin: 123456"
	regexp.MustCompile(`(?i)\bpin[\s:：]+(\d{4,8})`),
	// English: "token: 123456"
	regexp.MustCompile(`(?i)\btoken[\s:：]+(\d{4,8})`),
	// Chinese: 验证码、确认码、安全码、动态码、校验码
	regexp.MustCompile(`(?:验证码|确认码|安全码|动态码|校验码|短信码)[\s:：]*(\d{4,8})`),
	// Chinese: "验证码为123456", "验证码是123456"
	regexp.MustCompile(`(?:验证码|确认码|安全码|动态码|校验码)[为是][\s:]*(\d{4,8})`),
	// Chinese: "代码为123456", "代码是123456", "代码: 123456"
	regexp.MustCompile(`代码[\s:：]*[为是]?[\s:：]*(\d{4,8})`),
	// Chinese: "输入123456", "填写123456"
	regexp.MustCompile(`(?:输入|填写|使用)[的]?(?:验证码|校验码)?[\s:：]*(\d{4,8})`),
	// HTML: code inside <b>, <strong>, <span> with common class patterns
	regexp.MustCompile(`<(?:b|strong|span[^>]*class="[^"]*(?:code|otp|verify)[^"]*")[^>]*>\s*(\d{4,8})\s*</(?:b|strong|span)>`),
	// Standalone: 4-8 digits on their own line (common in code-only emails)
	regexp.MustCompile(`(?m)^\s*(\d{4,8})\s*$`),
	// Digits with spaces: "123 456" → captures "123456"
	regexp.MustCompile(`(?i)(?:code|otp|pin|验证码)[\s:：]*(\d{2,4}\s+\d{2,6})`),
	// Chinese: "你的 ChatGPT 代码为 624591" — subject-style patterns
	regexp.MustCompile(`(?:代码|验证码|校验码)\s*[为是]\s*(\d{4,8})`),
}

var linkPattern = regexp.MustCompile(`https?://[^\s<>"')\]]+`)

// htmlTagPattern strips HTML tags to get plain text for better code extraction.
var htmlTagPattern = regexp.MustCompile(`(?s)<[^>]*>`)

// Extract parses text and html bodies to find verification codes and URLs.
// Returns deduplicated slices of codes and links.
func Extract(textBody, htmlBody string) (codes []string, links []string) {
	// Normalize line endings and strip HTML tags for better matching.
	normalizedText := strings.ReplaceAll(textBody, "\r\n", "\n")
	strippedHTML := htmlTagPattern.ReplaceAllString(htmlBody, " ")
	strippedHTML = strings.ReplaceAll(strippedHTML, "\r\n", "\n")

	combined := normalizedText + "\n" + strippedHTML

	codeSet := make(map[string]struct{})
	for _, re := range codePatterns {
		for _, m := range re.FindAllStringSubmatch(combined, -1) {
			if len(m) > 1 {
				code := strings.Join(strings.Fields(m[1]), "") // strip inner spaces
				if code != "" && len(code) >= 4 {
					codeSet[code] = struct{}{}
				}
			}
		}
	}

	linkSet := make(map[string]struct{})
	for _, m := range linkPattern.FindAllString(combined, -1) {
		m = strings.TrimRight(m, ".,;:!?)")
		if m != "" {
			linkSet[m] = struct{}{}
		}
	}

	for c := range codeSet {
		codes = append(codes, c)
	}
	for l := range linkSet {
		links = append(links, l)
	}
	return
}
