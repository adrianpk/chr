package workflow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) regenerateWorkflowMerger(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, reader workflowIntegrationReader, state MergerRepairState, repair IntegrationTreeRepair) ([]MergerRegenerationEvidence, error) {
	evidence := []MergerRegenerationEvidence{}

	tree := repair.ProposalTree
	for _, conflict := range repair.Input.Conflicts {
		if !workflowMergerProcedureRequired(conflict) {
			continue
		}

		if len(evidence) >= 8 {
			return evidence, workflowstate.ErrConflict
		}

		observation, err := s.runWorkflowMergerRegeneration(ctx, claim, prepared, provider, reader, state, conflict, tree)

		evidence = append(evidence, observation)
		if err != nil {
			return evidence, err
		}

		if err := reader.checkMergerObservedPaths(ctx, repair.Input, repair.Proposal, tree, observation.OutputTree, false); err != nil {
			return evidence, err
		}

		tree = observation.OutputTree
	}

	return evidence, nil
}

func (s *Store) runWorkflowMergerRegeneration(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, reader workflowIntegrationReader, state MergerRepairState, conflict MergerConflict, tree string) (MergerRegenerationEvidence, error) {
	evidence := MergerRegenerationEvidence{ConflictID: conflict.ID, InputTree: tree}

	config, err := os.ReadFile(filepath.Join(reader.Directory, ".git", "config"))
	if err != nil {
		return evidence, err
	}

	commandInput := prepared
	commandInput.mergerProcedure = true
	commandInput.directory = prepared.input.Lane.Worktree

	evidence.Command, err = s.runWorkflowMergerProcedure(ctx, claim, commandInput, provider, conflict.RegenerationCommand)
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

	evidence.Validation, err = s.runWorkflowMergerProcedure(ctx, claim, commandInput, provider, conflict.ValidationCommand)
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

func (s *Store) runWorkflowMergerProcedure(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, command []string) (CommandResult, error) {
	if err := s.guardWorkflowMergerAcceptance(ctx, claim, prepared, provider); err != nil {
		return CommandResult{}, err
	}

	if !validWorkflowMergerCommand(command) || time.Now().After(prepared.leaseUntil) {
		return CommandResult{}, workflowstate.ErrConflict
	}

	result, err := workflowRunCommand(ctx, claim, prepared, command)
	if err != nil {
		return result, err
	}

	if result.ExitCode != 0 || result.Truncated {
		return result, workflowstate.ErrConflict
	}

	return result, s.guardWorkflowMergerAcceptance(ctx, claim, prepared, provider)
}

func checkWorkflowMergerLocalConfig(directory string, expected []byte) error {
	actual, err := os.ReadFile(filepath.Join(directory, ".git", "config"))
	if err != nil || !bytes.Equal(expected, actual) {
		return workflowstate.ErrConflict
	}

	return nil
}
