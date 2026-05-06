package mail

import (
	"database/sql"
	"encoding/json"
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
    extracted_codes TEXT NOT NULL DEFAULT '[]',
    extracted_links TEXT NOT NULL DEFAULT '[]',
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
	ID            int64     `json:"id"`
	ShortID       string    `json:"short_id"`
	FromAddr      string    `json:"from_addr"`
	ToAddr        string    `json:"to_addr"`
	Subject       string    `json:"subject"`
	TextBody      string    `json:"text_body"`
	HTMLBody      string    `json:"html_body"`
	RawSize       int64     `json:"raw_size"`
	IsRead        bool      `json:"is_read"`
	ExtractedCodes []string `json:"extracted_codes"`
	ExtractedLinks []string `json:"extracted_links"`
	CreatedAt     time.Time `json:"created_at"`
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

// scanMail scans a single row into a Mail struct. The scanner interface is
// satisfied by both *sql.Row and *sql.Rows.
func scanMail(scanner interface{ Scan(dest ...any) error }) (*Mail, error) {
	var m Mail
	var isRead int
	var codesJSON, linksJSON string
	if err := scanner.Scan(&m.ID, &m.ShortID, &m.FromAddr, &m.ToAddr, &m.Subject,
		&m.TextBody, &m.HTMLBody, &m.RawSize, &isRead, &codesJSON, &linksJSON, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.IsRead = isRead != 0
	json.Unmarshal([]byte(codesJSON), &m.ExtractedCodes)
	json.Unmarshal([]byte(linksJSON), &m.ExtractedLinks)
	if m.ExtractedCodes == nil {
		m.ExtractedCodes = []string{}
	}
	if m.ExtractedLinks == nil {
		m.ExtractedLinks = []string{}
	}
	return &m, nil
}

const mailSelectColumns = `id, short_id, from_addr, to_addr, subject, text_body, html_body, raw_size, is_read, extracted_codes, extracted_links, created_at`

// Save inserts a mail record into the database and sets mail.ID to the new row ID.
// It extracts verification codes and links before saving.
func (s *Store) Save(mail *Mail) error {
	codes, links := Extract(mail.Subject, mail.TextBody, mail.HTMLBody)
	mail.ExtractedCodes = codes
	mail.ExtractedLinks = links

	codesJSON, _ := json.Marshal(codes)
	linksJSON, _ := json.Marshal(links)

	now := time.Now()
	err := s.db.QueryRow(
		`INSERT INTO mails (short_id, from_addr, to_addr, subject, text_body, html_body, raw_size, extracted_codes, extracted_links, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, created_at`,
		mail.ShortID, mail.FromAddr, mail.ToAddr, mail.Subject, mail.TextBody, mail.HTMLBody, mail.RawSize, string(codesJSON), string(linksJSON), now,
	).Scan(&mail.ID, &mail.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// ListByShortID returns up to limit mails for the given short ID, ordered by created_at DESC.
func (s *Store) ListByShortID(shortID string, limit int) ([]Mail, error) {
	rows, err := s.db.Query(
		`SELECT `+mailSelectColumns+` FROM mails WHERE short_id = ? ORDER BY created_at DESC LIMIT ?`,
		shortID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		m, err := scanMail(rows)
		if err != nil {
			return nil, err
		}
		mails = append(mails, *m)
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
	return scanMail(s.db.QueryRow(
		`SELECT `+mailSelectColumns+` FROM mails WHERE id = ?`, id,
	))
}

// ListRecent returns the most recent mails across all short IDs.
func (s *Store) ListRecent(limit int) ([]Mail, error) {
	rows, err := s.db.Query(
		`SELECT `+mailSelectColumns+` FROM mails ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		m, err := scanMail(rows)
		if err != nil {
			return nil, err
		}
		mails = append(mails, *m)
	}
	return mails, rows.Err()
}

// CountByShortID returns the total number of mails for the given short ID.
func (s *Store) CountByShortID(shortID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM mails WHERE short_id = ?`, shortID).Scan(&count)
	return count, err
}

// Reextract re-runs code and link extraction on all mails for the given short ID
// and updates the database. Returns the number of mails updated.
func (s *Store) Reextract(shortID string) (int, error) {
	rows, err := s.db.Query(
		`SELECT id, subject, text_body, html_body FROM mails WHERE short_id = ?`, shortID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type mailRow struct {
		id       int64
		subject  string
		textBody string
		htmlBody string
	}
	var rows_data []mailRow
	for rows.Next() {
		var r mailRow
		if err := rows.Scan(&r.id, &r.subject, &r.textBody, &r.htmlBody); err != nil {
			return 0, err
		}
		rows_data = append(rows_data, r)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	updated := 0
	for _, r := range rows_data {
		codes, links := Extract(r.subject, r.textBody, r.htmlBody)
		codesJSON, _ := json.Marshal(codes)
		linksJSON, _ := json.Marshal(links)
		res, err := s.db.Exec(
			`UPDATE mails SET extracted_codes = ?, extracted_links = ? WHERE id = ?`,
			string(codesJSON), string(linksJSON), r.id,
		)
		if err != nil {
			continue
		}
		n, _ := res.RowsAffected()
		updated += int(n)
	}
	return updated, nil
}
