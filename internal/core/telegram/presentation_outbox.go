package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// QueuePresentationResponse freezes one bounded message and acknowledges its
// source atomically. Replayed sources cannot replace the original rendering.
func (s *ReceiverStore) QueuePresentationResponse(ctx context.Context, update ReceivedUpdate, message tgmessage.Blocks, configDigest string, now time.Time) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, update)
		if err != nil || !pending {
			return err
		}

		return queuePresentation(ctx, tx, update, message, configDigest, now)
	})
}

func queuePresentation(ctx context.Context, tx *sql.Tx, update ReceivedUpdate, message tgmessage.Blocks, configDigest string, now time.Time) error {
	presentation := tgmessage.Render(message)
	if presentation.PlainFallback == "" || len(configDigest) != 64 {
		return errors.New("invalid telegram presentation")
	}

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	if err := queueResponse(ctx, tx, update, presentation.PlainFallback, configDigest, now); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE bot_id=? AND event_key=?", string(payload), contenthash.Sum(payload), update.Peer.BotID, fmt.Sprintf("update:%d", update.UpdateID))
	if err != nil {
		return err
	}

	return checkIntakeCapacity(ctx, tx, update.Peer.BotID)
}

func decodePresentationHash(payload, hash string, delivery *ReceiverDelivery) error {
	presentation, err := tgmessage.DecodeRetained(tgmessage.RetainedInput{Payload: payload, Hash: hash, Text: delivery.Text, DocumentID: delivery.DocumentID})
	if err != nil {
		return err
	}

	if presentation != nil {
		delivery.Presentation = presentation
		delivery.PlainText = presentation.PlainFallback
	}

	return nil
}
