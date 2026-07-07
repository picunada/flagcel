package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/picunada/flagcel/internal/core"
)

func TestUsageQueryFromRequestDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/environments/env-test/flags/checkout/usage", nil)
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	query := usageQueryFromRequest(req, "env-test", "checkout", now)

	if query.EnvironmentID != "env-test" {
		t.Fatalf("environment = %q, want env-test", query.EnvironmentID)
	}
	if query.FlagKey != "checkout" {
		t.Fatalf("flag key = %q, want checkout", query.FlagKey)
	}
	if !query.Since.Equal(now.Add(-24 * time.Hour)) {
		t.Fatalf("since = %s, want %s", query.Since, now.Add(-24*time.Hour))
	}
	if query.Limit != 50 {
		t.Fatalf("limit = %d, want 50", query.Limit)
	}
}

func TestUsageQueryFromRequestAcceptsChartRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/environments/env-test/usage?hours=168&limit=25", nil)
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	query := usageQueryFromRequest(req, "env-test", "", now)

	if query.FlagKey != "" {
		t.Fatalf("flag key = %q, want empty", query.FlagKey)
	}
	if !query.Since.Equal(now.Add(-168 * time.Hour)) {
		t.Fatalf("since = %s, want %s", query.Since, now.Add(-168*time.Hour))
	}
	if query.Limit != 25 {
		t.Fatalf("limit = %d, want 25", query.Limit)
	}
}

func TestToFlagUsageResponseIncludesFlagKeyAndLatencyBuckets(t *testing.T) {
	ruleID := "rule-a"
	bucketStart := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	response := toFlagUsageResponse(
		[]*core.FlagUsageBucket{
			{
				EnvironmentID: "env-test",
				FlagKey:       "checkout",
				BucketStart:   bucketStart,
				ValueType:     core.ValueTypeBoolean,
				Value:         true,
				Reason:        "matched_rule",
				MatchedRuleID: &ruleID,
				Source:        "js-sdk",
				Count:         7,
			},
		},
		[]*core.FlagUsageLatencyBucket{
			{
				EnvironmentID: "env-test",
				FlagKey:       "checkout",
				BucketStart:   bucketStart,
				Count:         7,
				AvgLatency:    12 * time.Millisecond,
				P95Latency:    25 * time.Millisecond,
			},
		},
		nil,
	)

	if len(response.Buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(response.Buckets))
	}
	if response.Buckets[0].FlagKey != "checkout" {
		t.Fatalf("bucket flag key = %q, want checkout", response.Buckets[0].FlagKey)
	}
	if len(response.LatencyBuckets) != 1 {
		t.Fatalf("latency buckets = %d, want 1", len(response.LatencyBuckets))
	}
	latency := response.LatencyBuckets[0]
	if latency.FlagKey != "checkout" {
		t.Fatalf("latency flag key = %q, want checkout", latency.FlagKey)
	}
	if latency.AvgLatencyMs != 12 {
		t.Fatalf("avg latency = %f, want 12", latency.AvgLatencyMs)
	}
	if latency.P95LatencyMs != 25 {
		t.Fatalf("p95 latency = %f, want 25", latency.P95LatencyMs)
	}
}

func TestUsageHandlerGetEnvironmentUsage(t *testing.T) {
	store := &usageStoreStub{
		environmentBuckets: []*core.FlagUsageBucket{
			{
				EnvironmentID: "env-test",
				FlagKey:       "checkout",
				BucketStart:   time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
				ValueType:     core.ValueTypeBoolean,
				Value:         true,
				Reason:        "matched_rule",
				Count:         3,
			},
		},
		environmentLatencyBuckets: []*core.FlagUsageLatencyBucket{
			{
				EnvironmentID: "env-test",
				FlagKey:       "checkout",
				BucketStart:   time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
				Count:         3,
				AvgLatency:    4 * time.Millisecond,
				P95Latency:    8 * time.Millisecond,
			},
		},
	}
	handler := NewUsageHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/environments/env-test/usage?hours=168", nil)
	req.SetPathValue("environment_id", "env-test")
	rec := httptest.NewRecorder()

	handler.GetEnvironmentUsage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.environmentQuery.EnvironmentID != "env-test" {
		t.Fatalf("environment query id = %q, want env-test", store.environmentQuery.EnvironmentID)
	}
	if store.environmentQuery.FlagKey != "" {
		t.Fatalf("environment query flag = %q, want empty", store.environmentQuery.FlagKey)
	}
	if !store.environmentQuery.Since.Before(time.Now()) {
		t.Fatalf("environment query since = %s, want historical time", store.environmentQuery.Since)
	}

	var envelope struct {
		Data FlagUsageResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data.Buckets) != 1 {
		t.Fatalf("response buckets = %d, want 1", len(envelope.Data.Buckets))
	}
	if len(envelope.Data.LatencyBuckets) != 1 {
		t.Fatalf("response latency buckets = %d, want 1", len(envelope.Data.LatencyBuckets))
	}
}

type usageStoreStub struct {
	environmentQuery          core.FlagUsageQuery
	environmentBuckets        []*core.FlagUsageBucket
	environmentLatencyBuckets []*core.FlagUsageLatencyBucket
}

func (s *usageStoreStub) RecordFlagUsage(context.Context, core.FlagUsageEvent) error {
	return nil
}

func (s *usageStoreStub) ListFlagUsageBuckets(context.Context, core.FlagUsageQuery) ([]*core.FlagUsageBucket, error) {
	return nil, nil
}

func (s *usageStoreStub) ListEnvironmentUsageBuckets(_ context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageBucket, error) {
	s.environmentQuery = query
	return s.environmentBuckets, nil
}

func (s *usageStoreStub) ListFlagUsageLatencyBuckets(context.Context, core.FlagUsageQuery) ([]*core.FlagUsageLatencyBucket, error) {
	return nil, nil
}

func (s *usageStoreStub) ListEnvironmentUsageLatencyBuckets(_ context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageLatencyBucket, error) {
	s.environmentQuery = query
	return s.environmentLatencyBuckets, nil
}

func (s *usageStoreStub) ListFlagEvaluationEvents(context.Context, core.FlagUsageQuery) ([]*core.FlagUsageEvent, error) {
	return nil, nil
}
