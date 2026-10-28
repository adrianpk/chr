package merge

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SealInput validates and hashes a host-assembled immutable input.
// It preserves the caller's input and does not grant a capacity or budget slot.
func SealInput(input workflowstate.MergerInput, now time.Time) (workflowstate.MergerInput, error) {
	input.InputHash = ""
	if err := checkInputContent(input, now); err != nil {
		return workflowstate.MergerInput{}, err
	}

	hash, err := inputHash(input)
	if err != nil {
		return workflowstate.MergerInput{}, err
	}

	input.InputHash = hash

	return input, nil
}

func validDeadline(input workflowstate.MergerInput, now time.Time) bool {
	maximum := input.AdmittedAt.Add(time.Duration(input.WorkerTimeoutSeconds) * time.Second)

	return !input.AdmittedAt.IsZero() && !input.AdmittedAt.After(now) && input.Deadline.After(now) && !input.Deadline.After(maximum)
}

func validIdentity(input workflowstate.MergerInput) bool {
	keys := []string{input.ID, input.InvocationID, input.Unit.ID, input.CapacityReservationID, input.BudgetReservationID}
	for _, key := range keys {
		if !workflowstate.ValidRequestKey(key) || !filepath.IsLocal(key) || filepath.Base(key) != key || strings.ContainsRune(key, '\\') || key == "." {
			return false
		}
	}

	return input.ProtocolVersion == 1 && input.CandidateRevision > 0 && input.ClaimGeneration > 0 && slices.Contains([]string{"slice", "ticket"}, input.Unit.Kind)
}

func validLimits(input workflowstate.MergerInput) bool {
	return input.RemainingAttempts >= 1 && input.RemainingAttempts <= 10 && input.WorkerTimeoutSeconds >= 1 && input.WorkerTimeoutSeconds <= 86400
}

func validObjects(input workflowstate.MergerInput) bool {
	for _, hash := range []string{input.EntryID, input.CandidateHash, input.ScopeHash, input.OwnedWorkspaceID} {
		if !workflowstate.ValidFingerprint(hash) {
			return false
		}
	}

	return workflowstate.ValidCommitID(input.Source) && workflowstate.ValidCommitID(input.Base) && len(input.Source) == len(input.Base) && slices.Contains([]string{"merge", "rebase", "squash"}, input.Method)
}

func checkAllowance(input workflowstate.MergerInput, now time.Time) error {
	if !validLimits(input) {
		return fmt.Errorf("%w: merger requires finite attempts and runtime", workflowstate.ErrConflict)
	}

	if !validDeadline(input, now) {
		return fmt.Errorf("%w: merger deadline is stale or exceeds the worker timeout", workflowstate.ErrConflict)
	}

	allowance := input.TokenAllowance
	if allowance != nil && (!workflowstate.ValidTokenAllowance(*allowance) || allowance.Deadline != input.Deadline) {
		return fmt.Errorf("%w: merger token allowance is unenforceable", workflowstate.ErrConflict)
	}

	return nil
}

// CheckInput checks the full retained input, allowance and exact encoding hash.
func CheckInput(input workflowstate.MergerInput, now time.Time) error {
	if err := checkInputContent(input, now); err != nil {
		return err
	}

	hash, err := inputHash(input)
	if err != nil || !workflowstate.ValidFingerprint(input.InputHash) || hash != input.InputHash {
		return fmt.Errorf("%w: merger input hash changed", workflowstate.ErrConflict)
	}

	return nil
}

func checkInputContent(input workflowstate.MergerInput, now time.Time) error {
	if !validIdentity(input) || !validObjects(input) {
		return fmt.Errorf("%w: invalid merger identity or objects", workflowstate.ErrConflict)
	}

	if err := checkRepository(input); err != nil {
		return err
	}

	if err := checkAllowance(input, now); err != nil {
		return err
	}

	if err := CheckScope(input); err != nil {
		return err
	}

	if err := checkIntent(input); err != nil {
		return err
	}

	return checkReferences(input)
}

func checkIntent(input workflowstate.MergerInput) error {
	if len(input.Requirements) == 0 || len(input.Requirements) > workflowstate.PageLimit || len(input.Conflicts) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	if err := checkRequirements(input.Requirements); err != nil {
		return err
	}

	if err := checkConflicts(input); err != nil {
		return err
	}

	if err := checkReason(input); err != nil {
		return err
	}

	return checkCheckpoint(input)
}

func checkReason(input workflowstate.MergerInput) error {
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

func checkRepository(input workflowstate.MergerInput) error {
	repository, err := workflowstate.CanonicalRepository(input.Repository)
	if err != nil || repository != input.Repository || input.SourceRef == input.TargetRef {
		return workflowstate.ErrConflict
	}

	for _, ref := range []string{input.SourceRef, input.TargetRef} {
		if !strings.HasPrefix(ref, "refs/heads/") || !workflowstate.ValidBranch(strings.TrimPrefix(ref, "refs/heads/")) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func inputHash(input workflowstate.MergerInput) (string, error) {
	input.InputHash = ""

	data, err := json.Marshal(input)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return "", fmt.Errorf("%w: merger input exceeds file bound", workflowstate.ErrConflict)
	}

	return contenthash.Sum(data), nil
}
