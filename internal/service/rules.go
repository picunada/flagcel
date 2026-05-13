package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type RuleService struct {
	store *postgres.Store
}

func NewRuleService(store *postgres.Store) *RuleService {
	return &RuleService{store: store}
}

func (s *RuleService) ListRules(ctx context.Context, flagKey string) ([]core.Rule, error) {
	flag, err := s.store.GetFlag(ctx, flagKey)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to list rules %w", err)
	}
	return flag.Rules, nil
}

func (s *RuleService) GetRule(ctx context.Context, flagKey, ruleID string) (*core.Rule, error) {
	rule, err := s.store.GetRule(ctx, flagKey, ruleID)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to get rule %w", err)
	}
	return rule, nil
}

func (s *RuleService) CreateRule(ctx context.Context, flagKey string, rule core.Rule) (*core.Rule, error) {
	if rule.ID == "" {
		rule.ID = uuid.NewString()
	}
	if err := s.store.CreateRule(ctx, flagKey, rule); err != nil {
		return nil, fmt.Errorf("rule service: failed to create rule %w", err)
	}
	return &rule, nil
}

func (s *RuleService) UpdateRule(ctx context.Context, flagKey string, rule core.Rule) error {
	if err := s.store.UpdateRule(ctx, flagKey, rule); err != nil {
		return fmt.Errorf("rule service: failed to update rule %w", err)
	}
	return nil
}

func (s *RuleService) DeleteRule(ctx context.Context, flagKey, ruleID string) error {
	if err := s.store.DeleteRule(ctx, flagKey, ruleID); err != nil {
		return fmt.Errorf("rule service: failed to delete rule %w", err)
	}
	return nil
}

func (s *RuleService) ReorderRules(ctx context.Context, flagKey string, ruleIDs []string) error {
	if err := s.store.ReorderRules(ctx, flagKey, ruleIDs); err != nil {
		return fmt.Errorf("rule service: failed to reorder rules %w", err)
	}
	return nil
}
