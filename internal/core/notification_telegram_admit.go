package core

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// CreateTelegramNotification admits the event, immutable view and complete
// document atomically. A failed quota reservation leaves no partial event.
func (r *sqliteNotificationEventRepository) CreateTelegramNotification(ctx context.Context, input TelegramNotificationInput) error {
	if err := validateTelegramNotificationInput(input); err != nil {
		return err
	}

	return r.notificationTransaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, insertNotificationEventSQL(), (notificationEventRow{record: input.Record}).values()...); err != nil {
			return err
		}

		if err := purgeTelegramNotificationDocuments(ctx, tx, input.Record.Event.CreatedAt); err != nil {
			return err
		}

		view := RenderTelegramMessage(input.Message)
		if input.Document != nil {
			view = RenderTelegramDocument(input.Message)
			view.DocumentID = "tgndoc-" + contenthash.Sum([]byte(input.Record.Event.ID))[:32]

			view.DocumentHash = contenthash.Sum(input.Document.Payload)
			if err := retainTelegramNotificationDocument(ctx, tx, input, view.DocumentID); err != nil {
				return err
			}
		}

		payload, err := json.Marshal(view)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE notification_events SET telegram_presentation=?,telegram_presentation_hash=?,telegram_document_id=NULLIF(?,''),telegram_destination=?,telegram_config_digest=?,telegram_namespace=?,telegram_project_binding_digest=? WHERE id=?`, string(payload), contenthash.Sum(payload), view.DocumentID, input.Target.Destination, input.Target.ConfigDigest, input.Target.Namespace, input.Target.ProjectBindingDigest, input.Record.Event.ID)

		return err
	})
}

func validateTelegramNotificationInput(input TelegramNotificationInput) error {
	if err := ValidateNotificationAuditRecord(input.Record); err != nil {
		return err
	}

	if len(input.Message.InitialText)+len(input.Message.Content)+len(input.Message.SecondaryText) > telegramDocumentLimit || len(input.Message.Family) > 64 {
		return ErrTelegramDocumentUnavailable
	}

	if err := validateTelegramNotificationTarget(input); err != nil {
		return err
	}

	if RenderTelegramMessage(input.Message).PlainFallback == "" {
		return ErrTelegramDocumentUnavailable
	}

	return validateTelegramNotificationDocument(input.Document, input.Target)
}

func validateTelegramNotificationTarget(input TelegramNotificationInput) error {
	target := input.Target
	if input.Record.Channel != "telegram" || input.Record.Event.Status != NotificationStatusPending || input.Record.Event.ProjectID != target.ProjectID {
		return ErrTelegramNotificationTargetUnavailable
	}

	if !validTelegramDigest(target.ConfigDigest) || !validTelegramDigest(target.ProjectBindingDigest) || !opsnamespace.Valid(target.Namespace) || strings.TrimSpace(target.Destination) == "" || len(target.Destination) > 512 {
		return ErrTelegramNotificationTargetUnavailable
	}

	return nil
}

func validateTelegramNotificationDocument(document *TelegramDocument, target TelegramNotificationTarget) error {
	if document == nil {
		return nil
	}

	if err := validateTelegramDocument(*document); err != nil {
		return err
	}

	if document.ProjectID != target.ProjectID || document.Namespace != target.Namespace || document.ProjectBindingDigest != target.ProjectBindingDigest {
		return ErrTelegramDocumentUnavailable
	}

	return nil
}

func retainTelegramNotificationDocument(ctx context.Context, tx *sql.Tx, input TelegramNotificationInput, id string) error {
	row := telegramNotificationDocumentRow{id: id, owner: input.Record.Event.ID, target: input.Target, document: *input.Document, createdAt: input.Record.Event.CreatedAt}

	result, err := tx.ExecContext(ctx, insertTelegramNotificationDocumentSQL, row.values()...)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrTelegramDocumentUnavailable
	}

	return nil
}

func purgeTelegramNotificationDocuments(ctx context.Context, tx *sql.Tx, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_notification_documents SET payload=NULL,purged_at=? WHERE id IN (SELECT d.id FROM telegram_notification_documents d JOIN notification_events e ON e.id=d.owner_id WHERE d.payload IS NOT NULL AND d.terminal_at IS NOT NULL AND d.terminal_at<=? AND e.status IN ('delivered','failed') AND e.telegram_uncertain=0 ORDER BY d.terminal_at,d.id LIMIT 20)`, database.FormatTime(now), database.FormatTime(now.Add(-7*24*time.Hour)))

	return err
}

func validTelegramDigest(value string) bool {
	decoded, err := hex.DecodeString(value)

	return err == nil && len(decoded) == 32
}
