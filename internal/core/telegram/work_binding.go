package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// InvocationBinding references canonical native work without credentials.
type InvocationBinding struct {
	ID       string `json:"id"`
	Revision uint64 `json:"revision"`
}

// ValidateWorkConversation prevents admission replacing a bound run.
// Current peer/operator/project authority must be checked separately by the host.
func (s *ReceiverStore) ValidateWorkConversation(ctx context.Context, request OperationRequest) error {
	if request.OperationID != "work.admit" {
		return workflowstate.ErrConflict
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
		return workflowstate.ErrConflict
	}

	return nil
}

func projectWorkBinding(ctx context.Context, tx *sql.Tx, request OperationRequest, result ActionResult, now string) error {
	if result.Invocation == nil {
		return nil
	}

	if request.OperationID == "work.control" {
		return projectControlBinding(ctx, tx, request, result, now)
	}

	return projectAdmissionBinding(ctx, tx, request, result, now)
}

func projectAdmissionBinding(ctx context.Context, tx *sql.Tx, request OperationRequest, result ActionResult, now string) error {
	bound := result.Invocation
	if result.InvocationPhase != "" {
		return workflowstate.ErrConflict
	}

	if request.OperationID != "work.admit" || result.State != OperationConfirmed || !workflowstate.ValidRequestKey(bound.ID) || bound.Revision == 0 || bound.Revision > math.MaxInt64 || result.ReceiptRef != "workflow:"+bound.ID {
		return workflowstate.ErrConflict
	}

	if request.Target.ConversationID == "" {
		return nil
	}

	target := request.Target

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET invocation_id=?,last_observed_revision=?,authorization_ref=?,phase='executing',updated_at=? WHERE id=? AND revision=? AND project_id=? AND namespace=? AND project_binding_digest=? AND invocation_id IS NULL AND phase NOT IN ('completed','canceled','failed','executing','waiting')`, bound.ID, bound.Revision, request.ID, now, target.ConversationID, target.ConversationRevision, target.ProjectID, target.Namespace, target.ProjectBindingDigest)
	if err != nil {
		return err
	}

	return requireChanged(changed)
}

func projectControlBinding(ctx context.Context, tx *sql.Tx, request OperationRequest, result ActionResult, now string) error {
	if err := validateControlBinding(request, result); err != nil {
		return err
	}

	target := request.Target
	bound := result.Invocation

	if target.ConversationID == "" {
		return workflowstate.ErrConflict
	}

	changed, err := tx.ExecContext(ctx, `UPDATE telegram_conversations SET invocation_id=?,last_observed_revision=MAX(COALESCE(last_observed_revision,0),?),authorization_ref=COALESCE(NULLIF(authorization_ref,''),?),phase=?,updated_at=? WHERE id=? AND revision=? AND project_id=? AND namespace=? AND project_binding_digest=? AND (invocation_id IS NULL OR invocation_id=?) AND COALESCE(last_observed_revision,0)<=?`, bound.ID, bound.Revision, request.ID, result.InvocationPhase, now, target.ConversationID, target.ConversationRevision, target.ProjectID, target.Namespace, target.ProjectBindingDigest, bound.ID, bound.Revision)
	if err != nil {
		return err
	}

	// A committed native receipt remains reportable after a newer conversation
	// replaces this projection. Acknowledgement must not retarget that conversation.
	if result.ReceiptRef == workflowstate.MutationReceiptPath(request.ID) {
		return nil
	}

	return requireChanged(changed)
}

func validateControlBinding(request OperationRequest, result ActionResult) error {
	bound := result.Invocation
	if result.State != OperationConfirmed || !workflowstate.ValidRequestKey(bound.ID) || bound.Revision == 0 || bound.Revision > math.MaxInt64 {
		return workflowstate.ErrConflict
	}

	if !slices.Contains([]ConversationPhase{ConversationExecuting, ConversationWaiting, ConversationCompleted, ConversationCanceled}, result.InvocationPhase) {
		return workflowstate.ErrConflict
	}

	if result.ReceiptRef != workflowstate.MutationReceiptPath(request.ID) && result.ReceiptRef != fmt.Sprintf("workflow:%s:revision:%d", bound.ID, bound.Revision) {
		return workflowstate.ErrConflict
	}

	return nil
}
