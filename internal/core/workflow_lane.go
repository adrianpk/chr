package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTaskUnit(record WorkflowRecord) workflowstate.Ref {
	if record.Task.SliceID != "" {
		return workflowstate.Ref{Kind: "slice", ID: record.Task.SliceID}
	}

	return workflowstate.Ref{Kind: "ticket", ID: record.Task.Key}
}

func (s *WorkflowStore) retainedWorkflowLane(ctx context.Context, run workflowInvocationFile, record WorkflowRecord) (*WorkflowLane, error) {
	unit := workflowTaskUnit(record)
	if run.Lane != nil && run.Lane.Unit == unit {
		lane := *run.Lane

		return &lane, nil
	}

	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return nil, err
	}

	var latest *workflowInvocationFile

	for i := range runs {
		previous := &runs[i]
		if !workflowRunHasLane(*previous, run.Namespace, unit) {
			continue
		}

		if previous.Selection.Target.Repository != run.ExecutionSelection().Target.Repository {
			return nil, workflowstate.ErrConflict
		}

		if latest == nil || previous.UpdatedAt.After(latest.UpdatedAt) {
			latest = previous
		}
	}

	if latest == nil {
		return nil, nil
	}

	lane := *latest.Lane

	return &lane, nil
}

func workflowRunHasLane(run workflowInvocationFile, namespace string, unit workflowstate.Ref) bool {
	return run.Namespace == namespace && run.Lane != nil && run.Lane.Unit == unit
}

func (s *WorkflowStore) workflowLaneSelection(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, request workflowstate.PreviewRequest) (workflowstate.PreviewRequest, error) {
	lane, err := s.retainedWorkflowLane(ctx, run, record)
	if err != nil || lane == nil {
		return request, err
	}

	if _, err := s.observeWorkflowLane(ctx, *lane); err != nil {
		return request, err
	}

	request.CumulativeBranch, request.CumulativeCommit = lane.Branch, lane.HeadCommit

	return request, nil
}

func (s *WorkflowStore) newWorkflowLane(run workflowInvocationFile, record WorkflowRecord) WorkflowLane {
	unit := workflowTaskUnit(record)
	id := contenthash.Sum([]byte(run.Namespace + "\x00" + unit.Kind + "\x00" + unit.ID))[:24]

	return WorkflowLane{Unit: unit, Branch: "refs/heads/chr/workflow/" + id, Worktree: filepath.Join(s.layout.WorktreeRoot(), "workflow-"+id), BaseCommit: run.ExecutionSelection().Target.Commit, HeadCommit: run.ExecutionSelection().Target.Commit}
}

func (s *WorkflowStore) ensureWorkflowLane(ctx context.Context, lane WorkflowLane) error {
	git := executioncore.OSWorktreeGit{}
	branch := strings.TrimPrefix(lane.Branch, "refs/heads/")

	exists, err := git.BranchExists(ctx, s.layout.ProjectDir(), branch)
	if err != nil {
		return err
	}

	if !exists {
		if lane.HeadCommit != lane.BaseCommit {
			return fmt.Errorf("%w: retained lane branch is missing", workflowstate.ErrConflict)
		}

		if err := git.CreateBranch(ctx, s.layout.ProjectDir(), branch, lane.BaseCommit); err != nil {
			return err
		}
	}

	entries, err := git.WorktreeList(ctx, s.layout.ProjectDir())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Path == lane.Worktree {
			_, err := s.observeWorkflowLane(ctx, lane)

			return err
		}
	}

	if err := git.AddWorktree(ctx, s.layout.ProjectDir(), lane.Worktree, branch); err != nil {
		return err
	}

	_, err = s.observeWorkflowLane(ctx, lane)

	return err
}

func (s *WorkflowStore) observeWorkflowLane(ctx context.Context, lane WorkflowLane) (string, error) {
	commit, tree, err := s.workflowLaneCandidate(ctx, lane)
	if err == nil && commit != lane.HeadCommit {
		err = fmt.Errorf("%w: retained lane moved without a checkpoint", workflowstate.ErrConflict)
	}

	return tree, err
}

func (s *WorkflowStore) workflowLaneCandidate(ctx context.Context, lane WorkflowLane) (string, string, error) {
	if err := s.checkWorkflowLaneRegistration(ctx, lane); err != nil {
		return "", "", err
	}

	status, err := (executioncore.OSWorktreeGit{}).WorktreeStatus(ctx, lane.Worktree)
	if err != nil || len(status) != 0 {
		return "", "", fmt.Errorf("workflow lane must be clean: %w", errorsOrConflict(err))
	}

	reader := workflowProofReader{store: s}

	commit, err := (executioncore.OSWorktreeGit{}).ResolveRef(ctx, lane.Worktree, "HEAD^{commit}")
	if err != nil || !reader.isAncestor(ctx, lane.HeadCommit, commit) {
		return "", "", fmt.Errorf("%w: lane history does not descend from its accepted head", workflowstate.ErrConflict)
	}

	if err := s.checkWorkflowCoordinatorTarget(ctx, workflowstate.PreviewTarget{Commit: commit}); err != nil {
		return "", "", err
	}

	tree, err := (executioncore.OSWorktreeGit{}).ResolveRef(ctx, lane.Worktree, "HEAD^{tree}")

	return commit, tree, err
}

func errorsOrConflict(err error) error {
	if err != nil {
		return err
	}

	return workflowstate.ErrConflict
}

func (s *WorkflowStore) checkWorkflowLaneRegistration(ctx context.Context, lane WorkflowLane) error {
	if !validWorkflowLane(lane) {
		return workflowstate.ErrConflict
	}

	common, err := projectcore.ResolveRepositoryDirectory(ctx, s.layout.ProjectDir())
	if err != nil {
		return err
	}

	actual, err := projectcore.ResolveRepositoryDirectory(ctx, lane.Worktree)
	if err != nil || actual != common {
		return fmt.Errorf("%w: lane belongs to a different coordinator", workflowstate.ErrConflict)
	}

	entries, err := (executioncore.OSWorktreeGit{}).WorktreeList(ctx, s.layout.ProjectDir())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Path == lane.Worktree && "refs/heads/"+entry.Branch == lane.Branch && !entry.Prunable {
			return nil
		}
	}

	return fmt.Errorf("%w: lane branch or worktree registration changed", workflowstate.ErrConflict)
}

func validWorkflowLane(lane WorkflowLane) bool {
	return (lane.Unit.Kind == "slice" || lane.Unit.Kind == "ticket") && lane.Unit.ID != "" && strings.HasPrefix(lane.Branch, "refs/heads/chr/workflow/") && filepath.IsAbs(lane.Worktree) && workflowCommitID(lane.BaseCommit) && workflowCommitID(lane.HeadCommit)
}
