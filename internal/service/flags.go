package service

import (
	"context"
	"fmt"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type FlagService struct {
	store        *postgres.Store
	onFlagChange func(string)
}

func NewFlagService(store *postgres.Store, onFlagChange ...func(string)) *FlagService {
	s := &FlagService{store: store}
	if len(onFlagChange) > 0 {
		s.onFlagChange = onFlagChange[0]
	}
	return s
}

func (s *FlagService) GetFlags(ctx context.Context) ([]*core.FlagConfig, error) {
	// Fetch flag from store
	flags, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failder to list flags %w",
			err,
		)
	}
	return flags, nil
}

func (s *FlagService) GetFlag(ctx context.Context, key string) (*core.FlagConfig, error) {
	flag, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failed to get flag %w",
			err,
		)
	}
	return flag, nil
}

func (s *FlagService) CreateFlag(ctx context.Context, flag *core.FlagConfig) error {
	schema, err := s.contextForFlag(ctx, flag)
	if err != nil {
		return fmt.Errorf("flag service: failed to load context %w", err)
	}
	for i, rule := range flag.Rules {
		rule = normalizeRule(rule)
		if err := validateRule(rule, schema); err != nil {
			return prefixValidationError(err, fmt.Sprintf("rules[%d].", i))
		}
		flag.Rules[i] = rule
	}
	if err := s.store.SaveFlag(ctx, flag); err != nil {
		return fmt.Errorf(
			"flag service: failed to create flag %w",
			err,
		)
	}
	s.invalidate(flag.Key)
	return nil
}

func (s *FlagService) DeleteFlag(context context.Context, id string) error {
	if err := s.store.DeleteFlag(context, id); err != nil {
		return fmt.Errorf(
			"flag service: failed to delete flag %w",
			err,
		)
	}
	s.invalidate(id)
	return nil
}

func (s *FlagService) invalidate(key string) {
	if s.onFlagChange != nil {
		s.onFlagChange(key)
	}
}

func (s *FlagService) contextForFlag(ctx context.Context, flag *core.FlagConfig) (*core.ContextSchema, error) {
	if flag.ContextID == nil || *flag.ContextID == "" {
		return nil, nil
	}
	return s.store.GetContext(ctx, *flag.ContextID)
}

func prefixValidationError(err error, fieldPrefix string) error {
	validationErr, ok := err.(*core.ValidationError)
	if !ok {
		return err
	}
	issues := make([]core.ValidationIssue, len(validationErr.Issues))
	for i, issue := range validationErr.Issues {
		issue.Field = fieldPrefix + issue.Field
		issues[i] = issue
	}
	return &core.ValidationError{
		Message: validationErr.Message,
		Issues:  issues,
	}
}
