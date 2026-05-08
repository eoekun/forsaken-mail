package service

import (
	"encoding/json"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
)

type hubAdminView interface {
	ClientCount() int
	UpdateBlacklist([]string)
}

// Status is the typed admin status read model.
type Status struct {
	Uptime    string `json:"uptime"`
	MailCount int    `json:"mail_count"`
	WSClients int    `json:"ws_clients"`
	DBPath    string `json:"db_path"`
	MailHost  string `json:"mail_host"`
}

// AdminService owns admin-facing reads and writes.
type AdminService struct {
	settings  *settings.Service
	audit     *audit.Store
	mails     *mail.Store
	hub       hubAdminView
	startTime time.Time
	dbPath    string
}

// NewAdminService creates a new admin service.
func NewAdminService(settingsService *settings.Service, auditStore *audit.Store, mailStore *mail.Store, hub hubAdminView, startTime time.Time, dbPath string) *AdminService {
	return &AdminService{
		settings:  settingsService,
		audit:     auditStore,
		mails:     mailStore,
		hub:       hub,
		startTime: startTime,
		dbPath:    dbPath,
	}
}

// QueryAuditLogs returns paginated audit logs.
func (s *AdminService) QueryAuditLogs(event string, offset, limit int) ([]audit.Log, int, error) {
	return s.audit.Query(event, offset, limit)
}

// GetSettings returns the typed runtime settings snapshot.
func (s *AdminService) GetSettings() (settings.Values, error) {
	return s.settings.Load()
}

// UpdateSettings validates, persists, and audits a settings update.
func (s *AdminService) UpdateSettings(input map[string]any, actor, ip string) (settings.Values, error) {
	values, changed, err := s.settings.Update(input)
	if err != nil {
		return settings.Values{}, err
	}

	if len(changed) == 0 {
		return values, nil
	}

	if s.hub != nil {
		s.hub.UpdateBlacklist(values.KeywordBlacklistList())
	}

	detailBytes, _ := json.Marshal(map[string]any{
		"changed_keys": changed,
	})
	if err := s.audit.Record("CONFIG_CHANGED", actor, string(detailBytes), ip); err != nil {
		return settings.Values{}, err
	}

	return values, nil
}

// Status returns the current system status snapshot.
func (s *AdminService) Status() (Status, error) {
	mailCount, err := s.mails.Count()
	if err != nil {
		mailCount = -1
	}

	values, err := s.settings.Load()
	if err != nil {
		return Status{}, err
	}

	wsClients := 0
	if s.hub != nil {
		wsClients = s.hub.ClientCount()
	}

	return Status{
		Uptime:    s.startTime.Format("2006-01-02T15:04:05Z"),
		MailCount: mailCount,
		WSClients: wsClients,
		DBPath:    s.dbPath,
		MailHost:  values.MailHost,
	}, nil
}
