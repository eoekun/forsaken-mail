package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port              int
	MailinHost        string
	MailinPort        int
	MailHost          string
	SiteTitle         string
	DBPath            string
	OAuthProvider     string
	OAuthClientID     string
	OAuthClientSecret string
	SessionSecret     string
	LogLevel          string
	LogFile           string
	LogMaxSizeMB      int
	LogMaxBackups     int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          3000,
		MailinHost:    "0.0.0.0",
		MailinPort:    25,
		SiteTitle:     "Forsaken Mail",
		DBPath:        "./data/forsaken-mail.db",
		OAuthProvider: "github",
		LogLevel:      "info",
		LogMaxSizeMB:  10,
		LogMaxBackups: 3,
	}

	if v := os.Getenv("PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	if v := os.Getenv("MAILIN_HOST"); v != "" {
		cfg.MailinHost = v
	}

	if v := os.Getenv("MAILIN_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAILIN_PORT: %w", err)
		}
		cfg.MailinPort = port
	}

	cfg.MailHost = os.Getenv("MAIL_HOST")

	if v := os.Getenv("SITE_TITLE"); v != "" {
		cfg.SiteTitle = v
	}

	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}

	if v := os.Getenv("OAUTH_PROVIDER"); v != "" {
		cfg.OAuthProvider = v
	}

	cfg.OAuthClientID = os.Getenv("OAUTH_CLIENT_ID")
	cfg.OAuthClientSecret = os.Getenv("OAUTH_CLIENT_SECRET")
	cfg.SessionSecret = os.Getenv("SESSION_SECRET")

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	cfg.LogFile = os.Getenv("LOG_FILE")

	if v := os.Getenv("LOG_MAX_SIZE_MB"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid LOG_MAX_SIZE_MB: %w", err)
		}
		cfg.LogMaxSizeMB = n
	}

	if v := os.Getenv("LOG_MAX_BACKUPS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid LOG_MAX_BACKUPS: %w", err)
		}
		cfg.LogMaxBackups = n
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be 1-65535, got %d", c.Port)
	}
	if c.MailinPort < 1 || c.MailinPort > 65535 {
		return fmt.Errorf("MAILIN_PORT must be 1-65535, got %d", c.MailinPort)
	}
	if c.MailHost == "" {
		return fmt.Errorf("MAIL_HOST is required")
	}
	if c.OAuthClientID == "" {
		return fmt.Errorf("OAUTH_CLIENT_ID is required")
	}
	if c.OAuthClientSecret == "" {
		return fmt.Errorf("OAUTH_CLIENT_SECRET is required")
	}
	if c.SessionSecret == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	return nil
}
