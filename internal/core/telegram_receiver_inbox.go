package core

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"time"
)

// RetainBatch commits every normalized receipt and the next sparse cursor together.
func (s *TelegramReceiverStore) RetainBatch(ctx context.Context, botID int64, updates []TelegramReceivedUpdate, now time.Time) error {
	if len(updates) > 100 {
		return errors.New("telegram batch exceeds 100 updates")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		var (
			epoch int64
			next  sql.NullInt64
		)
		// Acquire SQLite's write reservation before checking quotas or cursor state.
		if _, err := tx.ExecContext(ctx, "UPDATE telegram_receivers SET updated_at=updated_at WHERE bot_id=?", botID); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx, "SELECT cursor_epoch,next_update_id FROM telegram_receivers WHERE bot_id=? AND schema_version=1", botID).Scan(&epoch, &next); err != nil {
			return err
		}

		var err error

		next, err = retainTelegramBatchUpdates(ctx, tx, botID, epoch, next, updates, now)
		if err != nil {
			return err
		}

		if err := checkTelegramIntakeCapacity(ctx, tx, botID); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_receivers SET next_update_id=?,last_poll_at=?,health='ready',
retry_attempt=0,retry_at=NULL,last_error_code='',updated_at=? WHERE bot_id=?`, next, formatProjectTime(now), formatProjectTime(now), botID)

		return err
	})
}

func retainTelegramBatchUpdates(ctx context.Context, tx *sql.Tx, botID, epoch int64, next sql.NullInt64, updates []TelegramReceivedUpdate, now time.Time) (sql.NullInt64, error) {
	for _, update := range updates {
		if err := retainTelegramUpdate(ctx, tx, botID, epoch, update, now); err != nil {
			return next, err
		}

		if !next.Valid || update.UpdateID >= next.Int64 {
			next = sql.NullInt64{Int64: update.UpdateID + 1, Valid: true}
		}
	}

	return next, nil
}

func retainTelegramUpdate(ctx context.Context, tx *sql.Tx, botID, epoch int64, update TelegramReceivedUpdate, now time.Time) error {
	if err := validateTelegramRetainedUpdate(botID, update); err != nil {
		return err
	}

	var prior string

	err := tx.QueryRowContext(ctx, "SELECT payload_digest FROM telegram_updates WHERE bot_id=? AND update_id=?", botID, update.UpdateID).Scan(&prior)
	if err == nil {
		if prior != update.Digest {
			return errTelegramReceiverConflict
		}

		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	state := "pending"
	if update.Kind == "ignored" || update.Kind == "rejected" {
		state = update.Kind
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,
chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at,reply_to_message_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		botID, update.UpdateID, epoch, update.Digest, update.Kind, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID,
		update.MessageID, update.ChatType, formatProjectTime(now), update.Text, state, update.Reason, formatProjectTime(now), update.ReplyToMessageID)
	if err != nil {
		return err
	}

	return revokeRejectedTelegramActions(ctx, tx, update, now)
}

func validateTelegramRetainedUpdate(botID int64, update TelegramReceivedUpdate) error {
	if !validTelegramReceiptIdentity(botID, update) {
		return errors.New("invalid telegram receipt identity")
	}

	switch update.Kind {
	case "message", "callback_query":
		if update.Peer.UserID <= 0 || update.Peer.ChatID == 0 || len(update.Text) == 0 || len(update.Text) > 16*1024 {
			return errors.New("invalid telegram receipt payload")
		}
	case "ignored", "rejected":
		if update.Text != "" {
			return errors.New("denied telegram payload cannot be retained")
		}
	default:
		return errors.New("unknown telegram update kind")
	}

	return nil
}

func validTelegramReceiptIdentity(botID int64, update TelegramReceivedUpdate) bool {
	return botID > 0 && update.Peer.BotID == botID && update.UpdateID >= 0 && update.UpdateID != math.MaxInt64 && regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(update.Digest)
}

func checkTelegramIntakeCapacity(ctx context.Context, tx *sql.Tx, botID int64) error {
	var pending, total, bytes int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(state IN ('pending','processing')),0),COALESCE(SUM(length(CAST(payload AS BLOB))),0)
FROM telegram_updates WHERE bot_id=?`, botID).Scan(&total, &pending, &bytes); err != nil {
		return err
	}

	var outboundBytes int64
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(SUM(length(CAST(text AS BLOB))),0) FROM telegram_outbox WHERE bot_id=?", botID).Scan(&outboundBytes); err != nil {
		return err
	}

	retainedBytes, err := telegramOperationCapacity(ctx, tx, botID)
	if err != nil {
		return err
	}

	if pending > 1024 || total > 10000 || bytes+outboundBytes+retainedBytes > 64*1024*1024 {
		return errors.New("telegram intake capacity exhausted")
	}

	return nil
}

// PendingUpdates returns at most 50 retained updates in source order.
func (s *TelegramReceiverStore) PendingUpdates(ctx context.Context, botID int64) ([]TelegramReceivedUpdate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT update_id,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,payload,reply_to_message_id
FROM telegram_updates WHERE bot_id=? AND state='pending' ORDER BY update_id LIMIT 50`, botID)
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck // rows.Err is checked after the bounded iteration.
		_ = rows.Close()
	}()

	var result []TelegramReceivedUpdate

	for rows.Next() {
		update := TelegramReceivedUpdate{Peer: TelegramPeerIdentity{BotID: botID}}
		if err := rows.Scan(&update.UpdateID, &update.Digest, &update.Kind, &update.Peer.ChatID, &update.Peer.TopicID, &update.Peer.UserID, &update.MessageID, &update.ChatType, &update.Text, &update.ReplyToMessageID); err != nil {
			return nil, err
		}

		result = append(result, update)
	}

	return result, rows.Err()
}
