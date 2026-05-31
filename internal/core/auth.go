package core

import (
	"errors"
	"time"
)

var (
	ErrAuthNotConfigured  = errors.New("auth is not configured")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotAllowed     = errors.New("user is not allowed")
	ErrSessionNotFound    = errors.New("session not found")
	ErrAPIKeyNotFound     = errors.New("api key not found")
)

type User struct {
	ID          string
	OIDCSubject string
	Email       string
	Name        string
	Description string
	Admin       bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *string
	DeletedBy   *string
}

type APIKey struct {
	ID          string
	Name        string
	Description string
	Prefix      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastUsedAt  *time.Time
	RevokedAt   *time.Time
	CreatedBy   *string
	DeletedBy   *string
}
