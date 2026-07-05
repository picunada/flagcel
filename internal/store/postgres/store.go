package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
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

func (s *Store) NotifyAPIKeyCacheInvalidated(ctx context.Context, payload string) error {
	_, err := s.pool.Exec(ctx, "SELECT pg_notify('flagcel_api_key_cache', $1)", payload)
	return err
}

func (s *Store) ListEnvironments(ctx context.Context) ([]*core.Environment, error) {
	rows, err := s.q.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*core.Environment, 0, len(rows))
	for _, row := range rows {
		out = append(out, environmentRowToCore(row))
	}
	return out, nil
}

func (s *Store) GetEnvironment(ctx context.Context, id string) (*core.Environment, error) {
	uid, err := stringToUUID(id)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	row, err := s.q.GetEnvironment(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrEnvironmentNotFound
		}
		return nil, err
	}
	return environmentRowToCore(row), nil
}

func (s *Store) GetEnvironmentByKey(ctx context.Context, key string) (*core.Environment, error) {
	row, err := s.q.GetEnvironmentByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrEnvironmentNotFound
		}
		return nil, err
	}
	return environmentRowToCore(row), nil
}

func (s *Store) CreateEnvironment(ctx context.Context, env *core.Environment) error {
	id, err := stringToUUID(env.ID)
	if err != nil {
		return err
	}
	if err := s.q.InsertEnvironment(ctx, sqlcgen.InsertEnvironmentParams{
		ID:          id,
		Key:         env.Key,
		Name:        env.Name,
		Description: env.Description,
	}); err != nil {
		if isUniqueViolation(err) {
			return core.ErrEnvironmentKeyTaken
		}
		return err
	}
	return nil
}

func (s *Store) UpdateEnvironment(ctx context.Context, env *core.Environment) error {
	id, err := stringToUUID(env.ID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	n, err := s.q.UpdateEnvironment(ctx, sqlcgen.UpdateEnvironmentParams{
		ID:          id,
		Key:         env.Key,
		Name:        env.Name,
		Description: env.Description,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return core.ErrEnvironmentKeyTaken
		}
		return err
	}
	if n == 0 {
		return core.ErrEnvironmentNotFound
	}
	return nil
}

func (s *Store) DeleteEnvironment(ctx context.Context, id string) error {
	uid, err := stringToUUID(id)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	n, err := s.q.DeleteEnvironment(ctx, uid)
	if err != nil {
		if isFKViolation(err) {
			return core.ErrEnvironmentInUse
		}
		return err
	}
	if n == 0 {
		return core.ErrEnvironmentNotFound
	}
	return nil
}

func (s *Store) ListenAPIKeyCacheInvalidations(ctx context.Context, handle func(payload string)) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN flagcel_api_key_cache"); err != nil {
		return err
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		handle(notification.Payload)
	}
}

func (s *Store) GetFlag(ctx context.Context, environmentID, key string) (*core.FlagConfig, error) {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	return loadFlag(ctx, s.q, envID, key)
}

// loadFlag reads a flag together with its rules using the given query handle,
// which may be transactional. It is used both for reads and for building audit
// snapshots inside a mutation's transaction.
func loadFlag(ctx context.Context, q *sqlcgen.Queries, envID pgtype.UUID, key string) (*core.FlagConfig, error) {
	flagRow, err := q.GetFlag(ctx, sqlcgen.GetFlagParams{
		EnvironmentID: envID,
		Key:           key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrFlagNotFound
		}
		return nil, err
	}

	ruleRows, err := q.ListRulesForFlag(ctx, sqlcgen.ListRulesForFlagParams{
		EnvironmentID: envID,
		FlagKey:       key,
	})
	if err != nil {
		return nil, err
	}

	rules := make([]core.Rule, len(ruleRows))
	for i, r := range ruleRows {
		rule, err := ruleRowToCore(r)
		if err != nil {
			return nil, err
		}
		rules[i] = rule
	}
	defaultValue, err := unmarshalJSONValue(flagRow.DefaultValue)
	if err != nil {
		return nil, fmt.Errorf("decode flag default_value: %w", err)
	}

	return &core.FlagConfig{
		Key:          flagRow.Key,
		Description:  flagRow.Description,
		Type:         core.ValueType(flagRow.ValueType),
		Enabled:      flagRow.Enabled,
		DefaultValue: defaultValue,
		Rules:        rules,
		ContextID:    uuidToStringPtr(flagRow.ContextID),
		CreatedAt:    timestamptzToTime(flagRow.CreatedAt),
		UpdatedAt:    timestamptzToTime(flagRow.UpdatedAt),
		CreatedBy:    uuidToStringPtr(flagRow.CreatedBy),
		UpdatedBy:    uuidToStringPtr(flagRow.UpdatedBy),
		DeletedBy:    uuidToStringPtr(flagRow.DeletedBy),
	}, nil
}

func (s *Store) ListFlags(ctx context.Context, environmentID string) ([]*core.FlagConfig, error) {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	flagRows, err := s.q.ListFlags(ctx, envID)
	if err != nil {
		return nil, err
	}

	ruleRows, err := s.q.ListAllRules(ctx, envID)
	if err != nil {
		return nil, err
	}

	rulesByFlag := make(map[string][]core.Rule, len(flagRows))
	for _, r := range ruleRows {
		rule, err := ruleRowToCore(r)
		if err != nil {
			return nil, err
		}
		rulesByFlag[r.FlagKey] = append(rulesByFlag[r.FlagKey], rule)
	}

	flags := make([]*core.FlagConfig, len(flagRows))
	for i, f := range flagRows {
		defaultValue, err := unmarshalJSONValue(f.DefaultValue)
		if err != nil {
			return nil, fmt.Errorf("decode flag default_value: %w", err)
		}
		flags[i] = &core.FlagConfig{
			Key:          f.Key,
			Description:  f.Description,
			Type:         core.ValueType(f.ValueType),
			Enabled:      f.Enabled,
			DefaultValue: defaultValue,
			Rules:        rulesByFlag[f.Key],
			ContextID:    uuidToStringPtr(f.ContextID),
			CreatedAt:    timestamptzToTime(f.CreatedAt),
			UpdatedAt:    timestamptzToTime(f.UpdatedAt),
			CreatedBy:    uuidToStringPtr(f.CreatedBy),
			UpdatedBy:    uuidToStringPtr(f.UpdatedBy),
			DeletedBy:    uuidToStringPtr(f.DeletedBy),
		}
	}
	return flags, nil
}

func (s *Store) SaveFlag(ctx context.Context, environmentID string, flag *core.FlagConfig) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	contextID, err := stringPtrToUUID(flag.ContextID)
	if err != nil {
		return err
	}
	defaultValue, err := marshalJSONValue(flag.DefaultValue)
	if err != nil {
		return err
	}

	actor := actorUUID(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	action := core.AuditActionCreated
	if _, err := qtx.GetFlag(ctx, sqlcgen.GetFlagParams{EnvironmentID: envID, Key: flag.Key}); err == nil {
		action = core.AuditActionUpdated
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if err := qtx.UpsertFlag(ctx, sqlcgen.UpsertFlagParams{
		EnvironmentID: envID,
		Key:           flag.Key,
		ValueType:     string(flag.Type),
		Enabled:       flag.Enabled,
		DefaultValue:  defaultValue,
		ContextID:     contextID,
		Description:   flag.Description,
		CreatedBy:     actor,
		UpdatedBy:     actor,
	}); err != nil {
		if isFKViolation(err) {
			if fkConstraintName(err) == "flags_environment_id_fkey" {
				return core.ErrEnvironmentNotFound
			}
			return core.ErrContextNotFound
		}
		return err
	}

	if err := qtx.DeleteRulesForFlag(ctx, sqlcgen.DeleteRulesForFlagParams{
		EnvironmentID: envID,
		FlagKey:       flag.Key,
	}); err != nil {
		return err
	}

	for i, r := range flag.Rules {
		value, err := marshalJSONValue(r.Value)
		if err != nil {
			return err
		}
		if err := qtx.InsertRule(ctx, sqlcgen.InsertRuleParams{
			ID:                r.ID,
			EnvironmentID:     envID,
			FlagKey:           flag.Key,
			Expression:        r.Expression,
			RolloutPercentage: int32(r.Rollout.Percentage),
			RolloutBucketBy:   r.Rollout.BucketBy,
			Position:          int32(i),
			Value:             value,
			Description:       r.Description,
			CreatedBy:         actor,
			UpdatedBy:         actor,
		}); err != nil {
			return err
		}
	}

	if err := s.recordFlagAudit(ctx, qtx, envID, flag.Key, action); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) DeleteFlag(ctx context.Context, environmentID, key string) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	// Confirm the flag exists before recording a deletion. Mirrors the prior
	// idempotent behaviour: deleting a missing flag is a no-op.
	if _, err := qtx.GetFlag(ctx, sqlcgen.GetFlagParams{EnvironmentID: envID, Key: key}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	if err := qtx.DeleteFlag(ctx, sqlcgen.DeleteFlagParams{
		EnvironmentID: envID,
		Key:           key,
	}); err != nil {
		return err
	}

	if err := s.recordFlagDeletion(ctx, qtx, envID, key); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) GetRule(ctx context.Context, environmentID, flagKey, ruleID string) (*core.Rule, error) {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	row, err := s.q.GetRule(ctx, sqlcgen.GetRuleParams{
		EnvironmentID: envID,
		FlagKey:       flagKey,
		ID:            ruleID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrRuleNotFound
		}
		return nil, err
	}
	rule, err := ruleRowToCore(row)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *Store) CreateRule(ctx context.Context, environmentID, flagKey string, rule core.Rule) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	value, err := marshalJSONValue(rule.Value)
	if err != nil {
		return err
	}
	actor := actorUUID(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	if err := qtx.InsertRuleAtEnd(ctx, sqlcgen.InsertRuleAtEndParams{
		ID:                rule.ID,
		EnvironmentID:     envID,
		FlagKey:           flagKey,
		Expression:        rule.Expression,
		RolloutPercentage: int32(rule.Rollout.Percentage),
		RolloutBucketBy:   rule.Rollout.BucketBy,
		Value:             value,
		Description:       rule.Description,
		CreatedBy:         actor,
		UpdatedBy:         actor,
	}); err != nil {
		return err
	}
	if err := s.touchFlagWithQueries(ctx, qtx, envID, flagKey, actor); err != nil {
		return err
	}
	if err := s.recordFlagAudit(ctx, qtx, envID, flagKey, core.AuditActionUpdated); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateRule(ctx context.Context, environmentID, flagKey string, rule core.Rule) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	value, err := marshalJSONValue(rule.Value)
	if err != nil {
		return err
	}
	actor := actorUUID(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	n, err := qtx.UpdateRule(ctx, sqlcgen.UpdateRuleParams{
		EnvironmentID:     envID,
		FlagKey:           flagKey,
		ID:                rule.ID,
		Expression:        rule.Expression,
		RolloutPercentage: int32(rule.Rollout.Percentage),
		RolloutBucketBy:   rule.Rollout.BucketBy,
		Value:             value,
		Description:       rule.Description,
		UpdatedBy:         actor,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrRuleNotFound
	}
	if err := s.touchFlagWithQueries(ctx, qtx, envID, flagKey, actor); err != nil {
		return err
	}
	if err := s.recordFlagAudit(ctx, qtx, envID, flagKey, core.AuditActionUpdated); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) DeleteRule(ctx context.Context, environmentID, flagKey, ruleID string) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	actor := actorUUID(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	n, err := qtx.DeleteRule(ctx, sqlcgen.DeleteRuleParams{
		EnvironmentID: envID,
		FlagKey:       flagKey,
		ID:            ruleID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrRuleNotFound
	}
	if err := s.touchFlagWithQueries(ctx, qtx, envID, flagKey, actor); err != nil {
		return err
	}
	if err := s.recordFlagAudit(ctx, qtx, envID, flagKey, core.AuditActionUpdated); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ReorderRules(ctx context.Context, environmentID, flagKey string, ruleIDs []string) error {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	actor := actorUUID(ctx)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	for i, id := range ruleIDs {
		n, err := qtx.SetRulePosition(ctx, sqlcgen.SetRulePositionParams{
			EnvironmentID: envID,
			FlagKey:       flagKey,
			ID:            id,
			Position:      int32(i),
			UpdatedBy:     actor,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return core.ErrRuleNotFound
		}
	}

	if err := s.touchFlagWithQueries(ctx, qtx, envID, flagKey, actor); err != nil {
		return err
	}
	if err := s.recordFlagAudit(ctx, qtx, envID, flagKey, core.AuditActionUpdated); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) touchFlagWithQueries(ctx context.Context, q *sqlcgen.Queries, environmentID pgtype.UUID, flagKey string, updatedBy pgtype.UUID) error {
	n, err := q.TouchFlag(ctx, sqlcgen.TouchFlagParams{
		EnvironmentID: environmentID,
		Key:           flagKey,
		UpdatedBy:     updatedBy,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrFlagNotFound
	}
	return nil
}

// recordFlagAudit appends a new version to the flag's change history, snapshotting
// the flag's full config (including rules) as it exists after the change. It must
// be called inside the same transaction as the change so the two commit atomically.
func (s *Store) recordFlagAudit(ctx context.Context, q *sqlcgen.Queries, environmentID pgtype.UUID, flagKey, action string) error {
	snapshot, err := loadFlag(ctx, q, environmentID, flagKey)
	if err != nil {
		return err
	}
	return insertAudit(ctx, q, environmentID, flagKey, action, snapshot)
}

// recordFlagDeletion records a deletion with a nil snapshot; the prior state is
// the previous version's snapshot.
func (s *Store) recordFlagDeletion(ctx context.Context, q *sqlcgen.Queries, environmentID pgtype.UUID, flagKey string) error {
	return insertAudit(ctx, q, environmentID, flagKey, core.AuditActionDeleted, nil)
}

func insertAudit(ctx context.Context, q *sqlcgen.Queries, environmentID pgtype.UUID, flagKey, action string, snapshot *core.FlagConfig) error {
	id, err := stringToUUID(uuid.NewString())
	if err != nil {
		return err
	}
	var snap []byte
	if snapshot != nil {
		snap, err = marshalJSONValue(snapshot)
		if err != nil {
			return err
		}
	}
	_, err = q.InsertAuditLog(ctx, sqlcgen.InsertAuditLogParams{
		ID:            id,
		EnvironmentID: environmentID,
		ResourceType:  core.ResourceTypeFlag,
		ResourceID:    flagKey,
		Action:        action,
		Snapshot:      snap,
		ActorID:       actorUUID(ctx),
		ActorLabel:    actorLabel(ctx),
	})
	return err
}

func (s *Store) ListFlagAuditLog(ctx context.Context, environmentID, flagKey string) ([]*core.AuditEntry, error) {
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	rows, err := s.q.ListFlagAuditLog(ctx, sqlcgen.ListFlagAuditLogParams{
		EnvironmentID: envID,
		ResourceID:    flagKey,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*core.AuditEntry, 0, len(rows))
	for _, r := range rows {
		entry, err := auditRowToCore(r)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}

func (s *Store) ListContexts(ctx context.Context) ([]*core.ContextSchema, error) {
	rows, err := s.q.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*core.ContextSchema, 0, len(rows))
	for _, r := range rows {
		c, err := contextRowToCore(r.ID, r.Name, r.Description, r.Fields, r.CreatedAt, r.UpdatedAt, r.CreatedBy, r.DeletedBy)
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
	return contextRowToCore(row.ID, row.Name, row.Description, row.Fields, row.CreatedAt, row.UpdatedAt, row.CreatedBy, row.DeletedBy)
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

func (s *Store) UpsertUserByOIDC(ctx context.Context, user *core.User) (*core.User, error) {
	uid, err := stringToUUID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	row, err := s.q.UpsertUserByOIDC(ctx, sqlcgen.UpsertUserByOIDCParams{
		ID:          uid,
		OidcSubject: user.OIDCSubject,
		Email:       user.Email,
		Name:        user.Name,
		Admin:       user.Admin,
	})
	if err != nil {
		return nil, err
	}
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Description, row.Admin, row.CreatedAt, row.UpdatedAt, row.CreatedBy, row.DeletedBy), nil
}

func (s *Store) UpsertLocalAdmin(ctx context.Context, user *core.User, passwordHash string) (*core.User, error) {
	uid, err := stringToUUID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	row, err := s.q.UpsertLocalAdmin(ctx, sqlcgen.UpsertLocalAdminParams{
		ID:           uid,
		OidcSubject:  user.OIDCSubject,
		Email:        user.Email,
		Name:         user.Name,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Description, row.Admin, row.CreatedAt, row.UpdatedAt, row.CreatedBy, row.DeletedBy), nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*core.User, string, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", core.ErrInvalidCredentials
		}
		return nil, "", err
	}
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Description, row.Admin, row.CreatedAt, row.UpdatedAt, row.CreatedBy, row.DeletedBy), row.PasswordHash, nil
}

func (s *Store) CreateSession(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time) error {
	sessionID, err := stringToUUID(id)
	if err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}
	uid, err := stringToUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	return s.q.CreateSession(ctx, sqlcgen.CreateSessionParams{
		ID:        sessionID,
		UserID:    uid,
		TokenHash: tokenHash,
		ExpiresAt: timeToTimestamptz(expiresAt),
	})
}

func (s *Store) GetUserBySessionHash(ctx context.Context, tokenHash string) (*core.User, error) {
	row, err := s.q.GetUserBySessionHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrSessionNotFound
		}
		return nil, err
	}
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Description, row.Admin, row.CreatedAt, row.UpdatedAt, row.CreatedBy, row.DeletedBy), nil
}

func (s *Store) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	return s.q.DeleteSessionByHash(ctx, tokenHash)
}

func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	return s.q.DeleteExpiredSessions(ctx)
}

func (s *Store) CreateAPIKey(ctx context.Context, id, name, description, prefix, secretHash, environmentID string) (*core.APIKey, error) {
	keyID, err := stringToUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid api key id: %w", err)
	}
	envID, err := stringToUUID(environmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	row, err := s.q.CreateAPIKey(ctx, sqlcgen.CreateAPIKeyParams{
		ID:            keyID,
		Name:          name,
		Description:   description,
		Prefix:        prefix,
		SecretHash:    secretHash,
		EnvironmentID: envID,
	})
	if err != nil {
		if isFKViolation(err) {
			return nil, core.ErrEnvironmentNotFound
		}
		return nil, err
	}
	return apiKeyRowToCore(row.ID, row.Name, row.Description, row.Prefix, row.EnvironmentID, row.CreatedAt, row.UpdatedAt, row.LastUsedAt, row.RevokedAt, row.CreatedBy, row.DeletedBy), nil
}

func (s *Store) ListAPIKeys(ctx context.Context) ([]*core.APIKey, error) {
	rows, err := s.q.ListAPIKeys(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]*core.APIKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, apiKeyRowToCore(row.ID, row.Name, row.Description, row.Prefix, row.EnvironmentID, row.CreatedAt, row.UpdatedAt, row.LastUsedAt, row.RevokedAt, row.CreatedBy, row.DeletedBy))
	}
	return keys, nil
}

func (s *Store) GetAPIKeyByHash(ctx context.Context, secretHash string) (*core.APIKey, error) {
	row, err := s.q.GetActiveAPIKeyByHash(ctx, secretHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return apiKeyRowToCore(row.ID, row.Name, row.Description, row.Prefix, row.EnvironmentID, row.CreatedAt, row.UpdatedAt, row.LastUsedAt, row.RevokedAt, row.CreatedBy, row.DeletedBy), nil
}

func (s *Store) RevokeAPIKey(ctx context.Context, id string) error {
	keyID, err := stringToUUID(id)
	if err != nil {
		return core.ErrAPIKeyNotFound
	}
	n, err := s.q.RevokeAPIKey(ctx, keyID)
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrAPIKeyNotFound
	}
	return nil
}

func (s *Store) TouchAPIKey(ctx context.Context, id string) error {
	keyID, err := stringToUUID(id)
	if err != nil {
		return core.ErrAPIKeyNotFound
	}
	return s.q.TouchAPIKey(ctx, keyID)
}

func (s *Store) RecordFlagUsage(ctx context.Context, event core.FlagUsageEvent) error {
	envID, err := stringToUUID(event.EnvironmentID)
	if err != nil {
		return core.ErrEnvironmentNotFound
	}
	value, err := marshalJSONValue(event.Value)
	if err != nil {
		return err
	}
	contextValue, err := marshalUsageContext(event.Context)
	if err != nil {
		return err
	}
	observedAt := event.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now()
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	apiKeyID := stringFromPtr(event.APIKeyID)
	matchedRuleID := stringFromPtr(event.MatchedRuleID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := s.q.WithTx(tx)

	if err := qtx.UpsertFlagUsageBucket(ctx, sqlcgen.UpsertFlagUsageBucketParams{
		EnvironmentID: envID,
		FlagKey:       event.FlagKey,
		BucketStart:   timeToTimestamptz(observedAt.Truncate(time.Hour)),
		ValueType:     string(event.ValueType),
		ValueKey:      string(value),
		Value:         value,
		Reason:        event.Reason,
		MatchedRuleID: matchedRuleID,
		ApiKeyID:      apiKeyID,
		Source:        event.Source,
		Count:         1,
	}); err != nil {
		return err
	}

	if err := qtx.InsertFlagEvaluationEvent(ctx, sqlcgen.InsertFlagEvaluationEventParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		EnvironmentID: envID,
		FlagKey:       event.FlagKey,
		ObservedAt:    timeToTimestamptz(observedAt),
		ValueType:     string(event.ValueType),
		Value:         value,
		Reason:        event.Reason,
		MatchedRuleID: matchedRuleID,
		ApiKeyID:      apiKeyID,
		Source:        event.Source,
		LatencyMs:     float64(event.Latency.Microseconds()) / 1000,
		Context:       contextValue,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) ListFlagUsageBuckets(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageBucket, error) {
	envID, err := stringToUUID(query.EnvironmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	rows, err := s.q.ListFlagUsageBuckets(ctx, sqlcgen.ListFlagUsageBucketsParams{
		EnvironmentID: envID,
		FlagKey:       query.FlagKey,
		BucketStart:   timeToTimestamptz(query.Since),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*core.FlagUsageBucket, 0, len(rows))
	for _, row := range rows {
		bucket, err := usageBucketRowToCore(row)
		if err != nil {
			return nil, err
		}
		out = append(out, bucket)
	}
	return out, nil
}

func (s *Store) ListFlagEvaluationEvents(ctx context.Context, query core.FlagUsageQuery) ([]*core.FlagUsageEvent, error) {
	envID, err := stringToUUID(query.EnvironmentID)
	if err != nil {
		return nil, core.ErrEnvironmentNotFound
	}
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.q.ListFlagEvaluationEvents(ctx, sqlcgen.ListFlagEvaluationEventsParams{
		EnvironmentID: envID,
		FlagKey:       query.FlagKey,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*core.FlagUsageEvent, 0, len(rows))
	for _, row := range rows {
		event, err := usageEventRowToCore(row)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, nil
}

func usageBucketRowToCore(r sqlcgen.ListFlagUsageBucketsRow) (*core.FlagUsageBucket, error) {
	value, err := unmarshalJSONValue(r.Value)
	if err != nil {
		return nil, fmt.Errorf("decode usage bucket value: %w", err)
	}
	return &core.FlagUsageBucket{
		EnvironmentID: uuidToString(r.EnvironmentID),
		FlagKey:       r.FlagKey,
		BucketStart:   timestamptzToTime(r.BucketStart),
		ValueType:     core.ValueType(r.ValueType),
		Value:         value,
		Reason:        r.Reason,
		MatchedRuleID: stringToPtr(r.MatchedRuleID),
		APIKeyID:      stringToPtr(r.ApiKeyID),
		Source:        r.Source,
		Count:         r.Count,
	}, nil
}

func usageEventRowToCore(r sqlcgen.FlagEvaluationEvent) (*core.FlagUsageEvent, error) {
	value, err := unmarshalJSONValue(r.Value)
	if err != nil {
		return nil, fmt.Errorf("decode usage event value: %w", err)
	}
	usageContext, err := unmarshalUsageContext(r.Context)
	if err != nil {
		return nil, fmt.Errorf("decode usage event context: %w", err)
	}
	return &core.FlagUsageEvent{
		ID:            uuidToString(r.ID),
		EnvironmentID: uuidToString(r.EnvironmentID),
		FlagKey:       r.FlagKey,
		ObservedAt:    timestamptzToTime(r.ObservedAt),
		ValueType:     core.ValueType(r.ValueType),
		Value:         value,
		Reason:        r.Reason,
		MatchedRuleID: stringToPtr(r.MatchedRuleID),
		APIKeyID:      stringToPtr(r.ApiKeyID),
		Source:        r.Source,
		Latency:       time.Duration(r.LatencyMs * float64(time.Millisecond)),
		Context:       usageContext,
	}, nil
}

func ruleRowToCore(r sqlcgen.Rule) (core.Rule, error) {
	value, err := unmarshalJSONValue(r.Value)
	if err != nil {
		return core.Rule{}, fmt.Errorf("decode rule value: %w", err)
	}
	return core.Rule{
		ID:          r.ID,
		Description: r.Description,
		Expression:  r.Expression,
		Rollout: core.Rollout{
			Percentage: int(r.RolloutPercentage),
			BucketBy:   r.RolloutBucketBy,
		},
		Value:     value,
		CreatedAt: timestamptzToTime(r.CreatedAt),
		UpdatedAt: timestamptzToTime(r.UpdatedAt),
		CreatedBy: uuidToStringPtr(r.CreatedBy),
		UpdatedBy: uuidToStringPtr(r.UpdatedBy),
		DeletedBy: uuidToStringPtr(r.DeletedBy),
	}, nil
}

func auditRowToCore(r sqlcgen.AuditLog) (*core.AuditEntry, error) {
	var snapshot *core.FlagConfig
	if len(r.Snapshot) > 0 {
		var f core.FlagConfig
		if err := json.Unmarshal(r.Snapshot, &f); err != nil {
			return nil, fmt.Errorf("decode audit snapshot: %w", err)
		}
		snapshot = &f
	}
	return &core.AuditEntry{
		ID:            uuidToString(r.ID),
		EnvironmentID: uuidToString(r.EnvironmentID),
		ResourceType:  r.ResourceType,
		ResourceID:    r.ResourceID,
		Action:        r.Action,
		Version:       int(r.Version),
		Snapshot:      snapshot,
		ActorID:       uuidToStringPtr(r.ActorID),
		ActorLabel:    r.ActorLabel,
		CreatedAt:     timestamptzToTime(r.CreatedAt),
	}, nil
}

// actorUUID returns the current actor's ID as a (possibly invalid/NULL) UUID,
// read from the request context. Invalid when no actor is present (auth disabled).
func actorUUID(ctx context.Context) pgtype.UUID {
	actor := core.ActorFromContext(ctx)
	if actor == nil || strings.TrimSpace(actor.ID) == "" {
		return pgtype.UUID{}
	}
	u, err := stringToUUID(actor.ID)
	if err != nil {
		return pgtype.UUID{}
	}
	return u
}

// actorLabel returns the current actor's display label (e.g. email), captured at
// write time so it survives later user deletion. Nil when no actor is present.
func actorLabel(ctx context.Context) *string {
	actor := core.ActorFromContext(ctx)
	if actor == nil || actor.Label == "" {
		return nil
	}
	label := actor.Label
	return &label
}

func environmentRowToCore(r sqlcgen.Environment) *core.Environment {
	return &core.Environment{
		ID:          uuidToString(r.ID),
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   timestamptzToTime(r.CreatedAt),
		UpdatedAt:   timestamptzToTime(r.UpdatedAt),
		CreatedBy:   uuidToStringPtr(r.CreatedBy),
		DeletedBy:   uuidToStringPtr(r.DeletedBy),
	}
}

func contextRowToCore(id pgtype.UUID, name, description string, raw []byte, createdAt, updatedAt pgtype.Timestamptz, createdBy, deletedBy pgtype.UUID) (*core.ContextSchema, error) {
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
		CreatedAt:   timestamptzToTime(createdAt),
		UpdatedAt:   timestamptzToTime(updatedAt),
		CreatedBy:   uuidToStringPtr(createdBy),
		DeletedBy:   uuidToStringPtr(deletedBy),
	}, nil
}

func userRowToCore(id pgtype.UUID, oidcSubject, email, name, description string, admin bool, createdAt, updatedAt pgtype.Timestamptz, createdBy, deletedBy pgtype.UUID) *core.User {
	return &core.User{
		ID:          uuidToString(id),
		OIDCSubject: oidcSubject,
		Email:       email,
		Name:        name,
		Description: description,
		Admin:       admin,
		CreatedAt:   timestamptzToTime(createdAt),
		UpdatedAt:   timestamptzToTime(updatedAt),
		CreatedBy:   uuidToStringPtr(createdBy),
		DeletedBy:   uuidToStringPtr(deletedBy),
	}
}

func apiKeyRowToCore(id pgtype.UUID, name, description, prefix string, environmentID pgtype.UUID, createdAt, updatedAt, lastUsedAt, revokedAt pgtype.Timestamptz, createdBy, deletedBy pgtype.UUID) *core.APIKey {
	return &core.APIKey{
		ID:            uuidToString(id),
		Name:          name,
		Description:   description,
		Prefix:        prefix,
		EnvironmentID: uuidToString(environmentID),
		CreatedAt:     timestamptzToTime(createdAt),
		UpdatedAt:     timestamptzToTime(updatedAt),
		LastUsedAt:    timestamptzToTimePtr(lastUsedAt),
		RevokedAt:     timestamptzToTimePtr(revokedAt),
		CreatedBy:     uuidToStringPtr(createdBy),
		DeletedBy:     uuidToStringPtr(deletedBy),
	}
}

func marshalFields(fields []core.ContextField) ([]byte, error) {
	if fields == nil {
		fields = []core.ContextField{}
	}
	return json.Marshal(fields)
}

func marshalJSONValue(value any) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode json value: %w", err)
	}
	return b, nil
}

func marshalUsageContext(value map[string]any) ([]byte, error) {
	if value == nil {
		value = map[string]any{}
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode usage context: %w", err)
	}
	return b, nil
}

func unmarshalUsageContext(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value map[string]any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil {
		value = map[string]any{}
	}
	return value, nil
}

func unmarshalJSONValue(raw []byte) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
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

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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

func timeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func timestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func fkConstraintName(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ""
	}
	return pgErr.ConstraintName
}
