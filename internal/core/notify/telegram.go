package notify

import (
	"context"
	"database/sql"
	"errors"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TelegramTarget binds the existing configured sender, project root
// and destination. ConfigDigest contains no credential values.
type TelegramTarget struct {
	ProjectID            string
	Namespace            string
	ProjectBindingDigest string
	Destination          string
	ConfigDigest         string
}

// TelegramInput freezes one view and optional complete export in the
// notification repository before its existing dispatch owner performs IO.
type TelegramInput struct {
	Record   AuditRecord
	Target   TelegramTarget
	Message  tgmessage.Blocks
	Document *tgmessage.Document
}

// TelegramDelivery is retained presentation and transport state,
// independent of native work validation or integration receipts.
type TelegramDelivery struct {
	Presentation    tgmessage.Presentation
	Document        *tgmessage.Document
	Attempts        int
	PlainFallback   bool
	Uncertain       bool
	Acknowledgement tgmessage.Acknowledgement
	Status          Status
}

func (r *sqliteRepository) transaction(ctx context.Context, action func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err = action(tx); err != nil {
		return errors.Join(database.TransactionFailure(ctx, err), database.TransactionFailure(ctx, tx.Rollback()))
	}

	return database.TransactionFailure(ctx, tx.Commit())
}

// ErrTelegramTargetUnavailable refuses changed or unavailable
// sender, destination, namespace and project-root bindings before transmission.
var ErrTelegramTargetUnavailable = errors.New("telegram notification target unavailable")
