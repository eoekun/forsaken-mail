package mail

import (
	"regexp"
	"strings"
)

var codePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:verification|verify|confirm|security|auth)\s*(?:code|pin|number)[\s:]*(\d{4,8})`),
	regexp.MustCompile(`(?i)(?:验证码|确认码|安全码|动态码)[\s:：]*(\d{4,8})`),
	regexp.MustCompile(`(?i)(?:your|the)\s+(?:code|pin)\s+(?:is|:)\s*(\d{4,8})`),
	regexp.MustCompile(`(?i)code[\s:]+(\d{4,8})`),
	regexp.MustCompile(`(?i)pin[\s:]+(\d{4,8})`),
}

var linkPattern = regexp.MustCompile(`https?://[^\s<>"')\]]+`)

// Extract parses text and html bodies to find verification codes and URLs.
// Returns deduplicated slices of codes and links.
func Extract(textBody, htmlBody string) (codes []string, links []string) {
	combined := textBody + "\n" + htmlBody

	codeSet := make(map[string]struct{})
	for _, re := range codePatterns {
		for _, m := range re.FindAllStringSubmatch(combined, -1) {
			if len(m) > 1 {
				code := strings.TrimSpace(m[1])
				if code != "" {
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
