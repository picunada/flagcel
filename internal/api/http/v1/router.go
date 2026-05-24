package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/docs"
	"github.com/picunada/flagcel/web"
)

type Handlers struct {
	Flags    *FlagsHandler
	Rules    *RulesHandler
	Contexts *ContextsHandler
}

func NewRouter(h *Handlers) http.Handler {
	v1 := http.NewServeMux()
	h.Flags.Register(v1)
	h.Rules.Register(v1)
	h.Contexts.Register(v1)

	root := http.NewServeMux()
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	docs.Register(root)
	root.Handle("/", web.Handler())
	return root
}
