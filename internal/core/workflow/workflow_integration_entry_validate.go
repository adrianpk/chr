package workflow

import (
	"fmt"
	"slices"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowIntegrationEntry(entry IntegrationEntry) error {
	if !validWorkflowIntegrationEntryHeader(entry) {
		return workflowstate.IntegrationConflict("has invalid version, identity or revision")
	}

	if !validWorkflowIntegrationTarget(entry) {
		return workflowstate.IntegrationConflict("has invalid canonical target")
	}

	if !validWorkflowIntegrationState(entry.State) || entry.BudgetRef != workflowIntegrationBudgetPath(entry.InvocationID) {
		return workflowstate.IntegrationConflict("has invalid phase or budget owner")
	}

	for _, ref := range []*workflowstate.EvidenceRef{entry.CandidateRef, entry.ReceiptRef, entry.DecisionRef} {
		if ref != nil {
			if err := workflowstate.CheckEvidenceRef(entry.InvocationID, *ref); err != nil {
				return err
			}
		}
	}

	return validateWorkflowIntegrationEntryPhase(entry)
}

func validWorkflowIntegrationEntryHeader(entry IntegrationEntry) bool {
	return validWorkflowIntegrationIdentity(entry) && validWorkflowIntegrationRevisions(entry) && entry.Unit.ID != "" && (entry.Unit.Kind == "ticket" || entry.Unit.Kind == "slice")
}

func validWorkflowIntegrationIdentity(entry IntegrationEntry) bool {
	return entry.SchemaVersion == 1 && workflowstate.ValidFingerprint(entry.ID) && workflowstate.ValidRequestKey(entry.InvocationID) && entry.ProjectID != "" && opsnamespace.Valid(entry.Namespace) && workflowstate.ValidFingerprint(entry.CoordinatorID) && workflowstate.ValidFingerprint(entry.FrozenScopeHash) && workflowstate.ValidRequestKey(entry.AdmissionKey)
}

func validWorkflowIntegrationRevisions(entry IntegrationEntry) bool {
	return entry.Revision > 0 && entry.StateRevision > 0 && entry.StateRevision <= entry.Revision && entry.ClaimGeneration < entry.StateRevision && entry.QueueSequence > 0 && !entry.CreatedAt.IsZero() && !entry.UpdatedAt.Before(entry.CreatedAt)
}

func validWorkflowIntegrationTarget(entry IntegrationEntry) bool {
	repository, err := workflowstate.CanonicalRepository(entry.Repository)

	return err == nil && repository == entry.Repository && strings.HasPrefix(entry.TargetRef, "refs/heads/") && workflowstate.ValidBranch(strings.TrimPrefix(entry.TargetRef, "refs/heads/"))
}

func validWorkflowIntegrationState(state IntegrationState) bool {
	return slices.Contains([]IntegrationState{IntegrationQueued, IntegrationPreparing, IntegrationMerging, IntegrationValidating, IntegrationReviewing, IntegrationWaitingReview, IntegrationReadyToPublish, IntegrationPublishing, IntegrationObserving, IntegrationFinalizing, IntegrationIntegrated, IntegrationBlocked, IntegrationPaused, IntegrationCancelled}, state)
}

func validateWorkflowIntegrationEntryPhase(entry IntegrationEntry) error {
	if entry.PreparationReserved && workflowIntegrationTerminal(entry.State) {
		return workflowstate.IntegrationConflict("terminal entry retains speculative preparation capacity")
	}

	if !slices.Contains([]string{"", "pause", "cancel"}, entry.StopIntent) {
		return workflowstate.IntegrationConflict("has invalid stop intent")
	}

	if err := validateWorkflowIntegrationBlocker(entry); err != nil {
		return err
	}

	if slices.Contains([]IntegrationState{IntegrationReviewing, IntegrationWaitingReview, IntegrationReadyToPublish, IntegrationPublishing, IntegrationObserving, IntegrationFinalizing, IntegrationIntegrated}, entry.State) && entry.CandidateRef == nil {
		return workflowstate.IntegrationConflict("requires exact candidate evidence")
	}

	return validateWorkflowIntegrationPublicationRefs(entry)
}

func validateWorkflowIntegrationPublicationRefs(entry IntegrationEntry) error {
	if slices.Contains([]IntegrationState{IntegrationPublishing, IntegrationObserving, IntegrationFinalizing, IntegrationIntegrated}, entry.State) && (entry.SubmissionRef == nil || !validWorkflowSubmissionProjection(*entry.SubmissionRef)) {
		return workflowstate.IntegrationConflict("requires retained submission authority")
	}

	if (entry.State == IntegrationFinalizing || entry.State == IntegrationIntegrated) && entry.ReceiptRef == nil {
		return workflowstate.IntegrationConflict("requires verified receipt evidence")
	}

	return nil
}

func validateWorkflowIntegrationBlocker(entry IntegrationEntry) error {
	if entry.State == IntegrationBlocked {
		if !validWorkflowIntegrationBlocker(entry.Blocker) {
			return workflowstate.IntegrationConflict("requires a typed reconciliation blocker")
		}
	} else if entry.Blocker != nil {
		return workflowstate.IntegrationConflict("retains a blocker outside blocked state")
	}

	return nil
}

func validWorkflowIntegrationBlocker(blocker *IntegrationBlocker) bool {
	return blocker != nil && workflowstate.ValidRequestKey(blocker.Reason) && workflowstate.ValidRequestKey(blocker.RequiredAction) && validWorkflowIntegrationState(blocker.ResumeState) && !slices.Contains([]IntegrationState{IntegrationBlocked, IntegrationCancelled, IntegrationIntegrated}, blocker.ResumeState)
}

func validateWorkflowIntegrationBindings(run workflowInvocationFile) error {
	if err := validateWorkflowIntegrationBudgetBinding(run); err != nil {
		return err
	}

	if len(run.IntegrationEntries) > workflowstate.RecordLimit {
		return workflowstate.IntegrationConflict("history exceeds bounds")
	}

	seen := map[workflowstate.Ref]bool{}

	current := run.Integration == nil
	for _, binding := range run.IntegrationEntries {
		if seen[binding.Unit] || !validWorkflowIntegrationBinding(run, binding) {
			return workflowstate.IntegrationConflict("has duplicate or foreign invocation binding")
		}

		if err := workflowstate.CheckEvidenceRef(run.ID, binding.EntryRef); err != nil {
			return err
		}

		seen[binding.Unit] = true
		current = current || workflowIntegrationCurrentBinding(run, binding)
	}

	if !current {
		return workflowstate.IntegrationConflict("current pointer differs from retained history")
	}

	return nil
}

func validWorkflowIntegrationBinding(run workflowInvocationFile, binding workflowstate.IntegrationBinding) bool {
	return binding.EntryID == workflowIntegrationEntryID(run, binding.Unit) && binding.Revision > 0 && binding.EntryRef.Path == workflowIntegrationEntryPath(run.ID, binding.EntryID)
}

func workflowIntegrationCurrentBinding(run workflowInvocationFile, binding workflowstate.IntegrationBinding) bool {
	return run.Integration != nil && *run.Integration == binding
}

func workflowIntegrationNextState(entry IntegrationEntry, next IntegrationState) error {
	allowed := workflowIntegrationSuccessors(entry.State)
	if entry.State == IntegrationBlocked && entry.Blocker != nil {
		allowed = []IntegrationState{entry.Blocker.ResumeState}
	}

	if !slices.Contains(allowed, next) {
		return fmt.Errorf("%w: integration phase %s cannot advance to %s", workflowstate.ErrConflict, entry.State, next)
	}

	return nil
}

func workflowIntegrationSuccessors(state IntegrationState) []IntegrationState {
	switch state {
	case IntegrationQueued:
		return []IntegrationState{IntegrationPreparing, IntegrationPaused, IntegrationCancelled, IntegrationBlocked}
	case IntegrationPreparing, IntegrationMerging, IntegrationValidating, IntegrationReviewing, IntegrationWaitingReview:
		return workflowIntegrationPreparationSuccessors(state)
	case IntegrationReadyToPublish:
		return []IntegrationState{IntegrationPublishing, IntegrationQueued, IntegrationBlocked, IntegrationPaused, IntegrationCancelled}
	case IntegrationPublishing:
		return []IntegrationState{IntegrationObserving, IntegrationQueued, IntegrationBlocked}
	case IntegrationObserving:
		return []IntegrationState{IntegrationFinalizing, IntegrationBlocked}
	case IntegrationFinalizing:
		return []IntegrationState{IntegrationIntegrated, IntegrationBlocked}
	case IntegrationPaused:
		return []IntegrationState{IntegrationQueued, IntegrationCancelled}
	default:
		return nil
	}
}

func workflowIntegrationPreparationSuccessors(state IntegrationState) []IntegrationState {
	allowed := []IntegrationState{IntegrationQueued, IntegrationBlocked, IntegrationPaused, IntegrationCancelled}

	switch state {
	case IntegrationPreparing:
		return append(allowed, IntegrationMerging, IntegrationValidating)
	case IntegrationMerging:
		return append(allowed, IntegrationValidating)
	case IntegrationValidating:
		return append(allowed, IntegrationReviewing, IntegrationMerging)
	case IntegrationReviewing:
		return append(allowed, IntegrationWaitingReview, IntegrationReadyToPublish, IntegrationMerging)
	case IntegrationWaitingReview:
		return append(allowed, IntegrationReadyToPublish, IntegrationMerging)
	default:
		return nil
	}
}

func validateWorkflowIntegrationBudgetBinding(run workflowInvocationFile) error {
	b := run.IntegrationBudget
	if b == nil {
		return nil
	}

	if b.Revision == 0 || b.Ref.Path != workflowIntegrationBudgetPath(run.ID) || !workflowstate.ValidFingerprint(b.Ref.Hash) {
		return workflowstate.ErrConflict
	}

	return nil
}
