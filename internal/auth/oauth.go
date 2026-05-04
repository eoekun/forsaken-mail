package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"forsaken-mail/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

const (
	EventLogin       = "LOGIN"
	EventLogout      = "LOGOUT"
	EventLoginFailed = "LOGIN_FAILED"
)

type Provider interface {
	AuthCodeURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (*oauth2.Token, error)
	GetEmail(ctx context.Context, token *oauth2.Token) (string, error)
	Name() string
}

func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.OAuthProvider {
	case "github":
		return NewGitHubProvider(cfg.OAuthClientID, cfg.OAuthClientSecret), nil
	case "google":
		return NewGoogleProvider(cfg.OAuthClientID, cfg.OAuthClientSecret), nil
	default:
		return nil, fmt.Errorf("unsupported oauth provider: %s", cfg.OAuthProvider)
	}
}

func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type GitHubProvider struct {
	clientID     string
	clientSecret string
}

type GoogleProvider struct {
	clientID     string
	clientSecret string
}

func NewGitHubProvider(clientID, clientSecret string) *GitHubProvider {
	return &GitHubProvider{clientID: clientID, clientSecret: clientSecret}
}

func NewGoogleProvider(clientID, clientSecret string) *GoogleProvider {
	return &GoogleProvider{clientID: clientID, clientSecret: clientSecret}
}

func (p *GitHubProvider) AuthCodeURL(state, redirectURI string) string {
	conf := p.config(redirectURI)
	return conf.AuthCodeURL(state)
}

func (p *GitHubProvider) Exchange(ctx context.Context, code, redirectURI string) (*oauth2.Token, error) {
	conf := p.config(redirectURI)
	return conf.Exchange(ctx, code)
}

func (p *GitHubProvider) Name() string { return "github" }

func (p *GitHubProvider) GetEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no primary verified email found")
}

func (p *GitHubProvider) config(redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectURI,
	}
}

func (p *GoogleProvider) AuthCodeURL(state, redirectURI string) string {
	conf := p.config(redirectURI)
	return conf.AuthCodeURL(state)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code, redirectURI string) (*oauth2.Token, error) {
	conf := p.config(redirectURI)
	return conf.Exchange(ctx, code)
}

func (p *GoogleProvider) Name() string { return "google" }

func (p *GoogleProvider) GetEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google api returned status %d", resp.StatusCode)
	}

	var user struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if user.Email == "" {
		return "", fmt.Errorf("no email found in google response")
	}

	return user.Email, nil
}

func (p *GoogleProvider) config(redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURI,
	}
}
