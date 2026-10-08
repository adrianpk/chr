package core

import (
	"fmt"
	"path/filepath"
	"slices"
)

func validateWorkflowInvocation(run workflowInvocationFile) error {
	if err := (WorkflowRecord{WorkflowIdentity: run.WorkflowIdentity}).validateIdentity(); err != nil {
		return err
	}

	if err := validateWorkflowInvocationHeader(run); err != nil {
		return err
	}

	request, err := normalizeWorkflowStart(WorkflowStartRequest{RequestKey: run.RequestKey, Selection: run.Selection, Policy: run.Policy})
	if err != nil {
		return err
	}

	hash, err := workflowStartHash(request)
	if err != nil || hash != run.RequestHash {
		return fmt.Errorf("%w: invocation request image changed", ErrWorkflowConflict)
	}

	if !validWorkflowInvocationBounds(run) {
		return fmt.Errorf("workflow invocation member, cursor or attempt bounds exceeded")
	}

	return validateWorkflowInvocationMembers(run)
}

func validateWorkflowInvocationHeader(run workflowInvocationFile) error {
	if run.Schema != "chr-invocation/v1" || run.Authority != "tickets" || !filepath.IsAbs(run.OwnerRoot) {
		return fmt.Errorf("unsupported workflow invocation schema, authority or owner")
	}

	if !slices.Contains([]string{"pending", "running", "waiting_capacity", "waiting_review", "waiting_integration", "blocked", "completed", "paused", "cancelled"}, run.Status) {
		return fmt.Errorf("unsupported workflow invocation status")
	}

	if !slices.Contains(run.RequestKeys, run.RequestKey) || len(run.RequestKeys) > WorkflowPageLimit {
		return fmt.Errorf("workflow request attachment keys are incomplete or oversized")
	}

	return nil
}

func validWorkflowInvocationBounds(run workflowInvocationFile) bool {
	return len(run.Members) >= 1 && len(run.Members) <= WorkflowRecordLimit && run.Cursor >= 0 && run.Cursor <= len(run.Members) && len(run.Attempts) <= WorkflowRecordLimit && len(run.SetImages) <= WorkflowRecordLimit
}

func validateWorkflowInvocationMembers(run workflowInvocationFile) error {
	keys := map[string]bool{}

	for _, member := range run.Members {
		if keys[member.Admission.Key] || !validWorkflowInvocationMember(member) {
			return fmt.Errorf("invalid workflow frozen task identity or image")
		}

		keys[member.Admission.Key] = true
	}

	return validateWorkflowInvocationAttempts(run, keys)
}

func validWorkflowInvocationMember(member WorkflowMember) bool {
	input := member.Admission

	return input.Key != "" && input.Revision > 0 && input.SourceHash != "" && input.ContractHash != "" && filepath.IsLocal(input.Path) && member.ExpectedRevision >= input.Revision && member.ExpectedSourceHash != ""
}

func validateWorkflowInvocationAttempts(run workflowInvocationFile, keys map[string]bool) error {
	ids := map[string]bool{}
	for _, attempt := range run.Attempts {
		if ids[attempt.ID] || !keys[attempt.TaskKey] || !validWorkflowAttemptIdentity(attempt, run.Owners[attempt.ID]) {
			return fmt.Errorf("invalid workflow attempt identity or fence")
		}

		ids[attempt.ID] = true
		if attempt.StartedAt.IsZero() || !attempt.LeaseUntil.After(attempt.StartedAt) || !slices.Contains([]string{"running", "checkpointed"}, attempt.Outcome) {
			return fmt.Errorf("invalid workflow attempt lease or outcome")
		}
	}

	if len(ids) != len(run.Owners) {
		return fmt.Errorf("workflow attempt credential bindings differ")
	}

	return nil
}

func validWorkflowAttemptIdentity(attempt WorkflowAttempt, ownerHash string) bool {
	return attempt.ID != "" && attempt.Generation > 0 && attempt.InputRevision > 0 && attempt.ContractHash != "" && len(ownerHash) == 64
}

func (s *WorkflowStore) guardWorkflowAdmission(runs []workflowInvocationFile, next workflowInvocationFile) error {
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
		if image.UseCase != "" {
			resources["use_case:"+image.UseCase] = true
		}
	}

	return resources
}
