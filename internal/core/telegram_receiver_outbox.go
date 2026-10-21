package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TelegramReceiverDelivery retains a response independently of its send outcome.
type TelegramReceiverDelivery struct {
	ID            string
	Peer          TelegramPeerIdentity
	Text          string
	State         string
	Attempts      int
	ConfigDigest  string
	ActionToken   string
	ActionChoices []string
	ParseMode     string
	PlainText     string
	PlainFallback bool
	DocumentID    string
	Presentation  *TelegramPresentation
}

// QueueReadResponse records a response and acknowledges its retained input atomically.
// Replaying an already processed source never replaces or duplicates its response.
func (s *TelegramReceiverStore) QueueReadResponse(ctx context.Context, update TelegramReceivedUpdate, text, configDigest string, now time.Time) error {
	return s.QueueReferencedResponse(ctx, update, text, fmt.Sprintf("update:%d", update.UpdateID), configDigest, now)
}

// QueueReferencedResponse retains bounded output with a canonical or paged reference.
func (s *TelegramReceiverStore) QueueReferencedResponse(ctx context.Context, update TelegramReceivedUpdate, text, reference, configDigest string, now time.Time) error {
	if text == "" || len(configDigest) != 64 {
		return errors.New("telegram response exceeds its Unicode bounds")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pendingSource, err := telegramReadResponsePending(ctx, tx, update)
		if err != nil || !pendingSource {
			return err
		}

		return queueTelegramReferencedResponse(ctx, tx, update, text, reference, configDigest, now)
	})
}

func queueTelegramResponse(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, text, configDigest string, now time.Time) error {
	return queueTelegramReferencedResponse(ctx, tx, update, text, fmt.Sprintf("update:%d", update.UpdateID), configDigest, now)
}

func queueTelegramReferencedResponse(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, text, reference, configDigest string, now time.Time) error {
	var pending int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND state!='delivered'", update.Peer.BotID).Scan(&pending); err != nil {
		return err
	}

	parts := TelegramResponseParts(text, reference)
	if pending+len(parts) > 1024 {
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
	for index, part := range parts {
		sum := sha256.Sum256(fmt.Appendf(nil, "%d:%s:%d", update.Peer.BotID, key, index+1))
		id := "tgd-" + hex.EncodeToString(sum[:16])

		_, err = tx.ExecContext(ctx, `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,part_index,part_count,state,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,'result',?,?,?,'pending',?,?)`, id, update.Peer.BotID, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID, key, configDigest, part, index+1, len(parts), formatProjectTime(now), formatProjectTime(now))
		if err != nil {
			return err
		}
	}

	return checkTelegramIntakeCapacity(ctx, tx, update.Peer.BotID)
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
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if err := revokeTelegramPeerActions(ctx, tx, update.Peer, now); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, `UPDATE telegram_updates SET state='rejected',payload='',reason_code='authorization_revoked',updated_at=? WHERE bot_id=? AND update_id=? AND state='pending'`, formatProjectTime(now), update.Peer.BotID, update.UpdateID)

		return err
	})
}

// DueDeliveries is bounded; sending rows after a crash retain uncertainty and budgets.
func (s *TelegramReceiverStore) DueDeliveries(ctx context.Context, botID int64, now time.Time) ([]TelegramReceiverDelivery, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT o.id,o.chat_id,o.topic_id,o.user_id,o.text,o.state,o.attempts,o.config_digest,COALESCE(o.action_token,''),COALESCE((SELECT allowed_choices FROM telegram_actions WHERE token=o.action_token),'[]'),o.parse_mode,o.plain_text,(o.plain_fallback OR o.presentation_fallback),COALESCE(o.presentation,''),COALESCE(o.presentation_hash,''),COALESCE(o.document_id,'') FROM telegram_outbox o
WHERE o.bot_id=? AND o.state IN ('pending','sending','uncertain') AND o.attempts<5 AND (o.retry_at IS NULL OR o.retry_at<=?)
AND NOT EXISTS(SELECT 1 FROM telegram_outbox earlier WHERE earlier.bot_id=o.bot_id AND earlier.chat_id=o.chat_id AND earlier.topic_id=o.topic_id AND earlier.user_id=o.user_id AND earlier.event_key=o.event_key AND earlier.part_index<o.part_index AND earlier.state!='delivered')
AND (o.kind!='progress' OR o.attempts>0 OR NOT EXISTS(SELECT 1 FROM telegram_outbox recent WHERE recent.conversation_id=o.conversation_id AND recent.kind='progress' AND recent.total_attempts>0 AND recent.event_key!=o.event_key AND recent.updated_at>?))
ORDER BY CASE (SELECT u.processing_lane FROM telegram_updates u WHERE u.bot_id=o.bot_id AND u.update_id=CAST(substr(o.event_key,8) AS INTEGER) AND o.event_key='update:'||u.update_id) WHEN 'urgent' THEN 0 WHEN 'decision' THEN 1 ELSE 2 END,o.created_at,o.rowid LIMIT 50`, botID, formatProjectTime(now), formatProjectTime(now.Add(-30*time.Second)))
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

		var choices, presentation, presentationHash string
		if err := rows.Scan(&delivery.ID, &delivery.Peer.ChatID, &delivery.Peer.TopicID, &delivery.Peer.UserID, &delivery.Text, &delivery.State, &delivery.Attempts, &delivery.ConfigDigest, &delivery.ActionToken, &choices, &delivery.ParseMode, &delivery.PlainText, &delivery.PlainFallback, &presentation, &presentationHash, &delivery.DocumentID); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(choices), &delivery.ActionChoices); err != nil {
			return nil, err
		}

		if err := decodeTelegramPresentationHash(presentation, presentationHash, &delivery); err != nil {
			return nil, err
		}

		result = append(result, delivery)
	}

	return result, rows.Err()
}

// ReserveDelivery persists the attempt before sendMessage; a lost acknowledgement cannot reset it.
func (s *TelegramReceiverStore) ReserveDelivery(ctx context.Context, id string, now time.Time) error {
	reserved, err := s.TryReserveDelivery(ctx, id, now)
	if err != nil {
		return err
	}

	if !reserved {
		return errors.New("telegram retry is not due, blocked or exhausted")
	}

	return nil
}

// TryReserveDelivery skips a stale read when routine progress was coalesced.
// Successful reservation persists its attempt before the API call.
func (s *TelegramReceiverStore) TryReserveDelivery(ctx context.Context, id string, now time.Time) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_outbox AS o SET state='sending',attempts=attempts+1,total_attempts=total_attempts+1,
retry_at=?,updated_at=? WHERE id=? AND state IN ('pending','sending','uncertain') AND attempts<5 AND (retry_at IS NULL OR retry_at<=?)
AND (o.kind!='progress' OR o.attempts>0 OR NOT EXISTS(SELECT 1 FROM telegram_outbox recent WHERE recent.conversation_id=o.conversation_id AND recent.kind='progress' AND recent.total_attempts>0 AND recent.event_key!=o.event_key AND recent.updated_at>?))`,
		formatProjectTime(now.Add(16*time.Second)), formatProjectTime(now), id, formatProjectTime(now), formatProjectTime(now.Add(-30*time.Second)))
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()

	return count == 1, err
}

// DeliverySent records definite API acknowledgement without applying any operation again.
func (s *TelegramReceiverStore) DeliverySent(ctx context.Context, id string, messageID int64, now time.Time) error {
	if messageID <= 0 {
		return errors.New("telegram delivery requires a definite message ID")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, "UPDATE telegram_outbox SET state='delivered',telegram_message_id=?,retry_at=NULL,last_error_code='',updated_at=? WHERE id=? AND state='sending'", messageID, formatProjectTime(now), id)
		if err != nil {
			return err
		}

		if err := telegramRequireChanged(result); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_document_snapshots SET terminal_at=? WHERE owner_id=? AND terminal_at IS NULL`, formatProjectTime(now), id); err != nil {
			return err
		}

		if err := retainTelegramReply(ctx, tx, id, messageID, now); err != nil {
			return err
		}

		return refreshTelegramReplyLanes(ctx, tx, id, messageID)
	})
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
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "UPDATE telegram_outbox SET state='blocked',last_error_code='authorization_revoked',updated_at=? WHERE id=? AND state!='delivered'", formatProjectTime(now), id); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, "UPDATE telegram_document_snapshots SET terminal_at=? WHERE owner_id=? AND terminal_at IS NULL", formatProjectTime(now), id)

		return err
	})
}
