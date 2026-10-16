package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TelegramInterpretation retains resolution independently of application effects.
type TelegramInterpretation struct {
	Attempts  int
	RetryAt   time.Time
	ErrorCode string
	Decision  json.RawMessage
}

// ResolveTelegramInterpretation reserves an attempt before calling a provider.
// OS exclusion outlives every database transaction and is released on process
// death. A timestamp never steals an active caller's interpretation claim.
func (s *TelegramReceiverStore) ResolveTelegramInterpretation(ctx context.Context, env UserEnvironment, update TelegramReceivedUpdate, now time.Time, invoke func(context.Context) (json.RawMessage, error)) (result TelegramInterpretation, err error) {
	if err := validateTelegramInterpretationInput(update, invoke); err != nil {
		return result, err
	}

	lock, err := acquireTelegramFileLock(env, fmt.Sprintf("interpretation-%d.lock", update.Peer.BotID))
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()

	result, ready, err := s.reserveTelegramInterpretation(ctx, update, now)
	if err != nil || !ready {
		return result, err
	}

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	decision, callErr := invoke(callCtx)
	if callErr == nil && (len(decision) > 65536 || !json.Valid(decision)) {
		callErr = errors.New("invalid typed interpretation")
	}
	// The attempt is already durable if shutdown prevents the outcome write.
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finishCancel()

	if callErr != nil {
		result.ErrorCode = telegramInterpretationFailure(callErr)
		result.RetryAt = time.Now().UTC().Add(time.Second * time.Duration(1<<(result.Attempts-1)))
		_, err = s.db.ExecContext(finishCtx, `UPDATE telegram_updates SET interpretation_retry_at=?,interpretation_error_code=? WHERE bot_id=? AND update_id=? AND payload_digest=? AND interpretation_attempt=?`, formatProjectTime(result.RetryAt), result.ErrorCode, update.Peer.BotID, update.UpdateID, update.Digest, result.Attempts)

		return result, err
	}

	result.Decision = decision
	result.RetryAt = time.Time{}
	result.ErrorCode = ""
	err = s.transaction(finishCtx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(finishCtx, `UPDATE telegram_updates SET interpretation=?,interpretation_retry_at=NULL,interpretation_error_code='' WHERE bot_id=? AND update_id=? AND payload_digest=? AND interpretation_attempt=?`, string(decision), update.Peer.BotID, update.UpdateID, update.Digest, result.Attempts)
		if err != nil {
			return err
		}

		return checkTelegramIntakeCapacity(finishCtx, tx, update.Peer.BotID)
	})

	return result, err
}

func (s *TelegramReceiverStore) reserveTelegramInterpretation(ctx context.Context, update TelegramReceivedUpdate, now time.Time) (TelegramInterpretation, bool, error) {
	var result TelegramInterpretation

	ready := false
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, update)
		if err != nil || !pending {
			return err
		}

		var retry, decision string

		err = tx.QueryRowContext(ctx, `SELECT interpretation_attempt,COALESCE(interpretation_retry_at,''),interpretation_error_code,COALESCE(interpretation,'') FROM telegram_updates WHERE bot_id=? AND update_id=? AND payload_digest=? AND chat_id=? AND topic_id=? AND user_id=? AND payload=? AND update_kind=?`, update.Peer.BotID, update.UpdateID, update.Digest, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, update.Text, update.Kind).Scan(&result.Attempts, &retry, &result.ErrorCode, &decision)
		if err != nil {
			return err
		}

		result.Decision = json.RawMessage(decision)
		if retry != "" {
			result.RetryAt, err = time.Parse(time.RFC3339Nano, retry)
			if err != nil {
				return err
			}
		}

		if len(result.Decision) > 0 || result.Attempts >= 5 || now.Before(result.RetryAt) {
			return nil
		}

		result.Attempts++
		result.RetryAt = now.Add(61 * time.Second)
		_, err = tx.ExecContext(ctx, `UPDATE telegram_updates SET interpretation_attempt=?,interpretation_retry_at=?,interpretation_error_code='interrupted' WHERE bot_id=? AND update_id=? AND payload_digest=?`, result.Attempts, formatProjectTime(result.RetryAt), update.Peer.BotID, update.UpdateID, update.Digest)
		ready = err == nil

		return err
	})

	return result, ready, err
}

func telegramInterpretationFailure(err error) string {
	switch {
	case errors.Is(err, ErrStructuredModelCapacity):
		return "provider_capacity"
	case errors.Is(err, context.DeadlineExceeded):
		return "provider_timeout"
	case errors.Is(err, context.Canceled):
		return "provider_canceled"
	default:
		return "provider_unavailable"
	}
}

func validateTelegramInterpretationInput(update TelegramReceivedUpdate, invoke func(context.Context) (json.RawMessage, error)) error {
	if update.Kind != "message" || strings.HasPrefix(strings.TrimSpace(update.Text), "/") || invoke == nil {
		return errTelegramReceiverConflict
	}

	return validateTelegramRetainedUpdate(update.Peer.BotID, update)
}
