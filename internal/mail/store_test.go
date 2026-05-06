package mail

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(db)
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return s
}

func insertTestMail(t *testing.T, s *Store, shortID, from, subject string) {
	t.Helper()
	m := &Mail{
		ShortID:  shortID,
		FromAddr: from,
		ToAddr:   shortID + "@test.com",
		Subject:  subject,
		TextBody: "body",
	}
	if err := s.Save(m); err != nil {
		t.Fatal(err)
	}
}

func TestBuildMailFilters_NoFilters(t *testing.T) {
	where, args := buildMailFilters("", "", "")
	if where != "" {
		t.Errorf("expected empty where, got %q", where)
	}
	if len(args) != 0 {
		t.Errorf("expected no args, got %d", len(args))
	}
}

func TestBuildMailFilters_ShortIDOnly(t *testing.T) {
	where, args := buildMailFilters("test", "", "")
	if where != " WHERE short_id = ?" {
		t.Errorf("unexpected where: %q", where)
	}
	if len(args) != 1 || args[0] != "test" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestBuildMailFilters_AllFilters(t *testing.T) {
	where, args := buildMailFilters("abc", "sender", "hello")
	if where != " WHERE short_id = ? AND from_addr LIKE ? AND subject LIKE ?" {
		t.Errorf("unexpected where: %q", where)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestListAll_Pagination(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		insertTestMail(t, s, "user", "from@test.com", "subject")
	}

	mails, total, err := s.ListAll(1, 3, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(mails) != 3 {
		t.Errorf("expected 3 mails, got %d", len(mails))
	}

	mails2, _, err := s.ListAll(2, 3, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(mails2) != 2 {
		t.Errorf("expected 2 mails on page 2, got %d", len(mails2))
	}
}

func TestListAll_FilterByShortID(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "alice", "x@test.com", "hello")
	insertTestMail(t, s, "bob", "y@test.com", "hello")
	insertTestMail(t, s, "alice", "z@test.com", "world")

	mails, total, err := s.ListAll(1, 20, "alice", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	for _, m := range mails {
		if m.ShortID != "alice" {
			t.Errorf("expected short_id alice, got %s", m.ShortID)
		}
	}
}

func TestListAll_FilterByFrom(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "u1", "openai@tm.openai.com", "a")
	insertTestMail(t, s, "u2", "google@gmail.com", "b")
	insertTestMail(t, s, "u3", "support@openai.com", "c")

	mails, total, err := s.ListAll(1, 20, "", "openai", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	for _, m := range mails {
		if m.FromAddr != "openai@tm.openai.com" && m.FromAddr != "support@openai.com" {
			t.Errorf("unexpected from: %s", m.FromAddr)
		}
	}
}

func TestListAll_FilterByQuery(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "u1", "a@test.com", "verification code")
	insertTestMail(t, s, "u2", "b@test.com", "welcome")
	insertTestMail(t, s, "u3", "c@test.com", "your verification code 123")

	mails, total, err := s.ListAll(1, 20, "", "", "verification")
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

func TestListAll_CombinedFilters(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "alice", "openai@test.com", "code")
	insertTestMail(t, s, "alice", "google@test.com", "code")
	insertTestMail(t, s, "bob", "openai@test.com", "code")

	mails, total, err := s.ListAll(1, 20, "alice", "openai", "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(mails) == 1 && (mails[0].ShortID != "alice" || mails[0].FromAddr != "openai@test.com") {
		t.Errorf("unexpected mail: %+v", mails[0])
	}
}

func TestListDistinctSenders(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "u1", "alice@test.com", "a")
	insertTestMail(t, s, "u2", "bob@test.com", "b")
	insertTestMail(t, s, "u3", "alice@test.com", "c")

	senders, err := s.ListDistinctSenders(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(senders) != 2 {
		t.Errorf("expected 2 distinct senders, got %d: %v", len(senders), senders)
	}
}

func TestListDistinctRecipients(t *testing.T) {
	s := newTestStore(t)
	insertTestMail(t, s, "alice", "x@test.com", "a")
	insertTestMail(t, s, "bob", "y@test.com", "b")
	insertTestMail(t, s, "alice", "z@test.com", "c")

	recipients, err := s.ListDistinctRecipients(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(recipients) != 2 {
		t.Errorf("expected 2 distinct recipients, got %d: %v", len(recipients), recipients)
	}
}
