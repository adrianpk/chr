package core

import (
	"fmt"
	"slices"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowIntegrationEntry(entry WorkflowIntegrationEntry) error {
	if !validWorkflowIntegrationEntryHeader(entry) {
		return workflowIntegrationConflict("has invalid version, identity or revision")
	}

	if !validWorkflowIntegrationTarget(entry) {
		return workflowIntegrationConflict("has invalid canonical target")
	}

	if !validWorkflowIntegrationState(entry.State) || entry.BudgetRef != workflowIntegrationBudgetPath(entry.InvocationID) {
		return workflowIntegrationConflict("has invalid phase or budget owner")
	}

	for _, ref := range []*WorkflowEvidenceRef{entry.CandidateRef, entry.ReceiptRef, entry.DecisionRef} {
		if ref != nil {
			if err := validateWorkflowEvidenceRef(entry.InvocationID, *ref); err != nil {
				return err
			}
		}
	}

	return validateWorkflowIntegrationEntryPhase(entry)
}

func validWorkflowIntegrationEntryHeader(entry WorkflowIntegrationEntry) bool {
	return validWorkflowIntegrationIdentity(entry) && validWorkflowIntegrationRevisions(entry) && entry.Unit.ID != "" && (entry.Unit.Kind == "ticket" || entry.Unit.Kind == "slice")
}

func validWorkflowIntegrationIdentity(entry WorkflowIntegrationEntry) bool {
	return entry.SchemaVersion == 1 && validWorkflowNativeFingerprint(entry.ID) && workflowstate.ValidRequestKey(entry.InvocationID) && entry.ProjectID != "" && opsnamespace.Valid(entry.Namespace) && validWorkflowNativeFingerprint(entry.CoordinatorID) && validWorkflowNativeFingerprint(entry.FrozenScopeHash) && workflowstate.ValidRequestKey(entry.AdmissionKey)
}

func validWorkflowIntegrationRevisions(entry WorkflowIntegrationEntry) bool {
	return entry.Revision > 0 && entry.StateRevision > 0 && entry.StateRevision <= entry.Revision && entry.ClaimGeneration < entry.StateRevision && entry.QueueSequence > 0 && !entry.CreatedAt.IsZero() && !entry.UpdatedAt.Before(entry.CreatedAt)
}

func validWorkflowIntegrationTarget(entry WorkflowIntegrationEntry) bool {
	repository, err := workflowCanonicalRepository(entry.Repository)

	return err == nil && repository == entry.Repository && strings.HasPrefix(entry.TargetRef, "refs/heads/") && validWorkflowPRBranch(strings.TrimPrefix(entry.TargetRef, "refs/heads/"))
}

func validWorkflowIntegrationState(state string) bool {
	return slices.Contains([]string{"queued", "preparing", "merging", "validating", "reviewing", "waiting_review", "ready_to_publish", "publishing", "observing", "finalizing", "integrated", "blocked", "paused", "cancelled"}, state)
}

func validateWorkflowIntegrationEntryPhase(entry WorkflowIntegrationEntry) error {
	if entry.PreparationReserved && workflowIntegrationTerminal(entry.State) {
		return workflowIntegrationConflict("terminal entry retains speculative preparation capacity")
	}

	if !slices.Contains([]string{"", "pause", "cancel"}, entry.StopIntent) {
		return workflowIntegrationConflict("has invalid stop intent")
	}

	if err := validateWorkflowIntegrationBlocker(entry); err != nil {
		return err
	}

	if slices.Contains([]string{"reviewing", "waiting_review", "ready_to_publish", "publishing", "observing", "finalizing", "integrated"}, entry.State) && entry.CandidateRef == nil {
		return workflowIntegrationConflict("requires exact candidate evidence")
	}

	return validateWorkflowIntegrationPublicationRefs(entry)
}

func validateWorkflowIntegrationPublicationRefs(entry WorkflowIntegrationEntry) error {
	if slices.Contains([]string{"publishing", "observing", "finalizing", "integrated"}, entry.State) && (entry.SubmissionRef == nil || !validWorkflowSubmissionProjection(*entry.SubmissionRef)) {
		return workflowIntegrationConflict("requires retained submission authority")
	}

	if (entry.State == "finalizing" || entry.State == "integrated") && entry.ReceiptRef == nil {
		return workflowIntegrationConflict("requires verified receipt evidence")
	}

	return nil
}

func validateWorkflowIntegrationBlocker(entry WorkflowIntegrationEntry) error {
	if entry.State == "blocked" {
		if !validWorkflowIntegrationBlocker(entry.Blocker) {
			return workflowIntegrationConflict("requires a typed reconciliation blocker")
		}
	} else if entry.Blocker != nil {
		return workflowIntegrationConflict("retains a blocker outside blocked state")
	}

	return nil
}

func validWorkflowIntegrationBlocker(blocker *WorkflowIntegrationBlocker) bool {
	return blocker != nil && workflowstate.ValidRequestKey(blocker.Reason) && workflowstate.ValidRequestKey(blocker.RequiredAction) && validWorkflowIntegrationState(blocker.ResumeState) && !slices.Contains([]string{"blocked", "cancelled", "integrated"}, blocker.ResumeState)
}

func validateWorkflowIntegrationBindings(run workflowInvocationFile) error {
	if err := validateWorkflowIntegrationBudgetBinding(run); err != nil {
		return err
	}

	if len(run.IntegrationEntries) > WorkflowRecordLimit {
		return workflowIntegrationConflict("history exceeds bounds")
	}

	seen := map[WorkflowRef]bool{}

	current := run.Integration == nil
	for _, binding := range run.IntegrationEntries {
		if seen[binding.Unit] || !validWorkflowIntegrationBinding(run, binding) {
			return workflowIntegrationConflict("has duplicate or foreign invocation binding")
		}

		if err := validateWorkflowEvidenceRef(run.ID, binding.EntryRef); err != nil {
			return err
		}

		seen[binding.Unit] = true
		current = current || workflowIntegrationCurrentBinding(run, binding)
	}

	if !current {
		return workflowIntegrationConflict("current pointer differs from retained history")
	}

	return nil
}

func validWorkflowIntegrationBinding(run workflowInvocationFile, binding WorkflowIntegrationBinding) bool {
	return binding.EntryID == workflowIntegrationEntryID(run, binding.Unit) && binding.Revision > 0 && binding.EntryRef.Path == workflowIntegrationEntryPath(run.ID, binding.EntryID)
}

func workflowIntegrationCurrentBinding(run workflowInvocationFile, binding WorkflowIntegrationBinding) bool {
	return run.Integration != nil && *run.Integration == binding
}

func workflowIntegrationNextState(entry WorkflowIntegrationEntry, next string) error {
	allowed := workflowIntegrationSuccessors(entry.State)
	if entry.State == "blocked" && entry.Blocker != nil {
		allowed = []string{entry.Blocker.ResumeState}
	}

	if !slices.Contains(allowed, next) {
		return fmt.Errorf("%w: integration phase %s cannot advance to %s", workflowstate.ErrConflict, entry.State, next)
	}

	return nil
}

func workflowIntegrationSuccessors(state string) []string {
	switch state {
	case "queued":
		return []string{"preparing", "paused", "cancelled", "blocked"}
	case "preparing", "merging", "validating", "reviewing", "waiting_review":
		return workflowIntegrationPreparationSuccessors(state)
	case "ready_to_publish":
		return []string{"publishing", "queued", "blocked", "paused", "cancelled"}
	case "publishing":
		return []string{"observing", "queued", "blocked"}
	case "observing":
		return []string{"finalizing", "blocked"}
	case "finalizing":
		return []string{"integrated", "blocked"}
	case "paused":
		return []string{"queued", "cancelled"}
	default:
		return nil
	}
}

func workflowIntegrationPreparationSuccessors(state string) []string {
	allowed := []string{"queued", "blocked", "paused", "cancelled"}

	switch state {
	case "preparing":
		return append(allowed, "merging", "validating")
	case "merging":
		return append(allowed, "validating")
	case "validating":
		return append(allowed, "reviewing", "merging")
	case "reviewing":
		return append(allowed, "waiting_review", "ready_to_publish", "merging")
	case "waiting_review":
		return append(allowed, "ready_to_publish", "merging")
	default:
		return nil
	}
}

func validateWorkflowIntegrationBudgetBinding(run workflowInvocationFile) error {
	b := run.IntegrationBudget
	if b == nil {
		return nil
	}

	if b.Revision == 0 || b.Ref.Path != workflowIntegrationBudgetPath(run.ID) || !validWorkflowNativeFingerprint(b.Ref.Hash) {
		return workflowstate.ErrConflict
	}

	return nil
}
