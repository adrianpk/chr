package core

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const workflowPublicationDirectory = "chr-workflow-publications"

type workflowSubmissionJournal struct {
	Schema       string               `json:"schema"`
	Key          string               `json:"key"`
	OwnerRoot    string               `json:"owner_root"`
	Namespace    string               `json:"namespace"`
	InvocationID string               `json:"invocation_id"`
	Unit         workflowstate.Ref    `json:"unit"`
	TargetKey    string               `json:"target_key"`
	Attempts     []WorkflowSubmission `json:"attempts"`
	sourceHash   string
}

func workflowSubmissionJournalKey(run workflowInvocationFile) string {
	return contenthash.Sum([]byte(run.OwnerRoot + "\n" + run.Namespace + "\n" + run.ID + "\n" + run.Lane.Unit.Kind + "\n" + run.Lane.Unit.ID))
}

func workflowSubmissionTargetKey(pr WorkflowPullRequestSnapshot) (string, error) {
	repository, err := workflowCanonicalRepository(pr.Repository)
	if err != nil || !validWorkflowPRBranch(pr.BaseBranch) {
		return "", workflowstate.ErrConflict
	}

	return contenthash.Sum([]byte(repository + "\nrefs/heads/" + pr.BaseBranch)), nil
}

func workflowSubmissionPath(ctx context.Context, projectDir, key string) (string, error) {
	if _, err := hex.DecodeString(key); err != nil || len(key) != 64 {
		return "", workflowstate.ErrConflict
	}

	common, err := workflowRepositoryDirectory(ctx, projectDir)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(common, workflowPublicationDirectory)

	path := filepath.Join(dir, key+".json")
	for _, checked := range []string{dir, path} {
		info, err := os.Lstat(checked)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("publication journal cannot traverse a symlink")
		}
	}

	return path, nil
}

func readWorkflowSubmissionFile(path string) (journal workflowSubmissionJournal, err error) {
	data, err := opsfile.Read(path)
	if err != nil {
		return journal, err
	}

	if err := json.Unmarshal(data, &journal); err != nil {
		return journal, err
	}

	journal.sourceHash = contenthash.Sum(data)
	if filepath.Base(path) != journal.Key+".json" {
		return journal, workflowstate.ErrConflict
	}

	return journal, validateWorkflowSubmissionJournal(journal)
}

func (s *WorkflowStore) readWorkflowSubmissionJournal(ctx context.Context, run workflowInvocationFile) (workflowSubmissionJournal, error) {
	if run.Lane == nil {
		return workflowSubmissionJournal{}, os.ErrNotExist
	}

	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), workflowSubmissionJournalKey(run))
	if err != nil {
		return workflowSubmissionJournal{}, err
	}

	journal, err := readWorkflowSubmissionFile(path)
	if err == nil && (journal.OwnerRoot != run.OwnerRoot || journal.Namespace != run.Namespace || journal.InvocationID != run.ID || journal.Unit != run.Lane.Unit) {
		err = workflowstate.ErrConflict
	}

	return journal, err
}

func (s *WorkflowStore) writeWorkflowSubmissionJournal(ctx context.Context, journal *workflowSubmissionJournal) error {
	if err := validateWorkflowSubmissionJournal(*journal); err != nil {
		return err
	}

	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), journal.Key)
	if err != nil {
		return err
	}

	if err := checkWorkflowSubmissionImage(path, journal.sourceHash); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(journal, "", "  ")
	if err != nil {
		return err
	}

	encoded = append(encoded, '\n')
	if err := opsfile.WriteAtomic(path, encoded); err != nil {
		return err
	}

	journal.sourceHash = contenthash.Sum(encoded)

	if s.afterReplace != nil {
		return s.afterReplace(-1)
	}

	return nil
}

func checkWorkflowSubmissionImage(path, expected string) error {
	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) && expected == "" {
		return nil
	}

	if err != nil {
		return err
	}

	if contenthash.Sum(data) != expected {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) guardWorkflowSubmissionTarget(ctx context.Context, key, target string) error {
	if err := s.checkWorkflowProjectedSubmissions(ctx, target); err != nil {
		return err
	}

	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), key)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if len(entries) >= workflowstate.RecordLimit {
		return fmt.Errorf("repository publication journal bound exceeded")
	}

	return checkWorkflowSubmissionTargets(filepath.Dir(path), entries, key, target)
}

func (s *WorkflowStore) checkWorkflowProjectedSubmissions(ctx context.Context, target string) error {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if run.Submission == nil || run.Submission.TargetKey != target {
			continue
		}

		if err := s.checkWorkflowSubmissionProjection(ctx, run); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowSubmissionProjection(ctx context.Context, run workflowInvocationFile) error {
	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), run.Submission.JournalKey)
	if err != nil {
		return err
	}

	journal, err := readWorkflowSubmissionFile(path)
	if err != nil {
		return fmt.Errorf("%w: projected submission journal is absent or invalid: %w", workflowstate.ErrConflict, err)
	}

	if journal.InvocationID != run.ID || journal.Unit != run.Submission.Unit || journal.OwnerRoot != run.OwnerRoot || journal.TargetKey != run.Submission.TargetKey {
		return workflowstate.ErrConflict
	}

	last := journal.Attempts[len(journal.Attempts)-1]
	if last.State != "awaiting_observation" && journal.sourceHash != run.Submission.JournalHash {
		return fmt.Errorf("%w: %w: resolved journal needs fresh owner observation", ErrWorkflowOwned, ErrWorkflowObservationRequired)
	}

	return nil
}

func checkWorkflowSubmissionTargets(dir string, entries []os.DirEntry, key, target string) error {
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".workflow-") {
			continue
		}

		if !entry.Type().IsRegular() {
			return workflowstate.ErrConflict
		}

		journal, err := readWorkflowSubmissionFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}

		last := journal.Attempts[len(journal.Attempts)-1]
		if journal.Key != key && journal.TargetKey == target && last.State == "awaiting_observation" {
			return fmt.Errorf("%w: target excluded by submission %s", ErrWorkflowOwned, last.ID)
		}
	}

	return nil
}
