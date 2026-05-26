// In package engine — what evaluation operates on.
package engine

import "github.com/google/cel-go/cel"

// In memory highly accessed types
type Flag struct {
	Key          string
	Enabled      bool
	Rules        []CompiledRule // pre-compiled, ready to evaluate
	DefaultValue bool
}

type CompiledRule struct {
	ID      string      // persisted rule id, used for diagnostics
	Source  string      //  for diagnostics/logging
	Program cel.Program // exported — accessed by Engine.Evaluate
	Rollout Rollout
}

type Rollout struct {
	Percentage int
	BucketBy   string
}
