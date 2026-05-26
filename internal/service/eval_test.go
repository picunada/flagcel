package service

import (
	"testing"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/engine"
)

func TestCompiledFlagCacheGetOrCompileReusesUnchangedFlag(t *testing.T) {
	eng := newTestEngine(t)
	cache := &compiledFlagCache{
		engine: eng,
		flags:  make(map[string]cachedFlag),
	}

	cfg := core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
		Rules: []core.Rule{
			{
				ID:         "rule-a",
				Expression: `user.tier == "pro"`,
				Rollout:    core.Rollout{Percentage: 100, BucketBy: "id"},
			},
		},
	}

	first, err := cache.GetOrCompile(cfg, userContextSchema())
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	second, err := cache.GetOrCompile(cfg, userContextSchema())
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}

	if first != second {
		t.Fatal("expected unchanged flag config to reuse cached compiled flag")
	}
}

func TestCompiledFlagCacheGetOrCompileRecompilesChangedFlag(t *testing.T) {
	eng := newTestEngine(t)
	cache := &compiledFlagCache{
		engine: eng,
		flags:  make(map[string]cachedFlag),
	}

	cfg := core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
	}

	first, err := cache.GetOrCompile(cfg, nil)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	cfg.DefaultValue = true
	second, err := cache.GetOrCompile(cfg, nil)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}

	if first == second {
		t.Fatal("expected changed flag config to replace cached compiled flag")
	}
}

func TestCompiledFlagCacheGetOrCompileRecompilesChangedContext(t *testing.T) {
	eng := newTestEngine(t)
	cache := &compiledFlagCache{
		engine: eng,
		flags:  make(map[string]cachedFlag),
	}

	contextID := "context-a"
	cfg := core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
		ContextID:    &contextID,
		Rules: []core.Rule{
			{
				ID:         "rule-a",
				Expression: `request.path == "/checkout"`,
				Rollout:    core.Rollout{Percentage: 100, BucketBy: "id"},
			},
		},
	}

	schema := &core.ContextSchema{
		ID: contextID,
		Fields: []core.ContextField{
			{Path: "request.path", Type: core.ContextTypeString},
		},
	}

	first, err := cache.GetOrCompile(cfg, schema)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	schema.Fields = append(schema.Fields, core.ContextField{
		Path: "request.method",
		Type: core.ContextTypeString,
	})
	second, err := cache.GetOrCompile(cfg, schema)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}

	if first == second {
		t.Fatal("expected changed context schema to replace cached compiled flag")
	}
}

func TestCompiledFlagCacheGetOrCompileLazyDoesNotLoadContextOnHit(t *testing.T) {
	eng := newTestEngine(t)
	cache := &compiledFlagCache{
		engine: eng,
		flags:  make(map[string]cachedFlag),
	}

	cfg := userFlagConfig()
	if _, err := cache.GetOrCompile(cfg, userContextSchema()); err != nil {
		t.Fatalf("first compile: %v", err)
	}

	_, err := cache.GetOrCompileLazy(cfg, func() (*core.ContextSchema, error) {
		t.Fatal("schema loader should not run on cache hit")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("cached compile: %v", err)
	}
}

func TestCompiledFlagCacheInvalidateContext(t *testing.T) {
	eng := newTestEngine(t)
	cache := &compiledFlagCache{
		engine: eng,
		flags:  make(map[string]cachedFlag),
	}

	cfg := userFlagConfig()
	first, err := cache.GetOrCompile(cfg, userContextSchema())
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	cache.InvalidateContext("context-user")

	second, err := cache.GetOrCompileLazy(cfg, func() (*core.ContextSchema, error) {
		return userContextSchema(), nil
	})
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}
	if first == second {
		t.Fatal("expected context invalidation to force recompile")
	}
}

func newTestEngine(t *testing.T) *engine.Engine {
	t.Helper()
	env, err := engine.NewCELEnv()
	if err != nil {
		t.Fatalf("new cel env: %v", err)
	}
	return engine.NewEngine(env)
}

func userFlagConfig() core.FlagConfig {
	contextID := "context-user"
	return core.FlagConfig{
		Key:          "feature-a",
		Enabled:      true,
		DefaultValue: false,
		ContextID:    &contextID,
		Rules: []core.Rule{
			{
				ID:         "rule-a",
				Expression: `user.tier == "pro"`,
				Rollout:    core.Rollout{Percentage: 100, BucketBy: "id"},
			},
		},
	}
}

func userContextSchema() *core.ContextSchema {
	return &core.ContextSchema{
		ID: "context-user",
		Fields: []core.ContextField{
			{Path: "user.tier", Type: core.ContextTypeString},
		},
	}
}
