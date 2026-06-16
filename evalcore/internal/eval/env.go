package eval

import (
	"sort"
	"strings"
	"unicode"

	"github.com/google/cel-go/cel"
)

func NewCELEnv() (*cel.Env, error) {
	return NewCELEnvForContext(nil)
}

func NewCELEnvForContext(schema *ContextSchema) (*cel.Env, error) {
	opts := []cel.EnvOption{
		cel.HomogeneousAggregateLiterals(),
	}

	roots := map[string]*cel.Type{}

	if schema != nil {
		for _, field := range schema.Fields {
			root := field.Path
			if before, _, ok := strings.Cut(field.Path, "."); ok {
				root = before
			}
			if isCELIdentifier(root) {
				if strings.Contains(field.Path, ".") {
					roots[root] = cel.MapType(cel.StringType, cel.DynType)
				} else if _, exists := roots[root]; !exists {
					roots[root] = celTypeForContextType(field.Type)
				}
			}
		}
	}

	names := make([]string, 0, len(roots))
	for root := range roots {
		names = append(names, root)
	}
	sort.Strings(names)
	for _, root := range names {
		opts = append(opts, cel.Variable(root, roots[root]))
	}

	return cel.NewEnv(opts...)
}

func celTypeForContextType(contextType ContextType) *cel.Type {
	switch contextType {
	case ContextTypeString, ContextTypeTimestamp:
		return cel.StringType
	case ContextTypeInt:
		return cel.IntType
	case ContextTypeDouble:
		return cel.DoubleType
	case ContextTypeBool:
		return cel.BoolType
	case ContextTypeList:
		return cel.ListType(cel.DynType)
	case ContextTypeMap:
		return cel.MapType(cel.StringType, cel.DynType)
	default:
		return cel.DynType
	}
}

func isCELIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
