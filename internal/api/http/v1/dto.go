package v1

type FlagResponse struct {
	Key          string         `json:"key"`
	Enabled      bool           `json:"enabled"`
	Rules        []RuleResponse `json:"rules"`
	DefaultValue bool           `json:"default_value"`
	ContextID    *string        `json:"context_id,omitempty"`
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
	Key          string              `json:"key"`
	Enabled      bool                `json:"enabled"`
	Rules        []CreateRuleRequest `json:"rules"`
	DefaultValue bool                `json:"default_value"`
	ContextID    *string             `json:"context_id,omitempty"`
}

type CreateRuleRequest struct {
	Expression string          `json:"expression"`
	Rollout    RolloutResponse `json:"rollout"`
}

type UpdateRuleRequest struct {
	Expression string          `json:"expression"`
	Rollout    RolloutResponse `json:"rollout"`
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
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

type EvalAllResponse struct {
	Flags map[string]bool `json:"flags"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Admin bool   `json:"admin"`
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
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	RevokedAt  *string `json:"revoked_at,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	APIKeyResponse
	Token string `json:"token"`
}
