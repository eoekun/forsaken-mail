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

	shortID := req.ShortID
	if shortID == "" {
		shortID = "test"
	}
	recipient := shortID + "@" + mailHost

	subject := fmt.Sprintf("SMTP Test - %s", time.Now().Format("2006-01-02 15:04:05"))
	body := fmt.Sprintf("This is a test email sent via QQ SMTP to forsaken-mail.\n\nSender: %s\nRecipient: %s\nTime: %s\n", req.SenderEmail, recipient, time.Now().Format(time.RFC3339))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		req.SenderEmail, recipient, subject, body)

	// Connect to QQ SMTP via SSL (port 465)
	addr := "smtp.qq.com:465"
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "smtp_connect_failed", err),
		})
		return
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "smtp.qq.com"})
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "tls_handshake_failed", err),
		})
		return
	}

	client, err := smtp.NewClient(tlsConn, "smtp.qq.com")
	if err != nil {
		tlsConn.Close()
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "smtp_client_error", err),
		})
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", req.SenderEmail, req.AuthCode, "smtp.qq.com")
	if err := client.Auth(auth); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "smtp_auth_failed", err),
		})
		return
	}

	if err := client.Mail(req.SenderEmail); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "mail_from_failed", err),
		})
		return
	}

	if err := client.Rcpt(recipient); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "rcpt_to_failed", err),
		})
		return
	}

	wc, err := client.Data()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "data_command_failed", err),
		})
		return
	}

	if _, err := wc.Write([]byte(msg)); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "write_message_failed", err),
		})
		return
	}

	if err := wc.Close(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": i18n.Tfmt(lang, "close_message_failed", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   i18n.Tfmt(lang, "test_email_sent", req.SenderEmail, recipient),
		"sender":    req.SenderEmail,
		"recipient": recipient,
	})
}
