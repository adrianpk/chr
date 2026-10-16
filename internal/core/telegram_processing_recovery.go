package core

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

	for _, update := range updates {
		lane, control, err := telegramInputLane(ctx, tx, update)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_updates SET processing_lane=?,control_input=? WHERE bot_id=? AND update_id=?`, lane, control, botID, update.UpdateID); err != nil {
			return err
		}
	}

	return nil
}
