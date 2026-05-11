package v1

import "net/http"

type Handlers struct {
	Flags *FlagsHandler
}

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()
	h.Flags.Register(mux)
	return mux
}
