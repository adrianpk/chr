package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// QueueFormattedReadResponse freezes both help renderings with source acknowledgement.
// Existing plain responses and replayed sources keep their original contents.
func (s *ReceiverStore) QueueFormattedReadResponse(ctx context.Context, update ReceivedUpdate, text, plain, configDigest string, now time.Time) error {
	if !validReceiverText(text) || !validReceiverText(plain) || len(tgmessage.SplitResponse(text)) != 1 || len(tgmessage.SplitResponse(plain)) != 1 || len(configDigest) != 64 {
		return errors.New("formatted telegram reply exceeds its single-message bounds")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, update)
		if err != nil || !pending {
			return err
		}

		key := fmt.Sprintf("update:%d", update.UpdateID)
		if err := queueResponseParts(ctx, tx, update, []string{text}, configDigest, now); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "UPDATE telegram_outbox SET parse_mode='Markdown',plain_text=? WHERE bot_id=? AND event_key=?", plain, update.Peer.BotID, key)
		if err != nil {
			return err
		}

		return checkIntakeCapacity(ctx, tx, update.Peer.BotID)
	})
}

// SelectPlainFallback retains one definite format rejection before another
// send can be reserved. It preserves response identity and all attempt counters.
func (s *ReceiverStore) SelectPlainFallback(ctx context.Context, id string, now time.Time) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_outbox SET plain_fallback=CASE WHEN parse_mode='Markdown' THEN 1 ELSE plain_fallback END,presentation_fallback=CASE WHEN presentation IS NOT NULL THEN 1 ELSE presentation_fallback END,state='pending',retry_at=NULL,last_error_code='format_rejected',updated_at=?
WHERE id=? AND state='sending' AND (parse_mode='Markdown' OR presentation IS NOT NULL) AND plain_fallback=0 AND presentation_fallback=0 AND attempts<5`, database.FormatTime(now), id)
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()

	return count == 1, err
}
