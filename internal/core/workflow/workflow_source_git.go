package workflow

import (
	"context"
	"fmt"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PublishWorkflowSource creates or updates only a validated owned source ref.
// An absent ref uses an empty lease; repairs require its retained previous head.
func (git OSWorkflowIntegrationGit) PublishWorkflowSource(ctx context.Context, query SourceQuery) error {
	branch := strings.TrimPrefix(query.Branch, "refs/heads/")
	if err := git.validateWorkflowSource(query, branch); err != nil {
		return err
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: query.Directory}}

	tree, err := reader.tree(ctx, query.Commit)
	if err != nil || tree != query.Tree {
		return errorsOrConflict(err)
	}

	ref := "refs/heads/" + branch

	previous, err := reader.workflowSourceHead(ctx, git.transport, ref)
	if err != nil || previous == query.Commit {
		return err
	}

	return reader.pushWorkflowSource(ctx, git.transport, query, ref, previous)
}

func (git OSWorkflowIntegrationGit) validateWorkflowSource(query SourceQuery, branch string) error {
	if err := git.checkWorkflowTargetQuery(TargetQuery{Repository: query.Repository, Directory: query.Directory, Branch: branch}); err != nil {
		return err
	}

	if !strings.HasPrefix(branch, "chr/workflow/") || !workflowCommitID(query.Commit) || !workflowCommitID(query.Tree) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (reader workflowIntegrationReader) pushWorkflowSource(ctx context.Context, transport string, query SourceQuery, ref, previous string) error {
	if previous != "" && previous != query.PreviousCommit {
		return fmt.Errorf("%w: canonical source branch changed", workflowstate.ErrConflict)
	}

	if _, err := reader.Output(ctx, "push", "--quiet", "--force-with-lease="+ref+":"+previous, "--", transport, query.Commit+":"+ref); err != nil {
		return err
	}

	observed, err := reader.workflowSourceHead(ctx, transport, ref)
	if err != nil || observed != query.Commit {
		return errorsOrConflict(err)
	}

	return nil
}

func (reader workflowIntegrationReader) workflowSourceHead(ctx context.Context, transport, ref string) (string, error) {
	output, err := reader.Output(ctx, "ls-remote", "--", transport, ref)
	if err != nil || output == "" {
		return "", err
	}

	fields := strings.Fields(output)
	if len(fields) != 2 || !workflowCommitID(fields[0]) || fields[1] != ref {
		return "", workflowstate.ErrConflict
	}

	return fields[0], nil
}
