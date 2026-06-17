package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type FlagService struct {
	store        *postgres.Store
	onFlagChange func(string, string)
}

func NewFlagService(store *postgres.Store, onFlagChange ...func(string, string)) *FlagService {
	s := &FlagService{store: store}
	if len(onFlagChange) > 0 {
		s.onFlagChange = onFlagChange[0]
	}
	return s
}

func (s *FlagService) GetFlags(ctx context.Context, environmentID string) ([]*core.FlagConfig, error) {
	// Fetch flag from store
	flags, err := s.store.ListFlags(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failder to list flags %w",
			err,
		)
	}
	return flags, nil
}

func (s *FlagService) GetFlag(ctx context.Context, environmentID, key string) (*core.FlagConfig, error) {
	flag, err := s.store.GetFlag(ctx, environmentID, key)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failed to get flag %w",
			err,
		)
	}
	return flag, nil
}

func (s *FlagService) GetFlagHistory(ctx context.Context, environmentID, key string) ([]*core.AuditEntry, error) {
	entries, err := s.store.ListFlagAuditLog(ctx, environmentID, key)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failed to list flag history %w",
			err,
		)
	}
	return entries, nil
}

func (s *FlagService) CreateFlag(ctx context.Context, environmentID string, flag *core.FlagConfig) error {
	*flag = normalizeFlag(*flag)
	// Rules are persisted by (environment_id, flag_key, id); a missing id would
	// collide on insert when a flag has more than one rule. Assign ids to new
	// rules while preserving any the client sent so rule identity stays stable.
	for i := range flag.Rules {
		if flag.Rules[i].ID == "" {
			flag.Rules[i].ID = uuid.NewString()
		}
	}
	schema, err := s.contextForFlag(ctx, flag)
	if err != nil {
		return fmt.Errorf("flag service: failed to load context %w", err)
	}
	if err := validateFlag(*flag, schema); err != nil {
		return err
	}
	if err := s.store.SaveFlag(ctx, environmentID, flag); err != nil {
		return fmt.Errorf(
			"flag service: failed to create flag %w",
			err,
		)
	}
	s.invalidate(environmentID, flag.Key)
	return nil
}

func (s *FlagService) DeleteFlag(context context.Context, environmentID, id string) error {
	if err := s.store.DeleteFlag(context, environmentID, id); err != nil {
		return fmt.Errorf(
			"flag service: failed to delete flag %w",
			err,
		)
	}
	s.invalidate(environmentID, id)
	return nil
}

func (s *FlagService) invalidate(environmentID, key string) {
	if s.onFlagChange != nil {
		s.onFlagChange(environmentID, key)
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
