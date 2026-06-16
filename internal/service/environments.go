package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

var environmentKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type EnvironmentService struct {
	store *postgres.Store
}

func NewEnvironmentService(store *postgres.Store) *EnvironmentService {
	return &EnvironmentService{store: store}
}

func (s *EnvironmentService) List(ctx context.Context) ([]*core.Environment, error) {
	return s.store.ListEnvironments(ctx)
}

func (s *EnvironmentService) Get(ctx context.Context, id string) (*core.Environment, error) {
	return s.store.GetEnvironment(ctx, id)
}

func (s *EnvironmentService) Default(ctx context.Context) (*core.Environment, error) {
	return s.store.GetEnvironmentByKey(ctx, core.DefaultEnvironmentKey)
}

func (s *EnvironmentService) Create(ctx context.Context, env *core.Environment) error {
	normalizeEnvironment(env)
	if env.ID == "" {
		env.ID = uuid.NewString()
	}
	if err := validateEnvironment(env); err != nil {
		return err
	}
	return s.store.CreateEnvironment(ctx, env)
}

func (s *EnvironmentService) Update(ctx context.Context, env *core.Environment) error {
	normalizeEnvironment(env)
	if env.ID == "" {
		return core.ErrEnvironmentNotFound
	}
	if err := validateEnvironment(env); err != nil {
		return err
	}
	existing, err := s.store.GetEnvironment(ctx, env.ID)
	if err != nil {
		return err
	}
	if existing.Key == core.DefaultEnvironmentKey && env.Key != core.DefaultEnvironmentKey {
		return core.ErrDefaultEnvironment
	}
	return s.store.UpdateEnvironment(ctx, env)
}

func (s *EnvironmentService) Delete(ctx context.Context, id string) error {
	env, err := s.store.GetEnvironment(ctx, id)
	if err != nil {
		return err
	}
	if env.Key == core.DefaultEnvironmentKey {
		return core.ErrDefaultEnvironment
	}
	return s.store.DeleteEnvironment(ctx, id)
}

func normalizeEnvironment(env *core.Environment) {
	env.Key = strings.ToLower(strings.TrimSpace(env.Key))
	env.Name = strings.TrimSpace(env.Name)
	env.Description = strings.TrimSpace(env.Description)
	if env.Name == "" {
		env.Name = env.Key
	}
}

func validateEnvironment(env *core.Environment) error {
	if env.Key == "" {
		return fmt.Errorf("key is required")
	}
	if !environmentKeyPattern.MatchString(env.Key) {
		return fmt.Errorf("key must use lowercase letters, numbers, and hyphens")
	}
	if env.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
