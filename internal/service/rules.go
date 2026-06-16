package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type RuleService struct {
	store        *postgres.Store
	onFlagChange func(string, string)
}

func NewRuleService(store *postgres.Store, onFlagChange ...func(string, string)) *RuleService {
	s := &RuleService{store: store}
	if len(onFlagChange) > 0 {
		s.onFlagChange = onFlagChange[0]
	}
	return s
}

func (s *RuleService) ListRules(ctx context.Context, environmentID, flagKey string) ([]core.Rule, error) {
	flag, err := s.store.GetFlag(ctx, environmentID, flagKey)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to list rules %w", err)
	}
	return flag.Rules, nil
}

func (s *RuleService) GetRule(ctx context.Context, environmentID, flagKey, ruleID string) (*core.Rule, error) {
	rule, err := s.store.GetRule(ctx, environmentID, flagKey, ruleID)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to get rule %w", err)
	}
	return rule, nil
}

func (s *RuleService) CreateRule(ctx context.Context, environmentID, flagKey string, rule core.Rule) (*core.Rule, error) {
	rule = normalizeRule(rule)
	flag, schema, err := s.flagAndContext(ctx, environmentID, flagKey)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to load flag context %w", err)
	}
	if err := validateRule(rule, schema); err != nil {
		return nil, err
	}
	if err := validateRuleValue(rule, flag.Type); err != nil {
		return nil, err
	}
	if rule.ID == "" {
		rule.ID = uuid.NewString()
	}
	if err := s.store.CreateRule(ctx, environmentID, flagKey, rule); err != nil {
		return nil, fmt.Errorf("rule service: failed to create rule %w", err)
	}
	s.invalidate(environmentID, flagKey)
	out, err := s.store.GetRule(ctx, environmentID, flagKey, rule.ID)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to load created rule %w", err)
	}
	return out, nil
}

func (s *RuleService) UpdateRule(ctx context.Context, environmentID, flagKey string, rule core.Rule) (*core.Rule, error) {
	rule = normalizeRule(rule)
	flag, schema, err := s.flagAndContext(ctx, environmentID, flagKey)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to load flag context %w", err)
	}
	if err := validateRule(rule, schema); err != nil {
		return nil, err
	}
	if err := validateRuleValue(rule, flag.Type); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRule(ctx, environmentID, flagKey, rule); err != nil {
		return nil, fmt.Errorf("rule service: failed to update rule %w", err)
	}
	s.invalidate(environmentID, flagKey)
	out, err := s.store.GetRule(ctx, environmentID, flagKey, rule.ID)
	if err != nil {
		return nil, fmt.Errorf("rule service: failed to load updated rule %w", err)
	}
	return out, nil
}

func (s *RuleService) DeleteRule(ctx context.Context, environmentID, flagKey, ruleID string) error {
	if err := s.store.DeleteRule(ctx, environmentID, flagKey, ruleID); err != nil {
		return fmt.Errorf("rule service: failed to delete rule %w", err)
	}
	s.invalidate(environmentID, flagKey)
	return nil
}

func (s *RuleService) ReorderRules(ctx context.Context, environmentID, flagKey string, ruleIDs []string) error {
	if err := s.store.ReorderRules(ctx, environmentID, flagKey, ruleIDs); err != nil {
		return fmt.Errorf("rule service: failed to reorder rules %w", err)
	}
	s.invalidate(environmentID, flagKey)
	return nil
}

func (s *RuleService) invalidate(environmentID, key string) {
	if s.onFlagChange != nil {
		s.onFlagChange(environmentID, key)
	}
}

func (s *RuleService) flagAndContext(ctx context.Context, environmentID, flagKey string) (*core.FlagConfig, *core.ContextSchema, error) {
	flag, err := s.store.GetFlag(ctx, environmentID, flagKey)
	if err != nil {
		return nil, nil, err
	}
	*flag = normalizeFlag(*flag)
	if flag.ContextID == nil || *flag.ContextID == "" {
		return flag, nil, nil
	}
	schema, err := s.store.GetContext(ctx, *flag.ContextID)
	if err != nil {
		return nil, nil, err
	}
	return flag, schema, nil
}
