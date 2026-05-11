package v1

type FlagResponse struct {
	Key          string         `json:"key"`
	Enabled      bool           `json:"enabled"`
	Rules        []RuleResponse `json:"rules"`
	DefaultValue bool           `json:"default_value"`
}

type RuleResponse struct {
	ID         string          `json:"id"`
	Expression string          `json:"expression"`
	Rollout    RolloutResponse `json:"rollout"`
}

type RolloutResponse struct {
	Percentage int    `json:"percentage"`
	BucketBy   string `json:"bucket_by,omitempty"`
}

type CreateFlagRequest struct {
	Key          string         `json:"key"`
	Enabled      bool           `json:"enabled"`
	Rules        []RuleResponse `json:"rules"`
	DefaultValue bool           `json:"default_value"`
}
