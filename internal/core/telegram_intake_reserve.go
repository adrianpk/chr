package core

import (
	"context"
	"database/sql"
	"time"
)

// The 32 reserved inputs are part of the 1,024 pending limit, not extra capacity.
// Rejected metadata and its diagnostic commit with the complete polling batch.
func rejectTelegramIntakeOverflow(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, control bool, now time.Time) error {
	var total, ordinary int

	err := tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(control_input=0),0) FROM telegram_updates WHERE bot_id=? AND state IN ('pending','processing')`, update.Peer.BotID).Scan(&total, &ordinary)
	if err != nil || (total <= 1024 && (control || ordinary <= 992)) {
		return err
	}

	var digest string
	if err := tx.QueryRowContext(ctx, "SELECT config_digest FROM telegram_receivers WHERE bot_id=?", update.Peer.BotID).Scan(&digest); err != nil {
		return err
	}

	text := "The input queue is full. This request was not accepted. Retry after pending requests finish. Existing work is unchanged; status, stop and reconciliation retain reserved capacity."
	if err := queueTelegramPresentation(ctx, tx, update, TelegramMessageBlocks{Family: "42", InitialText: text}, digest, now); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_updates SET state='rejected',payload='',reason_code='intake_capacity' WHERE bot_id=? AND update_id=?`, update.Peer.BotID, update.UpdateID)

	return err
}
