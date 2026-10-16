package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowBoundAdmission supplies a durable request key and the displayed policy.
// This mode admits without continuation; hosts activate the shared owner later.
type WorkflowBoundAdmission struct {
	ReadyTicket     *core.WorkflowTicketAdmission `json:"ready_ticket,omitempty"`
	ExecutionPolicy core.WorkflowExecutionPolicy  `json:"execution_policy"`
	RequestKey      string                        `json:"request_key"`
	Policy          core.WorkflowReviewPolicy     `json:"publication_policy"`
	Check           core.WorkflowAdmissionCheck   `json:"-"`
}

func (s service) admitBoundWorkExtent(ctx context.Context, selection core.WorkflowPreviewRequest, selected *core.WorkflowSelectionGuard, steps int, bindings []WorkflowBoundAdmission) (AgentWorkResult, error) {
	if len(bindings) != 1 || steps != 0 || len(selection.SnapshotHash) != 64 || selection.Offset != 0 {
		return AgentWorkResult{}, core.ErrWorkflowConflict
	}

	// Retained requests must reach the owner before re-reading mutable records,
	// targets or policy. The owner checks every guard for a new admission.
	request := core.WorkflowStartRequest{RequestKey: bindings[0].RequestKey, Selection: selection, Admission: &core.WorkflowAdmissionGuard{SnapshotHash: selection.SnapshotHash, Selected: selected, Policy: bindings[0].Policy, ExecutionPolicy: bindings[0].ExecutionPolicy, ReadyTicket: bindings[0].ReadyTicket}}

	store, err := s.workflowStore()
	if err != nil {
		return AgentWorkResult{}, err
	}

	started, err := store.StartWorkflowGuarded(ctx, request, s.boundAdmissionCheck(selection, bindings[0].Check))
	if err != nil {
		return AgentWorkResult{}, err
	}

	return s.nativeAgentWorkResult(started.Invocation)
}

func (s service) boundAdmissionCheck(selection core.WorkflowPreviewRequest, check core.WorkflowAdmissionCheck) core.WorkflowAdmissionCheck {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		config, err := s.workflowExtentConfig()
		if err != nil {
			return err
		}

		target, err := s.nativeAgentWorkTarget(config)
		if err != nil {
			return err
		}

		observed, err := s.observeNativeAgentWorkTarget(config, target)
		if err != nil {
			return err
		}

		if observed != selection.Target {
			return core.ErrWorkflowConflict
		}

		if check != nil {
			return check(ctx)
		}

		return nil
	}
}
