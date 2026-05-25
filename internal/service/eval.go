package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/engine"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type EvalService struct {
	store *postgres.Store
	cache *compiledFlagCache
}

func NewEvalService(store *postgres.Store, eng *engine.Engine) *EvalService {
	return &EvalService{
		store: store,
		cache: &compiledFlagCache{
			engine: eng,
			flags:  make(map[string]cachedFlag),
		},
	}
}

func (s *EvalService) Evaluate(ctx context.Context, key string, user engine.DataContext) (bool, error) {
	cfg, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return false, fmt.Errorf("eval service: get flag %w", err)
	}

	compiled, err := s.cache.GetOrCompile(*cfg)
	if err != nil {
		return false, fmt.Errorf("eval service: compile flag %w", err)
	}

	return s.cache.Evaluate(compiled, user), nil
}

func (s *EvalService) EvaluateAll(ctx context.Context, user engine.DataContext) (map[string]bool, error) {
	cfgs, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("eval service: list flags %w", err)
	}

	out := make(map[string]bool, len(cfgs))
	for _, cfg := range cfgs {
		compiled, err := s.cache.GetOrCompile(*cfg)
		if err != nil {
			continue
		}
		out[cfg.Key] = s.cache.Evaluate(compiled, user)
	}
	return out, nil
}

type compiledFlagCache struct {
	engine *engine.Engine

	mu    sync.RWMutex
	flags map[string]cachedFlag
}

type cachedFlag struct {
	signature uint64
	flag      *engine.Flag
}

func (c *compiledFlagCache) GetOrCompile(cfg core.FlagConfig) (*engine.Flag, error) {
	signature := flagSignature(cfg)

	c.mu.RLock()
	cached, ok := c.flags[cfg.Key]
	c.mu.RUnlock()
	if ok && cached.signature == signature {
		return cached.flag, nil
	}

	compiled, err := c.engine.CompileFlag(cfg.Key, cfg)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.flags[cfg.Key] = cachedFlag{
		signature: signature,
		flag:      compiled,
	}
	c.mu.Unlock()

	return compiled, nil
}

func (c *compiledFlagCache) Evaluate(flag *engine.Flag, user engine.DataContext) bool {
	return c.engine.Evaluate(flag, user)
}

func flagSignature(cfg core.FlagConfig) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s\x00%t\x00%t", cfg.Key, cfg.Enabled, cfg.DefaultValue)
	for _, rule := range cfg.Rules {
		_, _ = fmt.Fprintf(
			h,
			"\x00%s\x00%s\x00%d\x00%s",
			rule.ID,
			rule.Expression,
			rule.Rollout.Percentage,
			rule.Rollout.BucketBy,
		)
	}
	return h.Sum64()
}
