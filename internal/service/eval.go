package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type EvalService struct {
	store              *postgres.Store
	cache              *compiledFlagCache
	usage              core.FlagUsageRecorder
	definitionsVersion atomic.Uint64
}

func NewEvalService(store *postgres.Store, eng *evalcore.Engine) *EvalService {
	s := &EvalService{
		store: store,
		cache: &compiledFlagCache{
			engine:    eng,
			flags:     make(map[string]cachedFlag),
			allLoaded: make(map[string]bool),
		},
	}
	s.definitionsVersion.Store(uint64(time.Now().UnixNano()))
	return s
}

func (s *EvalService) SetUsageRecorder(recorder core.FlagUsageRecorder) {
	s.usage = recorder
}

func (s *EvalService) Evaluate(ctx context.Context, environmentID, key string, user evalcore.DataContext) (core.FlagValue, error) {
	return s.EvaluateWithUsage(ctx, environmentID, key, user, core.FlagUsageSource{})
}

func (s *EvalService) EvaluateWithUsage(ctx context.Context, environmentID, key string, user evalcore.DataContext, source core.FlagUsageSource) (core.FlagValue, error) {
	start := time.Now()
	if compiled, ok := s.cache.Get(environmentID, key); ok {
		trace := s.cache.EvaluateTrace(compiled, user)
		s.recordUsage(ctx, environmentID, trace, source, time.Since(start))
		return traceValue(trace), nil
	}

	cfg, err := s.store.GetFlag(ctx, environmentID, key)
	if err != nil {
		return core.FlagValue{}, fmt.Errorf("eval service: get flag %w", err)
	}

	compiled, err := s.cache.GetOrCompileLazy(environmentID, *cfg, func() (*core.ContextSchema, error) {
		return s.contextForFlag(ctx, cfg)
	})
	if err != nil {
		return core.FlagValue{}, fmt.Errorf("eval service: compile flag %w", err)
	}

	trace := s.cache.EvaluateTrace(compiled, user)
	s.recordUsage(ctx, environmentID, trace, source, time.Since(start))
	return traceValue(trace), nil
}

func (s *EvalService) EvaluateWithTrace(ctx context.Context, environmentID, key string, user evalcore.DataContext) (evalcore.EvaluationTrace, error) {
	cfg, err := s.store.GetFlag(ctx, environmentID, key)
	if err != nil {
		return evalcore.EvaluationTrace{}, fmt.Errorf("eval service: get flag %w", err)
	}
	schema, err := s.contextForFlag(ctx, cfg)
	if err != nil {
		return evalcore.EvaluationTrace{}, err
	}

	return s.cache.engine.EvaluateConfigForContext(*cfg, schema, user), nil
}

func (s *EvalService) InvalidateContext(id string) {
	s.cache.InvalidateContext(id)
	s.bumpDefinitionsVersion()
}

func (s *EvalService) InvalidateFlag(environmentID, key string) {
	s.cache.InvalidateFlag(environmentID, key)
	s.bumpDefinitionsVersion()
}

func (s *EvalService) DefinitionsETag() string {
	return fmt.Sprintf(`"%d"`, s.definitionsVersion.Load())
}

func (s *EvalService) Definitions(ctx context.Context, environmentID string) (evalcore.Definitions, string, error) {
	etag := s.DefinitionsETag()
	cfgs, err := s.store.ListFlags(ctx, environmentID)
	if err != nil {
		return evalcore.Definitions{}, etag, fmt.Errorf("eval service: list definition flags %w", err)
	}

	flags := make([]evalcore.FlagDefinition, len(cfgs))
	contexts := make([]evalcore.ContextSchema, 0)
	seenContexts := make(map[string]struct{})

	for i, cfg := range cfgs {
		flag := normalizeFlag(*cfg)
		if flag.Rules == nil {
			flag.Rules = []core.Rule{}
		}
		flags[i] = evalcore.FlagDefinition{FlagConfig: flag}
		if flag.ContextID == nil || *flag.ContextID == "" {
			continue
		}

		contextID := *flag.ContextID
		if _, ok := seenContexts[contextID]; ok {
			continue
		}
		schema, err := s.store.GetContext(ctx, contextID)
		if err != nil {
			return evalcore.Definitions{}, etag, fmt.Errorf("eval service: get definition context %w", err)
		}
		contexts = append(contexts, *schema)
		seenContexts[contextID] = struct{}{}
	}

	return evalcore.Definitions{
		Flags:    flags,
		Contexts: contexts,
	}, etag, nil
}

func (s *EvalService) bumpDefinitionsVersion() {
	s.definitionsVersion.Add(1)
}

func (s *EvalService) EvaluateAll(ctx context.Context, environmentID string, user evalcore.DataContext) (map[string]core.FlagValue, error) {
	return s.EvaluateAllWithUsage(ctx, environmentID, user, core.FlagUsageSource{})
}

func (s *EvalService) EvaluateAllWithUsage(ctx context.Context, environmentID string, user evalcore.DataContext, source core.FlagUsageSource) (map[string]core.FlagValue, error) {
	if flags, ok := s.cache.All(environmentID); ok {
		out := make(map[string]core.FlagValue, len(flags))
		for _, flag := range flags {
			start := time.Now()
			trace := s.cache.EvaluateTrace(flag, user)
			out[flag.Key] = traceValue(trace)
			s.recordUsage(ctx, environmentID, trace, source, time.Since(start))
		}
		return out, nil
	}

	cfgs, err := s.store.ListFlags(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("eval service: list flags %w", err)
	}

	out := make(map[string]core.FlagValue, len(cfgs))
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
		compiledFlags[cacheKey(environmentID, cfg.Key)] = cached
		compiled := cached.flag
		start := time.Now()
		trace := s.cache.EvaluateTrace(compiled, user)
		out[cfg.Key] = traceValue(trace)
		s.recordUsage(ctx, environmentID, trace, source, time.Since(start))
	}
	s.cache.SetAll(environmentID, compiledFlags)
	return out, nil
}

func (s *EvalService) recordUsage(ctx context.Context, environmentID string, trace evalcore.EvaluationTrace, source core.FlagUsageSource, latency time.Duration) {
	if s.usage == nil {
		return
	}
	if latency <= 0 {
		latency = time.Nanosecond
	}
	event := core.FlagUsageEvent{
		EnvironmentID: environmentID,
		FlagKey:       trace.Key,
		ValueType:     core.ValueType(trace.Type),
		Value:         trace.Value,
		Reason:        trace.Reason,
		Source:        source.Source,
		Latency:       latency,
		ObservedAt:    time.Now(),
	}
	if trace.MatchedRule != nil {
		event.MatchedRuleID = &trace.MatchedRule.ID
	}
	if source.APIKeyID != "" {
		event.APIKeyID = &source.APIKeyID
	}
	_ = s.usage.RecordFlagUsage(ctx, event)
}

func traceValue(trace evalcore.EvaluationTrace) core.FlagValue {
	return core.FlagValue{
		Type:  core.ValueType(trace.Type),
		Value: trace.Value,
	}
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
	engine *evalcore.Engine

	mu        sync.RWMutex
	flags     map[string]cachedFlag
	allLoaded map[string]bool
}

type cachedFlag struct {
	signature     uint64
	baseSignature uint64
	contextID     string
	flag          *evalcore.Flag
}

func (c *compiledFlagCache) Get(environmentID, key string) (*evalcore.Flag, bool) {
	c.mu.RLock()
	cached, ok := c.flags[cacheKey(environmentID, key)]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cached.flag, true
}

func (c *compiledFlagCache) All(environmentID string) ([]*evalcore.Flag, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.allLoaded[environmentID] {
		return nil, false
	}
	flags := make([]*evalcore.Flag, 0, len(c.flags))
	prefix := environmentID + "\x00"
	for key, cached := range c.flags {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		flags = append(flags, cached.flag)
	}
	return flags, true
}

func (c *compiledFlagCache) GetOrCompile(environmentID string, cfg core.FlagConfig, schema *core.ContextSchema) (*evalcore.Flag, error) {
	cfg = normalizeFlag(cfg)
	signature := flagSignature(cfg, schema)
	key := cacheKey(environmentID, cfg.Key)

	c.mu.RLock()
	cached, ok := c.flags[key]
	c.mu.RUnlock()
	if ok && cached.signature == signature {
		return cached.flag, nil
	}

	cached, err := c.Compile(cfg, schema)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.ensureMapsLocked()
	c.flags[key] = cached
	c.allLoaded[environmentID] = false
	c.mu.Unlock()

	return cached.flag, nil
}

func (c *compiledFlagCache) Compile(cfg core.FlagConfig, schema *core.ContextSchema) (cachedFlag, error) {
	cfg = normalizeFlag(cfg)
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

func (c *compiledFlagCache) SetAll(environmentID string, flags map[string]cachedFlag) {
	c.mu.Lock()
	c.ensureMapsLocked()
	prefix := environmentID + "\x00"
	for key := range c.flags {
		if strings.HasPrefix(key, prefix) {
			delete(c.flags, key)
		}
	}
	for key, flag := range flags {
		c.flags[key] = flag
	}
	c.allLoaded[environmentID] = true
	c.mu.Unlock()
}

func (c *compiledFlagCache) GetOrCompileLazy(environmentID string, cfg core.FlagConfig, loadSchema func() (*core.ContextSchema, error)) (*evalcore.Flag, error) {
	cfg = normalizeFlag(cfg)
	baseSignature := flagBaseSignature(cfg)
	key := cacheKey(environmentID, cfg.Key)

	c.mu.RLock()
	cached, ok := c.flags[key]
	c.mu.RUnlock()
	if ok && cached.baseSignature == baseSignature {
		return cached.flag, nil
	}

	schema, err := loadSchema()
	if err != nil {
		return nil, err
	}
	return c.GetOrCompile(environmentID, cfg, schema)
}

func (c *compiledFlagCache) InvalidateContext(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, cached := range c.flags {
		if cached.contextID == id {
			delete(c.flags, key)
		}
	}
	if c.allLoaded != nil {
		clear(c.allLoaded)
	}
}

func (c *compiledFlagCache) InvalidateFlag(environmentID, key string) {
	c.mu.Lock()
	c.ensureMapsLocked()
	delete(c.flags, cacheKey(environmentID, key))
	c.allLoaded[environmentID] = false
	c.mu.Unlock()
}

func (c *compiledFlagCache) Evaluate(flag *evalcore.Flag, user evalcore.DataContext) core.FlagValue {
	return c.engine.Evaluate(flag, user)
}

func (c *compiledFlagCache) EvaluateTrace(flag *evalcore.Flag, user evalcore.DataContext) evalcore.EvaluationTrace {
	return c.engine.EvaluateTrace(flag, user)
}

func cacheKey(environmentID, flagKey string) string {
	return environmentID + "\x00" + flagKey
}

func (c *compiledFlagCache) ensureMapsLocked() {
	if c.flags == nil {
		c.flags = make(map[string]cachedFlag)
	}
	if c.allLoaded == nil {
		c.allLoaded = make(map[string]bool)
	}
}

func flagBaseSignature(cfg core.FlagConfig) uint64 {
	cfg = normalizeFlag(cfg)
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s\x00%t\x00%s\x00%s", cfg.Key, cfg.Enabled, cfg.Type, flagContextID(cfg))
	writeValueSignature(h, cfg.DefaultValue)
	for _, rule := range cfg.Rules {
		_, _ = fmt.Fprintf(
			h,
			"\x00%s\x00%s\x00%d\x00%s\x00",
			rule.ID,
			rule.Expression,
			rule.Rollout.Percentage,
			rule.Rollout.BucketBy,
		)
		writeValueSignature(h, rule.Value)
	}
	return h.Sum64()
}

func writeValueSignature(h interface{ Write([]byte) (int, error) }, value any) {
	b, err := json.Marshal(value)
	if err != nil {
		_, _ = fmt.Fprintf(h, "%#v", value)
		return
	}
	_, _ = h.Write(b)
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
