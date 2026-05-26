package engine

import "testing"

func TestEvaluateDisabledFlagReturnsDefaultValue(t *testing.T) {
	e := &Engine{}

	tests := []struct {
		name         string
		defaultValue bool
		want         bool
	}{
		{
			name:         "default false",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "default true",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.Evaluate(&Flag{
				Key:          "feature-a",
				Enabled:      false,
				DefaultValue: tt.defaultValue,
				Rules: []CompiledRule{
					{
						Rollout: Rollout{Percentage: 100, BucketBy: "id"},
					},
				},
			}, DataContext{
				"id": "user-123",
			})

			if got != tt.want {
				t.Fatalf("Evaluate() = %t, want %t", got, tt.want)
			}
		})
	}
}
