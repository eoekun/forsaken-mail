package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"time"

	"forsaken-mail/internal/i18n"
)

type testEmailRequest struct {
	SenderEmail string `json:"sender_email"`
	AuthCode    string `json:"auth_code"`
	ShortID     string `json:"short_id"`
}

// writeSMTPError writes a standardized SMTP error response.
func writeSMTPError(w http.ResponseWriter, lang, i18nKey string, err error) {
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"ok":      false,
		"message": i18n.Tfmt(lang, i18nKey, err),
	})
}

func (rt *Router) handleTestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	lang := i18n.LangFromRequest(r)

	var req testEmailRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "invalid_request_body"))
		return
	}

	if req.SenderEmail == "" || req.AuthCode == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "sender_auth_required"))
		return
	}

	mailHost, err := rt.settings.Get("mail_host")
	if err != nil || mailHost == "" {
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "mail_host_not_configured"))
		return
	}

	// Get configurable SMTP host (default: smtp.qq.com)
	smtpHost, _ := rt.settings.Get("test_smtp_host")
	if smtpHost == "" {
		smtpHost = "smtp.qq.com"
	}
	smtpPort, _ := rt.settings.Get("test_smtp_port")
	if smtpPort == "" {
		smtpPort = "465"
	}

	shortID := req.ShortID
	if shortID == "" {
		shortID = "test"
	}
	recipient := shortID + "@" + mailHost

	subject := fmt.Sprintf("SMTP Test - %s", time.Now().Format("2006-01-02 15:04:05"))
	body := fmt.Sprintf("This is a test email sent via SMTP to forsaken-mail.\n\nSender: %s\nRecipient: %s\nTime: %s\n", req.SenderEmail, recipient, time.Now().Format(time.RFC3339))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		req.SenderEmail, recipient, subject, body)

	addr := net.JoinHostPort(smtpHost, smtpPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		writeSMTPError(w, lang, "smtp_connect_failed", err)
		return
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: smtpHost})
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		writeSMTPError(w, lang, "tls_handshake_failed", err)
		return
	}

	client, err := smtp.NewClient(tlsConn, smtpHost)
	if err != nil {
		tlsConn.Close()
		writeSMTPError(w, lang, "smtp_client_error", err)
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", req.SenderEmail, req.AuthCode, smtpHost)
	if err := client.Auth(auth); err != nil {
		writeSMTPError(w, lang, "smtp_auth_failed", err)
		return
	}

	if err := client.Mail(req.SenderEmail); err != nil {
		writeSMTPError(w, lang, "mail_from_failed", err)
		return
	}

	if err := client.Rcpt(recipient); err != nil {
		writeSMTPError(w, lang, "rcpt_to_failed", err)
		return
	}

	wc, err := client.Data()
	if err != nil {
		writeSMTPError(w, lang, "data_command_failed", err)
		return
	}

	if _, err := wc.Write([]byte(msg)); err != nil {
		writeSMTPError(w, lang, "write_message_failed", err)
		return
	}

	if err := wc.Close(); err != nil {
		writeSMTPError(w, lang, "close_message_failed", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   i18n.Tfmt(lang, "test_email_sent", req.SenderEmail, recipient),
		"sender":    req.SenderEmail,
		"recipient": recipient,
	})
}
