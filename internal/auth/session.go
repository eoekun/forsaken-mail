package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SessionData struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionManager struct {
	key          []byte
	cookieSecure bool
}

func NewSessionManager(secret string, cookieSecure bool) *SessionManager {
	hash := sha256.Sum256([]byte(secret))
	return &SessionManager{key: hash[:], cookieSecure: cookieSecure}
}

func (sm *SessionManager) Encrypt(data *SessionData) (string, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}

	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (sm *SessionManager) Decrypt(encoded string) (*SessionData, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(sm.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &data, nil
}

func (sm *SessionManager) SetCookie(w http.ResponseWriter, data *SessionData) {
	encoded, err := sm.Encrypt(data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   sm.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}

func (sm *SessionManager) GetCookie(r *http.Request) (*SessionData, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, fmt.Errorf("read cookie: %w", err)
	}

	return sm.Decrypt(cookie.Value)
}

func (sm *SessionManager) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   sm.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
