package v1

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/picunada/flagcel/internal/api/http/utils"
	"github.com/picunada/flagcel/internal/core"
)

type UsageStore interface {
	RecordFlagUsage(ctx context.Context, event core.FlagUsageEvent) error
	ListFlagUsageBuckets(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageBucket, error)
	ListEnvironmentUsageBuckets(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageBucket, error)
	ListFlagUsageLatencyBuckets(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageLatencyBucket, error)
	ListEnvironmentUsageLatencyBuckets(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageLatencyBucket, error)
	ListFlagEvaluationEvents(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageEvent, error)
}

type UsageHandler struct {
	store UsageStore
}

func NewUsageHandler(store UsageStore) *UsageHandler {
	return &UsageHandler{store: store}
}

func (h *UsageHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /environments/{environment_id}/usage", h.GetEnvironmentUsage)
	mux.HandleFunc("GET /environments/{environment_id}/flags/{key}/usage", h.GetFlagUsage)
}

func (h *UsageHandler) RegisterEval(mux *http.ServeMux) {
	mux.HandleFunc("POST /eval/usage", h.ReportUsage)
}

func (h *UsageHandler) GetFlagUsage(w http.ResponseWriter, r *http.Request) {
	query := usageQueryFromRequest(r, r.PathValue("environment_id"), r.PathValue("key"), time.Now())
	buckets, err := h.store.ListFlagUsageBuckets(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}
	events, err := h.store.ListFlagEvaluationEvents(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}
	latencyBuckets, err := h.store.ListFlagUsageLatencyBuckets(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagUsageResponse(buckets, latencyBuckets, events)); err != nil {
		WriteError(w, err)
	}
}

func (h *UsageHandler) GetEnvironmentUsage(w http.ResponseWriter, r *http.Request) {
	query := usageQueryFromRequest(r, r.PathValue("environment_id"), "", time.Now())
	buckets, err := h.store.ListEnvironmentUsageBuckets(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}
	latencyBuckets, err := h.store.ListEnvironmentUsageLatencyBuckets(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := utils.Encode(w, r, http.StatusOK, "success", toFlagUsageResponse(buckets, latencyBuckets, nil)); err != nil {
		WriteError(w, err)
	}
}

func (h *UsageHandler) ReportUsage(w http.ResponseWriter, r *http.Request) {
	environmentID, err := environmentIDFromAPIKey(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	req, err := utils.Decode[ReportUsageRequest](r)
	if err != nil {
		WriteError(w, InvalidRequest("invalid request body"))
		return
	}
	if len(req.Events) == 0 {
		WriteError(w, InvalidRequest("events are required"))
		return
	}
	key, _ := apiKeyFromRequest(r)
	for _, reported := range req.Events {
		event, err := toCoreUsageEvent(environmentID, key, r.UserAgent(), reported)
		if err != nil {
			WriteError(w, InvalidRequest(err.Error()))
			return
		}
		if err := h.store.RecordFlagUsage(r.Context(), event); err != nil {
			WriteError(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func usageQueryFromRequest(r *http.Request, environmentID, flagKey string, now time.Time) core.FlagUsageQuery {
	hours := 24
	if raw := r.URL.Query().Get("hours"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 24*90 {
			hours = parsed
		}
	}
	limit := int32(50)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = int32(parsed)
		}
	}
	return core.FlagUsageQuery{
		EnvironmentID: environmentID,
		FlagKey:       flagKey,
		Since:         now.Add(-time.Duration(hours) * time.Hour),
		Limit:         limit,
	}
}

func toCoreUsageEvent(environmentID string, key *core.APIKey, userAgent string, reported ReportUsageEventRequest) (core.FlagUsageEvent, error) {
	if reported.FlagKey == "" {
		return core.FlagUsageEvent{}, errUsageReport("flag_key is required")
	}
	if reported.ValueType == "" {
		return core.FlagUsageEvent{}, errUsageReport("value_type is required")
	}
	if reported.Reason == "" {
		return core.FlagUsageEvent{}, errUsageReport("reason is required")
	}
	observedAt := time.Now()
	if reported.ObservedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, reported.ObservedAt)
		if err != nil {
			return core.FlagUsageEvent{}, errUsageReport("observed_at must be RFC3339")
		}
		observedAt = parsed
	}
	source := reported.Source
	if source == "" {
		source = userAgent
	}
	event := core.FlagUsageEvent{
		EnvironmentID: environmentID,
		FlagKey:       reported.FlagKey,
		ValueType:     core.ValueType(reported.ValueType),
		Value:         reported.Value,
		Reason:        reported.Reason,
		MatchedRuleID: reported.MatchedRuleID,
		Source:        source,
		Latency:       time.Duration(reported.LatencyMs * float64(time.Millisecond)),
		ObservedAt:    observedAt,
		Context:       reported.Context,
	}
	if key != nil && key.ID != "" {
		event.APIKeyID = &key.ID
	}
	return event, nil
}

type usageReportError string

func (e usageReportError) Error() string {
	return string(e)
}

func errUsageReport(message string) error {
	return usageReportError(message)
}
