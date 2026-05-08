package settings

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"forsaken-mail/internal/config"
)

const (
	defaultSiteTitle          = "Tmail"
	defaultWebhookService     = "dingtalk"
	defaultWebhookMessage     = "new email received."
	defaultMailRetentionHours = 1
	defaultMailMaxCount       = 100
	defaultMaxMailSizeBytes   = 1048576
	defaultAuditRetentionDays = 7
	defaultAuditMaxCount      = 5000
	defaultTestSMTPHost       = "smtp.qq.com"
	defaultTestSMTPPort       = 465
)

// Values is the typed runtime settings model used by services and admin APIs.
type Values struct {
	MailHost           string            `json:"mail_host"`
	SiteTitle          string            `json:"site_title"`
	LoginWhitelist     string            `json:"login_whitelist"`
	KeywordBlacklist   string            `json:"keyword_blacklist"`
	WebhookEnabled     bool              `json:"webhook_enabled"`
	WebhookService     string            `json:"webhook_service"`
	WebhookConfig      map[string]string `json:"webhook_config"`
	WebhookMessage     string            `json:"webhook_message"`
	MailRetentionHours int               `json:"mail_retention_hours"`
	MailMaxCount       int               `json:"mail_max_count"`
	MaxMailSizeBytes   int               `json:"max_mail_size_bytes"`
	AuditMailReceived  bool              `json:"audit_mail_received"`
	AuditRetentionDays int               `json:"audit_retention_days"`
	AuditMaxCount      int               `json:"audit_max_count"`
	TestSMTPHost       string            `json:"test_smtp_host"`
	TestSMTPPort       int               `json:"test_smtp_port"`
}

// MailHostsList returns the normalized mail host list.
func (v Values) MailHostsList() []string {
	return splitCSV(v.MailHost, true)
}

// KeywordBlacklistList returns the normalized blacklist terms.
func (v Values) KeywordBlacklistList() []string {
	return splitCSV(v.KeywordBlacklist, true)
}

type runtimeDefaults struct {
	mailHost  string
	siteTitle string
}

// Service loads, validates, and persists runtime settings as a typed model.
type Service struct {
	store    *Store
	defaults runtimeDefaults
}

// NewService creates a typed runtime settings service.
func NewService(store *Store, cfg *config.Config) *Service {
	defaults := runtimeDefaults{
		siteTitle: defaultSiteTitle,
	}
	if cfg != nil {
		defaults.mailHost = strings.TrimSpace(cfg.MailHost)
		if strings.TrimSpace(cfg.SiteTitle) != "" {
			defaults.siteTitle = strings.TrimSpace(cfg.SiteTitle)
		}
	}

	return &Service{
		store:    store,
		defaults: defaults,
	}
}

// SeedFromEnv seeds the backing store from environment defaults.
func (s *Service) SeedFromEnv(cfg *config.Config) error {
	return s.store.SeedFromEnv(cfg)
}

// Load returns the normalized runtime settings snapshot.
func (s *Service) Load() (Values, error) {
	raw, err := s.store.GetAll()
	if err != nil {
		return Values{}, err
	}
	return decodeValues(raw, s.defaults), nil
}

// Update validates and persists a settings update payload.
func (s *Service) Update(input map[string]any) (Values, []string, error) {
	current, err := s.Load()
	if err != nil {
		return Values{}, nil, err
	}

	next, err := applyUpdate(current, unwrapValuesPayload(input), s.defaults)
	if err != nil {
		return Values{}, nil, err
	}

	currentRaw := encodeValues(current)
	nextRaw := encodeValues(next)

	changed := make([]string, 0, len(nextRaw))
	for key, value := range nextRaw {
		if currentRaw[key] != value {
			changed = append(changed, key)
		}
	}
	sort.Strings(changed)

	if len(changed) == 0 {
		return current, changed, nil
	}

	if err := s.store.SetAll(nextRaw); err != nil {
		return Values{}, nil, err
	}

	return next, changed, nil
}

// MailHosts returns the configured mail host list.
func (s *Service) MailHosts() ([]string, error) {
	values, err := s.Load()
	if err != nil {
		return nil, err
	}
	return splitCSV(values.MailHost, true), nil
}

// KeywordBlacklist returns the configured blacklist terms.
func (s *Service) KeywordBlacklist() ([]string, error) {
	values, err := s.Load()
	if err != nil {
		return nil, err
	}
	return splitCSV(values.KeywordBlacklist, true), nil
}

// LoginWhitelist returns the normalized whitelist CSV string.
func (s *Service) LoginWhitelist() (string, error) {
	values, err := s.Load()
	if err != nil {
		return "", err
	}
	return values.LoginWhitelist, nil
}

func unwrapValuesPayload(input map[string]any) map[string]any {
	if rawValues, ok := input["values"]; ok {
		if nested, ok := rawValues.(map[string]any); ok {
			return nested
		}
	}
	return input
}

func decodeValues(raw map[string]string, defaults runtimeDefaults) Values {
	values := Values{
		MailHost:           firstNonEmpty(raw["mail_host"], defaults.mailHost),
		SiteTitle:          firstNonEmpty(raw["site_title"], defaults.siteTitle),
		LoginWhitelist:     raw["login_whitelist"],
		KeywordBlacklist:   raw["keyword_blacklist"],
		WebhookEnabled:     parseStoredBool(raw["webhook_enabled"], false),
		WebhookService:     raw["webhook_service"],
		WebhookConfig:      parseStoredStringMap(raw["webhook_config"]),
		WebhookMessage:     raw["webhook_message"],
		MailRetentionHours: parseStoredInt(raw["mail_retention_hours"], defaultMailRetentionHours),
		MailMaxCount:       parseStoredInt(raw["mail_max_count"], defaultMailMaxCount),
		MaxMailSizeBytes:   parseStoredInt(raw["max_mail_size_bytes"], defaultMaxMailSizeBytes),
		AuditMailReceived:  parseStoredBool(raw["audit_mail_received"], true),
		AuditRetentionDays: parseStoredInt(raw["audit_retention_days"], defaultAuditRetentionDays),
		AuditMaxCount:      parseStoredInt(raw["audit_max_count"], defaultAuditMaxCount),
		TestSMTPHost:       firstNonEmpty(raw["test_smtp_host"], defaultTestSMTPHost),
		TestSMTPPort:       parseStoredInt(raw["test_smtp_port"], defaultTestSMTPPort),
	}

	normalized, err := normalizeValues(values, defaults)
	if err != nil {
		return fallbackValues(defaults)
	}
	return normalized
}

func applyUpdate(current Values, input map[string]any, defaults runtimeDefaults) (Values, error) {
	next := current

	for key, value := range input {
		switch key {
		case "mail_host":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.MailHost = v
		case "site_title":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.SiteTitle = v
		case "login_whitelist":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.LoginWhitelist = v
		case "keyword_blacklist":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.KeywordBlacklist = v
		case "webhook_enabled":
			v, err := parseBoolValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.WebhookEnabled = v
		case "webhook_service":
			v, err := parseWebhookServiceValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.WebhookService = v
		case "webhook_config":
			v, err := parseStringMapValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.WebhookConfig = v
		case "webhook_message":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.WebhookMessage = v
		case "mail_retention_hours":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.MailRetentionHours = v
		case "mail_max_count":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.MailMaxCount = v
		case "max_mail_size_bytes":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.MaxMailSizeBytes = v
		case "audit_mail_received":
			v, err := parseBoolValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.AuditMailReceived = v
		case "audit_retention_days":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.AuditRetentionDays = v
		case "audit_max_count":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.AuditMaxCount = v
		case "test_smtp_host":
			v, err := parseStringValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.TestSMTPHost = v
		case "test_smtp_port":
			v, err := parseIntValue(key, value)
			if err != nil {
				return Values{}, err
			}
			next.TestSMTPPort = v
		default:
			return Values{}, fmt.Errorf("unknown setting key: %s", key)
		}
	}

	return normalizeValues(next, defaults)
}

func normalizeValues(values Values, defaults runtimeDefaults) (Values, error) {
	values.MailHost = joinCSV(splitCSV(values.MailHost, true))
	if values.MailHost == "" {
		values.MailHost = joinCSV(splitCSV(defaults.mailHost, true))
	}
	if values.MailHost == "" {
		return Values{}, fmt.Errorf("mail_host is required")
	}

	values.SiteTitle = strings.TrimSpace(values.SiteTitle)
	if values.SiteTitle == "" {
		values.SiteTitle = defaults.siteTitle
	}
	if values.SiteTitle == "" {
		values.SiteTitle = defaultSiteTitle
	}

	values.LoginWhitelist = joinCSV(splitCSV(values.LoginWhitelist, true))
	values.KeywordBlacklist = joinCSV(splitCSV(values.KeywordBlacklist, true))
	values.WebhookService = normalizeWebhookService(values.WebhookService)
	values.WebhookConfig = normalizeStringMap(values.WebhookConfig)
	values.WebhookMessage = strings.TrimSpace(values.WebhookMessage)
	if values.WebhookMessage == "" {
		values.WebhookMessage = defaultWebhookMessage
	}

	if values.MailRetentionHours < 0 {
		return Values{}, fmt.Errorf("mail_retention_hours must be >= 0")
	}
	if values.MailMaxCount < 0 {
		return Values{}, fmt.Errorf("mail_max_count must be >= 0")
	}
	if values.MaxMailSizeBytes <= 0 {
		return Values{}, fmt.Errorf("max_mail_size_bytes must be > 0")
	}
	if values.AuditRetentionDays < 0 {
		return Values{}, fmt.Errorf("audit_retention_days must be >= 0")
	}
	if values.AuditMaxCount < 0 {
		return Values{}, fmt.Errorf("audit_max_count must be >= 0")
	}

	values.TestSMTPHost = strings.TrimSpace(values.TestSMTPHost)
	if values.TestSMTPHost == "" {
		values.TestSMTPHost = defaultTestSMTPHost
	}
	if values.TestSMTPPort == 0 {
		values.TestSMTPPort = defaultTestSMTPPort
	}
	if values.TestSMTPPort < 1 || values.TestSMTPPort > 65535 {
		return Values{}, fmt.Errorf("test_smtp_port must be between 1 and 65535")
	}

	return values, nil
}

func encodeValues(values Values) map[string]string {
	webhookConfigJSON, _ := json.Marshal(normalizeStringMap(values.WebhookConfig))

	return map[string]string{
		"mail_host":            values.MailHost,
		"site_title":           values.SiteTitle,
		"login_whitelist":      values.LoginWhitelist,
		"keyword_blacklist":    values.KeywordBlacklist,
		"webhook_enabled":      boolToStored(values.WebhookEnabled),
		"webhook_service":      values.WebhookService,
		"webhook_config":       string(webhookConfigJSON),
		"webhook_message":      values.WebhookMessage,
		"mail_retention_hours": strconv.Itoa(values.MailRetentionHours),
		"mail_max_count":       strconv.Itoa(values.MailMaxCount),
		"max_mail_size_bytes":  strconv.Itoa(values.MaxMailSizeBytes),
		"audit_mail_received":  boolToStored(values.AuditMailReceived),
		"audit_retention_days": strconv.Itoa(values.AuditRetentionDays),
		"audit_max_count":      strconv.Itoa(values.AuditMaxCount),
		"test_smtp_host":       values.TestSMTPHost,
		"test_smtp_port":       strconv.Itoa(values.TestSMTPPort),
	}
}

func fallbackValues(defaults runtimeDefaults) Values {
	values, _ := normalizeValues(Values{
		MailHost:           defaults.mailHost,
		SiteTitle:          defaults.siteTitle,
		WebhookEnabled:     false,
		WebhookService:     defaultWebhookService,
		WebhookConfig:      map[string]string{},
		WebhookMessage:     defaultWebhookMessage,
		MailRetentionHours: defaultMailRetentionHours,
		MailMaxCount:       defaultMailMaxCount,
		MaxMailSizeBytes:   defaultMaxMailSizeBytes,
		AuditMailReceived:  true,
		AuditRetentionDays: defaultAuditRetentionDays,
		AuditMaxCount:      defaultAuditMaxCount,
		TestSMTPHost:       defaultTestSMTPHost,
		TestSMTPPort:       defaultTestSMTPPort,
	}, defaults)
	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func splitCSV(value string, lower bool) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	seen := make(map[string]struct{})
	var result []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if lower {
			part = strings.ToLower(part)
		}
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return result
}

func joinCSV(parts []string) string {
	return strings.Join(parts, ",")
}

func normalizeWebhookService(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", defaultWebhookService:
		return defaultWebhookService
	case "telegram":
		return "telegram"
	case "slack":
		return "slack"
	default:
		return defaultWebhookService
	}
}

func normalizeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}

	result := make(map[string]string, len(input))
	for key, value := range input {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		result[trimmedKey] = strings.TrimSpace(value)
	}
	if len(result) == 0 {
		return map[string]string{}
	}
	return result
}

func parseStoredBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func parseStoredInt(value string, defaultValue int) int {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return defaultValue
	}
	return n
}

func parseStoredStringMap(value string) map[string]string {
	if strings.TrimSpace(value) == "" {
		return map[string]string{}
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return map[string]string{}
	}

	result := make(map[string]string, len(raw))
	for key, item := range raw {
		result[key] = fmt.Sprint(item)
	}
	return normalizeStringMap(result)
}

func boolToStored(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func parseStringValue(key string, value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("%s must be a string", key)
	}
}

func parseBoolValue(key string, value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true, nil
		case "0", "false", "no", "off":
			return false, nil
		default:
			return false, fmt.Errorf("%s must be a boolean", key)
		}
	case float64:
		if v == 1 {
			return true, nil
		}
		if v == 0 {
			return false, nil
		}
	}
	return false, fmt.Errorf("%s must be a boolean", key)
}

func parseIntValue(key string, value any) (int, error) {
	switch v := value.(type) {
	case float64:
		if float64(int(v)) != v {
			return 0, fmt.Errorf("%s must be an integer", key)
		}
		return int(v), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("%s must be an integer", key)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("%s must be an integer", key)
	}
}

func parseStringMapValue(key string, value any) (map[string]string, error) {
	switch v := value.(type) {
	case nil:
		return map[string]string{}, nil
	case string:
		return parseJSONStringMapStrict(key, v)
	case map[string]any:
		result := make(map[string]string, len(v))
		for itemKey, itemValue := range v {
			result[itemKey] = fmt.Sprint(itemValue)
		}
		return normalizeStringMap(result), nil
	case map[string]string:
		return normalizeStringMap(v), nil
	default:
		return nil, fmt.Errorf("%s must be an object or JSON string", key)
	}
}

func parseWebhookServiceValue(key string, value any) (string, error) {
	v, err := parseStringValue(key, value)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "dingtalk", "telegram", "slack":
		return v, nil
	default:
		return "", fmt.Errorf("%s must be one of dingtalk, telegram, slack", key)
	}
}

func parseJSONStringMapStrict(key, value string) (map[string]string, error) {
	if strings.TrimSpace(value) == "" {
		return map[string]string{}, nil
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return nil, fmt.Errorf("%s must be valid JSON", key)
	}

	result := make(map[string]string, len(raw))
	for itemKey, itemValue := range raw {
		result[itemKey] = fmt.Sprint(itemValue)
	}
	return normalizeStringMap(result), nil
}
