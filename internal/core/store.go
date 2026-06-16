package core

import (
	"context"
	"errors"
)

var (
	ErrFlagNotFound        = errors.New("flag not found")
	ErrRuleNotFound        = errors.New("rule not found")
	ErrEnvironmentNotFound = errors.New("environment not found")
	ErrEnvironmentKeyTaken = errors.New("environment key already in use")
	ErrEnvironmentInUse    = errors.New("environment is in use")
	ErrDefaultEnvironment  = errors.New("default environment cannot be deleted")
	ErrContextNotFound     = errors.New("context not found")
	ErrContextNameTaken    = errors.New("context name already in use")
	ErrContextHasReferers  = errors.New("context referenced by flags")
)

type Store interface {
	ListEnvironments(ctx context.Context) ([]*Environment, error)
	GetEnvironment(ctx context.Context, id string) (*Environment, error)
	GetEnvironmentByKey(ctx context.Context, key string) (*Environment, error)
	CreateEnvironment(ctx context.Context, env *Environment) error
	UpdateEnvironment(ctx context.Context, env *Environment) error
	DeleteEnvironment(ctx context.Context, id string) error

	GetFlag(ctx context.Context, environmentID, key string) (*FlagConfig, error)
	ListFlags(ctx context.Context, environmentID string) ([]*FlagConfig, error)
	SaveFlag(ctx context.Context, environmentID string, flag *FlagConfig) error
	DeleteFlag(ctx context.Context, environmentID, key string) error

	GetRule(ctx context.Context, environmentID, flagKey, ruleID string) (*Rule, error)
	CreateRule(ctx context.Context, environmentID, flagKey string, rule Rule) error
	UpdateRule(ctx context.Context, environmentID, flagKey string, rule Rule) error
	DeleteRule(ctx context.Context, environmentID, flagKey, ruleID string) error
	ReorderRules(ctx context.Context, environmentID, flagKey string, ruleIDs []string) error

	ListContexts(ctx context.Context) ([]*ContextSchema, error)
	GetContext(ctx context.Context, id string) (*ContextSchema, error)
	CreateContext(ctx context.Context, c *ContextSchema) error
	UpdateContext(ctx context.Context, c *ContextSchema) error
	DeleteContext(ctx context.Context, id string) error
}
