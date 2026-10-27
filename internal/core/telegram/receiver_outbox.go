package telegram

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// ReceiverDelivery retains a response independently of its send outcome.
type ReceiverDelivery struct {
	ID            string
	Peer          PeerIdentity
	Text          string
	State         string
	Attempts      int
	ConfigDigest  string
	ActionToken   string
	ActionChoices []string
	ParseMode     string
	PlainText     string
	PlainFallback bool
	DocumentID    string
	Presentation  *tgmessage.Presentation
}

// QueueReadResponse records a response and acknowledges its retained input atomically.
// Replaying an already processed source never replaces or duplicates its response.
func (s *ReceiverStore) QueueReadResponse(ctx context.Context, update ReceivedUpdate, text, configDigest string, now time.Time) error {
	return s.QueueReferencedResponse(ctx, ReferencedReply{Source: update, Text: text, Reference: fmt.Sprintf("update:%d", update.UpdateID), ConfigDigest: configDigest, At: now})
}

// QueueReferencedResponse retains bounded output with a canonical or paged reference.
func (s *ReceiverStore) QueueReferencedResponse(ctx context.Context, reply ReferencedReply) error {
	if reply.Text == "" || len(reply.ConfigDigest) != 64 {
		return errors.New("telegram response exceeds its Unicode bounds")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pendingSource, err := readResponsePending(ctx, tx, reply.Source)
		if err != nil || !pendingSource {
			return err
		}

		return queueReferencedResponse(ctx, tx, ReferencedReply{Source: reply.Source, Text: reply.Text, Reference: reply.Reference, ConfigDigest: reply.ConfigDigest, At: reply.At})
	})
}

func queueResponse(ctx context.Context, tx *sql.Tx, reply responseContent) error {
	return queueReferencedResponse(ctx, tx, ReferencedReply{Source: reply.Source, Text: reply.Text, Reference: fmt.Sprintf("update:%d", reply.Source.UpdateID), ConfigDigest: reply.ConfigDigest, At: reply.At})
}

func queueReferencedResponse(ctx context.Context, tx *sql.Tx, reply ReferencedReply) error {
	return queueResponseParts(ctx, tx, responseParts{Source: reply.Source, Parts: tgmessage.ResponseParts(reply.Text, reply.Reference), ConfigDigest: reply.ConfigDigest, At: reply.At})
}

func queueResponseParts(ctx context.Context, tx *sql.Tx, reply responseParts) error {
	var pending int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM telegram_outbox WHERE bot_id=? AND state!='delivered'", reply.Source.Peer.BotID).Scan(&pending); err != nil {
		return err
	}

	if pending+len(reply.Parts) > 1024 {
		return errors.New("telegram outbox capacity exhausted")
	}

	result, err := tx.ExecContext(ctx, "UPDATE telegram_updates SET state='processed',updated_at=? WHERE bot_id=? AND update_id=? AND payload_digest=? AND state='pending'", database.FormatTime(reply.At), reply.Source.Peer.BotID, reply.Source.UpdateID, reply.Source.Digest)
	if err != nil {
		return err
	}

	if err := requireChanged(result); err != nil {
		return err
	}

	key := fmt.Sprintf("update:%d", reply.Source.UpdateID)
	for index, part := range reply.Parts {
		sum := sha256.Sum256(fmt.Appendf(nil, "%d:%s:%d", reply.Source.Peer.BotID, key, index+1))
		id := "tgd-" + hex.EncodeToString(sum[:16])

		row := responsePartRow{id: id, peer: reply.Source.Peer, key: key, configDigest: reply.ConfigDigest, text: part, part: index + 1, parts: len(reply.Parts), at: reply.At}

		_, err = tx.ExecContext(ctx, insertResponsePartSQL, row.values()...)
		if err != nil {
			return err
		}
	}

	return checkIntakeCapacity(ctx, tx, reply.Source.Peer.BotID)
}

func readResponsePending(ctx context.Context, tx *sql.Tx, update ReceivedUpdate) (bool, error) {
	if _, err := tx.ExecContext(ctx, "UPDATE telegram_updates SET updated_at=updated_at WHERE bot_id=? AND update_id=?", update.Peer.BotID, update.UpdateID); err != nil {
		return false, err
	}

	var digest, state string
	if err := tx.QueryRowContext(ctx, "SELECT payload_digest,state FROM telegram_updates WHERE bot_id=? AND update_id=?", update.Peer.BotID, update.UpdateID).Scan(&digest, &state); err != nil {
		return false, err
	}

	if digest != update.Digest {
		return false, errReceiverConflict
	}

	if state == "processed" {
		return false, nil
	}

	if state != "pending" {
		return false, errors.New("telegram source cannot produce a read response")
	}

	return true, nil
}

// RejectPendingUpdate discards a queued body when current authority is revoked.
func (s *ReceiverStore) RejectPendingUpdate(ctx context.Context, update ReceivedUpdate, now time.Time) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if err := revokePeerActions(ctx, tx, update.Peer, now); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, `UPDATE telegram_updates SET state='rejected',payload='',reason_code='authorization_revoked',updated_at=? WHERE bot_id=? AND update_id=? AND state='pending'`, database.FormatTime(now), update.Peer.BotID, update.UpdateID)

		return err
	})
}

// DueDeliveries is bounded; sending rows after a crash retain uncertainty and budgets.
func (s *ReceiverStore) DueDeliveries(ctx context.Context, botID int64, now time.Time) ([]ReceiverDelivery, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+receiverDeliveryColumns+` FROM telegram_outbox o
WHERE o.bot_id=? AND o.state IN ('pending','sending','uncertain') AND o.attempts<5 AND (o.retry_at IS NULL OR o.retry_at<=?)
AND NOT EXISTS(SELECT 1 FROM telegram_outbox earlier WHERE earlier.bot_id=o.bot_id AND earlier.chat_id=o.chat_id AND earlier.topic_id=o.topic_id AND earlier.user_id=o.user_id AND earlier.event_key=o.event_key AND earlier.part_index<o.part_index AND earlier.state!='delivered')
AND (o.kind!='progress' OR o.attempts>0 OR NOT EXISTS(SELECT 1 FROM telegram_outbox recent WHERE recent.conversation_id=o.conversation_id AND recent.kind='progress' AND recent.total_attempts>0 AND recent.event_key!=o.event_key AND recent.updated_at>?))
ORDER BY CASE (SELECT u.processing_lane FROM telegram_updates u WHERE u.bot_id=o.bot_id AND u.update_id=CAST(substr(o.event_key,8) AS INTEGER) AND o.event_key='update:'||u.update_id) WHEN 'urgent' THEN 0 WHEN 'decision' THEN 1 ELSE 2 END,o.created_at,o.rowid LIMIT 50`, botID, database.FormatTime(now), database.FormatTime(now.Add(-30*time.Second)))
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck // rows.Err is checked after the bounded iteration.
		_ = rows.Close()
	}()

	var result []ReceiverDelivery

	for rows.Next() {
		row := receiverDeliveryRow{delivery: ReceiverDelivery{Peer: PeerIdentity{BotID: botID}}}
		if err := rows.Scan(row.scanTargets()...); err != nil {
			return nil, err
		}

		delivery := row.delivery
		choices, presentation, presentationHash := row.choices, row.presentation, row.presentationHash

		if err := json.Unmarshal([]byte(choices), &delivery.ActionChoices); err != nil {
			return nil, err
		}

		if err := decodePresentationHash(presentation, presentationHash, &delivery); err != nil {
			return nil, err
		}

		result = append(result, delivery)
	}

	return result, rows.Err()
}

// ReserveDelivery persists the attempt before sendMessage; a lost acknowledgement cannot reset it.
func (s *ReceiverStore) ReserveDelivery(ctx context.Context, id string, now time.Time) error {
	reserved, err := s.TryReserveDelivery(ctx, id, now)
	if err != nil {
		return err
	}

	if !reserved {
		return errors.New("telegram retry is not due, blocked or exhausted")
	}

	return nil
}

// TryReserveDelivery skips a stale read when routine progress was coalesced.
// Successful reservation persists its attempt before the API call.
func (s *ReceiverStore) TryReserveDelivery(ctx context.Context, id string, now time.Time) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_outbox AS o SET state='sending',attempts=attempts+1,total_attempts=total_attempts+1,
retry_at=?,updated_at=? WHERE id=? AND state IN ('pending','sending','uncertain') AND attempts<5 AND (retry_at IS NULL OR retry_at<=?)
AND (o.kind!='progress' OR o.attempts>0 OR NOT EXISTS(SELECT 1 FROM telegram_outbox recent WHERE recent.conversation_id=o.conversation_id AND recent.kind='progress' AND recent.total_attempts>0 AND recent.event_key!=o.event_key AND recent.updated_at>?))`,
		database.FormatTime(now.Add(16*time.Second)), database.FormatTime(now), id, database.FormatTime(now), database.FormatTime(now.Add(-30*time.Second)))
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()

	return count == 1, err
}

// DeliverySent records definite API acknowledgement without applying any operation again.
func (s *ReceiverStore) DeliverySent(ctx context.Context, id string, messageID int64, now time.Time) error {
	return s.deliverySent(ctx, id, messageID, "", now)
}

// DeliveryDocumentSent retains the original upload acknowledgement without
// changing its immutable source image or allocating another operation.
func (s *ReceiverStore) DeliveryDocumentSent(ctx context.Context, id string, ack tgmessage.Acknowledgement, now time.Time) error {
	if ack.FileID == "" || len(ack.FileID) > 256 {
		return errors.New("telegram document requires a definite file identity")
	}

	return s.deliverySent(ctx, id, ack.MessageID, ack.FileID, now)
}

func (s *ReceiverStore) deliverySent(ctx context.Context, id string, messageID int64, fileID string, now time.Time) error {
	if messageID <= 0 {
		return errors.New("telegram delivery requires a definite message ID")
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, "UPDATE telegram_outbox SET state='delivered',telegram_message_id=?,telegram_file_id=NULLIF(?,''),retry_at=NULL,last_error_code='',updated_at=? WHERE id=? AND state='sending' AND ((document_id IS NULL AND ?='') OR (document_id IS NOT NULL AND ?!=''))", messageID, fileID, database.FormatTime(now), id, fileID, fileID)
		if err != nil {
			return err
		}

		if err := requireChanged(result); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE telegram_document_snapshots SET terminal_at=? WHERE owner_id=? AND terminal_at IS NULL`, database.FormatTime(now), id); err != nil {
			return err
		}

		if err := retainReply(ctx, tx, id, messageID, now); err != nil {
			return err
		}

		return refreshReplyLanes(ctx, tx, id, messageID)
	})
}

// DeliveryFailed retains uncertainty or blocks a permanent/exhausted failure.
func (s *ReceiverStore) DeliveryFailed(ctx context.Context, delivery ReceiverDelivery, failure error, now time.Time) error {
	policy := failureState(failure, delivery.Attempts+1)
	code, permanent, delay := policy.Code, policy.Permanent, policy.Delay
	state := "pending"

	var apiErr *APIError
	if errors.As(failure, &apiErr) && apiErr.Uncertain {
		state = "uncertain"
	}

	if permanent || delivery.Attempts+1 >= 5 {
		state = "blocked"
	}

	_, err := s.db.ExecContext(ctx, "UPDATE telegram_outbox SET state=?,last_error_code=?,retry_at=?,updated_at=? WHERE id=? AND state='sending'", state, code, database.FormatTime(now.Add(delay)), database.FormatTime(now), delivery.ID)

	return err
}

// BlockDelivery prevents delivery to a revoked peer without creating another request.
func (s *ReceiverStore) BlockDelivery(ctx context.Context, id string, now time.Time) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "UPDATE telegram_outbox SET state='blocked',last_error_code='authorization_revoked',updated_at=? WHERE id=? AND state!='delivered'", database.FormatTime(now), id); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, "UPDATE telegram_document_snapshots SET terminal_at=? WHERE owner_id=? AND terminal_at IS NULL", database.FormatTime(now), id)

		return err
	})
}

// ReferencedReply retains one peer-bound response with its paging/canonical reference.
type ReferencedReply struct {
	Source                        ReceivedUpdate
	Text, Reference, ConfigDigest string
	At                            time.Time
}

type responseContent struct {
	Source             ReceivedUpdate
	Text, ConfigDigest string
	At                 time.Time
}

type responseParts struct {
	Source       ReceivedUpdate
	Parts        []string
	ConfigDigest string
	At           time.Time
}
