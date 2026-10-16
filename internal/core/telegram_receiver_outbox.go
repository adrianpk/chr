package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// TelegramReceiverDelivery retains a read response independently of its send outcome.
type TelegramReceiverDelivery struct {
	ID           string
	Peer         TelegramPeerIdentity
	Text         string
	State        string
	Attempts     int
	ConfigDigest string
}

// QueueReadResponse records a response and acknowledges its retained input atomically.
// Replaying an already processed source never replaces or duplicates its response.
func (s *TelegramReceiverStore) QueueReadResponse(ctx context.Context, update TelegramReceivedUpdate, text, configDigest string, now time.Time) error {
	if !validTelegramReceiverText(text) || len(configDigest) != 64 {
		return errors.New("telegram response exceeds its Unicode bounds")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pendingSource, err := telegramReadResponsePending(ctx, tx, update)
		if err != nil || !pendingSource {
			return err
		}

		var pending int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND state!='delivered'", update.Peer.BotID).Scan(&pending); err != nil {
			return err
		}

		if pending >= 1024 {
			return errors.New("telegram outbox capacity exhausted")
		}

		result, err := tx.ExecContext(ctx, "UPDATE telegram_updates SET state='processed',updated_at=? WHERE bot_id=? AND update_id=? AND payload_digest=? AND state='pending'", formatProjectTime(now), update.Peer.BotID, update.UpdateID, update.Digest)
		if err != nil {
			return err
		}

		if err := telegramRequireChanged(result); err != nil {
			return err
		}

		key := fmt.Sprintf("update:%d", update.UpdateID)
		sum := sha256.Sum256(fmt.Appendf(nil, "%d:%s", update.Peer.BotID, key))
		id := "tgd-" + hex.EncodeToString(sum[:16])

		_, err = tx.ExecContext(ctx, `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,'result',?,'pending',?,?)`, id, update.Peer.BotID, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, key, configDigest, text, formatProjectTime(now), formatProjectTime(now))
		if err != nil {
			return err
		}

		return checkTelegramIntakeCapacity(ctx, tx, update.Peer.BotID)
	})
}

func telegramReadResponsePending(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate) (bool, error) {
	if _, err := tx.ExecContext(ctx, "UPDATE telegram_updates SET updated_at=updated_at WHERE bot_id=? AND update_id=?", update.Peer.BotID, update.UpdateID); err != nil {
		return false, err
	}

	var digest, state string
	if err := tx.QueryRowContext(ctx, "SELECT payload_digest,state FROM telegram_updates WHERE bot_id=? AND update_id=?", update.Peer.BotID, update.UpdateID).Scan(&digest, &state); err != nil {
		return false, err
	}

	if digest != update.Digest {
		return false, errTelegramReceiverConflict
	}

	if state == "processed" {
		return false, nil
	}

	if state != "pending" {
		return false, errors.New("telegram source cannot produce a read response")
	}

	return true, nil
}

// RejectPendingUpdate discards a queued body when current authority is revoked.
func (s *TelegramReceiverStore) RejectPendingUpdate(ctx context.Context, update TelegramReceivedUpdate, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE telegram_updates SET state='rejected',payload='',reason_code='authorization_revoked',updated_at=?
WHERE bot_id=? AND update_id=? AND state='pending'`, formatProjectTime(now), update.Peer.BotID, update.UpdateID)

	return err
}

// DueDeliveries is bounded; sending rows after a crash retain uncertainty and budgets.
func (s *TelegramReceiverStore) DueDeliveries(ctx context.Context, botID int64, now time.Time) ([]TelegramReceiverDelivery, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,chat_id,topic_id,user_id,text,state,attempts,config_digest FROM telegram_outbox
WHERE bot_id=? AND state IN ('pending','sending','uncertain') AND attempts<5 AND (retry_at IS NULL OR retry_at<=?) ORDER BY created_at,rowid LIMIT 50`, botID, formatProjectTime(now))
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck // rows.Err is checked after the bounded iteration.
		_ = rows.Close()
	}()

	var result []TelegramReceiverDelivery

	for rows.Next() {
		delivery := TelegramReceiverDelivery{Peer: TelegramPeerIdentity{BotID: botID}}
		if err := rows.Scan(&delivery.ID, &delivery.Peer.ChatID, &delivery.Peer.TopicID, &delivery.Peer.UserID, &delivery.Text, &delivery.State, &delivery.Attempts, &delivery.ConfigDigest); err != nil {
			return nil, err
		}

		result = append(result, delivery)
	}

	return result, rows.Err()
}

// ReserveDelivery persists the attempt before sendMessage; a lost acknowledgement cannot reset it.
func (s *TelegramReceiverStore) ReserveDelivery(ctx context.Context, id string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_outbox SET state='sending',attempts=attempts+1,total_attempts=total_attempts+1,
retry_at=?,updated_at=? WHERE id=? AND state IN ('pending','sending','uncertain') AND attempts<5 AND (retry_at IS NULL OR retry_at<=?)`,
		formatProjectTime(now.Add(16*time.Second)), formatProjectTime(now), id, formatProjectTime(now))
	if err != nil {
		return err
	}

	return telegramRequireChanged(result)
}

// DeliverySent records definite API acknowledgement without applying any operation again.
func (s *TelegramReceiverStore) DeliverySent(ctx context.Context, id string, messageID int64, now time.Time) error {
	if messageID <= 0 {
		return errors.New("telegram delivery requires a definite message ID")
	}

	result, err := s.db.ExecContext(ctx, "UPDATE telegram_outbox SET state='delivered',telegram_message_id=?,retry_at=NULL,last_error_code='',updated_at=? WHERE id=? AND state='sending'", messageID, formatProjectTime(now), id)
	if err != nil {
		return err
	}

	return telegramRequireChanged(result)
}

// DeliveryFailed retains uncertainty or blocks a permanent/exhausted failure.
func (s *TelegramReceiverStore) DeliveryFailed(ctx context.Context, delivery TelegramReceiverDelivery, failure error, now time.Time) error {
	code, permanent, delay := telegramFailureState(failure, delivery.Attempts+1)
	state := "pending"

	var apiErr *TelegramAPIError
	if errors.As(failure, &apiErr) && apiErr.Uncertain {
		state = "uncertain"
	}

	if permanent || delivery.Attempts+1 >= 5 {
		state = "blocked"
	}

	_, err := s.db.ExecContext(ctx, "UPDATE telegram_outbox SET state=?,last_error_code=?,retry_at=?,updated_at=? WHERE id=? AND state='sending'", state, code, formatProjectTime(now.Add(delay)), formatProjectTime(now), delivery.ID)

	return err
}

// BlockDelivery prevents delivery to a revoked peer without creating another request.
func (s *TelegramReceiverStore) BlockDelivery(ctx context.Context, id string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE telegram_outbox SET state='blocked',last_error_code='authorization_revoked',updated_at=? WHERE id=? AND state!='delivered'", formatProjectTime(now), id)

	return err
}
