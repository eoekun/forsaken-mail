package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"time"
)

type testEmailRequest struct {
	SenderEmail string `json:"sender_email"`
	AuthCode    string `json:"auth_code"`
	ShortID     string `json:"short_id"`
}

func (rt *Router) handleTestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req testEmailRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SenderEmail == "" || req.AuthCode == "" {
		writeError(w, http.StatusBadRequest, "sender_email and auth_code are required")
		return
	}

	mailHost, err := rt.settings.Get("mail_host")
	if err != nil || mailHost == "" {
		writeError(w, http.StatusInternalServerError, "mail_host not configured")
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
			"message": "Failed to connect to QQ SMTP: " + err.Error(),
		})
		return
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "smtp.qq.com"})
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "TLS handshake failed: " + err.Error(),
		})
		return
	}

	client, err := smtp.NewClient(tlsConn, "smtp.qq.com")
	if err != nil {
		tlsConn.Close()
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "SMTP client error: " + err.Error(),
		})
		return
	}
	defer client.Close()

	auth := smtp.PlainAuth("", req.SenderEmail, req.AuthCode, "smtp.qq.com")
	if err := client.Auth(auth); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "SMTP auth failed (check email and auth code): " + err.Error(),
		})
		return
	}

	if err := client.Mail(req.SenderEmail); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "MAIL FROM failed: " + err.Error(),
		})
		return
	}

	if err := client.Rcpt(recipient); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "RCPT TO failed: " + err.Error(),
		})
		return
	}

	wc, err := client.Data()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "DATA command failed: " + err.Error(),
		})
		return
	}

	if _, err := wc.Write([]byte(msg)); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "Failed to write message: " + err.Error(),
		})
		return
	}

	if err := wc.Close(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"message": "Failed to close message: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"message":   fmt.Sprintf("Test email sent from %s to %s via QQ SMTP", req.SenderEmail, recipient),
		"sender":    req.SenderEmail,
		"recipient": recipient,
	})
}
