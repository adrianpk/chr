package core

import (
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func freezeWorkflowControlMutation(request workflowstate.ControlMutationRequest) (workflowstate.ControlMutationRequest, string, error) {
	if err := validateGuidedControlProof(request); err != nil {
		return request, "", err
	}

	if err := request.Binding.Validate(); err != nil {
		return request, "", err
	}

	if !workflowstate.ValidRequestKey(request.ID) || request.ExpectedRevision == 0 || !slices.Contains([]string{"approve", "grant", "answer_decision", "pause", "cancel", "resume", "continue", "reconcile"}, request.Action) {
		return request, "", workflowstate.ErrConflict
	}

	data, err := json.Marshal(request)
	if err != nil || len(data) > 32768 {
		return request, "", workflowstate.ErrConflict
	}

	var frozen workflowstate.ControlMutationRequest
	if err := json.Unmarshal(data, &frozen); err != nil {
		return frozen, "", err
	}

	if err := validateWorkflowControlPayload(frozen); err != nil {
		return frozen, "", err
	}

	return frozen, contenthash.Sum(data), nil
}

func validateWorkflowControlPayload(request workflowstate.ControlMutationRequest) error {
	if request.StepLimit != 0 && request.Action != "continue" {
		return workflowstate.ErrConflict
	}

	switch request.Action {
	case "pause", "cancel", "resume", "reconcile", "continue":
		return validateWorkflowLifecyclePayload(request)
	case "approve":
		if request.ReviewHash == "" || request.Grant != nil || request.Answer != nil {
			return workflowstate.ErrConflict
		}
	case "grant":
		return validateWorkflowControlGrant(request)
	case "answer_decision":
		return validateWorkflowControlAnswer(request)
	}

	return nil
}
