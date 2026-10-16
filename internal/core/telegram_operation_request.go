package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TelegramOperationRequest retains one complete transport binding before dispatch.
// PreparedInput is the caller's serialized, frozen shared-owner input.
type TelegramOperationRequest struct {
	ID             string
	Source         TelegramReceivedUpdate
	Ordinal        int
	Operator       string
	Target         TelegramOperationTarget
	OperationID    string
	Arguments      json.RawMessage
	ArgumentDigest string
	PreparedInput  json.RawMessage
	State          string
	ReceiptRef     string
	Result         json.RawMessage
}

const telegramOperationColumns = `r.id,r.operation_ordinal,r.operator_id,r.project_id,r.project_binding_digest,r.namespace,COALESCE(r.conversation_id,''),COALESCE(r.expected_conversation_revision,0),r.operation_id,r.arguments,r.argument_digest,COALESCE(r.prepared_input,''),r.domain_guard,r.state,COALESCE(r.domain_receipt_ref,''),COALESCE(r.result,'')`

// TelegramOperation reads an exact retained source binding, without resolving a new target.
func (s *TelegramReceiverStore) TelegramOperation(ctx context.Context, update TelegramReceivedUpdate, ordinal int) (TelegramOperationRequest, bool, error) {
	request, err := scanTelegramOperation(s.db.QueryRowContext(ctx, `SELECT `+telegramOperationColumns+` FROM telegram_requests r JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id WHERE r.bot_id=? AND r.update_id=? AND r.operation_ordinal=? AND u.payload_digest=? AND u.chat_id=? AND u.topic_id=? AND u.user_id=?`, update.Peer.BotID, update.UpdateID, ordinal, update.Digest, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return request, false, nil
	}

	request.Source = update

	return request, err == nil, err
}

func scanTelegramOperation(row interface{ Scan(...any) error }) (TelegramOperationRequest, error) {
	var (
		request                            TelegramOperationRequest
		arguments, prepared, result, guard string
	)

	err := row.Scan(&request.ID, &request.Ordinal, &request.Operator, &request.Target.ProjectID, &request.Target.ProjectBindingDigest, &request.Target.Namespace, &request.Target.ConversationID, &request.Target.ConversationRevision, &request.OperationID, &arguments, &request.ArgumentDigest, &prepared, &guard, &request.State, &request.ReceiptRef, &result)

	request.Arguments, request.PreparedInput, request.Result = json.RawMessage(arguments), json.RawMessage(prepared), json.RawMessage(result)
	if err == nil && prepared != "" && workflowHash([]byte(prepared)) != guard {
		return request, errTelegramReceiverConflict
	}

	return request, err
}

// ReserveTelegramOperation permanently binds normalized arguments, including failed attempts.
func (s *TelegramReceiverStore) ReserveTelegramOperation(ctx context.Context, request TelegramOperationRequest, now time.Time) (TelegramOperationRequest, error) {
	request, err := normalizeTelegramOperation(request)
	if err != nil {
		return request, err
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := telegramReadResponsePending(ctx, tx, request.Source)
		if err != nil {
			return err
		}

		if !pending {
			return errTelegramReceiverConflict
		}

		if err := insertTelegramOperation(ctx, tx, request, now); err != nil {
			return err
		}

		return checkTelegramIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
	})
	if err != nil {
		return request, err
	}

	retained, found, err := s.TelegramOperation(ctx, request.Source, request.Ordinal)
	if err != nil {
		return request, err
	}

	if !found || retained.ArgumentDigest != request.ArgumentDigest || retained.Operator != request.Operator {
		return request, errTelegramReceiverConflict
	}

	return retained, nil
}

func normalizeTelegramOperation(request TelegramOperationRequest) (TelegramOperationRequest, error) {
	if !validTelegramOperationEnvelope(request) {
		return request, errTelegramReceiverConflict
	}

	if err := validateWorkflowMutationBinding(WorkflowMutationBinding{RequestID: "request", Operator: request.Operator}); err != nil {
		return request, err
	}

	if len(request.Arguments) > 65536 {
		return request, errTelegramReceiverConflict
	}

	var arguments any
	if err := json.Unmarshal(request.Arguments, &arguments); err != nil {
		return request, err
	}

	normalized, err := json.Marshal(arguments)
	if err != nil {
		return request, err
	}

	request.Arguments = normalized

	binding, err := json.Marshal(struct {
		Target              TelegramOperationTarget
		Operation, Operator string
		Arguments           json.RawMessage
	}{request.Target, request.OperationID, request.Operator, request.Arguments})
	if err != nil {
		return request, err
	}

	sum := sha256.Sum256(binding)
	request.ArgumentDigest = hex.EncodeToString(sum[:])
	id := sha256.Sum256(fmt.Appendf(nil, "%d:%d:%d", request.Source.Peer.BotID, request.Source.UpdateID, request.Ordinal))
	request.ID = "tgr-" + hex.EncodeToString(id[:16])

	return request, nil
}

func validTelegramOperationEnvelope(request TelegramOperationRequest) bool {
	if request.Ordinal < 0 || request.Ordinal > 7 || request.OperationID == "" || len(request.OperationID) > 128 {
		return false
	}

	return IsValidOpsNamespace(request.Target.Namespace) && request.Target.ProjectID != "" && len(request.Target.ProjectBindingDigest) == 64
}

func insertTelegramOperation(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO telegram_requests(id,bot_id,update_id,operation_ordinal,operator_id,project_id,project_binding_digest,namespace,conversation_id,expected_conversation_revision,operation_id,arguments,argument_digest,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,0),?,?,?,'prepared',?,?) ON CONFLICT(bot_id,update_id,operation_ordinal) DO NOTHING`, request.ID, request.Source.Peer.BotID, request.Source.UpdateID, request.Ordinal, request.Operator, request.Target.ProjectID, request.Target.ProjectBindingDigest, request.Target.Namespace, request.Target.ConversationID, request.Target.ConversationRevision, request.OperationID, string(request.Arguments), request.ArgumentDigest, formatProjectTime(now), formatProjectTime(now))
	if err != nil {
		return err
	}

	return nil
}
