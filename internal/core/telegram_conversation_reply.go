package core

import (
	"context"
	"database/sql"
	"time"
)

func retainTelegramReply(ctx context.Context, tx *sql.Tx, id string, messageID int64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO telegram_replies(bot_id,chat_id,topic_id,message_id,user_id,conversation_id,purpose,conversation_revision,question_binding,created_at,invalidated_at)
SELECT o.bot_id,o.chat_id,o.topic_id,?,o.user_id,o.conversation_id,'capture',o.conversation_revision,o.question_binding,?,
CASE WHEN c.revision=o.conversation_revision AND c.question_binding=o.question_binding AND c.phase IN ('collecting','clarifying') THEN NULL ELSE ? END
FROM telegram_outbox o JOIN telegram_conversations c ON c.id=o.conversation_id WHERE o.id=?`, messageID, formatProjectTime(now), formatProjectTime(now), id)
	if err != nil {
		return err
	}

	return pruneTelegramReplies(ctx, tx, id, now)
}

func pruneTelegramReplies(ctx context.Context, tx *sql.Tx, id string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE rowid IN (
SELECT rowid FROM telegram_replies WHERE conversation_id=(SELECT conversation_id FROM telegram_outbox WHERE id=?) AND invalidated_at IS NULL ORDER BY created_at DESC,rowid DESC LIMIT -1 OFFSET 64)`, formatProjectTime(now), id)

	return err
}
