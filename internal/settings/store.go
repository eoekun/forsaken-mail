package settings

import (
	"database/sql"
	"fmt"
	"os"

	"forsaken-mail/internal/config"
)

var seedKeys = map[string]struct {
	envKey  string
	defaultValue string
}{
	"mail_host":                {envKey: "MAIL_HOST", defaultValue: ""},
	"site_title":               {envKey: "SITE_TITLE", defaultValue: "Forsaken Mail"},
	"allowed_emails":           {envKey: "ALLOWED_EMAILS", defaultValue: ""},
	"keyword_blacklist":        {envKey: "KEYWORD_BLACKLIST", defaultValue: "admin,postmaster,system,webmaster,administrator,hostmaster,service,server,root"},
	"dingtalk_webhook_token":   {envKey: "DINGTALK_WEBHOOK_TOKEN", defaultValue: ""},
	"dingtalk_webhook_message": {envKey: "DINGTALK_WEBHOOK_MESSAGE", defaultValue: "new email received."},
	"mail_retention_hours":     {envKey: "MAIL_RETENTION_HOURS", defaultValue: "1"},
	"mail_max_count":           {envKey: "MAIL_MAX_COUNT", defaultValue: "100"},
	"max_mail_size_bytes":      {envKey: "MAX_MAIL_SIZE_BYTES", defaultValue: "1048576"},
	"audit_retention_days":     {envKey: "AUDIT_RETENTION_DAYS", defaultValue: "7"},
	"audit_max_count":          {envKey: "AUDIT_MAX_COUNT", defaultValue: "5000"},
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (s *Store) Get(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return value, nil
}

func (s *Store) Set(key, value string) error {
	_, err := s.db.Exec(
		"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

func (s *Store) GetAll() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting row: %w", err)
		}
		result[key] = value
	}
	return result, rows.Err()
}

func (s *Store) SetAll(kvs map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at",
	)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for key, value := range kvs {
		if _, err := stmt.Exec(key, value); err != nil {
			return fmt.Errorf("set setting %q: %w", key, err)
		}
	}

	return tx.Commit()
}

func (s *Store) SeedFromEnv(cfg *config.Config) error {
	for key, seed := range seedKeys {
		existing, err := s.Get(key)
		if err != nil {
			return err
		}
		if existing != "" {
			continue
		}

		value := os.Getenv(seed.envKey)
		if value == "" {
			value = seed.defaultValue
		}

		if key == "mail_host" && value == "" {
			value = cfg.MailHost
		}

		if err := s.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}
