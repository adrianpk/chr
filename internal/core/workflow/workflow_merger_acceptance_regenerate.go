package workflow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) regenerateWorkflowMerger(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider, reader workflowIntegrationReader, input workflowMergerRepairInput) ([]workflowstate.MergerRegenerationEvidence, error) {
	state, repair := input.State, input.Repair
	evidence := []workflowstate.MergerRegenerationEvidence{}

	tree := repair.ProposalTree
	for _, conflict := range repair.Input.Conflicts {
		if !workflowmerge.ProcedureRequired(conflict) {
			continue
		}

		if len(evidence) >= 8 {
			return evidence, workflowstate.ErrConflict
		}

		observation, err := s.runWorkflowMergerRegeneration(ctx, claimed, provider, reader, workflowMergerRegenerationInput{State: state, Conflict: conflict, Tree: tree})

		evidence = append(evidence, observation)
		if err != nil {
			return evidence, err
		}

		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, workflowMergerPathCheck{Before: tree, After: observation.OutputTree, Mode: workflowMergerDeclaredPaths}); err != nil {
			return evidence, err
		}

		tree = observation.OutputTree
	}

	return evidence, nil
}

func (s *Store) runWorkflowMergerRegeneration(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider, reader workflowIntegrationReader, input workflowMergerRegenerationInput) (workflowstate.MergerRegenerationEvidence, error) {
	state, conflict, tree := input.State, input.Conflict, input.Tree
	claim, prepared := claimed.Claim, claimed.Prepared
	evidence := workflowstate.MergerRegenerationEvidence{ConflictID: conflict.ID, InputTree: tree}

	config, err := os.ReadFile(filepath.Join(reader.Directory, ".git", "config"))
	if err != nil {
		return evidence, err
	}

	commandInput := prepared
	commandInput.mergerProcedure = true
	commandInput.directory = prepared.input.Lane.Worktree

	evidence.Command, err = s.runWorkflowMergerProcedure(ctx, workflowClaimedMerger{Claim: claim, Prepared: commandInput}, provider, conflict.RegenerationCommand)
	if err != nil {
		return evidence, err
	}

	if err := checkWorkflowMergerLocalConfig(reader.Directory, config); err != nil {
		return evidence, err
	}

	evidence.OutputTree, err = workflowMergerSnapshot(ctx, reader, state)
	if err != nil {
		return evidence, err
	}

	evidence.Validation, err = s.runWorkflowMergerProcedure(ctx, workflowClaimedMerger{Claim: claim, Prepared: commandInput}, provider, conflict.ValidationCommand)
	if err != nil {
		return evidence, err
	}

	if err := checkWorkflowMergerLocalConfig(reader.Directory, config); err != nil {
		return evidence, err
	}

	after, err := workflowMergerSnapshot(ctx, reader, state)
	if err != nil || after != evidence.OutputTree {
		return evidence, workflowstate.ErrConflict
	}

	return evidence, nil
}

func (s *Store) runWorkflowMergerProcedure(ctx context.Context, claimed workflowClaimedMerger, provider PullRequestProvider, command []string) (workflowstate.CommandResult, error) {
	claim, prepared := claimed.Claim, claimed.Prepared
	if err := s.guardWorkflowMergerAcceptance(ctx, claimed, provider); err != nil {
		return workflowstate.CommandResult{}, err
	}

	if !workflowmerge.ValidCommand(command) || time.Now().After(prepared.leaseUntil) {
		return workflowstate.CommandResult{}, workflowstate.ErrConflict
	}

	result, err := workflowRunCommand(ctx, claim, prepared, command)
	if err != nil {
		return result, err
	}

	if result.ExitCode != 0 || result.Truncated {
		return result, workflowstate.ErrConflict
	}

	return result, s.guardWorkflowMergerAcceptance(ctx, claimed, provider)
}

func checkWorkflowMergerLocalConfig(directory string, expected []byte) error {
	actual, err := os.ReadFile(filepath.Join(directory, ".git", "config"))
	if err != nil || !bytes.Equal(expected, actual) {
		return workflowstate.ErrConflict
	}

	return nil
}
