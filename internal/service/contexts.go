package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	celast "github.com/google/cel-go/common/ast"
	"github.com/google/uuid"

	"github.com/picunada/flagcel/evalcore"
	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type ContextService struct {
	store           *postgres.Store
	onContextChange func(string)
}

func NewContextService(store *postgres.Store, onContextChange ...func(string)) *ContextService {
	s := &ContextService{store: store}
	if len(onContextChange) > 0 {
		s.onContextChange = onContextChange[0]
	}
	return s
}

func (s *ContextService) ListContexts(ctx context.Context) ([]*core.ContextSchema, error) {
	out, err := s.store.ListContexts(ctx)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to list contexts %w", err)
	}
	return out, nil
}

func (s *ContextService) GetContext(ctx context.Context, id string) (*core.ContextSchema, error) {
	out, err := s.store.GetContext(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to get context %w", err)
	}
	return out, nil
}

func (s *ContextService) ListReferences(ctx context.Context) ([]core.ContextReference, error) {
	references, err := s.store.ListContextReferences(ctx)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to list context references %w", err)
	}
	contexts, err := s.store.ListContexts(ctx)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to load referenced contexts %w", err)
	}
	byID := make(map[string]*core.ContextSchema, len(contexts))
	for _, schema := range contexts {
		byID[schema.ID] = schema
	}
	for i := range references {
		references[i].ReferencedFields = referencedFields(references[i].Flag, byID[references[i].ContextID])
	}
	return references, nil
}

func (s *ContextService) CreateContext(ctx context.Context, c *core.ContextSchema) (*core.ContextSchema, error) {
	if err := normalize(c); err != nil {
		return nil, err
	}
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if err := s.store.CreateContext(ctx, c); err != nil {
		return nil, fmt.Errorf("context service: failed to create context %w", err)
	}
	s.invalidate(c.ID)
	out, err := s.store.GetContext(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to load created context %w", err)
	}
	return out, nil
}

func (s *ContextService) UpdateContext(ctx context.Context, c *core.ContextSchema) (*core.ContextSchema, error) {
	if c.ID == "" {
		return nil, core.ErrContextNotFound
	}
	if err := normalize(c); err != nil {
		return nil, err
	}
	if err := s.validateReferences(ctx, c); err != nil {
		return nil, err
	}
	if err := s.store.UpdateContext(ctx, c); err != nil {
		return nil, fmt.Errorf("context service: failed to update context %w", err)
	}
	s.invalidate(c.ID)
	out, err := s.store.GetContext(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("context service: failed to load updated context %w", err)
	}
	return out, nil
}

func (s *ContextService) DeleteContext(ctx context.Context, id string) error {
	references, err := s.store.ListContextReferences(ctx)
	if err != nil {
		return fmt.Errorf("context service: failed to check context references %w", err)
	}
	for _, reference := range references {
		if reference.ContextID == id {
			return core.ErrContextHasReferers
		}
	}
	if err := s.store.DeleteContext(ctx, id); err != nil {
		return fmt.Errorf("context service: failed to delete context %w", err)
	}
	s.invalidate(id)
	return nil
}

func (s *ContextService) validateReferences(ctx context.Context, schema *core.ContextSchema) error {
	references, err := s.store.ListContextReferences(ctx)
	if err != nil {
		return fmt.Errorf("context service: failed to validate context references %w", err)
	}
	issues := make([]core.ValidationIssue, 0)
	for _, reference := range references {
		if reference.ContextID != schema.ID {
			continue
		}
		if err := validateFlag(reference.Flag, schema); err != nil {
			validationErr, ok := err.(*core.ValidationError)
			if !ok {
				return err
			}
			for _, issue := range validationErr.Issues {
				issue.Field = reference.EnvironmentKey + "/" + reference.FlagKey + "/" + issue.Field
				issue.Message = fmt.Sprintf("%s / %s: %s", reference.EnvironmentKey, reference.FlagKey, issue.Message)
				issues = append(issues, issue)
			}
		}
	}
	if len(issues) > 0 {
		return &core.ContextSchemaConflictError{Issues: dedupeIssues(issues)}
	}
	return nil
}

func referencedFields(flag core.FlagConfig, schema *core.ContextSchema) []core.ContextFieldReference {
	if schema == nil {
		return nil
	}
	counts := make(map[string]int)
	for _, rule := range flag.Rules {
		paths := make(map[string]struct{})
		if rule.Expression != "" {
			env, err := evalcore.NewCELEnvForContext(schema)
			if err == nil {
				parsed, parseIssues := env.Parse(rule.Expression)
				if parseIssues == nil || parseIssues.Err() == nil {
					celast.PostOrderVisit(parsed.NativeRep().Expr(), celast.NewExprVisitor(func(expr celast.Expr) {
						parts, ok := expressionPathParts(expr)
						if !ok {
							return
						}
						if path := referencedSchemaField(schema, strings.Join(parts, ".")); path != "" {
							paths[path] = struct{}{}
						}
					}))
				}
			}
		}
		if path := referencedSchemaField(schema, rule.Rollout.BucketBy); path != "" {
			paths[path] = struct{}{}
		}
		for path := range paths {
			counts[path]++
		}
	}
	out := make([]core.ContextFieldReference, 0, len(counts))
	for path, ruleCount := range counts {
		out = append(out, core.ContextFieldReference{Path: path, RuleCount: ruleCount})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func referencedSchemaField(schema *core.ContextSchema, path string) string {
	if path == "" {
		return ""
	}
	for _, field := range schema.Fields {
		if field.Path == path {
			return field.Path
		}
		if field.Type == core.ContextTypeMap && strings.HasPrefix(path, field.Path+".") {
			return field.Path
		}
	}
	return ""
}

func (s *ContextService) invalidate(id string) {
	if s.onContextChange != nil {
		s.onContextChange(id)
	}
}

func normalize(c *core.ContextSchema) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	c.Description = strings.TrimSpace(c.Description)
	seen := make(map[string]struct{}, len(c.Fields))
	cleaned := make([]core.ContextField, 0, len(c.Fields))
	for _, f := range c.Fields {
		path := strings.TrimSpace(f.Path)
		if path == "" {
			continue
		}
		if !f.Type.Valid() {
			return fmt.Errorf("field %q has invalid type %q", path, f.Type)
		}
		if _, dup := seen[path]; dup {
			return fmt.Errorf("duplicate field path %q", path)
		}
		seen[path] = struct{}{}
		cleaned = append(cleaned, core.ContextField{Path: path, Type: f.Type})
	}
	c.Fields = cleaned
	return nil
}
