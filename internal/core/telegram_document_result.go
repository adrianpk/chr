package core

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func queueTelegramOperationResult(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, text, receipt, configDigest string, now time.Time) error {
	if request.Delivery == nil {
		if request.Message != nil {
			return queueTelegramPresentation(ctx, tx, request.Source, *request.Message, configDigest, now)
		}

		return queueTelegramReferencedResponse(ctx, tx, request.Source, text, request.ID+" "+receipt, configDigest, now)
	}

	if _, err := tx.ExecContext(ctx, "SAVEPOINT telegram_document_result"); err != nil {
		return err
	}

	err := queueTelegramDocument(ctx, tx, request.Source, *request.Delivery, configDigest, now)
	if err == nil {
		_, err = tx.ExecContext(ctx, "RELEASE telegram_document_result")

		return err
	}
	// Presentation admission can fail after the domain owner has committed. Keep
	// that receipt and project only a truthful text result; never repeat its effect.
	if _, rollbackErr := tx.ExecContext(ctx, "ROLLBACK TO telegram_document_result"); rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	if _, releaseErr := tx.ExecContext(ctx, "RELEASE telegram_document_result"); releaseErr != nil {
		return errors.Join(err, releaseErr)
	}

	if !errors.Is(err, ErrTelegramDocumentUnavailable) {
		return err
	}

	return queueTelegramPresentation(ctx, tx, request.Source, TelegramMessageBlocks{InitialText: text, SecondaryText: "The complete document could not be attached: its image is invalid or document storage is full. The result is saved. Use local project commands to inspect it."}, configDigest, now)
}
