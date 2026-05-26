package engine

import (
	"sort"
	"strings"
	"unicode"

	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/internal/core"
)

func NewCELEnv() (*cel.Env, error) {
	return NewCELEnvForContext(nil)
}

func NewCELEnvForContext(schema *core.ContextSchema) (*cel.Env, error) {
	opts := []cel.EnvOption{
		cel.HomogeneousAggregateLiterals(),
	}

	roots := map[string]struct{}{}

	if schema != nil {
		for _, field := range schema.Fields {
			root := field.Path
			if before, _, ok := strings.Cut(field.Path, "."); ok {
				root = before
			}
			if isCELIdentifier(root) {
				roots[root] = struct{}{}
			}
		}
	}

	names := make([]string, 0, len(roots))
	for root := range roots {
		names = append(names, root)
	}
	sort.Strings(names)
	for _, root := range names {
		opts = append(opts, cel.Variable(root, cel.MapType(cel.StringType, cel.DynType)))
	}

	return cel.NewEnv(opts...)
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
