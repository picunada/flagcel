package flagcel

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/picunada/flagcel/evalcore"
)

const providerName = "flagcel"

// Provider is an OpenFeature FeatureProvider backed by Flagcel definitions.
type Provider struct {
	client       *definitionsHTTPClient
	cache        definitionsCache
	pollInterval time.Duration

	lifecycleMu sync.Mutex
	cancel      context.CancelFunc
	done        chan struct{}
}

var _ openfeature.FeatureProvider = (*Provider)(nil)
var _ openfeature.StateHandler = (*Provider)(nil)

// NewProvider creates a Flagcel OpenFeature provider.
func NewProvider(endpoint, apiKey string, options ...Option) (*Provider, error) {
	cfg := newProviderConfig(options)
	client, err := newDefinitionsHTTPClient(endpoint, apiKey, cfg.httpClient)
	if err != nil {
		return nil, err
	}

	return &Provider{
		client:       client,
		pollInterval: cfg.pollInterval,
	}, nil
}

func (p *Provider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: providerName}
}

func (p *Provider) Hooks() []openfeature.Hook {
	return nil
}

// Init performs one best-effort definitions fetch and starts background polling.
func (p *Provider) Init(openfeature.EvaluationContext) error {
	p.lifecycleMu.Lock()
	defer p.lifecycleMu.Unlock()

	if p.cancel != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan struct{})

	_ = p.refresh(ctx)
	go p.poll(ctx, p.done)
	return nil
}

func (p *Provider) Shutdown() {
	p.lifecycleMu.Lock()
	cancel := p.cancel
	done := p.done
	p.cancel = nil
	p.done = nil
	p.lifecycleMu.Unlock()

	if cancel == nil {
		return
	}
	cancel()
	<-done
}

func (p *Provider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	value, detail := p.evaluate(flag, defaultValue, flatCtx, evalcore.ValueTypeBoolean)
	if typed, ok := value.(bool); ok {
		return openfeature.BoolResolutionDetail{Value: typed, ProviderResolutionDetail: detail}
	}
	return openfeature.BoolResolutionDetail{Value: defaultValue, ProviderResolutionDetail: typeMismatchDetail(flag, "boolean", value)}
}

func (p *Provider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail {
	value, detail := p.evaluate(flag, defaultValue, flatCtx, evalcore.ValueTypeString)
	if typed, ok := value.(string); ok {
		return openfeature.StringResolutionDetail{Value: typed, ProviderResolutionDetail: detail}
	}
	return openfeature.StringResolutionDetail{Value: defaultValue, ProviderResolutionDetail: typeMismatchDetail(flag, "string", value)}
}

func (p *Provider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail {
	value, detail := p.evaluate(flag, defaultValue, flatCtx, evalcore.ValueTypeNumber)
	if typed, ok := asFloat64(value); ok {
		return openfeature.FloatResolutionDetail{Value: typed, ProviderResolutionDetail: detail}
	}
	return openfeature.FloatResolutionDetail{Value: defaultValue, ProviderResolutionDetail: typeMismatchDetail(flag, "number", value)}
}

func (p *Provider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail {
	value, detail := p.evaluate(flag, defaultValue, flatCtx, evalcore.ValueTypeNumber)
	if typed, ok := asInt64(value); ok {
		return openfeature.IntResolutionDetail{Value: typed, ProviderResolutionDetail: detail}
	}
	return openfeature.IntResolutionDetail{Value: defaultValue, ProviderResolutionDetail: typeMismatchDetail(flag, "integer", value)}
}

func (p *Provider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail {
	value, detail := p.evaluate(flag, defaultValue, flatCtx, evalcore.ValueTypeJSON)
	if value == nil {
		return openfeature.InterfaceResolutionDetail{Value: defaultValue, ProviderResolutionDetail: typeMismatchDetail(flag, "object", value)}
	}
	return openfeature.InterfaceResolutionDetail{Value: value, ProviderResolutionDetail: detail}
}

func (p *Provider) evaluate(flag string, defaultValue any, flatCtx openfeature.FlattenedContext, expected evalcore.ValueType) (any, openfeature.ProviderResolutionDetail) {
	evaluator, _, lastErr, ready := p.cache.snapshot()
	if evaluator == nil {
		return defaultValue, defaultDetail(lastErr, ready)
	}

	result := evaluator.Evaluate(flag, evalcore.DataContext(flatCtx))
	if result.Error != "" {
		return defaultValue, errorDetail(result)
	}
	if result.ValueType != expected {
		return defaultValue, typeMismatchDetail(flag, string(expected), result.Value)
	}

	return result.Value, resolutionDetail(result)
}

func (p *Provider) poll(ctx context.Context, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.refresh(ctx)
		}
	}
}

func (p *Provider) refresh(ctx context.Context) error {
	currentETag := p.cache.etagValue()
	result, err := p.client.fetchDefinitions(ctx, currentETag)
	if err != nil {
		p.cache.markError(err)
		return err
	}
	if result.unchanged {
		p.cache.markUnchanged(result.etag)
		return nil
	}
	return p.cache.store(result.definitions, result.etag)
}

func resolutionDetail(result evalcore.EvaluationResult) openfeature.ProviderResolutionDetail {
	return openfeature.ProviderResolutionDetail{
		Reason:  mapReason(result.Reason),
		Variant: result.Variant,
		FlagMetadata: openfeature.FlagMetadata{
			"flagcelReason": result.Reason,
			"valueType":     string(result.ValueType),
		},
	}
}

func errorDetail(result evalcore.EvaluationResult) openfeature.ProviderResolutionDetail {
	if result.Reason == "not_found" {
		return openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: openfeature.NewFlagNotFoundResolutionError(result.Error),
		}
	}
	return openfeature.ProviderResolutionDetail{
		Reason:          openfeature.ErrorReason,
		Variant:         result.Variant,
		ResolutionError: openfeature.NewGeneralResolutionError(result.Error),
	}
}

func defaultDetail(lastErr error, ready bool) openfeature.ProviderResolutionDetail {
	if lastErr == nil {
		return openfeature.ProviderResolutionDetail{Reason: openfeature.DefaultReason}
	}
	if !ready {
		return openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: openfeature.NewProviderNotReadyResolutionError(lastErr.Error()),
		}
	}
	return openfeature.ProviderResolutionDetail{
		Reason:          openfeature.ErrorReason,
		ResolutionError: openfeature.NewGeneralResolutionError(lastErr.Error()),
	}
}

func typeMismatchDetail(flag, want string, got any) openfeature.ProviderResolutionDetail {
	return openfeature.ProviderResolutionDetail{
		Reason:          openfeature.ErrorReason,
		ResolutionError: openfeature.NewTypeMismatchResolutionError(fmt.Sprintf("%s: expected %s, got %T", flag, want, got)),
	}
}

func mapReason(reason string) openfeature.Reason {
	switch reason {
	case "matched_rule":
		return openfeature.TargetingMatchReason
	case "default_no_match":
		return openfeature.DefaultReason
	case "disabled":
		return openfeature.DisabledReason
	case "not_found", "cel_error":
		return openfeature.ErrorReason
	default:
		return openfeature.Reason(reason)
	}
}

func asFloat64(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		if typed > math.MaxInt64 {
			return 0, false
		}
		return float64(typed), true
	default:
		return 0, false
	}
}

func asInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		if uint64(typed) > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		if typed > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case float64:
		if math.Trunc(typed) != typed || typed < math.MinInt64 || typed > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case float32:
		value := float64(typed)
		if math.Trunc(value) != value || value < math.MinInt64 || value > math.MaxInt64 {
			return 0, false
		}
		return int64(value), true
	default:
		return 0, false
	}
}
