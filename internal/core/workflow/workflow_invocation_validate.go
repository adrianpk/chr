package workflow

import (
	"fmt"
	"path/filepath"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowInvocation(run workflowInvocationFile) error {
	if err := (Record{Identity: run.Identity}).validateIdentity(); err != nil {
		return err
	}

	if err := validateWorkflowInvocationHeader(run); err != nil {
		return err
	}

	request, err := normalizeWorkflowStart(StartRequest{StartInput: workflowstate.StartInput{
		RequestKey:        run.RequestKey,
		Selection:         run.Selection,
		Policy:            run.Policy,
		ReviewRestriction: run.ReviewRestriction,
		Admission:         run.Admission,
		HarnessBinding:    run.HarnessBinding,
	}, HarnessCommand: run.HarnessCommand})
	if err != nil {
		return err
	}

	hash, err := workflowStartHash(request)
	if err != nil || hash != run.RequestHash {
		return fmt.Errorf("%w: invocation request image changed", workflowstate.ErrConflict)
	}

	if !validWorkflowInvocationBounds(run) {
		return fmt.Errorf("workflow invocation member, cursor or attempt bounds exceeded")
	}

	return validateWorkflowInvocationMembers(run)
}

func validateWorkflowInvocationHeader(run workflowInvocationFile) error {
	if err := validateWorkflowInvocationHarness(run); err != nil {
		return err
	}

	if run.Authority != "tickets" || !filepath.IsAbs(run.OwnerRoot) {
		return fmt.Errorf("unsupported workflow invocation schema, authority or owner")
	}

	if !slices.Contains([]string{"pending", "running", "waiting_capacity", "waiting_review", "waiting_integration", "blocked", "completed", "paused", "cancelled"}, run.Status) {
		return fmt.Errorf("unsupported workflow invocation status")
	}

	if !slices.Contains(run.RequestKeys, run.RequestKey) || len(run.RequestKeys) > workflowstate.PageLimit {
		return fmt.Errorf("workflow request attachment keys are incomplete or oversized")
	}

	if err := validateWorkflowStopState(run); err != nil {
		return err
	}

	if err := validateWorkflowFrontier(run); err != nil {
		return err
	}

	if err := validateWorkflowContinuation(run); err != nil {
		return err
	}

	return validateWorkflowInvocationBindings(run)
}

func validateWorkflowInvocationBindings(run workflowInvocationFile) error {
	if run.Lane != nil && !validWorkflowLane(*run.Lane) {
		return fmt.Errorf("invalid workflow invocation lane")
	}

	if err := validateWorkflowCandidateRef(run); err != nil {
		return err
	}

	if err := validateWorkflowPullRequestBinding(run); err != nil {
		return err
	}

	if err := validateWorkflowDeliveryRef(run); err != nil {
		return err
	}

	if err := validateWorkflowFinalizationRef(run); err != nil {
		return err
	}

	if err := validateWorkflowReviewBinding(run); err != nil {
		return err
	}

	if err := validateWorkflowPreparationBinding(run); err != nil {
		return err
	}

	if err := validateWorkflowSubmissionBinding(run); err != nil {
		return err
	}

	return validateWorkflowIntegrationBindings(run)
}

func validateWorkflowCandidateRef(run workflowInvocationFile) error {
	if run.CandidateRef == "" && run.CandidateHash == "" {
		return nil
	}

	if run.Lane != nil && run.Lane.Unit.Kind == "ticket" && validWorkflowTicketCandidateRef(run) {
		return nil
	}

	for _, attempt := range run.Attempts {
		if workflowAttemptOwnsCandidate(run, attempt) {
			return nil
		}
	}

	return fmt.Errorf("invalid workflow slice candidate binding")
}

func validWorkflowInvocationBounds(run workflowInvocationFile) bool {
	return validWorkflowInvocationExtent(run) && len(run.Members) <= workflowstate.RecordLimit && run.Cursor >= 0 && run.Cursor <= len(run.Members) && len(run.Attempts) <= workflowstate.RecordLimit && len(run.SetImages) <= workflowstate.RecordLimit
}

func validWorkflowInvocationExtent(run workflowInvocationFile) bool {
	return len(run.Members) > 0 || len(run.SetImages) > 0 && (run.Selection.Selector.Extent == "slice_set" || run.Selection.Selector.Extent == "use_case")
}

func validateWorkflowInvocationMembers(run workflowInvocationFile) error {
	if err := validateWorkflowLooseQueue(run); err != nil {
		return err
	}

	keys := map[string]bool{}

	for _, member := range run.Members {
		if keys[member.Admission.Key] || !validWorkflowInvocationMember(member) {
			return fmt.Errorf("invalid workflow frozen task identity or image")
		}

		keys[member.Admission.Key] = true
	}

	return validateWorkflowInvocationAttempts(run, keys)
}

func validWorkflowInvocationMember(member Member) bool {
	input := member.Admission

	return input.Key != "" && input.Revision > 0 && input.SourceHash != "" && input.ContractHash != "" && filepath.IsLocal(input.Path) && member.ExpectedRevision >= input.Revision && member.ExpectedSourceHash != ""
}

func validateWorkflowInvocationAttempts(run workflowInvocationFile, keys map[string]bool) error {
	ids := map[string]bool{}
	previous := map[string]Attempt{}

	for _, attempt := range run.Attempts {
		if ids[attempt.ID] || !keys[attempt.TaskKey] || !validWorkflowAttemptIdentity(attempt, run.Owners[attempt.ID]) {
			return fmt.Errorf("invalid workflow attempt identity or fence")
		}

		if !validWorkflowReplacement(attempt, previous) {
			return fmt.Errorf("invalid workflow replacement history")
		}

		previous[attempt.ID] = attempt

		ids[attempt.ID] = true
		if err := validateWorkflowAttemptState(attempt); err != nil {
			return err
		}

		if err := validateWorkflowAttemptExecution(run.ID, attempt); err != nil {
			return err
		}
	}

	if len(ids) != len(run.Owners) {
		return fmt.Errorf("workflow attempt credential bindings differ")
	}

	return nil
}

func validateWorkflowAttemptExecution(invocationID string, attempt Attempt) error {
	if filepath.Base(attempt.ID) != attempt.ID {
		return fmt.Errorf("invalid workflow attempt path identity")
	}

	if attempt.InputRef != "" {
		if attempt.InputRef != workflowAttemptFile(invocationID, attempt.ID, "input.json") || len(attempt.InputHash) != 64 || attempt.Lane == nil {
			return fmt.Errorf("invalid workflow attempt input binding")
		}
	}

	if attempt.Lane != nil && !validWorkflowLane(*attempt.Lane) {
		return fmt.Errorf("invalid workflow attempt lane")
	}

	if err := validateWorkflowAttemptProcesses(invocationID, attempt); err != nil {
		return err
	}

	return validateWorkflowAttemptResult(invocationID, attempt)
}

func validateWorkflowAttemptResult(invocationID string, attempt Attempt) error {
	if attempt.ResultRef != "" && (attempt.ResultRef != workflowAttemptFile(invocationID, attempt.ID, "result.json") || len(attempt.ResultHash) != 64 || attempt.InputRef == "" || attempt.FinishedAt.IsZero()) {
		return fmt.Errorf("invalid workflow attempt result binding")
	}

	if slices.Contains([]string{"validated", "failed", "reviewed", "prepared"}, attempt.Outcome) && attempt.ResultRef == "" {
		return fmt.Errorf("workflow attempt outcome requires a retained result")
	}

	return nil
}

func validWorkflowAttemptIdentity(attempt Attempt, ownerHash string) bool {
	return attempt.ID != "" && attempt.Generation > 0 && attempt.InputRevision > 0 && attempt.ContractHash != "" && len(ownerHash) == 64
}

func (s *Store) guardWorkflowAdmission(runs []workflowInvocationFile, next workflowInvocationFile) error {
	resources := workflowClaimResources(next)

	for _, run := range runs {
		if run.Namespace != s.layout.Namespace() || workflowInvocationTerminal(run.Status) {
			continue
		}

		for resource := range workflowClaimResources(run) {
			if resources[resource] {
				return fmt.Errorf("%w: %s owned by invocation %s", ErrWorkflowOwned, resource, run.ID)
			}
		}
	}

	return nil
}

func workflowClaimResources(run workflowInvocationFile) map[string]bool {
	resources := map[string]bool{}

	for _, input := range run.Members {
		member := input.Admission

		resources["ticket:"+member.Key] = true
		if member.SliceID != "" {
			resources["slice:"+member.SliceID] = true
			resources["set:"+member.SetID] = true
		}
	}

	for _, image := range run.SetImages {
		resources["set:"+workflowInvocationSetID(image)] = true

		if image.UseCase != "" {
			resources["use_case:"+image.UseCase] = true
		}
	}

	return resources
}

func validateWorkflowStopState(run workflowInvocationFile) error {
	if !slices.Contains([]string{"", "pause", "cancel"}, run.StopIntent) {
		return fmt.Errorf("invalid workflow stop intent")
	}

	if !slices.Contains([]string{"", "pending", "waiting_capacity", "waiting_review", "waiting_integration", "blocked", "paused"}, run.ResumeStatus) {
		return fmt.Errorf("invalid workflow resume status")
	}

	if !slices.Contains([]string{"", "pending", "waiting_review", "waiting_integration", "blocked", "paused"}, run.CapacityResumeStatus) {
		return fmt.Errorf("invalid workflow capacity resume status")
	}

	return nil
}

func validWorkflowReplacement(attempt Attempt, previous map[string]Attempt) bool {
	if attempt.Replaces == "" {
		return true
	}

	replaced, exists := previous[attempt.Replaces]

	return exists && replaced.TaskKey == attempt.TaskKey && replaced.Purpose == attempt.Purpose && replaced.Generation < attempt.Generation && replaced.Outcome != "validated"
}

func validateWorkflowAttemptState(attempt Attempt) error {
	if !slices.Contains([]string{"", workflowSliceValidationPurpose, workflowSliceRepairPurpose, workflowSliceReviewPurpose, workflowIntegrationValidationPurpose, workflowMergerPurpose}, attempt.Purpose) {
		return fmt.Errorf("invalid workflow attempt purpose")
	}

	if attempt.Purpose == workflowSliceRepairPurpose {
		return validateWorkflowRepairAttempt(attempt)
	}

	if attempt.Outcome == "reviewed" && attempt.Purpose != workflowSliceReviewPurpose {
		return workflowstate.ErrConflict
	}

	if attempt.Outcome == "prepared" && attempt.Purpose != workflowIntegrationValidationPurpose {
		return workflowstate.ErrConflict
	}

	if attempt.StartedAt.IsZero() || !attempt.LeaseUntil.After(attempt.StartedAt) || !slices.Contains([]string{"running", "checkpointed", "validated", "failed", "paused", "cancelled", "interrupted", "recovered", "reviewed", "prepared"}, attempt.Outcome) {
		return fmt.Errorf("invalid workflow attempt lease or outcome")
	}

	return nil
}

func validateWorkflowRepairAttempt(attempt Attempt) error {
	if attempt.StartedAt.IsZero() || !attempt.LeaseUntil.After(attempt.StartedAt) || attempt.Outcome != "repaired" || attempt.FinishedAt.IsZero() || attempt.InputRef != "" || len(attempt.Processes) != 0 {
		return fmt.Errorf("invalid workflow repair attempt")
	}

	return nil
}
