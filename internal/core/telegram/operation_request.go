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

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OperationRequest retains one complete transport binding before dispatch.
// PreparedInput is the caller's serialized, frozen shared-owner input.
type OperationRequest struct {
	// Delivery and Message are response-only data, outside operation authority.
	Delivery       *tgmessage.DocumentResponse `json:"-"`
	Message        *tgmessage.Blocks           `json:"-"`
	ID             string
	Source         ReceivedUpdate
	Ordinal        int
	Operator       string
	Target         OperationTarget
	OperationID    string
	Arguments      json.RawMessage
	ArgumentDigest string
	PreparedInput  json.RawMessage
	State          OperationState
	ReceiptRef     string
	Result         json.RawMessage
}

const operationColumns = `r.id,r.operation_ordinal,r.operator_id,r.project_id,r.project_binding_digest,r.namespace,COALESCE(r.conversation_id,''),COALESCE(r.expected_conversation_revision,0),r.operation_id,r.arguments,r.argument_digest,COALESCE(r.prepared_input,''),r.domain_guard,r.state,COALESCE(r.domain_receipt_ref,''),COALESCE(r.result,'')`

// Operation reads an exact retained source binding, without resolving a new target.
func (s *ReceiverStore) Operation(ctx context.Context, update ReceivedUpdate, ordinal int) (OperationLookup, error) {
	request, err := scanOperation(s.db.QueryRowContext(ctx, `SELECT `+operationColumns+` FROM telegram_requests r JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id WHERE r.bot_id=? AND r.update_id=? AND r.operation_ordinal=? AND u.payload_digest=? AND u.chat_id=? AND u.topic_id=? AND u.user_id=?`, update.Peer.BotID, update.UpdateID, ordinal, update.Digest, update.Peer.ChatID, update.Peer.TopicID, update.Peer.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return OperationLookup{Request: request, Found: false}, nil
	}

	request.Source = update

	return OperationLookup{Request: request, Found: err == nil}, err
}

func scanOperation(row interface{ Scan(...any) error }) (OperationRequest, error) {
	var (
		request                            OperationRequest
		arguments, prepared, result, guard string
	)

	err := row.Scan(&request.ID, &request.Ordinal, &request.Operator, &request.Target.ProjectID, &request.Target.ProjectBindingDigest, &request.Target.Namespace, &request.Target.ConversationID, &request.Target.ConversationRevision, &request.OperationID, &arguments, &request.ArgumentDigest, &prepared, &guard, &request.State, &request.ReceiptRef, &result)

	request.Arguments, request.PreparedInput, request.Result = json.RawMessage(arguments), json.RawMessage(prepared), json.RawMessage(result)
	if err == nil && prepared != "" && contenthash.Sum([]byte(prepared)) != guard {
		return request, errReceiverConflict
	}

	return request, err
}

// ReserveOperation permanently binds normalized arguments, including failed attempts.
func (s *ReceiverStore) ReserveOperation(ctx context.Context, request OperationRequest, now time.Time) (OperationRequest, error) {
	request, err := normalizeOperation(request)
	if err != nil {
		return request, err
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, request.Source)
		if err != nil {
			return err
		}

		if !pending {
			return errReceiverConflict
		}

		if err := insertOperation(ctx, tx, request, now); err != nil {
			return err
		}

		return checkIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
	})
	if err != nil {
		return request, err
	}

	operationLookup, err := s.Operation(ctx, request.Source, request.Ordinal)
	retained, found := operationLookup.Request, operationLookup.Found

	if err != nil {
		return request, err
	}

	if !found || retained.ArgumentDigest != request.ArgumentDigest || retained.Operator != request.Operator {
		return request, errReceiverConflict
	}

	return retained, nil
}

func normalizeOperation(request OperationRequest) (OperationRequest, error) {
	if !validOperationEnvelope(request) {
		return request, errReceiverConflict
	}

	if err := (workflowstate.MutationBinding{RequestID: "request", Operator: request.Operator}).Validate(); err != nil {
		return request, err
	}

	if len(request.Arguments) > 65536 {
		return request, errReceiverConflict
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
		Target              OperationTarget
		Operation, Operator string
		Arguments           json.RawMessage
	}{request.Target, request.OperationID, request.Operator, request.Arguments})
	if err != nil {
		return request, err
	}

	sum := sha256.Sum256(binding)
	request.ArgumentDigest = hex.EncodeToString(sum[:])
	request.ID = OperationID(request.Source, request.Ordinal)

	return request, nil
}

// OperationID derives the opaque request binding before serializing its
// owner input. Reservation still validates and binds the complete source/target.
func OperationID(source ReceivedUpdate, ordinal int) string {
	id := sha256.Sum256(fmt.Appendf(nil, "%d:%d:%d", source.Peer.BotID, source.UpdateID, ordinal))

	return "tgr-" + hex.EncodeToString(id[:16])
}

// ReservePreparedOperation retains complete neutral owner input and its
// immutable transport binding in one transaction, before any provider call.
func (s *ReceiverStore) ReservePreparedOperation(ctx context.Context, request OperationRequest, input json.RawMessage, now time.Time) (OperationRequest, error) {
	request, err := normalizeOperation(request)
	if err != nil {
		return request, err
	}

	if len(input) == 0 || len(input) > 1048576 || !json.Valid(input) {
		return request, workflowstate.ErrConflict
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		return reservePreparedOperation(ctx, tx, request, input, now)
	})
	if err != nil {
		return request, err
	}

	operationLookup2, err := s.Operation(ctx, request.Source, request.Ordinal)
	retained, found := operationLookup2.Request, operationLookup2.Found

	if err != nil {
		return request, err
	}

	if !found || retained.ArgumentDigest != request.ArgumentDigest || string(retained.PreparedInput) != string(input) {
		return request, workflowstate.ErrConflict
	}

	return retained, nil
}

func reservePreparedOperation(ctx context.Context, tx *sql.Tx, request OperationRequest, input json.RawMessage, now time.Time) error {
	pending, err := readResponsePending(ctx, tx, request.Source)
	if err != nil || !pending {
		return errors.Join(err, workflowstate.ErrConflict)
	}

	if err := insertOperation(ctx, tx, request, now); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE telegram_requests SET prepared_input=?,domain_guard=?,state='ready',updated_at=? WHERE id=? AND argument_digest=? AND state='prepared' AND prepared_input IS NULL`, string(input), contenthash.Sum(input), database.FormatTime(now), request.ID, request.ArgumentDigest)
	if err != nil {
		return err
	}

	return checkIntakeCapacity(ctx, tx, request.Source.Peer.BotID)
}

func validOperationEnvelope(request OperationRequest) bool {
	if request.Ordinal < 0 || request.Ordinal > 7 || request.OperationID == "" || len(request.OperationID) > 128 {
		return false
	}

	return opsnamespace.Valid(request.Target.Namespace) && request.Target.ProjectID != "" && len(request.Target.ProjectBindingDigest) == 64
}

func insertOperation(ctx context.Context, tx *sql.Tx, request OperationRequest, now time.Time) error {
	if err := checkRequestAllowance(ctx, tx, request); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `INSERT INTO telegram_requests(id,bot_id,update_id,operation_ordinal,operator_id,project_id,project_binding_digest,namespace,conversation_id,expected_conversation_revision,operation_id,arguments,argument_digest,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,0),?,?,?,'prepared',?,?) ON CONFLICT(bot_id,update_id,operation_ordinal) DO NOTHING`, request.ID, request.Source.Peer.BotID, request.Source.UpdateID, request.Ordinal, request.Operator, request.Target.ProjectID, request.Target.ProjectBindingDigest, request.Target.Namespace, request.Target.ConversationID, request.Target.ConversationRevision, request.OperationID, string(request.Arguments), request.ArgumentDigest, database.FormatTime(now), database.FormatTime(now))
	if err != nil {
		return err
	}

	return bindInitialAuthoringWork(ctx, tx, request)
}
