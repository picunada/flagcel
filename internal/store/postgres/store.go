package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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
		ContextID:    uuidToStringPtr(flagRow.ContextID),
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
			ContextID:    uuidToStringPtr(f.ContextID),
		}
	}
	return flags, nil
}

func (s *Store) SaveFlag(ctx context.Context, flag *core.FlagConfig) error {
	contextID, err := stringPtrToUUID(flag.ContextID)
	if err != nil {
		return err
	}

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
		ContextID:    contextID,
	}); err != nil {
		if isFKViolation(err) {
			return core.ErrContextNotFound
		}
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

func (s *Store) ListContexts(ctx context.Context) ([]*core.ContextSchema, error) {
	rows, err := s.q.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*core.ContextSchema, 0, len(rows))
	for _, r := range rows {
		c, err := contextRowToCore(r.ID, r.Name, r.Description, r.Fields)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) GetContext(ctx context.Context, id string) (*core.ContextSchema, error) {
	uid, err := stringToUUID(id)
	if err != nil {
		return nil, core.ErrContextNotFound
	}
	row, err := s.q.GetContext(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrContextNotFound
		}
		return nil, err
	}
	return contextRowToCore(row.ID, row.Name, row.Description, row.Fields)
}

func (s *Store) CreateContext(ctx context.Context, c *core.ContextSchema) error {
	uid, err := stringToUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid context id: %w", err)
	}
	fields, err := marshalFields(c.Fields)
	if err != nil {
		return err
	}
	if err := s.q.InsertContext(ctx, sqlcgen.InsertContextParams{
		ID:          uid,
		Name:        c.Name,
		Description: c.Description,
		Fields:      fields,
	}); err != nil {
		if isUniqueViolation(err) {
			return core.ErrContextNameTaken
		}
		return err
	}
	return nil
}

func (s *Store) UpdateContext(ctx context.Context, c *core.ContextSchema) error {
	uid, err := stringToUUID(c.ID)
	if err != nil {
		return core.ErrContextNotFound
	}
	fields, err := marshalFields(c.Fields)
	if err != nil {
		return err
	}
	n, err := s.q.UpdateContext(ctx, sqlcgen.UpdateContextParams{
		ID:          uid,
		Name:        c.Name,
		Description: c.Description,
		Fields:      fields,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return core.ErrContextNameTaken
		}
		return err
	}
	if n == 0 {
		return core.ErrContextNotFound
	}
	return nil
}

func (s *Store) DeleteContext(ctx context.Context, id string) error {
	uid, err := stringToUUID(id)
	if err != nil {
		return core.ErrContextNotFound
	}
	n, err := s.q.DeleteContext(ctx, uid)
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrContextNotFound
	}
	return nil
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

func contextRowToCore(id pgtype.UUID, name, description string, raw []byte) (*core.ContextSchema, error) {
	var fields []core.ContextField
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, fmt.Errorf("decode context fields: %w", err)
		}
	}
	if fields == nil {
		fields = []core.ContextField{}
	}
	return &core.ContextSchema{
		ID:          uuidToString(id),
		Name:        name,
		Description: description,
		Fields:      fields,
	}, nil
}

func marshalFields(fields []core.ContextField) ([]byte, error) {
	if fields == nil {
		fields = []core.ContextField{}
	}
	return json.Marshal(fields)
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func uuidToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuidToString(u)
	return &s
}

func stringToUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func stringPtrToUUID(s *string) (pgtype.UUID, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return pgtype.UUID{}, nil
	}
	return stringToUUID(*s)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
