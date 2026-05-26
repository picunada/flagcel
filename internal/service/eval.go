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
	if compiled, ok := s.cache.Get(key); ok {
		return s.cache.Evaluate(compiled, user), nil
	}

	cfg, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return false, fmt.Errorf("eval service: get flag %w", err)
	}

	compiled, err := s.cache.GetOrCompileLazy(*cfg, func() (*core.ContextSchema, error) {
		return s.contextForFlag(ctx, cfg)
	})
	if err != nil {
		return false, fmt.Errorf("eval service: compile flag %w", err)
	}

	return s.cache.Evaluate(compiled, user), nil
}

func (s *EvalService) EvaluateWithTrace(ctx context.Context, key string, user engine.DataContext) (engine.EvaluationTrace, error) {
	cfg, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return engine.EvaluationTrace{}, fmt.Errorf("eval service: get flag %w", err)
	}
	schema, err := s.contextForFlag(ctx, cfg)
	if err != nil {
		return engine.EvaluationTrace{}, err
	}

	return s.cache.engine.EvaluateConfigForContext(*cfg, schema, user), nil
}

func (s *EvalService) InvalidateContext(id string) {
	s.cache.InvalidateContext(id)
}

func (s *EvalService) InvalidateFlag(key string) {
	s.cache.InvalidateFlag(key)
}

func (s *EvalService) EvaluateAll(ctx context.Context, user engine.DataContext) (map[string]bool, error) {
	if flags, ok := s.cache.All(); ok {
		out := make(map[string]bool, len(flags))
		for _, flag := range flags {
			out[flag.Key] = s.cache.Evaluate(flag, user)
		}
		return out, nil
	}

	cfgs, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("eval service: list flags %w", err)
	}

	out := make(map[string]bool, len(cfgs))
	schemas := make(map[string]*core.ContextSchema)
	compiledFlags := make(map[string]cachedFlag, len(cfgs))
	for _, cfg := range cfgs {
		schema, err := s.contextForFlagCached(ctx, cfg, schemas)
		if err != nil {
			continue
		}
		cached, err := s.cache.Compile(*cfg, schema)
		if err != nil {
			continue
		}
		compiledFlags[cfg.Key] = cached
		compiled := cached.flag
		out[cfg.Key] = s.cache.Evaluate(compiled, user)
	}
	s.cache.SetAll(compiledFlags)
	return out, nil
}

func (s *EvalService) contextForFlag(ctx context.Context, cfg *core.FlagConfig) (*core.ContextSchema, error) {
	if cfg.ContextID == nil || *cfg.ContextID == "" {
		return nil, nil
	}
	schema, err := s.store.GetContext(ctx, *cfg.ContextID)
	if err != nil {
		return nil, fmt.Errorf("eval service: get context %w", err)
	}
	return schema, nil
}

func (s *EvalService) contextForFlagCached(ctx context.Context, cfg *core.FlagConfig, schemas map[string]*core.ContextSchema) (*core.ContextSchema, error) {
	if cfg.ContextID == nil || *cfg.ContextID == "" {
		return nil, nil
	}
	id := *cfg.ContextID
	if schema, ok := schemas[id]; ok {
		return schema, nil
	}
	schema, err := s.contextForFlag(ctx, cfg)
	if err != nil {
		return nil, err
	}
	schemas[id] = schema
	return schema, nil
}

type compiledFlagCache struct {
	engine *engine.Engine

	mu        sync.RWMutex
	flags     map[string]cachedFlag
	allLoaded bool
}

type cachedFlag struct {
	signature     uint64
	baseSignature uint64
	contextID     string
	flag          *engine.Flag
}

func (c *compiledFlagCache) Get(key string) (*engine.Flag, bool) {
	c.mu.RLock()
	cached, ok := c.flags[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cached.flag, true
}

func (c *compiledFlagCache) All() ([]*engine.Flag, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.allLoaded {
		return nil, false
	}
	flags := make([]*engine.Flag, 0, len(c.flags))
	for _, cached := range c.flags {
		flags = append(flags, cached.flag)
	}
	return flags, true
}

func (c *compiledFlagCache) GetOrCompile(cfg core.FlagConfig, schema *core.ContextSchema) (*engine.Flag, error) {
	signature := flagSignature(cfg, schema)

	c.mu.RLock()
	cached, ok := c.flags[cfg.Key]
	c.mu.RUnlock()
	if ok && cached.signature == signature {
		return cached.flag, nil
	}

	cached, err := c.Compile(cfg, schema)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.flags[cfg.Key] = cached
	c.allLoaded = false
	c.mu.Unlock()

	return cached.flag, nil
}

func (c *compiledFlagCache) Compile(cfg core.FlagConfig, schema *core.ContextSchema) (cachedFlag, error) {
	compiled, err := c.engine.CompileFlagForContext(cfg.Key, cfg, schema)
	if err != nil {
		return cachedFlag{}, err
	}

	return cachedFlag{
		signature:     flagSignature(cfg, schema),
		baseSignature: flagBaseSignature(cfg),
		contextID:     flagContextID(cfg),
		flag:          compiled,
	}, nil
}

func (c *compiledFlagCache) SetAll(flags map[string]cachedFlag) {
	c.mu.Lock()
	c.flags = flags
	c.allLoaded = true
	c.mu.Unlock()
}

func (c *compiledFlagCache) GetOrCompileLazy(cfg core.FlagConfig, loadSchema func() (*core.ContextSchema, error)) (*engine.Flag, error) {
	baseSignature := flagBaseSignature(cfg)

	c.mu.RLock()
	cached, ok := c.flags[cfg.Key]
	c.mu.RUnlock()
	if ok && cached.baseSignature == baseSignature {
		return cached.flag, nil
	}

	schema, err := loadSchema()
	if err != nil {
		return nil, err
	}
	return c.GetOrCompile(cfg, schema)
}

func (c *compiledFlagCache) InvalidateContext(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, cached := range c.flags {
		if cached.contextID == id {
			delete(c.flags, key)
			c.allLoaded = false
		}
	}
}

func (c *compiledFlagCache) InvalidateFlag(key string) {
	c.mu.Lock()
	delete(c.flags, key)
	c.allLoaded = false
	c.mu.Unlock()
}

func (c *compiledFlagCache) Evaluate(flag *engine.Flag, user engine.DataContext) bool {
	return c.engine.Evaluate(flag, user)
}

func flagBaseSignature(cfg core.FlagConfig) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s\x00%t\x00%t\x00%s", cfg.Key, cfg.Enabled, cfg.DefaultValue, flagContextID(cfg))
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

func flagSignature(cfg core.FlagConfig, schema *core.ContextSchema) uint64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%d", flagBaseSignature(cfg))
	if schema != nil {
		_, _ = fmt.Fprintf(h, "\x00context\x00%s", schema.ID)
		for _, field := range schema.Fields {
			_, _ = fmt.Fprintf(h, "\x00%s\x00%s", field.Path, field.Type)
		}
	}
	return h.Sum64()
}

func flagContextID(cfg core.FlagConfig) string {
	if cfg.ContextID == nil {
		return ""
	}
	return *cfg.ContextID
}
