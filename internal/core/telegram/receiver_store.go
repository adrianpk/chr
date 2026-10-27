package telegram

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ReceiverStore extends the existing user database with transport receipts.
type ReceiverStore struct {
	db  *sql.DB
	env StateDirectory
}

// ReceiverState is local transport evidence, not workflow completion.
type ReceiverState struct {
	BotID              int64
	Epoch              int64
	NextUpdateID       *int64
	Health             string
	RetryAttempt       int
	RetryCycle         int64
	TotalRetryAttempts int64
	RetryActor         string
	RetryStartedAt     time.Time
	RetentionGapAt     time.Time
	RetryAt            time.Time
	LastPollAt         time.Time
	LastErrorCode      string
}

// NewReceiverStore applies additive migrations and rejects future schemas.
func NewReceiverStore(ctx context.Context, env StoreDirectories, assets fs.FS) (*ReceiverStore, error) {
	status, err := database.NewMigrationStatusChecker(assets).Check(ctx, env)
	if err != nil {
		return nil, err
	}

	if len(status.Unknown) != 0 {
		return nil, errors.New("telegram receiver refuses an unknown database schema")
	}

	if err := createDatabase(env); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", database.UserDatabasePath(env))
	if err != nil {
		return nil, err
	}

	if err := database.Configure(ctx, db); err != nil {
		return nil, database.CloseOnError(db, err)
	}

	if err := database.NewMigrator(assets).Migrate(ctx, db); err != nil {
		return nil, database.CloseOnError(db, err)
	}

	return &ReceiverStore{db: db, env: env}, nil
}

func createDatabase(env database.DataDirectory) error {
	if err := os.MkdirAll(env.DataDir(), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(database.UserDatabasePath(env), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}

	if err := file.Chmod(0o600); err != nil {
		return errors.Join(err, file.Close())
	}

	return file.Close()
}

// Close releases only this store's connection.
func (s *ReceiverStore) Close() error { return s.db.Close() }

// BindReceiver preserves an existing bot's cursor and retry budget.
func (s *ReceiverStore) BindReceiver(ctx context.Context, botID int64, config ReceiverConfig, now time.Time) error {
	if botID <= 0 {
		return errors.New("invalid telegram bot identity")
	}

	if err := config.Validate(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, `INSERT INTO telegram_receivers(bot_id, config_digest, updated_at) VALUES(?,?,?)
ON CONFLICT(bot_id) DO UPDATE SET config_digest=excluded.config_digest, updated_at=excluded.updated_at WHERE schema_version=1`, botID, config.Digest(), database.FormatTime(now))
	if err != nil {
		return err
	}

	if _, err := s.Receiver(ctx, botID); err != nil {
		return err
	}

	return s.restoreInputLanes(ctx, botID)
}

// Receiver loads retained readiness and rejects newer encoded receiver records.
func (s *ReceiverStore) Receiver(ctx context.Context, botID int64) (ReceiverState, error) {
	state := ReceiverState{BotID: botID}

	var (
		version                                       int
		next                                          sql.NullInt64
		retryAt, lastPoll, retryStarted, retentionGap sql.NullString
	)

	err := s.db.QueryRowContext(ctx, `SELECT schema_version,cursor_epoch,next_update_id,health,retry_attempt,retry_at,last_poll_at,last_error_code,retry_cycle,total_retry_attempts,retry_actor,retry_started_at,retention_gap_at
FROM telegram_receivers WHERE bot_id=?`, botID).Scan(&version, &state.Epoch, &next, &state.Health, &state.RetryAttempt, &retryAt, &lastPoll, &state.LastErrorCode, &state.RetryCycle, &state.TotalRetryAttempts, &state.RetryActor, &retryStarted, &retentionGap)
	if err != nil {
		return state, err
	}

	if version != 1 {
		return state, errors.New("telegram receiver schema is unsupported")
	}

	if next.Valid {
		state.NextUpdateID = &next.Int64
	}

	state.RetryAt, err = parseOptionalTime(retryAt)
	if err != nil {
		return state, err
	}

	state.LastPollAt, err = parseOptionalTime(lastPoll)
	if err != nil {
		return state, err
	}

	state.RetryStartedAt, err = parseOptionalTime(retryStarted)
	if err != nil {
		return state, err
	}

	state.RetentionGapAt, err = parseOptionalTime(retentionGap)

	return state, err
}

func parseOptionalTime(value sql.NullString) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, nil
	}

	return database.ParseTime(value.String)
}

func (s *ReceiverStore) transaction(ctx context.Context, action func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := action(tx); err != nil {
		return errors.Join(database.TransactionFailure(ctx, err), database.TransactionFailure(ctx, tx.Rollback()))
	}

	return database.TransactionFailure(ctx, tx.Commit())
}

// ReservePoll checkpoints an attempt before IO; restart cannot reset exhaustion.
func (s *ReceiverStore) ReservePoll(ctx context.Context, botID int64, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_receivers SET retry_attempt=retry_attempt+1,
total_retry_attempts=total_retry_attempts+1,retry_at=?,updated_at=?
WHERE bot_id=? AND schema_version=1 AND retry_attempt<5 AND health!='blocked' AND (retry_at IS NULL OR retry_at<=?)`,
		database.FormatTime(now.Add(16*time.Second)), database.FormatTime(now), botID, database.FormatTime(now))
	if err != nil {
		return err
	}

	return requireChanged(result)
}

func requireChanged(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("telegram retry is not due, blocked or exhausted")
	}

	return nil
}

// PollFailed stores only sanitized classifications and the retained retry deadline.
func (s *ReceiverStore) PollFailed(ctx context.Context, botID int64, failure error, now time.Time) error {
	state, err := s.Receiver(ctx, botID)
	if err != nil {
		return err
	}

	policy := failureState(failure, state.RetryAttempt)
	code, permanent, delay := policy.Code, policy.Permanent, policy.Delay

	health := "degraded"
	if permanent {
		health = "blocked"
	}

	_, err = s.db.ExecContext(ctx, `UPDATE telegram_receivers SET health=?,last_error_code=?,retry_at=?,updated_at=? WHERE bot_id=?`,
		health, code, database.FormatTime(now.Add(delay)), database.FormatTime(now), botID)

	return err
}

func failureState(failure error, attempt int) failurePolicy {
	delay := time.Second << min(4, max(0, attempt-1))

	var apiErr *APIError
	if errors.As(failure, &apiErr) {
		return failurePolicy{Code: apiErr.Code, Permanent: apiErr.Permanent, Delay: max(delay, apiErr.RetryAfter)}
	}

	return failurePolicy{Code: "local_failure", Permanent: true, Delay: delay}
}
