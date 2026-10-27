package telegram

import (
	"context"
	"database/sql"
	"errors"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func queueOperationResult(ctx context.Context, tx *sql.Tx, reply operationReply) error {
	if reply.Request.Delivery == nil {
		if reply.Request.Message != nil {
			return queuePresentation(ctx, tx, presentationReply{Source: reply.Request.Source, Message: *reply.Request.Message, ConfigDigest: reply.ConfigDigest, At: reply.At})
		}

		return queueReferencedResponse(ctx, tx, ReferencedReply{Source: reply.Request.Source, Text: reply.Text, Reference: reply.Request.ID + " " + reply.ReceiptRef, ConfigDigest: reply.ConfigDigest, At: reply.At})
	}

	if _, err := tx.ExecContext(ctx, "SAVEPOINT telegram_document_result"); err != nil {
		return err
	}

	err := queueDocument(ctx, tx, documentReply{Source: reply.Request.Source, Response: *reply.Request.Delivery, ConfigDigest: reply.ConfigDigest, At: reply.At})
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

	if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		return err
	}

	return queuePresentation(ctx, tx, presentationReply{Source: reply.Request.Source, Message: tgmessage.Blocks{InitialText: reply.Text, SecondaryText: "The complete document could not be attached: its image is invalid or document storage is full. The result is saved. Use local project commands to inspect it."}, ConfigDigest: reply.ConfigDigest, At: reply.At})
}

type operationReply struct {
	Request                        OperationRequest
	Text, ReceiptRef, ConfigDigest string
	At                             time.Time
}
