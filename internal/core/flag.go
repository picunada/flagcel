package core

type FlagConfig struct {
	Enabled      bool   `json:"enabled"`
	Rules        []Rule `json:"rules"`
	DefaultValue bool   `json:"default_value"`
}

type Rule struct {
	Expression string  `json:"expression"`
	Rollout    Rollout `json:"rollout"`
}

type Rollout struct {
	Percentage int    `json:"percentage"`
	BucketBy   string `json:"bucket_by,omitempty"`
}
