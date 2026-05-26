package engine

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/picunada/flagcel/internal/core"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	os.Exit(m.Run())
}

func newBenchEngine(b *testing.B) *Engine {
	b.Helper()
	env, err := NewCELEnv()
	if err != nil {
		b.Fatalf("new cel env: %v", err)
	}
	return NewEngine(env)
}

func makeFlagConfig(nRules int, matchIndex int) core.FlagConfig {
	rules := make([]core.Rule, nRules)
	for i := range rules {
		// Make the matching rule trivially true; all others are expressions that
		// must be evaluated but won't match the test context.
		var expr string
		if i == matchIndex {
			expr = `user.tier == "pro"`
		} else {
			expr = fmt.Sprintf(`user.region == "rgn-%d"`, i)
		}
		rules[i] = core.Rule{
			ID:         fmt.Sprintf("r-%d", i),
			Expression: expr,
			Rollout:    core.Rollout{Percentage: 100, BucketBy: "id"},
		}
	}
	return core.FlagConfig{
		Key:          "flag-bench",
		Enabled:      true,
		Rules:        rules,
		DefaultValue: false,
	}
}

func benchContext() DataContext {
	return DataContext{
		"user": map[string]any{
			"id":     "user-12345",
			"tier":   "pro",
			"region": "rgn-unknown",
			"email":  "x@example.com",
		},
	}
}

func benchContextSchema() *core.ContextSchema {
	return &core.ContextSchema{
		ID: "bench-context",
		Fields: []core.ContextField{
			{Path: "user.id", Type: core.ContextTypeString},
			{Path: "user.tier", Type: core.ContextTypeString},
			{Path: "user.region", Type: core.ContextTypeString},
			{Path: "user.email", Type: core.ContextTypeString},
		},
	}
}

func compileBenchFlag(b *testing.B, e *Engine, cfg core.FlagConfig) *Flag {
	b.Helper()
	flag, err := e.CompileFlagForContext(cfg.Key, cfg, benchContextSchema())
	if err != nil {
		b.Fatal(err)
	}
	return flag
}

// BenchmarkEvaluate_NoRules measures the fast path: enabled flag, no rules,
// falls through to DefaultValue. This is the floor for per-eval overhead.
func BenchmarkEvaluate_NoRules(b *testing.B) {
	e := newBenchEngine(b)
	flag, err := e.CompileFlag("k", core.FlagConfig{Enabled: true, DefaultValue: true})
	if err != nil {
		b.Fatal(err)
	}
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Evaluate(flag, ctx)
	}
}

// BenchmarkEvaluate_SingleRule_Match exercises one CEL program eval + bucket().
func BenchmarkEvaluate_SingleRule_Match(b *testing.B) {
	e := newBenchEngine(b)
	flag := compileBenchFlag(b, e, makeFlagConfig(1, 0))
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Evaluate(flag, ctx)
	}
}

// BenchmarkEvaluate_MultiRule_LastMatches is the worst-case linear scan over rules.
func BenchmarkEvaluate_MultiRule_LastMatches(b *testing.B) {
	const n = 10
	e := newBenchEngine(b)
	flag := compileBenchFlag(b, e, makeFlagConfig(n, n-1))
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Evaluate(flag, ctx)
	}
}

// BenchmarkCompileFlag isolates the cost of CEL compilation.
func BenchmarkCompileFlag(b *testing.B) {
	e := newBenchEngine(b)
	cfg := makeFlagConfig(5, 2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.CompileFlagForContext(cfg.Key, cfg, benchContextSchema()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkServicePath_CompilePerEval mirrors what EvalService.Evaluate does
// today (internal/service/eval.go): it compiles the flag on every call.
// Compare against BenchmarkEvaluate_SingleRule_Match to see the cost of that
// per-call compile.
func BenchmarkServicePath_CompilePerEval(b *testing.B) {
	e := newBenchEngine(b)
	cfg := makeFlagConfig(1, 0)
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flag, err := e.CompileFlagForContext(cfg.Key, cfg, benchContextSchema())
		if err != nil {
			b.Fatal(err)
		}
		_ = e.Evaluate(flag, ctx)
	}
}

// BenchmarkEvaluateAll_CompilePerEval simulates EvalService.EvaluateAll for
// a batch of N flags, each compiled on the fly. This is the dominant
// production cost today.
func BenchmarkEvaluateAll_CompilePerEval(b *testing.B) {
	const flagCount = 25
	e := newBenchEngine(b)
	cfgs := make([]core.FlagConfig, flagCount)
	for i := range cfgs {
		c := makeFlagConfig(2, 1)
		c.Key = fmt.Sprintf("flag-%d", i)
		cfgs[i] = c
	}
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := make(map[string]bool, flagCount)
		for _, c := range cfgs {
			flag, err := e.CompileFlagForContext(c.Key, c, benchContextSchema())
			if err != nil {
				b.Fatal(err)
			}
			out[c.Key] = e.Evaluate(flag, ctx)
		}
	}
}

// BenchmarkEvaluateAll_PreCompiled shows what the same batch costs if compile
// happens once up front (i.e. if compiled programs were cached).
func BenchmarkEvaluateAll_PreCompiled(b *testing.B) {
	const flagCount = 25
	e := newBenchEngine(b)
	flags := make([]*Flag, flagCount)
	for i := range flags {
		c := makeFlagConfig(2, 1)
		c.Key = fmt.Sprintf("flag-%d", i)
		f, err := e.CompileFlagForContext(c.Key, c, benchContextSchema())
		if err != nil {
			b.Fatal(err)
		}
		flags[i] = f
	}
	ctx := benchContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := make(map[string]bool, flagCount)
		for _, f := range flags {
			out[f.Key] = e.Evaluate(f, ctx)
		}
	}
}
