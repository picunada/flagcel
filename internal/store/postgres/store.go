package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (s *Store) NotifyAPIKeyCacheInvalidated(ctx context.Context, payload string) error {
	_, err := s.pool.Exec(ctx, "SELECT pg_notify('flagcel_api_key_cache', $1)", payload)
	return err
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
		UpdatedAt:    timestamptzToTime(flagRow.UpdatedAt),
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
			UpdatedAt:    timestamptzToTime(f.UpdatedAt),
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
	if err := s.q.InsertRuleAtEnd(ctx, sqlcgen.InsertRuleAtEndParams{
		ID:                rule.ID,
		FlagKey:           flagKey,
		Expression:        rule.Expression,
		RolloutPercentage: int32(rule.Rollout.Percentage),
		RolloutBucketBy:   rule.Rollout.BucketBy,
	}); err != nil {
		return err
	}
	return s.touchFlag(ctx, flagKey)
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
	return s.touchFlag(ctx, flagKey)
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
	return s.touchFlag(ctx, flagKey)
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

	if err := s.touchFlagWithQueries(ctx, qtx, flagKey); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) touchFlag(ctx context.Context, flagKey string) error {
	return s.touchFlagWithQueries(ctx, s.q, flagKey)
}

func (s *Store) touchFlagWithQueries(ctx context.Context, q *sqlcgen.Queries, flagKey string) error {
	n, err := q.TouchFlag(ctx, flagKey)
	if err != nil {
		return err
	}
	if n == 0 {
		return core.ErrFlagNotFound
	}
	return nil
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
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Admin), nil
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
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Admin), nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*core.User, string, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", core.ErrInvalidCredentials
		}
		return nil, "", err
	}
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Admin), row.PasswordHash, nil
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
	return userRowToCore(row.ID, row.OidcSubject, row.Email, row.Name, row.Admin), nil
}

func (s *Store) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	return s.q.DeleteSessionByHash(ctx, tokenHash)
}

func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	return s.q.DeleteExpiredSessions(ctx)
}

func (s *Store) CreateAPIKey(ctx context.Context, id, name, prefix, secretHash string) (*core.APIKey, error) {
	keyID, err := stringToUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid api key id: %w", err)
	}
	row, err := s.q.CreateAPIKey(ctx, sqlcgen.CreateAPIKeyParams{
		ID:         keyID,
		Name:       name,
		Prefix:     prefix,
		SecretHash: secretHash,
	})
	if err != nil {
		return nil, err
	}
	return apiKeyRowToCore(row.ID, row.Name, row.Prefix, row.CreatedAt, row.LastUsedAt, row.RevokedAt), nil
}

func (s *Store) ListAPIKeys(ctx context.Context) ([]*core.APIKey, error) {
	rows, err := s.q.ListAPIKeys(ctx)
	if err != nil {
		return nil, err
	}
	keys := make([]*core.APIKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, apiKeyRowToCore(row.ID, row.Name, row.Prefix, row.CreatedAt, row.LastUsedAt, row.RevokedAt))
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
	return apiKeyRowToCore(row.ID, row.Name, row.Prefix, row.CreatedAt, row.LastUsedAt, row.RevokedAt), nil
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

func userRowToCore(id pgtype.UUID, oidcSubject, email, name string, admin bool) *core.User {
	return &core.User{
		ID:          uuidToString(id),
		OIDCSubject: oidcSubject,
		Email:       email,
		Name:        name,
		Admin:       admin,
	}
}

func apiKeyRowToCore(id pgtype.UUID, name, prefix string, createdAt, lastUsedAt, revokedAt pgtype.Timestamptz) *core.APIKey {
	return &core.APIKey{
		ID:         uuidToString(id),
		Name:       name,
		Prefix:     prefix,
		CreatedAt:  timestamptzToTime(createdAt),
		LastUsedAt: timestamptzToTimePtr(lastUsedAt),
		RevokedAt:  timestamptzToTimePtr(revokedAt),
	}
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
