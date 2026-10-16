package core

import (
	"context"
	"database/sql"
	"time"
)

func retainTelegramReply(ctx context.Context, tx *sql.Tx, id string, messageID int64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO telegram_replies(bot_id,chat_id,topic_id,message_id,user_id,conversation_id,request_id,action_token,purpose,conversation_revision,question_binding,created_at,invalidated_at)
SELECT o.bot_id,o.chat_id,o.topic_id,?,o.user_id,o.conversation_id,o.request_id,o.action_token,
CASE WHEN o.action_token IS NULL THEN 'capture' ELSE 'action' END,o.conversation_revision,o.question_binding,?,
CASE WHEN o.action_token IS NOT NULL THEN
 CASE WHEN c.revision=o.conversation_revision AND a.conversation_revision=c.revision AND a.state='pending' AND a.expires_at>? THEN NULL ELSE ? END
 ELSE CASE WHEN c.revision=o.conversation_revision AND c.question_binding=o.question_binding AND c.phase IN ('collecting','clarifying') THEN NULL ELSE ? END END
FROM telegram_outbox o JOIN telegram_conversations c ON c.id=o.conversation_id LEFT JOIN telegram_actions a ON a.token=o.action_token WHERE o.id=?`, messageID, formatProjectTime(now), formatProjectTime(now), formatProjectTime(now), formatProjectTime(now), id)
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
