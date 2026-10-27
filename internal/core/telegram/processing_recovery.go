package telegram

import (
	"context"
	"database/sql"
)

// Existing pending sources gain the same classification as new intake. This
// also resolves a reply whose definite delivery binding committed after intake.
func (s *TelegramReceiverStore) restoreTelegramInputLanes(ctx context.Context, botID int64) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		return restoreTelegramBotLanes(ctx, tx, botID)
	})
}

func restoreTelegramBotLanes(ctx context.Context, tx *sql.Tx, botID int64) error {
	rows, err := tx.QueryContext(ctx, `SELECT update_id,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,payload,reply_to_message_id FROM telegram_updates WHERE bot_id=? AND state IN ('pending','processing') ORDER BY update_id LIMIT 1024`, botID)
	if err != nil {
		return err
	}

	updates, err := readTelegramPendingRows(rows, botID)
	if err != nil {
		return err
	}

	return restoreTelegramUpdateLanes(ctx, tx, updates)
}

func restoreTelegramUpdateLanes(ctx context.Context, tx *sql.Tx, updates []TelegramReceivedUpdate) error {
	for _, update := range updates {
		lane, control, err := telegramInputLane(ctx, tx, update)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_updates SET processing_lane=?,control_input=? WHERE bot_id=? AND update_id=?`, lane, control, update.Peer.BotID, update.UpdateID); err != nil {
			return err
		}
	}

	return nil
}

func refreshTelegramReplyLanes(ctx context.Context, tx *sql.Tx, id string, messageID int64) error {
	var botID int64
	if err := tx.QueryRowContext(ctx, `SELECT bot_id FROM telegram_outbox WHERE id=?`, id).Scan(&botID); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `SELECT u.update_id,u.payload_digest,u.update_kind,u.chat_id,u.topic_id,u.user_id,u.message_id,u.chat_type,u.payload,u.reply_to_message_id FROM telegram_updates u JOIN telegram_outbox o ON o.bot_id=u.bot_id AND o.chat_id=u.chat_id AND o.topic_id=u.topic_id AND o.user_id=u.user_id WHERE o.id=? AND u.reply_to_message_id=? AND u.state IN ('pending','processing') ORDER BY u.update_id LIMIT 1024`, id, messageID)
	if err != nil {
		return err
	}

	updates, err := readTelegramPendingRows(rows, botID)
	if err != nil {
		return err
	}

	return restoreTelegramUpdateLanes(ctx, tx, updates)
}
