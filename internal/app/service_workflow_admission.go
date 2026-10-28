package app

import (
	"context"
	"time"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowBoundAdmission supplies a durable request key and the displayed policy.
// This mode admits without continuation; hosts activate the shared owner later.
type WorkflowBoundAdmission struct {
	ReadyTicket     *workflowstate.TicketAdmission `json:"ready_ticket,omitempty"`
	ExecutionPolicy workflowstate.ExecutionPolicy  `json:"execution_policy"`
	RequestKey      string                         `json:"request_key"`
	Policy          workflowstate.ReviewPolicy     `json:"publication_policy"`
	Check           workflowcore.AdmissionCheck    `json:"-"`
}

func (s service) admitBoundWorkExtent(ctx context.Context, selection workflowstate.PreviewRequest, selected *workflowstate.SelectionGuard, steps int, bindings []WorkflowBoundAdmission) (AgentWorkResult, error) {
	if len(bindings) != 1 || steps != 0 || len(selection.SnapshotHash) != 64 || selection.Offset != 0 {
		return AgentWorkResult{}, workflowstate.ErrConflict
	}

	// Retained requests must reach the owner before re-reading mutable records,
	// targets or policy. The owner checks every guard for a new admission.
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: bindings[0].RequestKey,
		Selection:  selection,
		Admission:  &workflowstate.AdmissionGuard{SnapshotHash: selection.SnapshotHash, Selected: selected, Policy: bindings[0].Policy, ExecutionPolicy: bindings[0].ExecutionPolicy, ReadyTicket: bindings[0].ReadyTicket},
	}}

	store, err := s.workflowStore()
	if err != nil {
		return AgentWorkResult{}, err
	}

	request.ResolveHarness = s.prepareWorkflowHarness

	started, err := store.StartWorkflowGuarded(ctx, request, s.boundAdmissionCheck(selection, bindings[0].Check))
	if err != nil {
		return AgentWorkResult{}, err
	}

	return s.nativeAgentWorkResult(started.Invocation)
}

func (s service) boundAdmissionCheck(selection workflowstate.PreviewRequest, check workflowcore.AdmissionCheck) workflowcore.AdmissionCheck {
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
			return workflowstate.ErrConflict
		}

		if check != nil {
			return check(ctx)
		}

		return nil
	}
}
