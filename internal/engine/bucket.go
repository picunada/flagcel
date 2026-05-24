package engine

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log/slog"
)

func (e *Engine) bucket(flagKey string, user DataContext, rollout Rollout) bool {
	// Trivial cases
	if rollout.Percentage >= 100 {
		return true
	}
	if rollout.Percentage <= 0 {
		return false
	}

	bucketAttr := rollout.BucketBy
	if bucketAttr == "" {
		bucketAttr = "id"
	}

	bucketValue, ok := user[bucketAttr]
	if !ok {
		// Missing buckey attribute - we can't bucket deterministically.
		// Return false (treat as "not in rollout") rather than picking
		// a default, because picking arbitrarily would violate the
		// determinism guarantee.
		slog.Info("bucket: missing bucket by key/value in context")
		return false
	}

	// Convert to string for hashing
	bucketStr := fmt.Sprintf("%v", bucketValue)

	hashInput := flagKey + ":" + bucketStr
	hash := sha1.Sum([]byte(hashInput))

	// Take the first 4 bytes as uint32 and reduce to 0-99 for percantage correlation.
	bucketNum := binary.BigEndian.Uint32(hash[:4]) % 100

	return bucketNum < uint32(rollout.Percentage)
}
