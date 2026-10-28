package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowIntegrationLegacySubmissions(ctx context.Context, run workflowInvocationFile) ([]IntegrationCounter, error) {
	key := contenthash.Sum([]byte("budget-journal-scan"))

	path, err := workflowSubmissionPath(ctx, s.layout.ProjectDir(), key)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if len(entries) > workflowstate.RecordLimit {
		return nil, workflowstate.ErrConflict
	}

	return readWorkflowIntegrationLegacyCounters(filepath.Dir(path), entries, run)
}

func readWorkflowIntegrationLegacyCounters(dir string, entries []os.DirEntry, run workflowInvocationFile) ([]IntegrationCounter, error) {
	counters := []IntegrationCounter{}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".workflow-") {
			continue
		}

		if !entry.Type().IsRegular() {
			return nil, workflowstate.ErrConflict
		}

		journal, err := readWorkflowSubmissionFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		if journal.InvocationID != run.ID || journal.Namespace != run.Namespace || journal.OwnerRoot != run.OwnerRoot {
			continue
		}

		if err := appendWorkflowIntegrationLegacyCounter(&counters, journal); err != nil {
			return nil, err
		}
	}

	return counters, nil
}

func appendWorkflowIntegrationLegacyCounter(counters *[]IntegrationCounter, journal workflowSubmissionJournal) error {
	consumed := int64(0)

	for _, attempt := range journal.Attempts {
		if attempt.Native == nil {
			consumed++
		}
	}

	if consumed == 0 {
		return nil
	}

	if slices.ContainsFunc(*counters, func(c IntegrationCounter) bool { return c.Unit == journal.Unit }) {
		return workflowstate.ErrConflict
	}

	*counters = append(*counters, IntegrationCounter{Unit: journal.Unit, Submission: consumed, ExistingSubmissions: consumed})

	return nil
}

func checkWorkflowIntegrationLegacySubmissions(budget IntegrationBudget, counters []IntegrationCounter) error {
	for _, current := range counters {
		index := slices.IndexFunc(budget.UnitCounters, func(c IntegrationCounter) bool { return c.Unit == current.Unit })
		if index < 0 || budget.UnitCounters[index].ExistingSubmissions != current.ExistingSubmissions {
			return workflowstate.IntegrationConflict("legacy submission usage differs from admission snapshot")
		}
	}

	for _, retained := range budget.UnitCounters {
		if retained.ExistingSubmissions == 0 {
			continue
		}

		if !slices.ContainsFunc(counters, func(c IntegrationCounter) bool { return c.Unit == retained.Unit }) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
