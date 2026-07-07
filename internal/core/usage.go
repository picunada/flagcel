package core

import (
	"context"
	"time"
)

type FlagUsageSource struct {
	APIKeyID string
	Source   string
}

type FlagUsageEvent struct {
	ID            string
	EnvironmentID string
	FlagKey       string
	ValueType     ValueType
	Value         any
	Reason        string
	MatchedRuleID *string
	APIKeyID      *string
	Source        string
	Latency       time.Duration
	ObservedAt    time.Time
	Context       map[string]any
}

type FlagUsageBucket struct {
	EnvironmentID string
	FlagKey       string
	BucketStart   time.Time
	ValueType     ValueType
	Value         any
	Reason        string
	MatchedRuleID *string
	APIKeyID      *string
	APIKeyName    string
	Source        string
	Count         int64
}

type FlagUsageLatencyBucket struct {
	EnvironmentID string
	FlagKey       string
	Source        string
	BucketStart   time.Time
	Count         int64
	AvgLatency    time.Duration
	P95Latency    time.Duration
}

type FlagUsageQuery struct {
	EnvironmentID string
	FlagKey       string
	Since         time.Time
	Limit         int32
}

type FlagUsageRecorder interface {
	RecordFlagUsage(ctx context.Context, event FlagUsageEvent) error
}
