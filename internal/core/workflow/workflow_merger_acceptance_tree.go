package workflow

import (
	"context"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMergerSnapshot(ctx context.Context, reader workflowIntegrationReader, state MergerRepairState) (string, error) {
	if err := checkWorkflowMergerPrivateRepository(ctx, reader.Directory); err != nil {
		return "", err
	}

	head, err := reader.Output(ctx, "rev-parse", "HEAD")
	if err != nil || head != state.PreparedParent {
		return "", workflowstate.ErrConflict
	}

	if err := checkWorkflowMergerSnapshotFiles(reader.Directory); err != nil {
		return "", err
	}
	// Reset the index independently: an untrusted index cannot hide worktree edits.
	if _, err := reader.Output(ctx, "read-tree", "--reset", state.Tree); err != nil {
		return "", err
	}

	if _, err := reader.Output(ctx, "add", "--all", "--force", "--", "."); err != nil {
		return "", err
	}

	return reader.Output(ctx, "write-tree")
}

func checkWorkflowMergerSnapshotFiles(directory string) error {
	count := 0

	return filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == filepath.Join(directory, ".git") {
			return filepath.SkipDir
		}

		count++
		if count > workflowstate.RecordLimit {
			return workflowstate.ErrConflict
		}

		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if !workflowmerge.RegularFile(info) {
			return workflowstate.ErrConflict
		}

		return nil
	})
}

func (reader workflowIntegrationReader) mergerDiffPaths(ctx context.Context, before, after string) ([]string, error) {
	data, err := reader.RawOutput(ctx, "diff", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--no-textconv", before, after, "--")
	if err != nil {
		return nil, err
	}

	return workflowMergerChangedPaths(data)
}

func (reader workflowIntegrationReader) checkMergerObservedPaths(ctx context.Context, input workflowstate.MergerInput, proposal workflowstate.MergerResult, before, after string, exact bool) error {
	paths, err := reader.mergerDiffPaths(ctx, before, after)
	if err != nil {
		return err
	}

	declared := slices.Clone(proposal.ChangedPaths)
	slices.Sort(declared)

	if exact && !slices.Equal(paths, declared) {
		return workflowstate.ErrConflict
	}

	for _, path := range paths {
		if !slices.Contains(declared, path) || !workflowmerge.PathAllowed(input, path) {
			return workflowstate.ErrConflict
		}
	}

	return reader.checkMergerChangedKinds(ctx, input, paths, before, after)
}

func (reader workflowIntegrationReader) checkMergerChangedKinds(ctx context.Context, input workflowstate.MergerInput, paths []string, objects ...string) error {
	for _, path := range paths {
		for _, object := range objects {
			kind, err := reader.mergerFileKind(ctx, object, path)
			if err != nil {
				return err
			}

			declared := slices.ContainsFunc(input.Conflicts, func(c workflowstate.MergerConflict) bool {
				return c.Path == path && c.Kind == kind && workflowmerge.ProcedureRequired(c)
			})
			if kind != "text" && !declared {
				return workflowstate.ErrConflict
			}
		}
	}

	return nil
}

func (reader workflowIntegrationReader) mergerPathObject(ctx context.Context, tree, path string) (string, error) {
	return reader.RawOutput(ctx, "ls-tree", "-z", "--full-tree", tree, "--", path)
}

func (reader workflowIntegrationReader) checkMergerSideSelection(ctx context.Context, state MergerRepairState, input workflowstate.MergerInput, tree string) error {
	for _, conflict := range input.Conflicts {
		if workflowmerge.ProcedureRequired(conflict) {
			continue
		}

		if err := reader.checkMergerConflictSides(ctx, state, tree, conflict.Path); err != nil {
			return err
		}
	}

	return nil
}

func (reader workflowIntegrationReader) checkMergerConflictSides(ctx context.Context, state MergerRepairState, tree, path string) error {
	resolved, err := reader.mergerPathObject(ctx, tree, path)
	if err != nil || resolved == "" {
		return workflowstate.ErrConflict
	}

	content, err := reader.RawOutput(ctx, "show", tree+":"+path)
	if err != nil {
		return err
	}

	if strings.Contains(content, "<<<<<<<") && strings.Contains(content, ">>>>>>>") {
		return workflowstate.ErrConflict
	}

	for _, side := range []string{state.PreparedParent, state.CurrentSource} {
		if err := reader.checkMergerCopiedSide(ctx, side, path, resolved, content); err != nil {
			return err
		}
	}

	return nil
}

func (reader workflowIntegrationReader) checkMergerCopiedSide(ctx context.Context, side, path, resolved, content string) error {
	object, err := reader.mergerPathObject(ctx, side, path)
	if err != nil {
		return err
	}

	if object == "" {
		return nil
	}

	if object == resolved {
		return workflowstate.ErrConflict
	}

	other, err := reader.RawOutput(ctx, "show", side+":"+path)
	if err != nil {
		return err
	}

	if strings.TrimSpace(content) == strings.TrimSpace(other) {
		return workflowstate.ErrConflict
	}

	return nil
}
