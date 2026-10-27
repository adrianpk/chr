package core

import (
	"context"
	"database/sql"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

func revokeRejectedTelegramActions(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, now time.Time) error {
	if update.Kind != "rejected" || update.Reason != "unauthorized_peer" {
		return nil
	}

	return revokeTelegramPeerActions(ctx, tx, update.Peer, now)
}

func revokeTelegramPeerActions(ctx context.Context, tx *sql.Tx, peer TelegramPeerIdentity, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_actions SET state='superseded',updated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND state IN ('pending','applying')`, database.FormatTime(now), peer.BotID, peer.ChatID, peer.TopicID, peer.UserID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_requests SET state='blocked',updated_at=? WHERE state NOT IN ('confirmed','rejected') AND id IN (
SELECT request_id FROM telegram_actions WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND state='superseded'
UNION SELECT consumed_by_request FROM telegram_actions WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND state='superseded')`, database.FormatTime(now), peer.BotID, peer.ChatID, peer.TopicID, peer.UserID, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND invalidated_at IS NULL AND action_token IN (SELECT token FROM telegram_actions WHERE state='superseded')`, database.FormatTime(now), peer.BotID, peer.ChatID, peer.TopicID, peer.UserID)

	return err
}
