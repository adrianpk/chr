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

		return queuePresentation(ctx, tx, presentationReply{Source: update, Message: message, ConfigDigest: configDigest, At: now})
	})
}

func queuePresentation(ctx context.Context, tx *sql.Tx, reply presentationReply) error {
	presentation := tgmessage.Render(reply.Message)
	if presentation.PlainFallback == "" || len(reply.ConfigDigest) != 64 {
		return errors.New("invalid telegram presentation")
	}

	payload, err := json.Marshal(presentation)
	if err != nil {
		return err
	}

	if err := queueResponse(ctx, tx, responseContent{Source: reply.Source, Text: presentation.PlainFallback, ConfigDigest: reply.ConfigDigest, At: reply.At}); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE bot_id=? AND event_key=?", string(payload), contenthash.Sum(payload), reply.Source.Peer.BotID, fmt.Sprintf("update:%d", reply.Source.UpdateID))
	if err != nil {
		return err
	}

	return checkIntakeCapacity(ctx, tx, reply.Source.Peer.BotID)
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

type presentationReply struct {
	Source       ReceivedUpdate
	Message      tgmessage.Blocks
	ConfigDigest string
	At           time.Time
}
