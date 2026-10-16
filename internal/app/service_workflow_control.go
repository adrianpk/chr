package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowControlRequest carries the exact displayed invocation and revision.
type WorkflowControlRequest struct {
	AuthorizationCheck core.WorkflowAdmissionCheck   `json:"-"`
	RequestBinding     *core.WorkflowMutationBinding `json:"request_binding,omitempty"`
	Integration        *WorkflowIntegrationControlGuard
	BudgetGrant        *core.WorkflowIntegrationGrant
	DecisionAnswer     *core.WorkflowIntegrationAnswer
	DecisionRevision   uint64
	DecisionHash       string
	Action             string
	ID                 string
	ExpectedRevision   uint64
	BindingHash        string
	StepLimit          int
}

// ControlWorkExtent dispatches shared lifecycle operations and returns current
// read-only status even when the guarded action fails.
func (s service) ControlWorkExtent(ctx context.Context, request WorkflowControlRequest) (WorkflowWorkStatus, error) {
	if !validWorkflowControlIdentity(request.ID) || request.ExpectedRevision == 0 || request.StepLimit < 0 || request.StepLimit > 128 {
		return WorkflowWorkStatus{}, fmt.Errorf("workflow control requires the displayed identity and revision")
	}

	receipt, err := s.executeWorkflowControlResult(ctx, request)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	status, readErr := s.WorkflowStatus(cleanup, request.ID, WorkflowStatusRequest{Limit: 8})
	if err == nil {
		status.ControlReceipt = receipt
	}

	return status, errors.Join(err, readErr)
}

func (s service) executeWorkflowControl(ctx context.Context, request WorkflowControlRequest) error {
	switch request.Action {
	case "pause", "cancel":
		return s.executeGuardedWorkflowControl(ctx, request)
	case "resume":
		_, err := s.resumeWorkflowChecked(ctx, request.ID, request.AuthorizationCheck, request.ExpectedRevision)

		return err
	case "continue":
		_, err := s.continueWorkExtent(ctx, request.ID, request.StepLimit, request.AuthorizationCheck, request.ExpectedRevision)

		return err
	case "approve":
		if request.Integration != nil {
			if _, err := s.workflowIntegrationControlGuard(ctx, request); err != nil {
				return err
			}
		}

		_, err := s.ApproveWorkflowSlice(ctx, request.ID, request.BindingHash, request.ExpectedRevision)

		return err
	case "reconcile":
		return s.executeGuardedWorkflowControl(ctx, request)
	case "grant", "answer_decision":
		return s.executeWorkflowIntegrationControl(ctx, request)
	default:
		return fmt.Errorf("unsupported workflow control: %s", request.Action)
	}
}

func (s service) executeGuardedWorkflowControl(ctx context.Context, request WorkflowControlRequest) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	if request.Action == "reconcile" {
		_, err = store.ReconcileWorkflowGuarded(ctx, request.ID, request.ExpectedRevision, request.AuthorizationCheck)
	} else {
		_, err = store.StopWorkflowGuarded(ctx, request.ID, request.Action, request.ExpectedRevision, request.AuthorizationCheck)
	}

	return err
}

func validWorkflowControlIdentity(id string) bool {
	return id != "" && len(id) <= 128 && strings.TrimSpace(id) == id && !strings.ContainsAny(id, "\x00\n\r")
}
