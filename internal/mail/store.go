package mail

import (
	"database/sql"
	"fmt"
	"time"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS mails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_id TEXT NOT NULL,
    from_addr TEXT NOT NULL,
    to_addr TEXT NOT NULL,
    subject TEXT NOT NULL DEFAULT '',
    text_body TEXT NOT NULL DEFAULT '',
    html_body TEXT NOT NULL DEFAULT '',
    raw_size INTEGER NOT NULL DEFAULT 0,
    is_read INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const createIndexShortIDSQL = `
CREATE INDEX IF NOT EXISTS idx_mails_short_id ON mails(short_id);
`

const createIndexCreatedSQL = `
CREATE INDEX IF NOT EXISTS idx_mails_created ON mails(created_at);
`

// Mail represents a single email message stored in the database.
type Mail struct {
	ID        int64     `json:"id"`
	ShortID   string    `json:"short_id"`
	FromAddr  string    `json:"from_addr"`
	ToAddr    string    `json:"to_addr"`
	Subject   string    `json:"subject"`
	TextBody  string    `json:"text_body"`
	HTMLBody  string    `json:"html_body"`
	RawSize   int64     `json:"raw_size"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// Store provides SQLite-backed mail storage.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init creates the mails table and indexes if they do not already exist.
func (s *Store) Init() error {
	if _, err := s.db.Exec(createTableSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(createIndexShortIDSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(createIndexCreatedSQL); err != nil {
		return err
	}
	return nil
}

// Save inserts a mail record into the database and sets mail.ID to the new row ID.
func (s *Store) Save(mail *Mail) error {
	result, err := s.db.Exec(
		`INSERT INTO mails (short_id, from_addr, to_addr, subject, text_body, html_body, raw_size)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		mail.ShortID, mail.FromAddr, mail.ToAddr, mail.Subject, mail.TextBody, mail.HTMLBody, mail.RawSize,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	mail.ID = id
	return nil
}

// ListByShortID returns up to limit mails for the given short ID, ordered by created_at DESC.
func (s *Store) ListByShortID(shortID string, limit int) ([]Mail, error) {
	rows, err := s.db.Query(
		`SELECT id, short_id, from_addr, to_addr, subject, text_body, html_body, raw_size, is_read, created_at
		 FROM mails
		 WHERE short_id = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		shortID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		var m Mail
		var isRead int
		if err := rows.Scan(&m.ID, &m.ShortID, &m.FromAddr, &m.ToAddr, &m.Subject, &m.TextBody, &m.HTMLBody, &m.RawSize, &isRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.IsRead = isRead != 0
		mails = append(mails, m)
	}
	return mails, rows.Err()
}

// Count returns the total number of mail records.
func (s *Store) Count() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM mails`).Scan(&count)
	return count, err
}

// CleanupByAge deletes mails older than the given number of hours.
func (s *Store) CleanupByAge(hours int) error {
	dur := fmt.Sprintf("-%d hours", hours)
	_, err := s.db.Exec(`DELETE FROM mails WHERE created_at < datetime('now', ?)`, dur)
	return err
}

// CleanupByCount keeps at most maxCount mails, deleting the oldest ones.
func (s *Store) CleanupByCount(maxCount int) error {
	_, err := s.db.Exec(
		`DELETE FROM mails WHERE id NOT IN (
			SELECT id FROM mails ORDER BY created_at DESC LIMIT ?
		)`,
		maxCount,
	)
	return err
}

// MarkAsRead sets is_read=1 for the given mail ID.
func (s *Store) MarkAsRead(id int64) error {
	result, err := s.db.Exec(`UPDATE mails SET is_read = 1 WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteByShortID deletes all mails for the given short ID.
func (s *Store) DeleteByShortID(shortID string) error {
	_, err := s.db.Exec(`DELETE FROM mails WHERE short_id = ?`, shortID)
	return err
}

// DeleteByID deletes a single mail by its ID.
func (s *Store) DeleteByID(id int64) error {
	result, err := s.db.Exec(`DELETE FROM mails WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetByID returns a single mail by its ID.
func (s *Store) GetByID(id int64) (*Mail, error) {
	var m Mail
	var isRead int
	err := s.db.QueryRow(
		`SELECT id, short_id, from_addr, to_addr, subject, text_body, html_body, raw_size, is_read, created_at
		 FROM mails WHERE id = ?`, id,
	).Scan(&m.ID, &m.ShortID, &m.FromAddr, &m.ToAddr, &m.Subject, &m.TextBody, &m.HTMLBody, &m.RawSize, &isRead, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	m.IsRead = isRead != 0
	return &m, nil
}

// CountByShortID returns the total number of mails for the given short ID.
func (s *Store) CountByShortID(shortID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM mails WHERE short_id = ?`, shortID).Scan(&count)
	return count, err
}
