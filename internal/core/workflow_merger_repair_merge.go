package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowMergerTree struct {
	tree      string
	conflicts []string
}

func (reader workflowIntegrationReader) mergerTree(ctx context.Context, original, parent, source string) (workflowMergerTree, error) {
	command := reader.Process(ctx, "merge-tree", "--write-tree", "--name-only", "-z", "--merge-base="+original, parent, source)
	stdout, stderr := &opsfile.ProviderOutput{}, &opsfile.CommandOutput{}
	command.Stdout, command.Stderr = stdout, stderr

	err := command.Run()

	var exit *exec.ExitError
	if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
		return workflowMergerTree{}, fmt.Errorf("merger tree preparation: %w", err)
	}

	if stdout.Overflow {
		return workflowMergerTree{}, workflowstate.ErrConflict
	}

	return parseWorkflowMergerTree(stdout.Buffer.String(), err != nil)
}

func parseWorkflowMergerTree(data string, conflicted bool) (workflowMergerTree, error) {
	fields := strings.Split(data, "\x00")
	if len(fields) < 2 || !workflowCommitID(fields[0]) {
		return workflowMergerTree{}, workflowstate.ErrConflict
	}

	result := workflowMergerTree{tree: fields[0]}
	if !conflicted {
		return result, nil
	}

	for _, path := range fields[1:] {
		if path == "" {
			break
		}

		if !validWorkflowMergerPath(path) || len(result.conflicts) >= workflowstate.PageLimit {
			return workflowMergerTree{}, workflowstate.ErrConflict
		}

		result.conflicts = append(result.conflicts, path)
	}

	slices.Sort(result.conflicts)

	result.conflicts = slices.Compact(result.conflicts)
	if len(result.conflicts) == 0 {
		return workflowMergerTree{}, fmt.Errorf("%w: merge conflict has no bounded file diagnosis", workflowstate.ErrConflict)
	}

	return result, nil
}

func (reader workflowIntegrationReader) materializeMergerTree(ctx context.Context, parent, tree string) error {
	if _, err := reader.Output(ctx, "checkout", "--quiet", "--detach", parent); err != nil {
		return err
	}

	_, err := reader.Output(ctx, "read-tree", "--reset", "-u", tree)

	return err
}
