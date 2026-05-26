package engine

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"
)

func (e *Engine) bucket(flagKey string, user DataContext, rollout Rollout) bool {
	return e.bucketDetails(flagKey, user, rollout).InRollout
}

func (e *Engine) bucketDetails(flagKey string, user DataContext, rollout Rollout) BucketTrace {
	bucketAttr := rollout.BucketBy
	if bucketAttr == "" {
		bucketAttr = "id"
	}

	trace := BucketTrace{
		BucketBy:   bucketAttr,
		Percentage: rollout.Percentage,
	}

	bucketValue, ok := lookupPath(user, bucketAttr)
	if !ok {
		trace.Missing = true
		if rollout.Percentage > 0 && rollout.Percentage < 100 {
			slog.Info("bucket: missing bucket by key/value in context")
		}
		if rollout.Percentage >= 100 {
			trace.InRollout = true
		}
		return trace
	}

	// Convert to string for hashing
	bucketStr := fmt.Sprintf("%v", bucketValue)
	trace.BucketValue = bucketStr

	hashInput := flagKey + ":" + bucketStr
	hash := sha1.Sum([]byte(hashInput))

	// Take the first 4 bytes as uint32 and reduce to 0-99 for percantage correlation.
	bucketNum := binary.BigEndian.Uint32(hash[:4]) % 100
	bucketNumber := int(bucketNum)
	trace.BucketNumber = &bucketNumber

	switch {
	case rollout.Percentage >= 100:
		trace.InRollout = true
	case rollout.Percentage <= 0:
		trace.InRollout = false
	default:
		trace.InRollout = bucketNum < uint32(rollout.Percentage)
	}
	return trace
}

func lookupPath(data DataContext, path string) (any, bool) {
	if value, ok := data[path]; ok {
		return value, true
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, false
	}

	var current any = map[string]any(data)
	for _, part := range parts {
		switch typed := current.(type) {
		case DataContext:
			var ok bool
			current, ok = typed[part]
			if !ok {
				return nil, false
			}
		case map[string]any:
			var ok bool
			current, ok = typed[part]
			if !ok {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return current, true
}
