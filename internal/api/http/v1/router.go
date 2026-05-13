package v1

import "net/http"

type Handlers struct {
	Flags *FlagsHandler
	Rules *RulesHandler
}

func NewRouter(h *Handlers) http.Handler {
	v1 := http.NewServeMux()
	h.Flags.Register(v1)
	h.Rules.Register(v1)

	root := http.NewServeMux()
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	return root
}
