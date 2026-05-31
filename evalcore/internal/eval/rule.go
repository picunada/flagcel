// In package eval — what evaluation operates on.
package eval

import (
	"github.com/google/cel-go/cel"
)

// In memory highly accessed types
type Flag struct {
	Key          string
	Type         ValueType
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
