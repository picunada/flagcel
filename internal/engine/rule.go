// In package engine — what evaluation operates on.
package engine

import (
	"github.com/google/cel-go/cel"
	"github.com/picunada/flagcel/internal/core"
)

// In memory highly accessed types
type Flag struct {
	Key          string
	Type         core.ValueType
	Enabled      bool
	Rules        []CompiledRule // pre-compiled, ready to evaluate
	DefaultValue any
}

type CompiledRule struct {
	ID      string      // persisted rule id, used for diagnostics
	Source  string      //  for diagnostics/logging
	Program cel.Program // exported — accessed by Engine.Evaluate
	Rollout Rollout
	Value   any
}

type Rollout struct {
	Percentage int
	BucketBy   string
}
