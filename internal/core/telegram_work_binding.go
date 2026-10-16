package core

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"slices"
)

// TelegramInvocationBinding references canonical native work without credentials.
type TelegramInvocationBinding struct {
	ID       string `json:"id"`
	Revision uint64 `json:"revision"`
}

// ValidateTelegramWorkConversation prevents admission replacing a bound run.
// Current peer/operator/project authority must be checked separately by the host.
func (s *TelegramReceiverStore) ValidateTelegramWorkConversation(ctx context.Context, request TelegramOperationRequest) error {
	if request.OperationID != "work.admit" {
		return ErrWorkflowConflict
	}

	if request.Target.ConversationID == "" {
		return nil
	}

	var count int

	target := request.Target

	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_conversations WHERE id=? AND revision=? AND project_binding_digest=? AND invocation_id IS NULL AND phase NOT IN ('completed','canceled','failed','executing','waiting')`, target.ConversationID, target.ConversationRevision, target.ProjectBindingDigest).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrWorkflowConflict
	}

	return nil
}

func projectTelegramWorkBinding(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, result TelegramActionResult, now string) error {
	if result.Invocation == nil {
		return nil
	}

	if request.OperationID == "work.control" {
		return projectTelegramControlBinding(ctx, tx, request, result, now)
	}

	return projectTelegramAdmissionBinding(ctx, tx, request, result, now)
}

func projectTelegramAdmissionBinding(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, result TelegramActionResult, now string) error {
	bound := result.Invocation
	if result.InvocationPhase != "" {
		return ErrWorkflowConflict
	}

	if request.OperationID != "work.admit" || result.State != "confirmed" || !validWorkflowRequestKey(bound.ID) || bound.Revision == 0 || bound.Revision > math.MaxInt64 || result.ReceiptRef != "workflow:"+bound.ID {
		return ErrWorkflowConflict
	}

	if request.Target.ConversationID == "" {
		return nil
	}

	target := request.Target

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET invocation_id=?,last_observed_revision=?,authorization_ref=?,phase='executing',updated_at=? WHERE id=? AND revision=? AND project_id=? AND namespace=? AND project_binding_digest=? AND invocation_id IS NULL AND phase NOT IN ('completed','canceled','failed','executing','waiting')`, bound.ID, bound.Revision, request.ID, now, target.ConversationID, target.ConversationRevision, target.ProjectID, target.Namespace, target.ProjectBindingDigest)
	if err != nil {
		return err
	}

	return telegramRequireChanged(changed)
}

func projectTelegramControlBinding(ctx context.Context, tx *sql.Tx, request TelegramOperationRequest, result TelegramActionResult, now string) error {
	bound := result.Invocation
	if result.State != "confirmed" || !validWorkflowRequestKey(bound.ID) || bound.Revision == 0 || bound.Revision > math.MaxInt64 {
		return ErrWorkflowConflict
	}

	if !slices.Contains([]string{"executing", "waiting", "completed", "canceled"}, result.InvocationPhase) {
		return ErrWorkflowConflict
	}

	if result.ReceiptRef != workflowMutationReceiptPath(request.ID) && result.ReceiptRef != fmt.Sprintf("workflow:%s:revision:%d", bound.ID, bound.Revision) {
		return ErrWorkflowConflict
	}

	target := request.Target
	if target.ConversationID == "" {
		return ErrWorkflowConflict
	}

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET invocation_id=?,last_observed_revision=MAX(COALESCE(last_observed_revision,0),?),authorization_ref=COALESCE(NULLIF(authorization_ref,''),?),phase=?,updated_at=? WHERE id=? AND revision=? AND project_id=? AND namespace=? AND project_binding_digest=? AND (invocation_id IS NULL OR invocation_id=?)`, bound.ID, bound.Revision, request.ID, result.InvocationPhase, now, target.ConversationID, target.ConversationRevision, target.ProjectID, target.Namespace, target.ProjectBindingDigest, bound.ID)
	if err != nil {
		return err
	}

	return telegramRequireChanged(changed)
}
