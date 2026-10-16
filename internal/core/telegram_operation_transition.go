package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"
)

// PrepareTelegramOperation checkpoints the exact owner input before application.
func (s *TelegramReceiverStore) PrepareTelegramOperation(ctx context.Context, request TelegramOperationRequest, input json.RawMessage, now time.Time) error {
	if len(input) == 0 || len(input) > 1048576 || !json.Valid(input) {
		return errTelegramReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET prepared_input=?,domain_guard=?,state='ready',updated_at=? WHERE id=? AND argument_digest=? AND state='prepared' AND prepared_input IS NULL`, string(input), workflowHash(input), formatProjectTime(now), request.ID, request.ArgumentDigest)
		if err != nil {
			return err
		}

		if err := telegramRequireChanged(result); err != nil {
			return err
		}

		return checkTelegramIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
	})
}

// BeginTelegramOperation persists application intent without spanning the owner call.
func (s *TelegramReceiverStore) BeginTelegramOperation(ctx context.Context, request TelegramOperationRequest, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_requests SET state='applying',updated_at=? WHERE id=? AND argument_digest=? AND state IN ('ready','applying') AND prepared_input=? AND domain_guard=?`, formatProjectTime(now), request.ID, request.ArgumentDigest, string(request.PreparedInput), workflowHash(request.PreparedInput))
	if err != nil {
		return err
	}

	return telegramRequireChanged(result)
}

// FinishTelegramOperation records confirmed facts or an explicit failure with the
// source acknowledgement and outbox in one transaction. It never retries a domain effect.
func (s *TelegramReceiverStore) FinishTelegramOperation(ctx context.Context, request TelegramOperationRequest, state, receipt string, facts json.RawMessage, text, configDigest string, now time.Time) error {
	if !validTelegramOperationFinish(state, receipt, facts) {
		return errTelegramReceiverConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, request.Source)
		if err != nil || !pending {
			return err
		}

		return finishTelegramOperation(ctx, tx, request, state, receipt, facts, text, configDigest, now)
	})
}

func finishTelegramOperation(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, state, receipt string, facts json.RawMessage, text, configDigest string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE telegram_requests SET state=?,domain_receipt_ref=NULLIF(?,''),result=?,updated_at=? WHERE id=? AND argument_digest=? AND ((?='confirmed' AND state='applying') OR (?='rejected' AND state='prepared') OR (?='blocked' AND state IN ('prepared','ready','applying'))) `, state, receipt, string(facts), formatProjectTime(now), request.ID, request.ArgumentDigest, state, state, state)
	if err != nil {
		return err
	}

	if err := telegramRequireChanged(result); err != nil {
		return err
	}

	if err := queueTelegramReferencedResponse(ctx, tx, request.Source, text, request.ID+" "+receipt, configDigest, now); err != nil {
		return err
	}

	return bindTelegramOperationResponse(ctx, tx, request)
}

func validTelegramOperationFinish(state, receipt string, facts json.RawMessage) bool {
	if state != "confirmed" && state != "rejected" && state != "blocked" {
		return false
	}

	if state == "confirmed" && receipt == "" {
		return false
	}

	return len(receipt) <= 4096 && len(facts) <= 65536 && json.Valid(facts)
}

func bindTelegramOperationResponse(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest) error {
	_, err := tx.ExecContext(ctx, `UPDATE telegram_outbox SET request_id=?,conversation_id=NULLIF(?,''),conversation_revision=NULLIF(?,0) WHERE bot_id=? AND event_key=?`, request.ID, request.Target.ConversationID, request.Target.ConversationRevision, request.Source.Peer.BotID, "update:"+strconv.FormatInt(request.Source.UpdateID, 10))
	if err != nil {
		return err
	}

	return nil
}
