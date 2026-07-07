package service

import (
	"context"
	"testing"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/core"
)

type recordingUsageSink struct {
	events []core.FlagUsageEvent
}

func (s *recordingUsageSink) RecordFlagUsage(ctx context.Context, event core.FlagUsageEvent) error {
	s.events = append(s.events, event)
	return nil
}

func TestEvalServiceEvaluateWithUsageRecordsMatchedRule(t *testing.T) {
	svc := NewEvalService(nil, newTestEngine(t))
	sink := &recordingUsageSink{}
	svc.SetUsageRecorder(sink)

	cfg := core.FlagConfig{
		Key:          "checkout-copy",
		Type:         core.ValueTypeString,
		Enabled:      true,
		DefaultValue: "control",
		Rules: []core.Rule{
			{
				ID:         "pro-users",
				Expression: `user.tier == "pro"`,
				Rollout:    core.Rollout{Percentage: 100, BucketBy: "user.id"},
				Value:      "variant-a",
			},
		},
	}
	if _, err := svc.cache.GetOrCompile(testEnvironmentID, cfg, userContextSchema()); err != nil {
		t.Fatalf("compile flag: %v", err)
	}

	value, err := svc.EvaluateWithUsage(
		context.Background(),
		testEnvironmentID,
		cfg.Key,
		evalcore.DataContext{"user": map[string]any{"id": "u-1", "tier": "pro"}},
		core.FlagUsageSource{APIKeyID: "api-key-1", Source: "js-sdk"},
	)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if value.Value != "variant-a" {
		t.Fatalf("value = %v, want variant-a", value.Value)
	}

	if len(sink.events) != 1 {
		t.Fatalf("recorded events = %d, want 1", len(sink.events))
	}
	event := sink.events[0]
	if event.EnvironmentID != testEnvironmentID {
		t.Fatalf("environment = %q, want %q", event.EnvironmentID, testEnvironmentID)
	}
	if event.FlagKey != cfg.Key {
		t.Fatalf("flag key = %q, want %q", event.FlagKey, cfg.Key)
	}
	if event.Value != "variant-a" {
		t.Fatalf("event value = %v, want variant-a", event.Value)
	}
	if event.ValueType != core.ValueTypeString {
		t.Fatalf("value type = %q, want %q", event.ValueType, core.ValueTypeString)
	}
	if event.Reason != "matched_rule" {
		t.Fatalf("reason = %q, want matched_rule", event.Reason)
	}
	if event.MatchedRuleID == nil || *event.MatchedRuleID != "pro-users" {
		t.Fatalf("matched rule = %v, want pro-users", event.MatchedRuleID)
	}
	if event.APIKeyID == nil || *event.APIKeyID != "api-key-1" {
		t.Fatalf("api key id = %v, want api-key-1", event.APIKeyID)
	}
	if event.Source != "js-sdk" {
		t.Fatalf("source = %q, want js-sdk", event.Source)
	}
	if event.Latency <= 0 {
		t.Fatalf("latency = %s, want positive duration", event.Latency)
	}
}

func TestEvalServiceEvaluateAllWithUsageRecordsEachFlag(t *testing.T) {
	svc := NewEvalService(nil, newTestEngine(t))
	sink := &recordingUsageSink{}
	svc.SetUsageRecorder(sink)

	first := core.FlagConfig{
		Key:          "new-nav",
		Type:         core.ValueTypeBoolean,
		Enabled:      true,
		DefaultValue: true,
	}
	second := core.FlagConfig{
		Key:          "checkout-copy",
		Type:         core.ValueTypeString,
		Enabled:      false,
		DefaultValue: "control",
	}
	firstCached, err := svc.cache.Compile(first, nil)
	if err != nil {
		t.Fatalf("compile first flag: %v", err)
	}
	secondCached, err := svc.cache.Compile(second, nil)
	if err != nil {
		t.Fatalf("compile second flag: %v", err)
	}
	svc.cache.SetAll(testEnvironmentID, map[string]cachedFlag{
		cacheKey(testEnvironmentID, first.Key):  firstCached,
		cacheKey(testEnvironmentID, second.Key): secondCached,
	})

	values, err := svc.EvaluateAllWithUsage(
		context.Background(),
		testEnvironmentID,
		evalcore.DataContext{},
		core.FlagUsageSource{APIKeyID: "api-key-1", Source: "go-sdk"},
	)
	if err != nil {
		t.Fatalf("evaluate all: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("values = %d, want 2", len(values))
	}
	if len(sink.events) != 2 {
		t.Fatalf("recorded events = %d, want 2", len(sink.events))
	}

	seen := map[string]bool{}
	for _, event := range sink.events {
		seen[event.FlagKey] = true
		if event.APIKeyID == nil || *event.APIKeyID != "api-key-1" {
			t.Fatalf("api key id = %v, want api-key-1", event.APIKeyID)
		}
		if event.Source != "go-sdk" {
			t.Fatalf("source = %q, want go-sdk", event.Source)
		}
	}
	if !seen[first.Key] || !seen[second.Key] {
		t.Fatalf("recorded flags = %+v, want %s and %s", seen, first.Key, second.Key)
	}
}
