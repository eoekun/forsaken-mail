package service

import (
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
)

// PublicConfig is the typed payload returned to the SPA bootstrap request.
type PublicConfig struct {
	Host             string   `json:"host"`
	Hosts            []string `json:"hosts"`
	SiteTitle        string   `json:"site_title"`
	AuthMode         string   `json:"auth_mode"`
	KeywordBlacklist []string `json:"keyword_blacklist"`
	Email            string   `json:"email,omitempty"`
}

// PublicConfigService builds the read model used by GET /api/config.
type PublicConfigService struct {
	settings *settings.Service
	authMode string
}

// NewPublicConfigService creates a new public config service.
func NewPublicConfigService(settingsService *settings.Service, authMode string) *PublicConfigService {
	return &PublicConfigService{
		settings: settingsService,
		authMode: authMode,
	}
}

// Load returns the public site config payload for the current user.
func (s *PublicConfigService) Load(email string) (PublicConfig, error) {
	values, err := s.settings.Load()
	if err != nil {
		return PublicConfig{}, err
	}

	hosts := mail.ParseDomains(values.MailHost)
	if len(hosts) == 0 {
		hosts = []string{values.MailHost}
	}

	return PublicConfig{
		Host:             hosts[0],
		Hosts:            hosts,
		SiteTitle:        values.SiteTitle,
		AuthMode:         s.authMode,
		KeywordBlacklist: values.KeywordBlacklistList(),
		Email:            email,
	}, nil
}
