package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkflow observes a retained invocation; its private token hashes stay private.
func (s *WorkflowStore) ReadWorkflow(ctx context.Context, id string) (result WorkflowInvocation, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		run, _, readErr := s.locateWorkflowInvocation(ctx, id)
		result = run.WorkflowInvocation

		return readErr
	})

	return result, err
}

func (s *WorkflowStore) readWorkflowInvocation(id string) (workflowInvocationFile, error) {
	if !filepath.IsLocal(id) || filepath.Base(id) != id || id == "." {
		return workflowInvocationFile{}, fmt.Errorf("invalid workflow invocation identity")
	}

	path, err := s.safePath(workflowInvocationPath(id))
	if err != nil {
		return workflowInvocationFile{}, err
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return workflowInvocationFile{}, err
	}

	run := workflowInvocationFile{}
	if err := json.Unmarshal(data, &run); err != nil {
		return run, err
	}

	if err := json.Unmarshal(data, &run.fields); err != nil {
		return run, err
	}

	if !s.workflowInvocationScopeMatches(run, id) {
		return run, fmt.Errorf("foreign workflow invocation scope")
	}

	run.sourceHash = contenthash.Sum(data)

	return run, validateWorkflowInvocation(run)
}

func (s *WorkflowStore) locateWorkflowInvocation(ctx context.Context, id string) (workflowInvocationFile, *WorkflowStore, error) {
	return s.locateWorkflowInvocationPrepared(ctx, id, func(owner *WorkflowStore) error { return owner.prepareWorkflowRead(ctx) })
}

func (s *WorkflowStore) locateWorkflowInvocationPrepared(ctx context.Context, id string, prepare func(*WorkflowStore) error) (workflowInvocationFile, *WorkflowStore, error) {
	runs, err := s.repositoryWorkflowInvocationsPrepared(ctx, prepare)
	if err != nil {
		return workflowInvocationFile{}, nil, err
	}

	for _, run := range runs {
		if run.ID == id && run.Namespace == s.layout.Namespace() {
			owner, err := s.workflowOwnerStore(run)

			return run, owner, err
		}
	}

	return workflowInvocationFile{}, nil, fmt.Errorf("workflow invocation not found: %s", id)
}

func (s *WorkflowStore) workflowOwnerStore(run workflowInvocationFile) (*WorkflowStore, error) {
	projectDir := filepath.Dir(filepath.Dir(run.OwnerRoot))

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: run.Namespace})
	if err != nil {
		return nil, err
	}

	if layout.Root() != run.OwnerRoot {
		return nil, fmt.Errorf("workflow owner root is inconsistent")
	}

	owner := NewWorkflowStore(layout, run.ProjectID)
	owner.afterReplace = s.afterReplace

	return owner, nil
}

func (s *WorkflowStore) repositoryWorkflowInvocations(ctx context.Context) ([]workflowInvocationFile, error) {
	return s.repositoryWorkflowInvocationsPrepared(ctx, func(owner *WorkflowStore) error { return owner.prepareWorkflowRead(ctx) })
}

func (s *WorkflowStore) repositoryWorkflowInvocationsPrepared(ctx context.Context, prepare func(*WorkflowStore) error) ([]workflowInvocationFile, error) {
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return nil, err
	}

	runs := []workflowInvocationFile{}

	owners := map[string]bool{}
	for _, layout := range layouts {
		owners[layout.Root()] = true
	}

	for _, layout := range layouts {
		store := NewWorkflowStore(layout, s.projectID)
		if err := prepare(store); err != nil {
			return nil, err
		}

		if err := store.appendWorkflowInvocations(ctx, &runs, owners); err != nil {
			return nil, err
		}
	}

	return runs, nil
}

func (s *WorkflowStore) appendWorkflowInvocations(ctx context.Context, runs *[]workflowInvocationFile, owners map[string]bool) error {
	entries, err := os.ReadDir(filepath.Join(s.layout.Root(), "run"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if len(entries)+len(*runs) > WorkflowRecordLimit {
		return fmt.Errorf("repository invocation observation bound exceeded")
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		if !entry.IsDir() {
			continue
		}

		if err := s.appendObservedWorkflowInvocation(runs, entry.Name(), owners); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) appendObservedWorkflowInvocation(runs *[]workflowInvocationFile, id string, owners map[string]bool) error {
	run, err := s.readWorkflowInvocation(id)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	outside := !owners[run.OwnerRoot]
	if outside && !workflowInvocationTerminal(run.Status) {
		return fmt.Errorf("%w: invocation owner is outside this repository coordinator", workflowstate.ErrConflict)
	}
	// A Git copy cannot claim work. Terminal copies retain retry identity as history.
	if run.OwnerRoot == s.layout.Root() || outside {
		*runs = append(*runs, run)
	}

	return nil
}

func (s *WorkflowStore) workflowRepositoryLayouts(ctx context.Context) ([]OpsLayout, error) {
	worktrees, err := (OSWorktreeGit{}).WorktreeList(ctx, s.layout.ProjectDir())
	if err != nil {
		return nil, err
	}

	if len(worktrees) == 0 || len(worktrees) > WorkflowPageLimit {
		return nil, fmt.Errorf("workflow worktree observation bound exceeded")
	}

	result := []OpsLayout{}

	for _, worktree := range worktrees {
		if worktree.Bare || worktree.Prunable {
			continue
		}

		layouts, err := s.workflowWorktreeLayouts(worktree.Path)
		if err != nil {
			return nil, err
		}

		result = append(result, layouts...)
	}

	if len(result) > WorkflowRecordLimit {
		return nil, fmt.Errorf("workflow namespace observation bound exceeded")
	}

	return result, nil
}

func (s *WorkflowStore) workflowWorktreeLayouts(projectDir string) ([]OpsLayout, error) {
	entries, err := os.ReadDir(filepath.Join(projectDir, OpsDirectoryName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if len(entries) > WorkflowPageLimit {
		return nil, fmt.Errorf("workflow namespace observation bound exceeded")
	}

	names := []string{s.layout.Namespace(), opsnamespace.Default}

	for _, entry := range entries {
		if entry.IsDir() && opsnamespace.Valid(entry.Name()) {
			names = append(names, entry.Name())
		}
	}

	slices.Sort(names)

	names = slices.Compact(names)
	result := []OpsLayout{}

	for _, name := range names {
		layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: name})
		if err != nil {
			return nil, err
		}

		result = append(result, layout)
	}

	return result, nil
}

func (s *WorkflowStore) workflowInvocationScopeMatches(run workflowInvocationFile, id string) bool {
	return run.ID == id && run.Namespace == s.layout.Namespace()
}
