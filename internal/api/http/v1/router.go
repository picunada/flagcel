package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/docs"
	"github.com/picunada/flagcel/web"
)

type Handlers struct {
	Health       *HealthHandler
	Flags        *FlagsHandler
	Rules        *RulesHandler
	Environments *EnvironmentsHandler
	Contexts     *ContextsHandler
	Eval         *EvalHandler
	Auth         *AuthHandler
	APIKeys      *APIKeysHandler
	Usage        *UsageHandler
}

func NewRouter(h *Handlers) http.Handler {
	admin := http.NewServeMux()
	h.Flags.Register(admin)
	h.Rules.Register(admin)
	h.Environments.Register(admin)
	h.Contexts.Register(admin)
	h.APIKeys.Register(admin)
	h.Eval.RegisterAdmin(admin)
	h.Usage.Register(admin)

	eval := http.NewServeMux()
	h.Eval.Register(eval)
	h.Usage.RegisterEval(eval)

	adminProtected := h.Auth.AdminMiddleware(admin)
	evalProtected := h.Auth.APIKeyMiddleware(eval)

	v1 := http.NewServeMux()
	h.Health.Register(v1)
	h.Auth.RegisterAPI(v1)
	v1.Handle("/eval", evalProtected)
	v1.Handle("/eval/", evalProtected)
	for _, prefix := range []string{"/flags", "/flags/", "/rules", "/rules/", "/environments", "/environments/", "/contexts", "/contexts/", "/api-keys", "/api-keys/"} {
		v1.Handle(prefix, adminProtected)
	}

	root := http.NewServeMux()
	h.Health.Register(root)
	h.Auth.RegisterPublic(root)
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))
	docs.Register(root)
	root.Handle("/", web.Handler())
	return root
}
