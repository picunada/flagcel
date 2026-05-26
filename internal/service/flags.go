package service

import (
	"context"
	"fmt"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres"
)

type FlagService struct {
	store        *postgres.Store
	onFlagChange func(string)
}

func NewFlagService(store *postgres.Store, onFlagChange ...func(string)) *FlagService {
	s := &FlagService{store: store}
	if len(onFlagChange) > 0 {
		s.onFlagChange = onFlagChange[0]
	}
	return s
}

func (s *FlagService) GetFlags(ctx context.Context) ([]*core.FlagConfig, error) {
	// Fetch flag from store
	flags, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failder to list flags %w",
			err,
		)
	}
	return flags, nil
}

func (s *FlagService) GetFlag(ctx context.Context, key string) (*core.FlagConfig, error) {
	flag, err := s.store.GetFlag(ctx, key)
	if err != nil {
		return nil, fmt.Errorf(
			"flag service: failed to get flag %w",
			err,
		)
	}
	return flag, nil
}

func (s *FlagService) CreateFlag(ctx context.Context, flag *core.FlagConfig) error {
	if err := s.store.SaveFlag(ctx, flag); err != nil {
		return fmt.Errorf(
			"flag service: failed to create flag %w",
			err,
		)
	}
	s.invalidate(flag.Key)
	return nil
}

func (s *FlagService) DeleteFlag(context context.Context, id string) error {
	if err := s.store.DeleteFlag(context, id); err != nil {
		return fmt.Errorf(
			"flag service: failed to delete flag %w",
			err,
		)
	}
	s.invalidate(id)
	return nil
}

func (s *FlagService) invalidate(key string) {
	if s.onFlagChange != nil {
		s.onFlagChange(key)
	}
}
