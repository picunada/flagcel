package core

import (
	"context"
	"errors"
)

var (
	ErrFlagNotFound = errors.New("flag not found")
	ErrRuleNotFound = errors.New("rule not found")
)

type Store interface {
	GetFlag(ctx context.Context, key string) (*FlagConfig, error)
	ListFlags(ctx context.Context) ([]*FlagConfig, error)
	SaveFlag(ctx context.Context, flag *FlagConfig) error
	DeleteFlag(ctx context.Context, key string) error

	GetRule(ctx context.Context, flagKey, ruleID string) (*Rule, error)
	CreateRule(ctx context.Context, flagKey string, rule Rule) error
	UpdateRule(ctx context.Context, flagKey string, rule Rule) error
	DeleteRule(ctx context.Context, flagKey, ruleID string) error
	ReorderRules(ctx context.Context, flagKey string, ruleIDs []string) error
}
