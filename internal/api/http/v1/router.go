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
	Eval     *EvalHandler
	Auth     *AuthHandler
	APIKeys  *APIKeysHandler
}

func NewRouter(h *Handlers) http.Handler {
	admin := http.NewServeMux()
	h.Flags.Register(admin)
	h.Rules.Register(admin)
	h.Contexts.Register(admin)
	h.APIKeys.Register(admin)

	eval := http.NewServeMux()
	h.Eval.Register(eval)

	adminProtected := h.Auth.AdminMiddleware(admin)
	evalProtected := h.Auth.APIKeyMiddleware(eval)

	v1 := http.NewServeMux()
	h.Auth.RegisterAPI(v1)
	v1.Handle("/eval", evalProtected)
	v1.Handle("/eval/", evalProtected)
	for _, prefix := range []string{"/flags", "/flags/", "/rules", "/rules/", "/contexts", "/contexts/", "/api-keys", "/api-keys/"} {
		v1.Handle(prefix, adminProtected)
	}

	root := http.NewServeMux()
	h.Auth.RegisterPublic(root)
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	docs.Register(root)
	root.Handle("/", web.Handler())
	return root
}
