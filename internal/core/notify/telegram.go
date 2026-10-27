package notify

import (
	"context"
	"database/sql"
	"errors"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TelegramNotificationTarget binds the existing configured sender, project root
// and destination. ConfigDigest contains no credential values.
type TelegramNotificationTarget struct {
	ProjectID            string
	Namespace            string
	ProjectBindingDigest string
	Destination          string
	ConfigDigest         string
}

// TelegramNotificationInput freezes one view and optional complete export in the
// notification repository before its existing dispatch owner performs IO.
type TelegramNotificationInput struct {
	Record   NotificationAuditRecord
	Target   TelegramNotificationTarget
	Message  tgmessage.Blocks
	Document *tgmessage.Document
}

// TelegramNotificationDelivery is retained presentation and transport state,
// independent of native work validation or integration receipts.
type TelegramNotificationDelivery struct {
	Presentation    tgmessage.Presentation
	Document        *tgmessage.Document
	Attempts        int
	PlainFallback   bool
	Uncertain       bool
	Acknowledgement tgmessage.Acknowledgement
	Status          NotificationStatus
}

func (r *sqliteNotificationEventRepository) notificationTransaction(ctx context.Context, action func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err = action(tx); err != nil {
		return errors.Join(database.TransactionFailure(ctx, err), database.TransactionFailure(ctx, tx.Rollback()))
	}

	return database.TransactionFailure(ctx, tx.Commit())
}

// ErrTelegramNotificationTargetUnavailable refuses changed or unavailable
// sender, destination, namespace and project-root bindings before transmission.
var ErrTelegramNotificationTargetUnavailable = errors.New("telegram notification target unavailable")
