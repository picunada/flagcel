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

	first, err := cache.GetOrCompile(cfg)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	second, err := cache.GetOrCompile(cfg)
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

	first, err := cache.GetOrCompile(cfg)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	cfg.DefaultValue = true
	second, err := cache.GetOrCompile(cfg)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}

	if first == second {
		t.Fatal("expected changed flag config to replace cached compiled flag")
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
