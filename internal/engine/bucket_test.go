package engine

import "testing"

func TestEngine_bucket(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		flagKey string
		user    DataContext
		rollout Rollout
		want    bool
	}{
		{
			name:    "bucket returns false when rollout percentage is 0",
			flagKey: "",
			user:    DataContext{},
			rollout: Rollout{
				Percentage: 0,
			},
			want: false,
		},
		{
			name:    "bucket returns true when rollout percentage is 100",
			flagKey: "",
			user:    DataContext{},
			rollout: Rollout{
				Percentage: 100,
			},
			want: true,
		},
		{
			name:    "bucket returns false when rollout bucketBy is set to not-existing field in user context",
			flagKey: "test-flag",
			user:    DataContext{},
			rollout: Rollout{
				BucketBy: "non-existing-field",
			},
			want: false,
		},
		{
			name:    "bucket resolves dotted paths",
			flagKey: "test-flag",
			user: DataContext{
				"user": map[string]any{"id": "user-123"},
			},
			rollout: Rollout{
				Percentage: 100,
				BucketBy:   "user.id",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var e Engine
			got := e.bucket(tt.flagKey, tt.user, tt.rollout)
			if got != tt.want {
				t.Errorf("bucket() = %v, want %v", got, tt.want)
			}
		})
	}
}
