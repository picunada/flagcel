package service

import (
	"context"
	"fmt"

	"github.com/picunada/flagcel/internal/engine"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type EvalService struct {
	store  *postgres.Store
	engine *engine.Engine
}

func NewEvalService(store *postgres.Store, eng *engine.Engine) *EvalService {
	return &EvalService{store: store, engine: eng}
}

func (s *EvalService) Evaluate(ctx context.Context, key string, user engine.DataContext) (bool, error) {
	cfg, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return false, fmt.Errorf("eval service: get flag %w", err)
	}

	compiled, err := s.engine.CompileFlag(cfg.Key, *cfg)
	if err != nil {
		return false, fmt.Errorf("eval service: compile flag %w", err)
	}

	return s.engine.Evaluate(compiled, user), nil
}

func (s *EvalService) EvaluateAll(ctx context.Context, user engine.DataContext) (map[string]bool, error) {
	cfgs, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("eval service: list flags %w", err)
	}

	out := make(map[string]bool, len(cfgs))
	for _, cfg := range cfgs {
		compiled, err := s.engine.CompileFlag(cfg.Key, *cfg)
		if err != nil {
			continue
		}
		out[cfg.Key] = s.engine.Evaluate(compiled, user)
	}
	return out, nil
}
