package core

import (
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedMaterializationBinding retains owner identities without copying native task state.
func GuidedMaterializationBinding(result workflowcore.MaterializationResult, input generationcore.GeneratedWorkflowPlan) (authoringcore.GuidedMaterializationReference, error) {
	if result.Set.Set == nil || result.Set.Set.PlanRevision == nil || result.Set.Set.UseCase != input.UseCase || len(result.Tasks) == 0 || len(result.Tasks) > workflowstate.PageLimit {
		return authoringcore.GuidedMaterializationReference{}, workflowstate.ErrConflict
	}

	ref := authoringcore.GuidedMaterializationReference{ID: input.MaterializationID, ReceiptRef: result.ReceiptRef, SetID: result.Set.ID, SetPath: result.Set.Path, PlanRevision: *result.Set.Set.PlanRevision}
	for _, task := range result.Tasks {
		if task.Task == nil || task.Task.SetID != ref.SetID {
			return ref, workflowstate.ErrConflict
		}

		ref.Members = append(ref.Members, authoringcore.GuidedWorkMember{Key: task.Task.Key, Path: task.Path, ContractHash: task.Task.ContractHash, SourceHash: task.SourceHash, Revision: task.Revision})
	}

	return ref, nil
}
