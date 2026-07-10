package v1

import (
	"testing"

	"github.com/picunada/flagcel/internal/core"
)

func TestContextReferenceErrors(t *testing.T) {
	inUse := toAPIError(core.ErrContextHasReferers)
	if inUse.Status != 409 || inUse.Code != "CONTEXT_IN_USE" {
		t.Fatalf("in-use error = %#v", inUse)
	}

	conflict := toAPIError(&core.ContextSchemaConflictError{Issues: []core.ValidationIssue{{
		Code: core.ValidationIssueUnknownField, Path: "user.plan", Message: "referenced",
	}}})
	if conflict.Status != 409 || conflict.Code != "CONTEXT_SCHEMA_CONFLICT" || len(conflict.Details) != 1 {
		t.Fatalf("schema conflict error = %#v", conflict)
	}
}
