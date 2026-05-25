package v1

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type userContextKey struct{}

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) RegisterPublic(root *http.ServeMux) {
	root.HandleFunc("GET /auth/login", h.Login)
	root.HandleFunc("GET /auth/callback", h.Callback)
}

func (h *AuthHandler) RegisterAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/me", h.Me)
	mux.HandleFunc("POST /auth/login", h.PasswordLogin)
	mux.HandleFunc("POST /auth/logout", h.Logout)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.service.OIDCEnabled() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	state, err := randomCookieToken()
	if err != nil {
		WriteError(w, err)
		return
	}
	nonce, err := randomCookieToken()
	if err != nil {
		WriteError(w, err)
		return
	}
	loginURL, err := h.service.LoginURL(state, nonce)
	if err != nil {
		WriteError(w, err)
		return
	}
	setTransientCookie(w, service.OAuthStateCookieName, state, h.service.CookieSecure())
	setTransientCookie(w, service.OAuthNonceCookieName, nonce, h.service.CookieSecure())
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *AuthHandler) PasswordLogin(w http.ResponseWriter, r *http.Request) {
	req, err := utils.Decode[PasswordLoginRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	user, err := h.service.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := h.setSession(w, r, user.ID); err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", AuthMeResponse{
		AuthEnabled:   true,
		Mode:          h.service.Mode(),
		Authenticated: true,
		User:          toUserResponse(user),
	}); err != nil {
		WriteError(w, err)
	}
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if !h.service.OIDCEnabled() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	stateCookie, err := r.Cookie(service.OAuthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		WriteError(w, BadRequest("invalid oauth state"))
		return
	}
	nonceCookie, err := r.Cookie(service.OAuthNonceCookieName)
	if err != nil || nonceCookie.Value == "" {
		WriteError(w, BadRequest("missing oauth nonce"))
		return
	}
	clearCookie(w, service.OAuthStateCookieName, "/auth")
	clearCookie(w, service.OAuthNonceCookieName, "/auth")

	user, err := h.service.CompleteOIDCLogin(r.Context(), r.URL.Query().Get("code"), nonceCookie.Value)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := h.setSession(w, r, user.ID); err != nil {
		WriteError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(service.SessionCookieName); err == nil {
		_ = h.service.RevokeSession(r.Context(), cookie.Value)
	}
	clearCookie(w, service.SessionCookieName, "/")
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if !h.service.Enabled() {
		_ = utils.Encode(w, r, http.StatusOK, "success", AuthMeResponse{
			AuthEnabled:   false,
			Authenticated: false,
		})
		return
	}
	user, err := h.userFromRequest(r)
	if err != nil {
		if errors.Is(err, core.ErrSessionNotFound) {
			_ = utils.Encode(w, r, http.StatusOK, "success", AuthMeResponse{
				AuthEnabled:   true,
				Mode:          h.service.Mode(),
				Authenticated: false,
			})
			return
		}
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", AuthMeResponse{
		AuthEnabled:   true,
		Mode:          h.service.Mode(),
		Authenticated: true,
		User:          toUserResponse(user),
	}); err != nil {
		WriteError(w, err)
	}
}

func (h *AuthHandler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.service.Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		user, err := h.userFromRequest(r)
		if err != nil {
			WriteError(w, err)
			return
		}
		if !user.Admin {
			WriteError(w, core.ErrUserNotAllowed)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.service.Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			WriteError(w, ErrUnauthorized)
			return
		}
		if _, err := h.service.ValidateAPIKey(r.Context(), token); err != nil {
			WriteError(w, ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandler) userFromRequest(r *http.Request) (*core.User, error) {
	cookie, err := r.Cookie(service.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, core.ErrSessionNotFound
		}
		return nil, err
	}
	return h.service.UserForSessionToken(r.Context(), cookie.Value)
}

func (h *AuthHandler) setSession(w http.ResponseWriter, r *http.Request, userID string) error {
	token, expiresAt, err := h.service.CreateSession(r.Context(), userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     service.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.service.CookieSecure(),
	})
	return nil
}

func setTransientCookie(w http.ResponseWriter, name, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/auth",
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func clearCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomCookieToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
