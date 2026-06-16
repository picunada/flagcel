package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type RulesHandler struct {
	service      *service.RuleService
	environments *service.EnvironmentService
}

func NewRulesHandler(service *service.RuleService, environments *service.EnvironmentService) *RulesHandler {
	return &RulesHandler{service: service, environments: environments}
}

func (h *RulesHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /flags/{key}/rules", h.ListRules)
	mux.HandleFunc("POST /flags/{key}/rules", h.CreateRule)
	mux.HandleFunc("POST /flags/{key}/rules/reorder", h.ReorderRules)
	mux.HandleFunc("GET /flags/{key}/rules/{id}", h.GetRule)
	mux.HandleFunc("PUT /flags/{key}/rules/{id}", h.UpdateRule)
	mux.HandleFunc("DELETE /flags/{key}/rules/{id}", h.DeleteRule)
	mux.HandleFunc("GET /environments/{environment_id}/flags/{key}/rules", h.ListRules)
	mux.HandleFunc("POST /environments/{environment_id}/flags/{key}/rules", h.CreateRule)
	mux.HandleFunc("POST /environments/{environment_id}/flags/{key}/rules/reorder", h.ReorderRules)
	mux.HandleFunc("GET /environments/{environment_id}/flags/{key}/rules/{id}", h.GetRule)
	mux.HandleFunc("PUT /environments/{environment_id}/flags/{key}/rules/{id}", h.UpdateRule)
	mux.HandleFunc("DELETE /environments/{environment_id}/flags/{key}/rules/{id}", h.DeleteRule)
}

func (h *RulesHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	rules, err := h.service.ListRules(r.Context(), environmentID, flagKey)
	if err != nil {
		WriteError(w, err)
		return
	}

	out := make([]RuleResponse, len(rules))
	for i, rule := range rules {
		out[i] = toRuleResponse(rule)
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", out); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *RulesHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	ruleID := r.PathValue("id")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	rule, err := h.service.GetRule(r.Context(), environmentID, flagKey, ruleID)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toRuleResponse(*rule)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *RulesHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	req, err := utils.Decode[CreateRuleRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}

	coreRule, err := toCoreRule(req)
	if err != nil {
		WriteError(w, InvalidRequest(err.Error()))
		return
	}

	rule, err := h.service.CreateRule(r.Context(), environmentID, flagKey, coreRule)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toRuleResponse(*rule)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *RulesHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	ruleID := r.PathValue("id")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	req, err := utils.Decode[UpdateRuleRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}

	value, err := decodeValue(req.Value, true)
	if err != nil {
		WriteError(w, InvalidRequest("value: "+err.Error()))
		return
	}

	rule := core.Rule{
		ID:          ruleID,
		Description: req.Description,
		Expression:  req.Expression,
		Rollout:     toCoreRollout(req.Rollout),
		Value:       value,
	}

	saved, err := h.service.UpdateRule(r.Context(), environmentID, flagKey, rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := utils.Encode(w, r, http.StatusOK, "success", toRuleResponse(*saved)); err != nil {
		WriteError(w, err)
		return
	}
}

func (h *RulesHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	ruleID := r.PathValue("id")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.service.DeleteRule(r.Context(), environmentID, flagKey, ruleID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) ReorderRules(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")
	environmentID, err := h.environmentID(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	req, err := utils.Decode[ReorderRulesRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}

	if err := h.service.ReorderRules(r.Context(), environmentID, flagKey, req.RuleIDs); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) environmentID(r *http.Request) (string, error) {
	if id := r.PathValue("environment_id"); id != "" {
		return id, nil
	}
	env, err := h.environments.Default(r.Context())
	if err != nil {
		return "", err
	}
	return env.ID, nil
}
