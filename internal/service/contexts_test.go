package service

import (
	"reflect"
	"testing"

	"github.com/picunada/flagcel/internal/core"
)

func TestReferencedFieldsCountsDistinctRulesAndBucketFields(t *testing.T) {
	schema := &core.ContextSchema{
		Name: "web-user",
		Fields: []core.ContextField{
			{Path: "user.id", Type: core.ContextTypeString},
			{Path: "user.plan", Type: core.ContextTypeString},
			{Path: "attributes", Type: core.ContextTypeMap},
			{Path: "targetingKey", Type: core.ContextTypeString},
		},
	}
	flag := core.FlagConfig{Rules: []core.Rule{
		{Expression: `user.id == "a" || user.id == "b"`, Rollout: core.Rollout{BucketBy: "user.plan"}},
		{Expression: `attributes["region"] == "west" && targetingKey != ""`},
	}}

	got := referencedFields(flag, schema)
	want := []core.ContextFieldReference{
		{Path: "attributes", RuleCount: 1},
		{Path: "targetingKey", RuleCount: 1},
		{Path: "user.id", RuleCount: 1},
		{Path: "user.plan", RuleCount: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("referencedFields() = %#v, want %#v", got, want)
	}
}

func TestReferencedSchemaField(t *testing.T) {
	schema := &core.ContextSchema{Fields: []core.ContextField{
		{Path: "user.id", Type: core.ContextTypeString},
		{Path: "attributes", Type: core.ContextTypeMap},
	}}
	for _, test := range []struct {
		path string
		want string
	}{
		{path: "user.id", want: "user.id"},
		{path: "attributes.region", want: "attributes"},
		{path: "user.missing", want: ""},
	} {
		if got := referencedSchemaField(schema, test.path); got != test.want {
			t.Errorf("referencedSchemaField(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}
