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
	Integration      *WorkflowIntegrationControlGuard
	BudgetGrant      *core.WorkflowIntegrationGrant
	DecisionAnswer   *core.WorkflowIntegrationAnswer
	DecisionRevision uint64
	DecisionHash     string
	Action           string
	ID               string
	ExpectedRevision uint64
	BindingHash      string
	StepLimit        int
}

// ControlWorkExtent dispatches shared lifecycle operations and returns current
// read-only status even when the guarded action fails.
func (s service) ControlWorkExtent(ctx context.Context, request WorkflowControlRequest) (WorkflowWorkStatus, error) {
	if !validWorkflowControlIdentity(request.ID) || request.ExpectedRevision == 0 || request.StepLimit < 0 || request.StepLimit > 128 {
		return WorkflowWorkStatus{}, fmt.Errorf("workflow control requires the displayed identity and revision")
	}

	err := s.executeWorkflowControl(ctx, request)

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	status, readErr := s.WorkflowStatus(cleanup, request.ID, WorkflowStatusRequest{Limit: 8})

	return status, errors.Join(err, readErr)
}

func (s service) executeWorkflowControl(ctx context.Context, request WorkflowControlRequest) error {
	switch request.Action {
	case "pause", "cancel":
		_, err := s.StopWorkflow(ctx, request.ID, request.Action, request.ExpectedRevision)

		return err
	case "resume":
		_, err := s.ResumeWorkflow(ctx, request.ID, request.ExpectedRevision)

		return err
	case "continue":
		_, err := s.ContinueWorkExtent(ctx, request.ID, request.StepLimit, request.ExpectedRevision)

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
		_, err := s.ReconcileWorkflow(ctx, request.ID, request.ExpectedRevision)

		return err
	case "grant", "answer_decision":
		return s.executeWorkflowIntegrationControl(ctx, request)
	default:
		return fmt.Errorf("unsupported workflow control: %s", request.Action)
	}
}

func validWorkflowControlIdentity(id string) bool {
	return id != "" && len(id) <= 128 && strings.TrimSpace(id) == id && !strings.ContainsAny(id, "\x00\n\r")
}
