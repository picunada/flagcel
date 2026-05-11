package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/picunada/flagcel/internal/core"
	"github.com/picunada/flagcel/internal/store/postgres/sqlcgen"
)

type Store struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: sqlcgen.New(pool)}
}

func (s *Store) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.pool.Close()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("pgxpool close: %w", ctx.Err())
	}
}

func (s *Store) GetFlag(ctx context.Context, key string) (*core.FlagConfig, error) {
	flagRow, err := s.q.GetFlag(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrFlagNotFound
		}
		return nil, err
	}

	ruleRows, err := s.q.ListRulesForFlag(ctx, key)
	if err != nil {
		return nil, err
	}

	rules := make([]core.Rule, len(ruleRows))
	for i, r := range ruleRows {
		rules[i] = ruleRowToCore(r)
	}

	return &core.FlagConfig{
		Key:          flagRow.Key,
		Enabled:      flagRow.Enabled,
		DefaultValue: flagRow.DefaultValue,
		Rules:        rules,
	}, nil
}

func (s *Store) ListFlags(ctx context.Context) ([]*core.FlagConfig, error) {
	flagRows, err := s.q.ListFlags(ctx)
	if err != nil {
		return nil, err
	}

	ruleRows, err := s.q.ListAllRules(ctx)
	if err != nil {
		return nil, err
	}

	rulesByFlag := make(map[string][]core.Rule, len(flagRows))
	for _, r := range ruleRows {
		rulesByFlag[r.FlagKey] = append(rulesByFlag[r.FlagKey], ruleRowToCore(r))
	}

	flags := make([]*core.FlagConfig, len(flagRows))
	for i, f := range flagRows {
		flags[i] = &core.FlagConfig{
			Key:          f.Key,
			Enabled:      f.Enabled,
			DefaultValue: f.DefaultValue,
			Rules:        rulesByFlag[f.Key],
		}
	}
	return flags, nil
}

func (s *Store) SaveFlag(ctx context.Context, flag *core.FlagConfig) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	if err := qtx.UpsertFlag(ctx, sqlcgen.UpsertFlagParams{
		Key:          flag.Key,
		Enabled:      flag.Enabled,
		DefaultValue: flag.DefaultValue,
	}); err != nil {
		return err
	}

	if err := qtx.DeleteRulesForFlag(ctx, flag.Key); err != nil {
		return err
	}

	for i, r := range flag.Rules {
		if err := qtx.InsertRule(ctx, sqlcgen.InsertRuleParams{
			ID:                r.ID,
			FlagKey:           flag.Key,
			Expression:        r.Expression,
			RolloutPercentage: int32(r.Rollout.Percentage),
			RolloutBucketBy:   r.Rollout.BucketBy,
			Position:          int32(i),
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) DeleteFlag(ctx context.Context, key string) error {
	return s.q.DeleteFlag(ctx, key)
}

func (s *Store) GetRule(ctx context.Context, flagKey, ruleID string) (*core.Rule, error) {
	row, err := s.q.GetRule(ctx, sqlcgen.GetRuleParams{
		FlagKey: flagKey,
		ID:      ruleID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrRuleNotFound
		}
		return nil, err
	}
	rule := ruleRowToCore(row)
	return &rule, nil
}

func (s *Store) CreateRule(ctx context.Context, flagKey string, rule core.Rule) error {
	return s.q.InsertRuleAtEnd(ctx, sqlcgen.InsertRuleAtEndParams{
		ID:                rule.ID,
		FlagKey:           flagKey,
		Expression:        rule.Expression,
		RolloutPercentage: int32(rule.Rollout.Percentage),
		RolloutBucketBy:   rule.Rollout.BucketBy,
	})
}

func (s *Store) UpdateRule(ctx context.Context, flagKey string, rule core.Rule) error {
	n, err := s.q.UpdateRule(ctx, sqlcgen.UpdateRuleParams{
		FlagKey:           flagKey,
		ID:                rule.ID,
		Expression:        rule.Expression,
		RolloutPercentage: int32(rule.Rollout.Percentage),
		RolloutBucketBy:   rule.Rollout.BucketBy,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrRuleNotFound
	}
	return nil
}

func (s *Store) DeleteRule(ctx context.Context, flagKey, ruleID string) error {
	n, err := s.q.DeleteRule(ctx, sqlcgen.DeleteRuleParams{
		FlagKey: flagKey,
		ID:      ruleID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrRuleNotFound
	}
	return nil
}

func (s *Store) ReorderRules(ctx context.Context, flagKey string, ruleIDs []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	for i, id := range ruleIDs {
		n, err := qtx.SetRulePosition(ctx, sqlcgen.SetRulePositionParams{
			FlagKey:  flagKey,
			ID:       id,
			Position: int32(i),
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return core.ErrRuleNotFound
		}
	}

	return tx.Commit(ctx)
}

func ruleRowToCore(r sqlcgen.Rule) core.Rule {
	return core.Rule{
		ID:         r.ID,
		Expression: r.Expression,
		Rollout: core.Rollout{
			Percentage: int(r.RolloutPercentage),
			BucketBy:   r.RolloutBucketBy,
		},
	}
}
