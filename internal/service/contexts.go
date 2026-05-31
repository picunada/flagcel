package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

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
	if err := s.store.DeleteContext(ctx, id); err != nil {
		return fmt.Errorf("context service: failed to delete context %w", err)
	}
	s.invalidate(id)
	return nil
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
