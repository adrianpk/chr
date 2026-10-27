package core

import (
	"context"
	"database/sql"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// ReserveTelegramNotification permits one original dispatch and at most one
// definite-format fallback. Reopen or uncertainty never grants a fresh attempt.
func (r *sqliteNotificationEventRepository) ReserveTelegramNotification(ctx context.Context, id string, target TelegramNotificationTarget, plain bool) (bool, error) {
	expected := 0
	if plain {
		expected = 1
	}

	result, err := r.db.ExecContext(ctx, `UPDATE notification_events SET telegram_attempts=telegram_attempts+1,telegram_plain_fallback=?,telegram_uncertain=1 WHERE id=? AND project_id=? AND channel='telegram' AND status='pending' AND telegram_presentation IS NOT NULL AND telegram_destination=? AND telegram_config_digest=? AND telegram_namespace=? AND telegram_project_binding_digest=? AND telegram_attempts=? AND telegram_plain_fallback=0`, plain, id, target.ProjectID, target.Destination, target.ConfigDigest, target.Namespace, target.ProjectBindingDigest, expected)
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()

	return count == 1, err
}

// FinishTelegramNotification saves only transport evidence for the original
// claimed event; it cannot publish work or create a native delivery receipt.
func (r *sqliteNotificationEventRepository) FinishTelegramNotification(ctx context.Context, id string, target TelegramNotificationTarget, ack TelegramDocumentAcknowledgement, code string, uncertain bool, now time.Time) error {
	status := NotificationStatusFailed
	minimumAttempts := 0

	var delivered any

	if ack.MessageID > 0 {
		minimumAttempts = 1
		status = NotificationStatusDelivered
		delivered = database.FormatTime(now)
		uncertain = false
		code = ""
	}

	if len(code) > 128 {
		return ErrTelegramDocumentUnavailable
	}

	return r.notificationTransaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE notification_events SET status=?,failure_reason=?,delivered_at=?,updated_at=?,telegram_message_id=NULLIF(?,0),telegram_file_id=NULLIF(?,''),telegram_uncertain=? WHERE id=? AND project_id=? AND telegram_destination=? AND telegram_config_digest=? AND telegram_namespace=? AND telegram_project_binding_digest=? AND status='pending' AND telegram_attempts>=?`, string(status), code, delivered, database.FormatTime(now), ack.MessageID, ack.FileID, uncertain, id, target.ProjectID, target.Destination, target.ConfigDigest, target.Namespace, target.ProjectBindingDigest, minimumAttempts)
		if err != nil {
			return err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if count != 1 {
			return ErrWorkflowConflict
		}

		if uncertain {
			return nil
		}

		_, err = tx.ExecContext(ctx, "UPDATE telegram_notification_documents SET terminal_at=COALESCE(terminal_at,?) WHERE owner_id=?", database.FormatTime(now), id)

		return err
	})
}
