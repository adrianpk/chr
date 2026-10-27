package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
)

// PrepareOperation checkpoints the exact owner input before application.
func (s *ReceiverStore) PrepareOperation(ctx context.Context, request OperationRequest, input json.RawMessage, now time.Time) error {
	if len(input) == 0 || len(input) > 1048576 || !json.Valid(input) {
		return errReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET prepared_input=?,domain_guard=?,state='ready',updated_at=? WHERE id=? AND argument_digest=? AND state='prepared' AND prepared_input IS NULL`, string(input), contenthash.Sum(input), database.FormatTime(now), request.ID, request.ArgumentDigest)
		if err != nil {
			return err
		}

		if err := requireChanged(result); err != nil {
			return err
		}

		return checkIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
	})
}

// BeginOperation persists application intent without spanning the owner call.
func (s *ReceiverStore) BeginOperation(ctx context.Context, request OperationRequest, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_requests SET state='applying',updated_at=? WHERE id=? AND argument_digest=? AND state IN ('ready','applying') AND prepared_input=? AND domain_guard=?`, database.FormatTime(now), request.ID, request.ArgumentDigest, string(request.PreparedInput), contenthash.Sum(request.PreparedInput))
	if err != nil {
		return err
	}

	return requireChanged(result)
}

// FinishOperation records confirmed facts or an explicit failure with the
// source acknowledgement and outbox in one transaction. It never retries a domain effect.
func (s *ReceiverStore) FinishOperation(ctx context.Context, request OperationRequest, completion OperationCompletion) error {
	if !validOperationFinish(completion.State, completion.ReceiptRef, completion.Facts) {
		return errReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return err
		}

		return finishOperation(ctx, tx, request, OperationCompletion{State: completion.State, ReceiptRef: completion.ReceiptRef, Facts: completion.Facts, Text: completion.Text, ConfigDigest: completion.ConfigDigest, At: completion.At})
	})
}

func finishOperation(ctx context.Context, tx *sql.Tx, request OperationRequest, completion OperationCompletion) error {
	result, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET state=?,domain_receipt_ref=NULLIF(?,''),result=?,updated_at=? WHERE id=? AND argument_digest=? AND ((?='confirmed' AND state='applying') OR (?='rejected' AND state='prepared') OR (?='blocked' AND state IN ('prepared','ready','applying'))) `, completion.State, completion.ReceiptRef, string(completion.Facts), database.FormatTime(completion.At), request.ID, request.ArgumentDigest, completion.State, completion.State, completion.State)
	if err != nil {
		return err
	}

	if err := requireChanged(result); err != nil {
		return err
	}

	if err := finishHarnessConversation(ctx, tx, request, completion.State, completion.At); err != nil {
		return err
	}

	if err := queueOperationResult(ctx, tx, operationReply{Request: request, Text: completion.Text, ReceiptRef: completion.ReceiptRef, ConfigDigest: completion.ConfigDigest, At: completion.At}); err != nil {
		return err
	}

	return bindOperationResponse(ctx, tx, request)
}

func validOperationFinish(state OperationState, receipt string, facts json.RawMessage) bool {
	if state != OperationConfirmed && state != OperationRejected && state != OperationBlocked {
		return false
	}

	if state == OperationConfirmed && receipt == "" {
		return false
	}

	return len(receipt) <= 4096 && len(facts) <= 65536 && json.Valid(facts)
}

func bindOperationResponse(ctx context.Context, tx *sql.Tx, request OperationRequest) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_outbox SET request_id=?,conversation_id=NULLIF(?,''),conversation_revision=NULLIF(?,0) WHERE bot_id=? AND event_key=?`, request.ID, request.Target.ConversationID, request.Target.ConversationRevision, request.Source.Peer.BotID, "update:"+strconv.FormatInt(request.Source.UpdateID, 10))
	if err != nil {
		return err
	}

	return nil
}

// OperationCompletion records one guarded owner outcome and its retained reply.
type OperationCompletion struct {
	State              OperationState
	ReceiptRef         string
	Facts              json.RawMessage
	Text, ConfigDigest string
	At                 time.Time
}
