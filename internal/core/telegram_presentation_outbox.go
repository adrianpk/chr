package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"
)

// QueuePresentationResponse freezes one bounded message and acknowledges its
// source atomically. Replayed sources cannot replace the original rendering.
func (s *TelegramReceiverStore) QueuePresentationResponse(ctx context.Context, update TelegramReceivedUpdate, message TelegramMessageBlocks, configDigest string, now time.Time) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, update)
		if err != nil || !pending {
			return err
		}

		return queueTelegramPresentation(ctx, tx, update, message, configDigest, now)
	})
}

func queueTelegramPresentation(ctx context.Context, tx *sql.Tx, update TelegramReceivedUpdate, message TelegramMessageBlocks, configDigest string, now time.Time) error {
	presentation := RenderTelegramMessage(message)
	if presentation.PlainFallback == "" || len(configDigest) != 64 {
		return errors.New("invalid telegram presentation")
	}

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	if err := queueTelegramResponse(ctx, tx, update, presentation.PlainFallback, configDigest, now); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE bot_id=? AND event_key=?", string(payload), workflowHash(payload), update.Peer.BotID, fmt.Sprintf("update:%d", update.UpdateID))
	if err != nil {
		return err
	}

	return checkTelegramIntakeCapacity(ctx, tx, update.Peer.BotID)
}

func decodeTelegramPresentation(payload string, delivery *TelegramReceiverDelivery) error {
	if payload == "" {
		return nil
	}

	var presentation TelegramPresentation
	if err := json.Unmarshal([]byte(payload), &presentation); err != nil {
		return err
	}

	if presentation.Version != 1 || (presentation.TemplateRevision != "text-v1" && presentation.TemplateRevision != "document-v1") {
		return errors.New("unsupported telegram presentation version")
	}

	rendered := RenderTelegramMessage(presentation.TelegramMessageBlocks)
	if presentation.TemplateRevision == "document-v1" {
		rendered = RenderTelegramDocument(presentation.TelegramMessageBlocks)
	}

	if rendered.PlainFallback != presentation.PlainFallback || delivery.Text != presentation.PlainFallback || !slices.Equal(rendered.Entities, presentation.Entities) {
		return errors.New("invalid retained telegram presentation")
	}

	delivery.Presentation = &presentation
	delivery.PlainText = presentation.PlainFallback

	return nil
}

func decodeTelegramPresentationHash(payload, hash string, delivery *TelegramReceiverDelivery) error {
	if hash != "" && workflowHash([]byte(payload)) != hash {
		return errors.New("retained telegram presentation hash mismatch")
	}

	if err := decodeTelegramPresentation(payload, delivery); err != nil {
		return err
	}

	if delivery.DocumentID != "" && (delivery.Presentation == nil || delivery.Presentation.DocumentID != delivery.DocumentID || hash == "") {
		return ErrTelegramDocumentUnavailable
	}

	return nil
}
