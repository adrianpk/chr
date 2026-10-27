package core

import (
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowControlIntegrationGuard is the displayed integration image. Private
// claim credentials are derived by the native owner after these fields match.
type WorkflowControlIntegrationGuard struct {
	EntryID        string
	EntryRevision  uint64
	BudgetRevision uint64
	CandidateHash  string
}

// WorkflowControlMutationRequest binds one attributable control to its
// displayed invocation and any required candidate, budget or question images.
type WorkflowControlMutationRequest struct {
	Guided           *GuidedWorkflowControlProof `json:"guided,omitempty"`
	Binding          workflowstate.MutationBinding
	Integration      *WorkflowControlIntegrationGuard
	Grant            *WorkflowIntegrationGrant
	Answer           *WorkflowIntegrationAnswer
	Action           string
	ID               string
	ExpectedRevision uint64
	ReviewHash       string
	DecisionRevision uint64
	DecisionHash     string
	StepLimit        int `json:",omitempty"`
}

// WorkflowControlResult is the original confirmed mutation receipt. Its revision
// describes acceptance; callers observe current workflow state separately.
type WorkflowControlResult struct {
	RequestID      string `json:"request_id"`
	Action         string `json:"action"`
	InvocationID   string `json:"invocation_id"`
	ArgumentDigest string `json:"argument_digest"`
	ReceiptRef     string `json:"receipt_ref"`
	PriorRevision  uint64 `json:"prior_revision"`
	Revision       uint64 `json:"revision"`
}

func freezeWorkflowControlMutation(request WorkflowControlMutationRequest) (WorkflowControlMutationRequest, string, error) {
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

	var frozen WorkflowControlMutationRequest
	if err := json.Unmarshal(data, &frozen); err != nil {
		return frozen, "", err
	}

	if err := validateWorkflowControlPayload(frozen); err != nil {
		return frozen, "", err
	}

	return frozen, contenthash.Sum(data), nil
}

func validateWorkflowControlPayload(request WorkflowControlMutationRequest) error {
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
