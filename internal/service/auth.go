package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

const (
	SessionCookieName    = "flagcel_session"
	OAuthStateCookieName = "flagcel_oauth_state"
	OAuthNonceCookieName = "flagcel_oauth_nonce"
	apiKeyTouchInterval  = time.Minute
)

type AuthConfig struct {
	OIDCIssuerURL     string
	OIDCClientID      string
	OIDCSecret        string
	OIDCRedirectURL   string
	AdminEmails       string
	BootstrapEmail    string
	BootstrapPassword string
	BootstrapName     string
	SessionSecret     string
	CookieSecure      bool
	SessionTTL        time.Duration
}

type AuthService struct {
	cfg         AuthConfig
	store       *postgres.Store
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	oauth       *oauth2.Config
	adminEmails map[string]struct{}
	secret      []byte
	mode        string
	apiKeyMu    sync.RWMutex
	apiKeys     map[string]cachedAPIKey
}

type cachedAPIKey struct {
	key         *core.APIKey
	lastTouched time.Time
}

type OIDCUserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Nonce         string `json:"nonce"`
}

type CreatedAPIKey struct {
	Key   *core.APIKey
	Token string
}

func NewAuthService(ctx context.Context, cfg AuthConfig, store *postgres.Store) (*AuthService, error) {
	s := &AuthService{
		cfg:         cfg,
		store:       store,
		adminEmails: parseAdminEmails(cfg.AdminEmails),
		secret:      []byte(cfg.SessionSecret),
		apiKeys:     make(map[string]cachedAPIKey),
	}
	if cfg.SessionTTL == 0 {
		s.cfg.SessionTTL = 24 * time.Hour
	}
	if s.OIDCEnabled() {
		s.mode = "oidc"
	} else {
		s.mode = "password"
	}
	if !s.Enabled() {
		return s, nil
	}
	if err := s.validateConfig(); err != nil {
		return nil, err
	}

	if s.OIDCEnabled() {
		provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuerURL)
		if err != nil {
			return nil, fmt.Errorf("oidc provider: %w", err)
		}
		s.provider = provider
		s.verifier = provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})
		s.oauth = &oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		}
		return s, nil
	}
	if err := s.bootstrapLocalAdmin(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *AuthService) Enabled() bool {
	return true
}

func (s *AuthService) OIDCEnabled() bool {
	return strings.TrimSpace(s.cfg.OIDCIssuerURL) != "" ||
		strings.TrimSpace(s.cfg.OIDCClientID) != "" ||
		strings.TrimSpace(s.cfg.OIDCSecret) != "" ||
		strings.TrimSpace(s.cfg.OIDCRedirectURL) != ""
}

func (s *AuthService) PasswordEnabled() bool {
	return s.Enabled() && !s.OIDCEnabled()
}

func (s *AuthService) Mode() string {
	return s.mode
}

func (s *AuthService) CookieSecure() bool {
	return s.cfg.CookieSecure
}

func (s *AuthService) SessionTTL() time.Duration {
	return s.cfg.SessionTTL
}

func (s *AuthService) LoginURL(state, nonce string) (string, error) {
	if !s.Enabled() {
		return "", core.ErrAuthNotConfigured
	}
	return s.oauth.AuthCodeURL(state, oidc.Nonce(nonce)), nil
}

func (s *AuthService) CompleteOIDCLogin(ctx context.Context, code, nonce string) (*core.User, error) {
	if !s.OIDCEnabled() {
		return nil, core.ErrAuthNotConfigured
	}
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oidc code exchange: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("oidc response missing id_token")
	}
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc id_token verify: %w", err)
	}
	var claims OIDCUserInfo
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc claims: %w", err)
	}
	if claims.Nonce != nonce {
		return nil, fmt.Errorf("oidc nonce mismatch")
	}
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" || !claims.EmailVerified {
		return nil, core.ErrUserNotAllowed
	}
	if _, ok := s.adminEmails[email]; !ok {
		return nil, core.ErrUserNotAllowed
	}
	return s.store.UpsertUserByOIDC(ctx, &core.User{
		ID:          uuid.NewString(),
		OIDCSubject: claims.Subject,
		Email:       email,
		Name:        strings.TrimSpace(claims.Name),
		Admin:       true,
	})
}

func (s *AuthService) LoginWithPassword(ctx context.Context, email, password string) (*core.User, error) {
	if !s.PasswordEnabled() {
		return nil, core.ErrAuthNotConfigured
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil, core.ErrInvalidCredentials
	}
	user, passwordHash, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(user.OIDCSubject, "local:") || !user.Admin {
		return nil, core.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, core.ErrInvalidCredentials
	}
	return user, nil
}

func (s *AuthService) CreateSession(ctx context.Context, userID string) (string, time.Time, error) {
	if !s.Enabled() {
		return "", time.Time{}, core.ErrAuthNotConfigured
	}
	token, err := randomToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.cfg.SessionTTL)
	_ = s.store.DeleteExpiredSessions(ctx)
	if err := s.store.CreateSession(ctx, uuid.NewString(), userID, s.hash(token), expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (s *AuthService) UserForSessionToken(ctx context.Context, token string) (*core.User, error) {
	if !s.Enabled() {
		return nil, core.ErrAuthNotConfigured
	}
	if strings.TrimSpace(token) == "" {
		return nil, core.ErrSessionNotFound
	}
	return s.store.GetUserBySessionHash(ctx, s.hash(token))
}

func (s *AuthService) RevokeSession(ctx context.Context, token string) error {
	if !s.Enabled() {
		return core.ErrAuthNotConfigured
	}
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return s.store.DeleteSessionByHash(ctx, s.hash(token))
}

func (s *AuthService) CreateAPIKey(ctx context.Context, name string) (*CreatedAPIKey, error) {
	if !s.Enabled() {
		return nil, core.ErrAuthNotConfigured
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	prefixPart, err := randomToken(6)
	if err != nil {
		return nil, err
	}
	secretPart, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	prefix := "fc_" + prefixPart
	token := prefix + "_" + secretPart
	key, err := s.store.CreateAPIKey(ctx, uuid.NewString(), name, prefix, s.hash(token))
	if err != nil {
		return nil, err
	}
	s.cacheAPIKey(token, key)
	return &CreatedAPIKey{Key: key, Token: token}, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context) ([]*core.APIKey, error) {
	if !s.Enabled() {
		return []*core.APIKey{}, nil
	}
	return s.store.ListAPIKeys(ctx)
}

func (s *AuthService) RevokeAPIKey(ctx context.Context, id string) error {
	if !s.Enabled() {
		return core.ErrAuthNotConfigured
	}
	if err := s.store.RevokeAPIKey(ctx, id); err != nil {
		return err
	}
	s.apiKeyMu.Lock()
	clear(s.apiKeys)
	s.apiKeyMu.Unlock()
	return nil
}

func (s *AuthService) ValidateAPIKey(ctx context.Context, token string) (*core.APIKey, error) {
	if !s.Enabled() {
		return nil, core.ErrAuthNotConfigured
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, core.ErrAPIKeyNotFound
	}
	hash := s.hash(token)

	s.apiKeyMu.RLock()
	cached, ok := s.apiKeys[hash]
	s.apiKeyMu.RUnlock()
	if ok {
		s.touchCachedAPIKey(ctx, hash, cached)
		return cached.key, nil
	}

	key, err := s.store.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	_ = s.store.TouchAPIKey(ctx, key.ID)
	s.cacheAPIKeyHash(hash, key)

	return key, nil
}

func (s *AuthService) cacheAPIKey(token string, key *core.APIKey) {
	s.cacheAPIKeyHash(s.hash(token), key)
}

func (s *AuthService) cacheAPIKeyHash(hash string, key *core.APIKey) {
	s.apiKeyMu.Lock()
	s.apiKeys[hash] = cachedAPIKey{key: key, lastTouched: time.Now()}
	s.apiKeyMu.Unlock()
}

func (s *AuthService) touchCachedAPIKey(ctx context.Context, hash string, cached cachedAPIKey) {
	if time.Since(cached.lastTouched) < apiKeyTouchInterval {
		return
	}
	if err := s.store.TouchAPIKey(ctx, cached.key.ID); err != nil {
		return
	}
	cached.lastTouched = time.Now()
	s.apiKeyMu.Lock()
	s.apiKeys[hash] = cached
	s.apiKeyMu.Unlock()
}

func (s *AuthService) validateConfig() error {
	missing := []string{}
	if s.OIDCEnabled() {
		if strings.TrimSpace(s.cfg.OIDCIssuerURL) == "" {
			missing = append(missing, "AUTH_OIDC_ISSUER_URL")
		}
		if strings.TrimSpace(s.cfg.OIDCClientID) == "" {
			missing = append(missing, "AUTH_OIDC_CLIENT_ID")
		}
		if strings.TrimSpace(s.cfg.OIDCSecret) == "" {
			missing = append(missing, "AUTH_OIDC_CLIENT_SECRET")
		}
		if strings.TrimSpace(s.cfg.OIDCRedirectURL) == "" {
			missing = append(missing, "AUTH_OIDC_REDIRECT_URL")
		}
		if len(s.adminEmails) == 0 {
			missing = append(missing, "AUTH_ADMIN_EMAILS")
		}
	} else {
		if strings.TrimSpace(s.cfg.BootstrapEmail) == "" {
			missing = append(missing, "AUTH_BOOTSTRAP_ADMIN_EMAIL")
		}
		if strings.TrimSpace(s.cfg.BootstrapPassword) == "" {
			missing = append(missing, "AUTH_BOOTSTRAP_ADMIN_PASSWORD")
		}
	}
	if len(s.secret) < 32 {
		missing = append(missing, "AUTH_SESSION_SECRET (at least 32 bytes)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("auth config missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (s *AuthService) bootstrapLocalAdmin(ctx context.Context) error {
	email := strings.ToLower(strings.TrimSpace(s.cfg.BootstrapEmail))
	password := strings.TrimSpace(s.cfg.BootstrapPassword)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(s.cfg.BootstrapName)
	if name == "" {
		name = "Admin"
	}
	_, err = s.store.UpsertLocalAdmin(ctx, &core.User{
		ID:          uuid.NewString(),
		OIDCSubject: "local:" + email,
		Email:       email,
		Name:        name,
		Admin:       true,
	}, string(hash))
	return err
}

func (s *AuthService) hash(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func parseAdminEmails(raw string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, item := range strings.Split(raw, ",") {
		email := strings.ToLower(strings.TrimSpace(item))
		if email != "" {
			out[email] = struct{}{}
		}
	}
	return out
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func IsAuthError(err error) bool {
	return errors.Is(err, core.ErrAuthNotConfigured) ||
		errors.Is(err, core.ErrInvalidCredentials) ||
		errors.Is(err, core.ErrUserNotAllowed) ||
		errors.Is(err, core.ErrSessionNotFound) ||
		errors.Is(err, core.ErrAPIKeyNotFound)
}
