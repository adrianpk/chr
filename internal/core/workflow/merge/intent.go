package merge

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validCheckpointObjects(input workflowstate.MergerInput, checkpoint workflowstate.MergerCheckpoint) bool {
	return workflowstate.ValidCommitID(checkpoint.OriginalCommit) && workflowstate.ValidCommitID(checkpoint.PreparedParent) && len(checkpoint.OriginalCommit) == len(input.Source) && len(checkpoint.PreparedParent) == len(input.Base)
}

func validCheckpointPosition(input workflowstate.MergerInput, checkpoint workflowstate.MergerCheckpoint) bool {
	return input.Method == "rebase" && checkpoint.ReplayIndex >= 0 && checkpoint.ReplayIndex < workflowstate.RecordLimit && len(checkpoint.Predecessors) == checkpoint.ReplayIndex
}

func validConflictProcedure(conflict workflowstate.MergerConflict) bool {
	return slices.Contains([]string{"text", "generated", "lockfile", "binary", "submodule", "lfs", "semantic"}, conflict.Kind) && (len(conflict.RegenerationCommand) == 0 || ValidCommand(conflict.RegenerationCommand)) && (len(conflict.ValidationCommand) == 0 || ValidCommand(conflict.ValidationCommand))
}

func validPredecessor(replay workflowstate.IntegrationReplay, size int) bool {
	return workflowstate.ValidCommitID(replay.Source) && workflowstate.ValidCommitID(replay.Commit) && len(replay.Source) == size && len(replay.Commit) == size && slices.Contains([]string{"nonempty", "source_empty", "already_applied"}, replay.EmptyDecision)
}

func checkCheckpoint(input workflowstate.MergerInput) error {
	checkpoint := input.HistoryCheckpoint
	if checkpoint == nil {
		return nil
	}

	if !validCheckpointPosition(input, *checkpoint) {
		return workflowstate.ErrConflict
	}

	if !validCheckpointObjects(input, *checkpoint) {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{checkpoint.OriginalCommit: true}
	for _, replay := range checkpoint.Predecessors {
		if !validPredecessor(replay, len(input.Source)) || seen[replay.Source] {
			return workflowstate.ErrConflict
		}

		seen[replay.Source] = true
	}

	parent := input.Base
	if len(checkpoint.Predecessors) > 0 {
		parent = checkpoint.Predecessors[len(checkpoint.Predecessors)-1].Commit
	}

	if parent != checkpoint.PreparedParent {
		return workflowstate.ErrConflict
	}

	return nil
}

func checkConflicts(input workflowstate.MergerInput) error {
	seen := map[string]bool{}
	for _, conflict := range input.Conflicts {
		if !workflowstate.ValidRequestKey(conflict.ID) || seen[conflict.ID] || !PathAllowed(input, conflict.Path) {
			return workflowstate.ErrConflict
		}

		if err := checkRequirementIDs(conflict.RequirementIDs, input.Requirements); err != nil {
			return err
		}

		if !validConflictProcedure(conflict) {
			return workflowstate.ErrConflict
		}

		seen[conflict.ID] = true
	}

	return nil
}

func checkRequirementIDs(ids []string, requirements []workflowstate.MergerRequirement) error {
	if len(ids) == 0 || len(ids) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] || !slices.ContainsFunc(requirements, func(r workflowstate.MergerRequirement) bool { return r.ID == id }) {
			return workflowstate.ErrConflict
		}

		seen[id] = true
	}

	return nil
}

func checkRequirements(requirements []workflowstate.MergerRequirement) error {
	seen := map[string]bool{}
	source, target := false, false

	for _, requirement := range requirements {
		if !workflowstate.ValidRequestKey(requirement.ID) || seen[requirement.ID] || !slices.Contains([]string{"source", "target"}, requirement.Origin) {
			return workflowstate.ErrConflict
		}

		seen[requirement.ID] = true
		source = source || requirement.Origin == "source"
		target = target || requirement.Origin == "target"
	}

	if !source || !target {
		return workflowstate.ErrConflict
	}

	return checkIntentConflicts(requirements)
}
