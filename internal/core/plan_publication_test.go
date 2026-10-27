package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Two publishers bind the same observed absence; the process-safe owner admits
// one result and rejects the other without replacing the first plan.
func TestPlanPublicationConcurrent(t *testing.T) {
	project := t.TempDir()
	plan := planPublicationCandidate()

	revision, err := PlanTargetRevision(t.Context(), project, plan.Scope)
	if err != nil {
		t.Fatal(err)
	}

	plan.ExpectedRevision, plan.CreateOnly = revision, true
	results := make(chan error, 2)

	var group sync.WaitGroup

	for range 2 {
		group.Go(func() {
			_, err := WriteGeneratedPlan(t.Context(), project, plan)
			results <- err
		})
	}

	group.Wait()
	close(results)

	writes, conflicts := 0, 0

	for err := range results {
		if err == nil {
			writes++
		} else if errors.Is(err, workflowstate.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}

	if writes != 1 || conflicts != 1 {
		t.Fatal("concurrent publishers did not compare the observed target", writes, conflicts)
	}
}

// Archive and replacement are one recoverable publication. A retained journal
// blocks observation after interruption; recovery completes the same images.
func TestPlanPublicationRecovery(t *testing.T) {
	project := t.TempDir()
	plan := planPublicationCandidate()

	paths, err := WriteGeneratedPlan(t.Context(), project, plan)
	if err != nil {
		t.Fatal(err)
	}

	plan.ExpectedRevision, err = PlanTargetRevision(t.Context(), project, plan.Scope)
	if err != nil {
		t.Fatal(err)
	}

	plan.Title, plan.Body = "New plan", "New bounded work."
	plan.ArchiveAt = plan.CreatedAt.Add(time.Minute)

	r := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(int) error { return errors.New("publication interrupted") }}
	if err := WithAuthoringMutation(t.Context(), project, func(ctx context.Context) error { return r.publishGeneratedPlan(ctx, paths, plan) }); err == nil {
		t.Fatal("publication interruption was ignored")
	}

	if _, err := PlanTargetRevision(t.Context(), project, plan.Scope); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("pending publication was readable", err)
	}

	if err := WithAuthoringMutation(t.Context(), project, func(context.Context) error { return nil }); err != nil {
		t.Fatal("recover publication", err)
	}

	checkRecoveredPlanPublication(t, project, paths, plan)

	if _, err := WriteGeneratedPlan(t.Context(), project, plan); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("old target binding was accepted after recovery", err)
	}
}

func checkRecoveredPlanPublication(t *testing.T, project string, paths GeneratedPlanPaths, plan GeneratedPlan) {
	t.Helper()

	content, err := os.ReadFile(paths.Path)
	if err != nil || !strings.Contains(string(content), "New bounded work.") {
		t.Fatal("replacement did not recover", err)
	}

	archive, err := ArchivedPlanLayout(project, plan.Scope, plan.ArchiveAt)
	if err != nil {
		t.Fatal(err)
	}

	prior, err := os.ReadFile(archive.Path)
	if err != nil || !strings.Contains(string(prior), "Original bounded work.") {
		t.Fatal("prior plan was lost", err)
	}

	if _, err := os.Stat(filepath.Join(project, authoringRuntimeRef("pending.json"))); !os.IsNotExist(err) {
		t.Fatal("completed publication retained a journal", err)
	}
}

func planPublicationCandidate() GeneratedPlan {
	return GeneratedPlan{Scope: quizcore.Scope{ProjectID: "example", CaseSlug: "chosen"}, Title: "Original plan", Body: "Original bounded work.", SourceObservationHash: "observed", SourceInputs: []PlanSourceBinding{{Kind: "dump", Reference: "ops/default/dump/00000-context/artifact.md", Revision: "observed-source"}}, CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
}
