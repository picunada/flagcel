package core

import "strings"

const (
	ValidationIssueParseError        = "parse_error"
	ValidationIssueUnknownField      = "unknown_field"
	ValidationIssueNonBoolExpression = "non_bool_expression"
	ValidationIssueInvalidRollout    = "invalid_rollout"
	ValidationIssueMissingBucket     = "missing_bucket_field"
	ValidationIssueInvalidValueType  = "invalid_value_type"
	ValidationIssueInvalidValue      = "invalid_value"
)

type ValidationIssue struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type ValidationError struct {
	Message string            `json:"message"`
	Issues  []ValidationIssue `json:"issues"`
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	messages := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		if issue.Message != "" {
			messages = append(messages, issue.Message)
		}
	}
	return strings.Join(messages, "; ")
}
