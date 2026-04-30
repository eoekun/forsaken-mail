package audit

import (
	"database/sql"
	"fmt"
	"time"
)

type Log struct {
	ID        int64
	Event     string
	Email     string
	Detail    string
	IP        string
	CreatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Init() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event TEXT NOT NULL,
			email TEXT NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '{}',
			ip TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create audit_logs table: %w", err)
	}

	if _, err := s.db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_event ON audit_logs(event)"); err != nil {
		return fmt.Errorf("create idx_audit_event: %w", err)
	}

	if _, err := s.db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at)"); err != nil {
		return fmt.Errorf("create idx_audit_created: %w", err)
	}

	return nil
}

func (s *Store) Record(event, email, detail, ip string) error {
	_, err := s.db.Exec(
		"INSERT INTO audit_logs (event, email, detail, ip, created_at) VALUES (?, ?, ?, ?, ?)",
		event, email, detail, ip, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("record audit log: %w", err)
	}
	return nil
}

func (s *Store) Query(event string, offset, limit int) ([]Log, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs"
	countArgs := []any{}

	if event != "" {
		countQuery += " WHERE event = ?"
		countArgs = append(countArgs, event)
	}

	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	query := "SELECT id, event, email, detail, ip, created_at FROM audit_logs"
	args := []any{}

	if event != "" {
		query += " WHERE event = ?"
		args = append(args, event)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.Event, &l.Email, &l.Detail, &l.IP, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit logs: %w", err)
	}

	return logs, total, nil
}

func (s *Store) CleanupByAge(days int) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	_, err := s.db.Exec("DELETE FROM audit_logs WHERE created_at < ?", cutoff)
	if err != nil {
		return fmt.Errorf("cleanup audit logs by age: %w", err)
	}
	return nil
}

func (s *Store) CleanupByCount(maxCount int) error {
	_, err := s.db.Exec(`
		DELETE FROM audit_logs WHERE id IN (
			SELECT id FROM audit_logs ORDER BY created_at DESC LIMIT -1 OFFSET ?
		)
	`, maxCount)
	if err != nil {
		return fmt.Errorf("cleanup audit logs by count: %w", err)
	}
	return nil
}
