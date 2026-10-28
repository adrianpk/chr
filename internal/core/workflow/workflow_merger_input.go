package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SealWorkflowMergerInput validates and hashes a host-assembled immutable input.
// It preserves the caller's input and does not grant a capacity or budget slot.
func SealWorkflowMergerInput(input workflowstate.MergerInput, now time.Time) (workflowstate.MergerInput, error) {
	input.InputHash = ""
	if err := validateWorkflowMergerInputContent(input, now); err != nil {
		return workflowstate.MergerInput{}, err
	}

	hash, err := workflowMergerInputHash(input)
	if err != nil {
		return workflowstate.MergerInput{}, err
	}

	input.InputHash = hash

	return input, nil
}

func workflowMergerInputHash(input workflowstate.MergerInput) (string, error) {
	input.InputHash = ""

	data, err := json.Marshal(input)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return "", fmt.Errorf("%w: merger input exceeds file bound", workflowstate.ErrConflict)
	}

	return contenthash.Sum(data), nil
}

func validateWorkflowMergerInput(input workflowstate.MergerInput, now time.Time) error {
	if err := validateWorkflowMergerInputContent(input, now); err != nil {
		return err
	}

	hash, err := workflowMergerInputHash(input)
	if err != nil || !validWorkflowNativeFingerprint(input.InputHash) || hash != input.InputHash {
		return fmt.Errorf("%w: merger input hash changed", workflowstate.ErrConflict)
	}

	return nil
}

func validateWorkflowMergerInputContent(input workflowstate.MergerInput, now time.Time) error {
	if !validWorkflowMergerIdentity(input) || !validWorkflowMergerObjects(input) {
		return fmt.Errorf("%w: invalid merger identity or objects", workflowstate.ErrConflict)
	}

	if err := validateWorkflowMergerRepository(input); err != nil {
		return err
	}

	if err := validateWorkflowMergerAllowance(input, now); err != nil {
		return err
	}

	if err := workflowmerge.ValidateWorkflowMergerScope(input); err != nil {
		return err
	}

	if err := validateWorkflowMergerIntent(input); err != nil {
		return err
	}

	return validateWorkflowMergerReferences(input)
}

func validateWorkflowMergerRepository(input workflowstate.MergerInput) error {
	repository, err := workflowCanonicalRepository(input.Repository)
	if err != nil || repository != input.Repository || input.SourceRef == input.TargetRef {
		return workflowstate.ErrConflict
	}

	for _, ref := range []string{input.SourceRef, input.TargetRef} {
		if !strings.HasPrefix(ref, "refs/heads/") || !validWorkflowPRBranch(strings.TrimPrefix(ref, "refs/heads/")) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func validWorkflowMergerIdentity(input workflowstate.MergerInput) bool {
	keys := []string{input.ID, input.InvocationID, input.Unit.ID, input.CapacityReservationID, input.BudgetReservationID}
	for _, key := range keys {
		if !workflowstate.ValidRequestKey(key) || !filepath.IsLocal(key) || filepath.Base(key) != key || strings.ContainsRune(key, '\\') || key == "." {
			return false
		}
	}

	return input.ProtocolVersion == 1 && input.CandidateRevision > 0 && input.ClaimGeneration > 0 && slices.Contains([]string{"slice", "ticket"}, input.Unit.Kind)
}

func validWorkflowMergerObjects(input workflowstate.MergerInput) bool {
	for _, hash := range []string{input.EntryID, input.CandidateHash, input.ScopeHash, input.OwnedWorkspaceID} {
		if !validWorkflowNativeFingerprint(hash) {
			return false
		}
	}

	return workflowCommitID(input.Source) && workflowCommitID(input.Base) && len(input.Source) == len(input.Base) && slices.Contains([]string{"merge", "rebase", "squash"}, input.Method)
}

func validateWorkflowMergerAllowance(input workflowstate.MergerInput, now time.Time) error {
	if !validWorkflowMergerLimits(input) {
		return fmt.Errorf("%w: merger requires finite attempts and runtime", workflowstate.ErrConflict)
	}

	if !validWorkflowMergerDeadline(input, now) {
		return fmt.Errorf("%w: merger deadline is stale or exceeds the worker timeout", workflowstate.ErrConflict)
	}

	allowance := input.TokenAllowance
	if allowance != nil && (!validWorkflowTokenAllowance(*allowance) || allowance.Deadline != input.Deadline) {
		return fmt.Errorf("%w: merger token allowance is unenforceable", workflowstate.ErrConflict)
	}

	return nil
}

func validWorkflowMergerLimits(input workflowstate.MergerInput) bool {
	return input.RemainingAttempts >= 1 && input.RemainingAttempts <= 10 && input.WorkerTimeoutSeconds >= 1 && input.WorkerTimeoutSeconds <= 86400
}

func validWorkflowMergerDeadline(input workflowstate.MergerInput, now time.Time) bool {
	maximum := input.AdmittedAt.Add(time.Duration(input.WorkerTimeoutSeconds) * time.Second)

	return !input.AdmittedAt.IsZero() && !input.AdmittedAt.After(now) && input.Deadline.After(now) && !input.Deadline.After(maximum)
}

func validateWorkflowMergerIntent(input workflowstate.MergerInput) error {
	if len(input.Requirements) == 0 || len(input.Requirements) > workflowstate.PageLimit || len(input.Conflicts) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowMergerRequirements(input.Requirements); err != nil {
		return err
	}

	if err := validateWorkflowMergerConflicts(input); err != nil {
		return err
	}

	if err := validateWorkflowMergerReason(input); err != nil {
		return err
	}

	return validateWorkflowMergerCheckpoint(input)
}

func validateWorkflowMergerReason(input workflowstate.MergerInput) error {
	switch input.Reason {
	case "preparation_conflict":
		if len(input.Conflicts) == 0 {
			return workflowstate.ErrConflict
		}
	case "combined_gate_regression", "actionable_review_finding":
		if len(input.FindingRefs) == 0 {
			return workflowstate.ErrConflict
		}
	default:
		return fmt.Errorf("%w: clean or drift-only preparation cannot launch a merger", workflowstate.ErrConflict)
	}

	return nil
}
