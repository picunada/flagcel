package v1

import (
	"net/http"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/service"
)

type RulesHandler struct {
	service *service.RuleService
}

func NewRulesHandler(service *service.RuleService) *RulesHandler {
	return &RulesHandler{service: service}
}

func (h *RulesHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /flags/{key}/rules", h.ListRules)
	mux.HandleFunc("POST /flags/{key}/rules", h.CreateRule)
	mux.HandleFunc("POST /flags/{key}/rules/reorder", h.ReorderRules)
	mux.HandleFunc("GET /flags/{key}/rules/{id}", h.GetRule)
	mux.HandleFunc("PUT /flags/{key}/rules/{id}", h.UpdateRule)
	mux.HandleFunc("DELETE /flags/{key}/rules/{id}", h.DeleteRule)
}

func (h *RulesHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")

	rules, err := h.service.ListRules(r.Context(), flagKey)
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

	rule, err := h.service.GetRule(r.Context(), flagKey, ruleID)
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

	rule, err := h.service.CreateRule(r.Context(), flagKey, coreRule)
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

	saved, err := h.service.UpdateRule(r.Context(), flagKey, rule)
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

	if err := h.service.DeleteRule(r.Context(), flagKey, ruleID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) ReorderRules(w http.ResponseWriter, r *http.Request) {
	flagKey := r.PathValue("key")

	req, err := utils.Decode[ReorderRulesRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}

	if err := h.service.ReorderRules(r.Context(), flagKey, req.RuleIDs); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
