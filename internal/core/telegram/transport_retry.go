package telegram

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TelegramTransportRetry names one exhausted transport cycle. It cannot reset
// application model attempts, native work, confirmations or domain budgets.
type TelegramTransportRetry struct {
	BotID      int64
	Poll       bool
	DeliveryID string
	Actor      string
}

// RetryLocalTelegramTransport acquires receiver exclusion before opening or
// migrating the store. A running receiver's schema and attempts stay untouched.
func RetryLocalTelegramTransport(ctx context.Context, env TelegramStoreDirectories, assets fs.FS, request TelegramTransportRetry, now time.Time) (err error) {
	lock, err := AcquireTelegramConfigurationLock(env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	botID, err := TelegramBoundBotID(env)
	if err != nil || botID != request.BotID {
		return workflowstate.ErrConflict
	}

	store, err := NewTelegramReceiverStore(ctx, env, assets)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()

	return store.retryTelegramTransport(ctx, request, now)
}

// RetryTelegramTransport starts an attributable local cycle only while the
// receiver is offline. Lifetime counts, source identities and effects survive.
func (s *TelegramReceiverStore) RetryTelegramTransport(ctx context.Context, request TelegramTransportRetry, now time.Time) (err error) {
	lock, err := AcquireTelegramConfigurationLock(s.env)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	return s.retryTelegramTransport(ctx, request, now)
}

func (s *TelegramReceiverStore) retryTelegramTransport(ctx context.Context, request TelegramTransportRetry, now time.Time) error {
	if request.BotID <= 0 || request.Poll == (request.DeliveryID != "") || len(request.DeliveryID) > 128 || strings.TrimSpace(request.Actor) != request.Actor {
		return workflowstate.ErrConflict
	}

	if err := (workflowstate.MutationBinding{RequestID: "telegram-transport-retry", Operator: request.Actor}).Validate(); err != nil {
		return err
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE telegram_receivers SET updated_at=updated_at WHERE bot_id=? AND schema_version=1`, request.BotID); err != nil {
			return err
		}

		if request.Poll {
			return retryTelegramPoll(ctx, tx, request, now)
		}

		result, err := tx.ExecContext(ctx, `UPDATE telegram_outbox SET state='pending',attempts=0,retry_at=NULL,last_error_code='',retry_cycle=retry_cycle+1,retry_actor=?,retry_started_at=?,updated_at=? WHERE id=? AND bot_id=? AND state!='delivered' AND (state='blocked' OR attempts=5)`, request.Actor, database.FormatTime(now), database.FormatTime(now), request.DeliveryID, request.BotID)
		if err != nil {
			return err
		}

		return telegramRequireChanged(result)
	})
}

func retryTelegramPoll(ctx context.Context, tx *sql.Tx, request TelegramTransportRetry, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE telegram_receivers SET health='degraded',retry_attempt=0,retry_at=NULL,last_error_code='',retry_cycle=retry_cycle+1,retry_actor=?,retry_started_at=?,updated_at=? WHERE bot_id=? AND schema_version=1 AND (health='blocked' OR retry_attempt=5)`, request.Actor, database.FormatTime(now), database.FormatTime(now), request.BotID)
	if err != nil {
		return err
	}

	return telegramRequireChanged(result)
}
