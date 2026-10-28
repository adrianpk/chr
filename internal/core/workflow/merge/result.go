package merge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Fence is the host's current immutable acceptance identity.
// Callers obtain it from retained state; an agent-supplied fence is not trusted.
type Fence struct {
	EntryID, CandidateHash, ScopeHash, OwnedWorkspaceID string
	Repository, SourceRef, TargetRef                    string
	Source, Base, Method                                string
	Unit                                                workflowstate.Ref
	CandidateRevision, ClaimGeneration                  uint64
}

// ParseResult rejects malformed or stale proposals before host
// diff verification. Parsing a resolved proposal is never candidate approval.
func ParseResult(data []byte, input workflowstate.MergerInput, current Fence, now time.Time) (workflowstate.MergerResult, error) {
	var result workflowstate.MergerResult
	if err := CheckInput(input, now); err != nil {
		return result, err
	}

	if InputFence(input) != current {
		return result, fmt.Errorf("%w: merger generation or candidate is stale", workflowstate.ErrConflict)
	}

	if err := decodeResult(data, &result); err != nil {
		return workflowstate.MergerResult{}, err
	}

	if err := CheckResult(result, input); err != nil {
		return workflowstate.MergerResult{}, err
	}

	return result, nil
}

func decodeResult(data []byte, result *workflowstate.MergerResult) error {
	if len(data) == 0 || len(data) > workflowstate.FileLimit {
		return fmt.Errorf("%w: merger result exceeds file bound", workflowstate.ErrConflict)
	}

	if err := CheckJSON(data); err != nil {
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

// CheckReportedUsage checks reported whole-call usage against the retained allowance.
func CheckReportedUsage(usage *workflowstate.MergerUsage, allowance *workflowstate.TokenAllowance) error {
	if usage == nil {
		return nil
	}

	total, err := workflowstate.TokenSum(usage.Input, usage.Output)
	if err != nil || !usage.WholeCall || !workflowstate.ValidFingerprint(usage.CapabilityRevision) {
		return workflowstate.ErrConflict
	}

	if allowance != nil && (usage.CapabilityRevision != allowance.CapabilityRevision || total > allowance.Maximum) {
		return workflowstate.ErrConflict
	}

	return nil
}

// CheckResult checks proposal identity, scope, evidence, resolutions and outcome.
func CheckResult(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if result.ProtocolVersion != input.ProtocolVersion || result.InputHash != input.InputHash || !slices.Contains([]string{"resolved", "needs_decision", "failed"}, result.Outcome) {
		return fmt.Errorf("%w: merger result version or input is stale", workflowstate.ErrConflict)
	}

	if err := CheckPaths(result.ChangedPaths); err != nil {
		return err
	}

	for _, path := range result.ChangedPaths {
		if !PathAllowed(input, path) {
			return fmt.Errorf("%w: merger result escapes approved scope", workflowstate.ErrConflict)
		}
	}

	if err := checkResultEvidence(result, input); err != nil {
		return err
	}

	if err := checkResolutions(result, input); err != nil {
		return err
	}

	return checkOutcome(result, input)
}

func checkResultEvidence(result workflowstate.MergerResult, input workflowstate.MergerInput) error {
	if len(result.LocalCheckRefs) > 16 {
		return workflowstate.ErrConflict
	}

	for _, ref := range result.LocalCheckRefs {
		if err := workflowstate.CheckEvidenceRef(input.InvocationID, ref); err != nil {
			return err
		}
	}

	if len(result.PreservedRequirements) > 0 {
		if err := checkRequirementIDs(result.PreservedRequirements, input.Requirements); err != nil {
			return err
		}
	}

	return CheckReportedUsage(result.Usage, input.TokenAllowance)
}

// InputFence copies immutable acceptance identity from a retained merger input.
func InputFence(input workflowstate.MergerInput) Fence {
	return Fence{
		EntryID: input.EntryID, CandidateHash: input.CandidateHash, ScopeHash: input.ScopeHash,
		OwnedWorkspaceID: input.OwnedWorkspaceID, Source: input.Source, Base: input.Base,
		Method: input.Method, Unit: input.Unit, CandidateRevision: input.CandidateRevision,
		ClaimGeneration: input.ClaimGeneration,
		Repository:      input.Repository, SourceRef: input.SourceRef, TargetRef: input.TargetRef,
	}
}
