package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowMergerFence is the host's current immutable acceptance identity.
// Callers obtain it from retained state; an agent-supplied fence is not trusted.
type WorkflowMergerFence struct {
	EntryID, CandidateHash, ScopeHash, OwnedWorkspaceID string
	Repository, SourceRef, TargetRef                    string
	Source, Base, Method                                string
	Unit                                                WorkflowRef
	CandidateRevision, ClaimGeneration                  uint64
}

func workflowMergerFence(input WorkflowMergerInput) WorkflowMergerFence {
	return WorkflowMergerFence{
		EntryID: input.EntryID, CandidateHash: input.CandidateHash, ScopeHash: input.ScopeHash,
		OwnedWorkspaceID: input.OwnedWorkspaceID, Source: input.Source, Base: input.Base,
		Method: input.Method, Unit: input.Unit, CandidateRevision: input.CandidateRevision,
		ClaimGeneration: input.ClaimGeneration,
		Repository:      input.Repository, SourceRef: input.SourceRef, TargetRef: input.TargetRef,
	}
}

// ParseWorkflowMergerResult rejects malformed or stale proposals before host
// diff verification. Parsing a resolved proposal is never candidate approval.
func ParseWorkflowMergerResult(data []byte, input WorkflowMergerInput, current WorkflowMergerFence, now time.Time) (WorkflowMergerResult, error) {
	var result WorkflowMergerResult
	if err := validateWorkflowMergerInput(input, now); err != nil {
		return result, err
	}

	if workflowMergerFence(input) != current {
		return result, fmt.Errorf("%w: merger generation or candidate is stale", workflowstate.ErrConflict)
	}

	if err := decodeWorkflowMergerResult(data, &result); err != nil {
		return WorkflowMergerResult{}, err
	}

	if err := validateWorkflowMergerResult(result, input); err != nil {
		return WorkflowMergerResult{}, err
	}

	return result, nil
}

func decodeWorkflowMergerResult(data []byte, result *WorkflowMergerResult) error {
	if len(data) == 0 || len(data) > workflowstate.WorkflowFileLimit {
		return fmt.Errorf("%w: merger result exceeds file bound", workflowstate.ErrConflict)
	}

	if err := checkWorkflowMergerJSON(data); err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("%w: invalid merger result: %w", workflowstate.ErrConflict, err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: merger result contains trailing output", workflowstate.ErrConflict)
	}

	return nil
}

func validateWorkflowMergerResult(result WorkflowMergerResult, input WorkflowMergerInput) error {
	if result.ProtocolVersion != input.ProtocolVersion || result.InputHash != input.InputHash || !slices.Contains([]string{"resolved", "needs_decision", "failed"}, result.Outcome) {
		return fmt.Errorf("%w: merger result version or input is stale", workflowstate.ErrConflict)
	}

	if err := validateWorkflowMergerPaths(result.ChangedPaths); err != nil {
		return err
	}

	for _, path := range result.ChangedPaths {
		if !workflowMergerPathAllowed(input, path) {
			return fmt.Errorf("%w: merger result escapes approved scope", workflowstate.ErrConflict)
		}
	}

	if err := validateWorkflowMergerResultEvidence(result, input); err != nil {
		return err
	}

	if err := validateWorkflowMergerResolutions(result, input); err != nil {
		return err
	}

	return validateWorkflowMergerOutcome(result, input)
}

func validateWorkflowMergerResultEvidence(result WorkflowMergerResult, input WorkflowMergerInput) error {
	if len(result.LocalCheckRefs) > 16 {
		return workflowstate.ErrConflict
	}

	for _, ref := range result.LocalCheckRefs {
		if err := validateWorkflowEvidenceRef(input.InvocationID, ref); err != nil {
			return err
		}
	}

	if len(result.PreservedRequirements) > 0 {
		if err := validateWorkflowMergerRequirementIDs(result.PreservedRequirements, input.Requirements); err != nil {
			return err
		}
	}

	return validateWorkflowMergerReportedUsage(result.Usage, input.TokenAllowance)
}

func validateWorkflowMergerReportedUsage(usage *WorkflowMergerUsage, allowance *WorkflowTokenAllowance) error {
	if usage == nil {
		return nil
	}

	total, err := workflowTokenSum(usage.Input, usage.Output)
	if err != nil || !usage.WholeCall || !validWorkflowNativeFingerprint(usage.CapabilityRevision) {
		return workflowstate.ErrConflict
	}

	if allowance != nil && (usage.CapabilityRevision != allowance.CapabilityRevision || total > allowance.Maximum) {
		return workflowstate.ErrConflict
	}

	return nil
}
