package core

import "slices"

func validateWorkflowMergerRequirements(requirements []WorkflowMergerRequirement) error {
	seen := map[string]bool{}
	source, target := false, false

	for _, requirement := range requirements {
		if !validWorkflowRequestKey(requirement.ID) || seen[requirement.ID] || !slices.Contains([]string{"source", "target"}, requirement.Origin) {
			return ErrWorkflowConflict
		}

		seen[requirement.ID] = true
		source = source || requirement.Origin == "source"
		target = target || requirement.Origin == "target"
	}

	if !source || !target {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowMergerRequirementIDs(ids []string, requirements []WorkflowMergerRequirement) error {
	if len(ids) == 0 || len(ids) > WorkflowPageLimit {
		return ErrWorkflowConflict
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] || !slices.ContainsFunc(requirements, func(r WorkflowMergerRequirement) bool { return r.ID == id }) {
			return ErrWorkflowConflict
		}

		seen[id] = true
	}

	return nil
}

func validateWorkflowMergerConflicts(input WorkflowMergerInput) error {
	seen := map[string]bool{}
	for _, conflict := range input.Conflicts {
		if !validWorkflowRequestKey(conflict.ID) || seen[conflict.ID] || !workflowMergerPathAllowed(input, conflict.Path) {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowMergerRequirementIDs(conflict.RequirementIDs, input.Requirements); err != nil {
			return err
		}

		if !slices.Contains([]string{"text", "generated", "lockfile", "binary", "submodule", "lfs", "semantic"}, conflict.Kind) {
			return ErrWorkflowConflict
		}

		if len(conflict.RegenerationCommand) > 0 && !validWorkflowMergerCommand(conflict.RegenerationCommand) {
			return ErrWorkflowConflict
		}

		seen[conflict.ID] = true
	}

	return nil
}

func validateWorkflowMergerCheckpoint(input WorkflowMergerInput) error {
	checkpoint := input.HistoryCheckpoint
	if checkpoint == nil {
		return nil
	}

	if !validWorkflowMergerCheckpointPosition(input, *checkpoint) {
		return ErrWorkflowConflict
	}

	if !validWorkflowMergerCheckpointObjects(input, *checkpoint) {
		return ErrWorkflowConflict
	}

	seen := map[string]bool{checkpoint.OriginalCommit: true}
	for _, replay := range checkpoint.Predecessors {
		if !validWorkflowMergerPredecessor(replay, len(input.Source)) || seen[replay.Source] {
			return ErrWorkflowConflict
		}

		seen[replay.Source] = true
	}

	parent := input.Base
	if len(checkpoint.Predecessors) > 0 {
		parent = checkpoint.Predecessors[len(checkpoint.Predecessors)-1].Commit
	}

	if parent != checkpoint.PreparedParent {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowMergerCheckpointPosition(input WorkflowMergerInput, checkpoint WorkflowMergerCheckpoint) bool {
	return input.Method == "rebase" && checkpoint.ReplayIndex >= 0 && checkpoint.ReplayIndex < WorkflowRecordLimit && len(checkpoint.Predecessors) == checkpoint.ReplayIndex
}

func validWorkflowMergerCheckpointObjects(input WorkflowMergerInput, checkpoint WorkflowMergerCheckpoint) bool {
	return workflowCommitID(checkpoint.OriginalCommit) && workflowCommitID(checkpoint.PreparedParent) && len(checkpoint.OriginalCommit) == len(input.Source) && len(checkpoint.PreparedParent) == len(input.Base)
}

func validWorkflowMergerPredecessor(replay WorkflowIntegrationReplay, size int) bool {
	return workflowCommitID(replay.Source) && workflowCommitID(replay.Commit) && len(replay.Source) == size && len(replay.Commit) == size && slices.Contains([]string{"nonempty", "source_empty", "already_applied"}, replay.EmptyDecision)
}
