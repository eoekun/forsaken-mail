package smtp

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"

	goSmtp "github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
)

const defaultMaxMailSize = 1 << 20 // 1 MiB

// Server wraps a go-smtp server and connects it to the mail router.
type Server struct {
	server   *goSmtp.Server
	router   *mail.Router
	limiter  *RateLimiter
	settings *settings.Store
	audit    *audit.Store
}

// New creates an SMTP server wired to the given dependencies.
func New(router *mail.Router, limiter *RateLimiter, settingsStore *settings.Store, auditStore *audit.Store) *Server {
	s := &Server{
		router:   router,
		limiter:  limiter,
		settings: settingsStore,
		audit:    auditStore,
	}

	s.server = goSmtp.NewServer(s)
	s.server.AllowInsecureAuth = true
	s.server.ReadTimeout = 30 * time.Second
	s.server.WriteTimeout = 30 * time.Second
	s.server.MaxMessageBytes = 1 << 20 // 1 MiB, also enforced in Data()
	// Auth is disabled by not implementing go-smtp.AuthSession.

	return s
}

// ListenAndServe starts the SMTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	s.server.Addr = addr
	slog.Info("SMTP server starting", "addr", addr)
	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("smtp listen: %w", err)
	}
	return nil
}

// Close gracefully shuts down the SMTP server.
func (s *Server) Close() error {
	return s.server.Close()
}

// ---------------------------------------------------------------------------
// go-smtp Backend interface
// ---------------------------------------------------------------------------

// NewSession creates a new SMTP session for an incoming connection.
func (s *Server) NewSession(c *goSmtp.Conn) (goSmtp.Session, error) {
	ip := extractIP(c)
	if !s.limiter.Allow(ip) {
		slog.Warn("SMTP connection rate-limited", "ip", ip)
		return nil, fmt.Errorf("rate limit exceeded")
	}
	return &session{
		router:   s.router,
		settings: s.settings,
		audit:    s.audit,
		ip:       ip,
	}, nil
}

// ---------------------------------------------------------------------------
// go-smtp Session interface
// ---------------------------------------------------------------------------

type session struct {
	router   *mail.Router
	settings *settings.Store
	audit    *audit.Store
	ip       string
	from     string
	to       []string
}

func (s *session) Mail(from string, opts *goSmtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *goSmtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	maxSize := getMaxMailSize(s.settings)

	// Read up to maxSize+1 bytes so we can detect oversize messages.
	raw, err := io.ReadAll(io.LimitReader(r, int64(maxSize)+1))
	if err != nil {
		slog.Error("failed to read mail data", "from", s.from, "ip", s.ip, "error", err)
		return nil
	}

	if len(raw) > maxSize {
		slog.Warn("mail exceeds size limit", "from", s.from, "ip", s.ip, "size", len(raw), "limit", maxSize)
		_ = s.audit.Record("MAIL_DROPPED", "", fmt.Sprintf("size %d exceeds limit %d", len(raw), maxSize), s.ip)
		return nil // don't reject at SMTP level
	}

	env, err := enmime.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		slog.Error("failed to parse mail envelope", "from", s.from, "ip", s.ip, "error", err)
		_ = s.audit.Record("MAIL_DROPPED", "", err.Error(), s.ip)
		return nil
	}

	textBody := env.Text
	htmlBody := env.HTML
	if htmlBody == "" {
		htmlBody = textBody
	}

	s.router.Handle(s.from, s.to, env.GetHeader("Subject"), textBody, htmlBody, int64(len(raw)))
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractIP returns the remote IP address from the SMTP connection.
func extractIP(c *goSmtp.Conn) string {
	addr := c.Conn().RemoteAddr().String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// getMaxMailSize reads the max_mail_size_bytes setting, falling back to the
// default value of 1 MiB.
func getMaxMailSize(s *settings.Store) int {
	v, err := s.Get("max_mail_size_bytes")
	if err != nil || v == "" {
		return defaultMaxMailSize
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultMaxMailSize
	}
	return n
}

