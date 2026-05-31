package v1

import "encoding/json"

type FlagResponse struct {
	Key          string         `json:"key"`
	Description  string         `json:"description,omitempty"`
	Type         string         `json:"type"`
	Enabled      bool           `json:"enabled"`
	Rules        []RuleResponse `json:"rules"`
	DefaultValue any            `json:"default_value"`
	ContextID    *string        `json:"context_id,omitempty"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	CreatedBy    *string        `json:"created_by,omitempty"`
	DeletedBy    *string        `json:"deleted_by,omitempty"`
}

type RuleResponse struct {
	ID          string          `json:"id"`
	Description string          `json:"description,omitempty"`
	Expression  string          `json:"expression"`
	Rollout     RolloutResponse `json:"rollout"`
	Value       any             `json:"value"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	DeletedBy   *string         `json:"deleted_by,omitempty"`
}

type RolloutResponse struct {
	Percentage int    `json:"percentage"`
	BucketBy   string `json:"bucket_by,omitempty"`
}

type CreateFlagRequest struct {
	Key          string              `json:"key"`
	Description  string              `json:"description"`
	Type         string              `json:"type"`
	Enabled      bool                `json:"enabled"`
	Rules        []CreateRuleRequest `json:"rules"`
	DefaultValue json.RawMessage     `json:"default_value"`
	ContextID    *string             `json:"context_id,omitempty"`
}

type CreateRuleRequest struct {
	Description string          `json:"description"`
	Expression  string          `json:"expression"`
	Rollout     RolloutResponse `json:"rollout"`
	Value       json.RawMessage `json:"value"`
}

type UpdateRuleRequest struct {
	Description string          `json:"description"`
	Expression  string          `json:"expression"`
	Rollout     RolloutResponse `json:"rollout"`
	Value       json.RawMessage `json:"value"`
}

type ReorderRulesRequest struct {
	RuleIDs []string `json:"rule_ids"`
}

type ContextFieldDTO struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type ContextResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Fields      []ContextFieldDTO `json:"fields"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	CreatedBy   *string           `json:"created_by,omitempty"`
	DeletedBy   *string           `json:"deleted_by,omitempty"`
}

type CreateContextRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []ContextFieldDTO `json:"fields"`
}

type UpdateContextRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []ContextFieldDTO `json:"fields"`
}

type EvalRequest struct {
	Context map[string]any `json:"context"`
}

type EvalResponse struct {
	Key       string `json:"key"`
	ValueType string `json:"value_type"`
	Value     any    `json:"value"`
}

type EvalTraceResponse struct {
	Key          string                   `json:"key"`
	ValueType    string                   `json:"value_type"`
	Enabled      bool                     `json:"enabled"`
	DefaultValue any                      `json:"default_value"`
	Value        any                      `json:"value"`
	Reason       string                   `json:"reason"`
	Error        string                   `json:"error,omitempty"`
	MatchedRule  *EvalMatchedRuleResponse `json:"matched_rule,omitempty"`
	Bucket       *EvalBucketResponse      `json:"bucket,omitempty"`
	RuleResults  []EvalRuleResultResponse `json:"rule_results"`
}

type EvalMatchedRuleResponse struct {
	ID         string `json:"id"`
	Index      int    `json:"index"`
	Expression string `json:"expression"`
	Value      any    `json:"value"`
}

type EvalRuleResultResponse struct {
	ID         string `json:"id"`
	Index      int    `json:"index"`
	Expression string `json:"expression"`
	Value      any    `json:"value"`
	Matched    bool   `json:"matched"`
	Error      string `json:"error,omitempty"`
}

type EvalBucketResponse struct {
	BucketBy     string `json:"bucket_by"`
	BucketValue  string `json:"bucket_value,omitempty"`
	BucketNumber *int   `json:"bucket_number,omitempty"`
	Percentage   int    `json:"percentage"`
	InRollout    bool   `json:"in_rollout"`
	Missing      bool   `json:"missing"`
}

type EvalAllResponse struct {
	Flags map[string]EvalFlagValueResponse `json:"flags"`
}

type EvalFlagValueResponse struct {
	ValueType string `json:"value_type"`
	Value     any    `json:"value"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Admin       bool    `json:"admin"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	CreatedBy   *string `json:"created_by,omitempty"`
	DeletedBy   *string `json:"deleted_by,omitempty"`
}

type AuthMeResponse struct {
	AuthEnabled   bool          `json:"auth_enabled"`
	Mode          string        `json:"mode,omitempty"`
	Authenticated bool          `json:"authenticated"`
	User          *UserResponse `json:"user,omitempty"`
}

type PasswordLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type APIKeyResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Prefix      string  `json:"prefix"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
	RevokedAt   *string `json:"revoked_at,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	DeletedBy   *string `json:"deleted_by,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateAPIKeyResponse struct {
	APIKeyResponse
	Token string `json:"token"`
}
