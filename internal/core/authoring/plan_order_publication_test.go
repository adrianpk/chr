package authoring

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// New successful generations retain separate revision pairs even when every
// rendered plan byte is identical. Generation never allocates native tickets.
func TestPlanOrderHistory(t *testing.T) {
	project := t.TempDir()
	plan := orderedPlanCandidate()

	paths, err := WriteGeneratedPlan(t.Context(), project, plan)
	if err != nil {
		t.Fatal(err)
	}

	first := readTestPlanBinding(t, project, plan.Scope)
	firstBody := readTestPlanImage(t, project, first.BodyRef)

	firstTracker := readTestPlanImage(t, project, first.TrackerRef)
	checkDraftPlanMap(t, firstTracker)

	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal(err)
	}

	second := readTestPlanBinding(t, project, plan.Scope)
	if first.Revision != 1 || second.Revision != 2 || first.BodyHash != second.BodyHash || first.BodyRef == second.BodyRef {
		t.Fatal(first, second)
	}

	if !bytes.Equal(firstBody, readTestPlanImage(t, project, first.BodyRef)) || !bytes.Equal(firstTracker, readTestPlanImage(t, project, first.TrackerRef)) {
		t.Fatal("regeneration changed retained history")
	}

	verifyPlanManifest(t, paths, first, second)

	if _, err := os.Stat(filepath.Join(project, "ops", "default", "ticket")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("generation allocated tickets", err)
	}
}

func orderedPlanCandidate() generationcore.GeneratedPlan {
	plan := planPublicationCandidate()
	plan.RequireOrderMap = true
	plan.Order = []generationcore.PlanOrderSlice{{ID: "slice-1", Order: 1, Title: "Foundation", Tasks: []generationcore.PlanOrderTask{{ID: "task-t1.1", Order: 1, Title: "Implement foundation"}}}}

	return plan
}

func readTestPlanBinding(t *testing.T, project string, scope quizcore.Scope) workflowstate.PlanRevisionBinding {
	t.Helper()

	binding, err := ObservePlanRevision(t.Context(), project, scope)
	if err != nil {
		t.Fatal(err)
	}

	return binding
}

func readTestPlanImage(t *testing.T, project, reference string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(project, "ops", "default", filepath.FromSlash(reference)))
	if err != nil {
		t.Fatal(err)
	}

	return content
}

func verifyPlanManifest(t *testing.T, paths GeneratedPlanPaths, bindings ...workflowstate.PlanRevisionBinding) {
	t.Helper()

	encoded, err := os.ReadFile(filepath.Join(filepath.Dir(paths.Path), "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	var metadata struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(encoded, &metadata); err != nil {
		t.Fatal(err)
	}

	if len(metadata.Files) != 1+7*len(bindings) {
		t.Fatal("retained files missing from manifest", metadata.Files)
	}

	for _, binding := range bindings {
		for _, name := range []string{"plan.md", "order.json", "tracker.md", "binding.json", "planned-order.json", "planned-tracker.md", "planned-binding.json"} {
			expected := filepath.ToSlash(filepath.Join("files", "plan-revisions", filepath.Base(filepath.Dir(binding.BodyRef)), name))
			if !strings.Contains(strings.Join(metadata.Files, "\n"), expected) {
				t.Fatal("manifest missing", expected)
			}
		}
	}
}

// Every publication image can be interrupted. Pending reads fail; recovering
// the existing owner completes the exact plan, map, manifest and receipt pair.
func TestPlanOrderRecovery(t *testing.T) {
	for _, after := range []int{0, 3, 7, 9} {
		t.Run("after image "+strconv.Itoa(after), func(t *testing.T) {
			checkPlanOrderRecovery(t, after)
		})
	}
}

// Regeneration recovery retains the original pair and archive while completing
// the new revision after an interruption during canonical replacement.
func TestPlanOrderArchiveRecovery(t *testing.T) {
	project := t.TempDir()
	plan := orderedPlanCandidate()

	paths, err := WriteGeneratedPlan(t.Context(), project, plan)
	if err != nil {
		t.Fatal(err)
	}

	first := readTestPlanBinding(t, project, plan.Scope)
	plan.Body = "Replacement work with its own retained map."
	plan.ArchiveAt = plan.CreatedAt

	r := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == 10 {
			return errors.New("interrupted replacement")
		}

		return nil
	}}
	if err := WithAuthoringMutation(t.Context(), project, func(ctx context.Context) error { return r.publishGeneratedPlan(ctx, paths, plan) }); err == nil {
		t.Fatal("replacement did not interrupt")
	}

	if err := WithAuthoringMutation(t.Context(), project, func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}

	second := readTestPlanBinding(t, project, plan.Scope)
	if second.Revision != 2 || first.BodyHash == second.BodyHash {
		t.Fatal(first, second)
	}

	checkPlanOrderArchive(t, project, plan)

	if contenthash.Sum(readTestPlanImage(t, project, first.TrackerRef)) != first.TrackerHash {
		t.Fatal("regeneration altered historical pair")
	}
}

func checkDraftPlanMap(t *testing.T, tracker []byte) {
	t.Helper()

	if !bytes.Contains(tracker, []byte("Tickets are unbound")) || bytes.Contains(tracker, []byte("in_progress")) {
		t.Fatal("draft map acquired progress or execution authority")
	}
}

func checkPlanOrderRecovery(t *testing.T, after int) {
	t.Helper()

	project := t.TempDir()
	plan := orderedPlanCandidate()
	plan.CreateOnly = true
	plan.Operation = &generationcore.PlanPublicationOperation{ID: "recover-plan", Scope: plan.Scope}

	paths, err := GeneratedPlanLayout(project, plan.Scope)
	if err != nil {
		t.Fatal(err)
	}

	r := &fileAuthoringSessionRepository{projectDir: project, afterMutationReplace: func(index int) error {
		if index == after {
			return errors.New("interrupted")
		}

		return nil
	}}
	if err := WithAuthoringMutation(t.Context(), project, func(ctx context.Context) error { return r.publishGeneratedPlan(ctx, paths, plan) }); err == nil {
		t.Fatal("publication did not interrupt")
	}

	if _, err := ObservePlanRevision(t.Context(), project, plan.Scope); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("pending pair was readable", err)
	}

	result, err := LookupPlanPublication(t.Context(), project, *plan.Operation)
	if err != nil || !result.Found || result.Result.Binding.Revision != 1 {
		t.Fatal("receipt did not recover", result.Result, result.Found, err)
	}

	readTestPlanBinding(t, project, plan.Scope)

	if _, err := WriteGeneratedPlan(t.Context(), project, plan); err != nil {
		t.Fatal("committed retry failed", err)
	}
}

func checkPlanOrderArchive(t *testing.T, project string, plan generationcore.GeneratedPlan) {
	t.Helper()

	archive, err := ArchivedPlanLayout(project, plan.Scope, plan.ArchiveAt)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(archive.Path)
	if err != nil || !strings.Contains(string(content), "Original bounded work.") {
		t.Fatal("archive was lost", err)
	}

	metadataBytes, err := os.ReadFile(filepath.Join(filepath.Dir(archive.Path), "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	var metadata struct {
		Files       []string                          `json:"files"`
		SourceFiles []string                          `json:"source_files"`
		Binding     workflowstate.PlanRevisionBinding `json:"plan_binding"`
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		t.Fatal(err)
	}

	if len(metadata.Files) != 1 || len(metadata.SourceFiles) != 8 || metadata.Binding.Revision != 1 {
		t.Fatal("archive lost its source binding or acquired phantom local files", metadata)
	}
}
