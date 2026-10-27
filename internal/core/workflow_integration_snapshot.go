package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) readWorkflowIntegrationSnapshot(ctx context.Context) (workflowIntegrationSnapshot, error) {
	snapshot := workflowIntegrationSnapshot{}

	runs, err := s.repositoryWorkflowInvocationsPrepared(ctx, (*WorkflowStore).requireNoWorkflowTransition)
	if err != nil {
		return snapshot, err
	}

	snapshot.Runs = runs

	snapshot.Journals, err = s.workflowIntegrationJournals(ctx)
	if err != nil {
		return snapshot, integrationRecordError("publication journals", err)
	}

	coordinator, err := s.checkWorkflowCoordinator(ctx)
	if err != nil {
		return snapshot, err
	}

	entries, err := s.repositoryWorkflowIntegrationEntries(runs, coordinator)
	if err != nil {
		return snapshot, err
	}

	for _, run := range runs {
		if err := s.appendWorkflowIntegrationViews(ctx, &snapshot, run, entries); err != nil {
			return snapshot, integrationRecordError(workflowInvocationPath(run.ID), err)
		}
	}

	if err := s.checkWorkflowIntegrationJournals(ctx, snapshot); err != nil {
		return snapshot, err
	}

	slices.SortFunc(snapshot.Views, compareWorkflowIntegrationViews)

	snapshot.Hash, err = workflowIntegrationSnapshotHash(snapshot)

	return snapshot, err
}

func compareWorkflowIntegrationViews(a, b WorkflowIntegrationView) int {
	if a.Entry.QueueSequence < b.Entry.QueueSequence {
		return -1
	}

	if a.Entry.QueueSequence > b.Entry.QueueSequence {
		return 1
	}

	return strings.Compare(a.Entry.ID, b.Entry.ID)
}

func workflowIntegrationSnapshotHash(snapshot workflowIntegrationSnapshot) (string, error) {
	images := []string{}
	for _, run := range snapshot.Runs {
		images = append(images, run.OwnerRoot+"\x00"+workflowInvocationPath(run.ID)+"\x00"+run.sourceHash)
	}

	for _, journal := range snapshot.Journals {
		images = append(images, journal.Key+"\x00"+journal.sourceHash)
	}

	slices.Sort(images)

	data, err := json.Marshal(struct {
		Images []string
		Views  []WorkflowIntegrationView
	}{images, snapshot.Views})

	return contenthash.Sum(data), err
}

func (s *WorkflowStore) appendWorkflowIntegrationViews(ctx context.Context, snapshot *workflowIntegrationSnapshot, run workflowInvocationFile, entries []WorkflowIntegrationEntry) error {
	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return err
	}

	if err := owner.checkWorkflowIntegrationFiles(run); err != nil {
		return err
	}

	if len(run.IntegrationEntries) == 0 && run.IntegrationBudget == nil {
		return nil
	}

	budget, err := owner.readWorkflowIntegrationBudget(run)
	if err != nil {
		return err
	}

	if err := owner.checkWorkflowIntegrationBudgetRecords(ctx, run, budget, entries); err != nil {
		return err
	}

	usage, err := owner.workflowIntegrationUsage(ctx, run)
	if err != nil {
		return err
	}

	return owner.appendWorkflowIntegrationRunEntries(ctx, snapshot, run, entries, budget, usage)
}

func (s *WorkflowStore) appendWorkflowIntegrationRunEntries(ctx context.Context, snapshot *workflowIntegrationSnapshot, run workflowInvocationFile, entries []WorkflowIntegrationEntry, budget WorkflowIntegrationBudget, usage WorkflowIntegrationUsage) error {
	for _, entry := range entries {
		if entry.InvocationID == run.ID && entry.Namespace == run.Namespace {
			if err := s.checkWorkflowIntegrationEntryRecords(ctx, run, entry); err != nil {
				return integrationRecordError(workflowIntegrationEntryPath(run.ID, entry.ID), err)
			}

			snapshot.Views = append(snapshot.Views, WorkflowIntegrationView{Entry: entry, Budget: budget, CurrentUsage: usage})
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowIntegrationFiles(run workflowInvocationFile) error {
	relative := filepath.ToSlash(filepath.Join("run", run.ID, "extension", "chr", "integration"))

	path, err := s.safePath(relative)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		if len(run.IntegrationEntries) == 0 && run.IntegrationBudget == nil {
			return nil
		}

		return workflowstate.ErrConflict
	}

	if err != nil || len(files) > workflowstate.WorkflowRecordLimit+1 {
		return errorsOrConflict(err)
	}

	return checkWorkflowIntegrationDirectoryFiles(run, files)
}

func checkWorkflowIntegrationDirectoryFiles(run workflowInvocationFile, files []os.DirEntry) error {
	for _, file := range files {
		if file.Name() == "budget.json" && run.IntegrationBudget != nil && file.Type().IsRegular() {
			continue
		}

		known := slices.ContainsFunc(run.IntegrationEntries, func(b WorkflowIntegrationBinding) bool { return b.EntryID == file.Name() })
		if !known || !file.IsDir() || file.Type()&os.ModeSymlink != 0 {
			return workflowIntegrationConflict("unbound integration image requires explicit guarded migration")
		}
	}

	return nil
}

func (s *WorkflowStore) workflowIntegrationJournals(ctx context.Context) ([]workflowSubmissionJournal, error) {
	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), contenthash.Sum([]byte("integration-query")))
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil || len(files) > workflowstate.WorkflowRecordLimit {
		return nil, errorsOrConflict(err)
	}

	result := []workflowSubmissionJournal{}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".workflow-") {
			continue
		}

		if !file.Type().IsRegular() {
			return nil, workflowstate.ErrConflict
		}

		journal, err := readWorkflowSubmissionFile(filepath.Join(filepath.Dir(path), file.Name()))
		if err != nil {
			return nil, integrationRecordError(file.Name(), err)
		}

		result = append(result, journal)
	}

	return result, nil
}
