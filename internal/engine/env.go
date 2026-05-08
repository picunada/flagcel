package engine

import (
	"github.com/google/cel-go/cel"
)

func NewCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("user", cel.MapType(cel.StringType, cel.DynType)),
		cel.HomogeneousAggregateLiterals(),
	)
}
