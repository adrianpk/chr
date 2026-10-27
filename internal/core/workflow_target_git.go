package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowTargetQuery identifies the host checkout receiving canonical objects.
// Fetching objects must not move that checkout's branch or working tree.
type WorkflowTargetQuery struct {
	Repository string
	Branch     string
	Directory  string
}

// WorkflowTargetGit refreshes a canonical branch and makes its commit available.
// It is trusted host IO, never a worker-supplied target or source of authority.
type WorkflowTargetGit interface {
	ObserveWorkflowTarget(context.Context, WorkflowTargetQuery) (workflowstate.PreviewTarget, error)
}

// WorkflowTargetReader observes a remote target without hydrating local objects.
type WorkflowTargetReader interface {
	ReadWorkflowTarget(context.Context, WorkflowTargetQuery) (workflowstate.PreviewTarget, error)
}

// ReadWorkflowTarget leaves local refs, objects and working files unchanged.
func (git OSWorkflowIntegrationGit) ReadWorkflowTarget(ctx context.Context, query WorkflowTargetQuery) (workflowstate.PreviewTarget, error) {
	if err := git.checkWorkflowTargetQuery(query); err != nil {
		return workflowstate.PreviewTarget{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader := workflowIntegrationReader{dir: query.Directory}
	commit, err := reader.remoteTarget(ctx, git.transport, query.Branch)

	return workflowstate.PreviewTarget{Repository: query.Repository, Commit: commit}, err
}

// ObserveWorkflowTarget fetches objects while preserving checkout refs and files.
// It observes the exact branch again after IO and rejects a moved observation.
func (git OSWorkflowIntegrationGit) ObserveWorkflowTarget(ctx context.Context, query WorkflowTargetQuery) (workflowstate.PreviewTarget, error) {
	if err := git.checkWorkflowTargetQuery(query); err != nil {
		return workflowstate.PreviewTarget{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	reader := workflowIntegrationReader{dir: query.Directory}

	commit, err := reader.remoteTarget(ctx, git.transport, query.Branch)
	if err != nil {
		return workflowstate.PreviewTarget{}, err
	}

	if _, err := reader.output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-write-fetch-head", "--", git.transport, "refs/heads/"+query.Branch); err != nil {
		return workflowstate.PreviewTarget{}, err
	}

	if _, err := reader.output(ctx, "cat-file", "-e", commit+"^{commit}"); err != nil {
		return workflowstate.PreviewTarget{}, err
	}

	again, err := reader.remoteTarget(ctx, git.transport, query.Branch)
	if err != nil || again != commit {
		return workflowstate.PreviewTarget{}, errorsOrConflict(err)
	}

	return workflowstate.PreviewTarget{Repository: query.Repository, Commit: commit}, nil
}

func (git OSWorkflowIntegrationGit) checkWorkflowTargetQuery(query WorkflowTargetQuery) error {
	repository, err := workflowCanonicalRepository(git.repository)
	if err != nil || repository != query.Repository || git.transport == "" || strings.HasPrefix(git.transport, "-") || strings.ContainsAny(git.transport, "\x00\r\n") || !validWorkflowPRBranch(query.Branch) || query.Directory == "" {
		return fmt.Errorf("canonical continuation target identity mismatch")
	}

	return nil
}

func (reader workflowIntegrationReader) remoteTarget(ctx context.Context, transport, branch string) (string, error) {
	output, err := reader.output(ctx, "ls-remote", "--exit-code", "--", transport, "refs/heads/"+branch)
	if err != nil {
		return "", err
	}

	fields := strings.Fields(output)
	if len(fields) != 2 || !workflowCommitID(fields[0]) || fields[1] != "refs/heads/"+branch {
		return "", workflowstate.ErrConflict
	}

	return fields[0], nil
}
