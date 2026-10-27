package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// HarnessSelectionOperation identifies deterministic configuration selection.
const HarnessSelectionOperation = "harnesses.change"

// ReserveHarnessOperation atomically allocates a source-owned prompt
// conversation and freezes prepared input against the current selected project.
func (s *ReceiverStore) ReserveHarnessOperation(ctx context.Context, input ConversationInput, request OperationRequest, prepared json.RawMessage, now time.Time) (OperationRequest, error) {
	if !validHarnessReservation(input, request, prepared) {
		return request, workflowstate.ErrConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

		target, err := op.harnessConfirmationTarget(request.Target)
		if err != nil {
			return err
		}

		request.Target = target

		request, err = normalizeOperation(request)
		if err != nil {
			return err
		}

		return reservePreparedOperation(ctx, tx, request, prepared, now)
	})
	if err != nil {
		return request, err
	}

	retained, found, err := s.Operation(ctx, request.Source, 0)
	if err != nil {
		return request, err
	}

	if !found || retained.ArgumentDigest != request.ArgumentDigest || string(retained.PreparedInput) != string(prepared) {
		return request, workflowstate.ErrConflict
	}

	return retained, nil
}

func validHarnessReservation(input ConversationInput, request OperationRequest, prepared json.RawMessage) bool {
	user, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	if !allowed || user.OperatorID != request.Operator || request.Source != input.Update || request.OperationID != HarnessSelectionOperation || request.Ordinal != 0 {
		return false
	}

	return len(prepared) > 0 && len(prepared) <= 65536 && json.Valid(prepared)
}

func (op conversationTransaction) harnessConfirmationTarget(expected OperationTarget) (OperationTarget, error) {
	project, err := op.selectedHarnessProject()
	if err != nil {
		return OperationTarget{}, err
	}

	op.input.Project = project.ID

	current, err := op.operationTarget(nil)
	if err != nil {
		return OperationTarget{}, err
	}

	if current != expected {
		return OperationTarget{}, workflowstate.ErrConflict
	}

	conversation, err := op.sourceActionConversation("prepare", "Harness defaults")
	if err != nil {
		return OperationTarget{}, err
	}

	if conversation.ID != conversationID(op.input.Update) {
		return OperationTarget{}, workflowstate.ErrConflict
	}

	return op.operationTarget(conversation)
}

// ValidateHarnessTarget rechecks mapped identity, current project choice,
// registry root and the source-owned confirmation conversation before publication.
func (s *ReceiverStore) ValidateHarnessTarget(ctx context.Context, input ConversationInput, request OperationRequest) error {
	user, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType)
	if !allowed || user.OperatorID != request.Operator || request.Source.Peer != input.Update.Peer || request.OperationID != HarnessSelectionOperation || request.Target.ConversationID != conversationID(request.Source) {
		return workflowstate.ErrConflict
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input}

		project, err := op.selectedHarnessProject()
		if err != nil {
			return err
		}

		if project.ID != request.Target.ProjectID {
			return workflowstate.ErrConflict
		}

		conversation, err := op.conversation(request.Target.ConversationID)
		if err != nil {
			return err
		}

		if err := validateBoundConversation(*conversation, request.Target, request.Target.ProjectBindingDigest); err != nil {
			return err
		}

		return op.validateTarget(conversation)
	})
}

func finishHarnessConversation(ctx context.Context, tx *sql.Tx, request OperationRequest, state string, now time.Time) error {
	if request.OperationID != HarnessSelectionOperation {
		return nil
	}

	if request.Target.ConversationID != conversationID(request.Source) {
		return workflowstate.ErrConflict
	}

	phase := "failed"
	if state == "confirmed" {
		phase = "completed"
	}

	_, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET phase=?,updated_at=? WHERE id=? AND revision=? AND project_binding_digest=? AND phase='prepared'`, phase, database.FormatTime(now), request.Target.ConversationID, request.Target.ConversationRevision, request.Target.ProjectBindingDigest)

	return err
}
