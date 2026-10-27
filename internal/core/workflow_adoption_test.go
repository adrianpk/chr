package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowPlanAdoptionFixture(t *testing.T) (*WorkflowStore, WorkflowPlanAdoptionOptions) {
	t.Helper()

	store := workflowTestStore(t)
	plan := workflowPlanFixture()

	artifact, err := opscore.CreateArtifact(opscore.ArtifactOptions{Layout: store.layout, Type: "plan", Title: "Original plan", Producer: "chr"})
	if err != nil {
		t.Fatal(err)
	}

	plan.PlanRef = "plan/" + artifact.ID + "/artifact.md"

	body := []byte("Original approved plan\n")
	if err := os.WriteFile(filepath.Join(store.layout.Root(), plan.PlanRef), body, 0o644); err != nil {
		t.Fatal(err)
	}

	plan.PlanHash = contenthash.Sum(body)

	metadata, err := os.ReadFile(filepath.Join(artifact.Dir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	opts := WorkflowPlanAdoptionOptions{Plan: plan, PlanPath: "plan/" + artifact.ID + "/metadata.json", SourceHash: contenthash.Sum(metadata) + ":" + plan.PlanHash, CheckQuiescent: func(context.Context, []string) error { return nil }}
	for i := range opts.Plan.Slices {
		for j := range opts.Plan.Slices[i].Tasks {
			task := &opts.Plan.Slices[i].Tasks[j]
			selection := workflowLegacyTicketFixture(t, store, *task, i*3+j+1)

			data, err := os.ReadFile(filepath.Join(store.layout.Root(), selection.Path))
			if err != nil {
				t.Fatal(err)
			}

			parsedMarkdown, err := workflowstate.RequireMarkdown(data)
			original := parsedMarkdown.Body

			if err != nil {
				t.Fatal(err)
			}

			task.Body = string(original)

			opts.Tasks = append(opts.Tasks, selection)
		}
	}

	return store, opts
}

func workflowLegacyTicketFixture(t *testing.T, store *WorkflowStore, task generationcore.GeneratedWorkflowTask, number int) WorkflowAdoptionTask {
	t.Helper()

	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	text, err := NewTicketRenderer().RenderTicket(Ticket{Key: fmt.Sprintf("PRJ-%d", number), Title: task.Title, Synopsis: task.Outcome, Status: TicketStatusReadyForDev, Body: "Original instructions and history\n", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}

	text = strings.Replace(text, "---\n\n", "custom: |\n  Preserve original metadata\n---\n\n", 1)
	path := fmt.Sprintf("ticket/legacy-%d.md", number)

	absolute := filepath.Join(store.layout.Root(), path)
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(absolute, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	return WorkflowAdoptionTask{TaskID: task.ID, Path: path, SourceHash: contenthash.Sum([]byte(text))}
}

// TestWorkflowPlanAdoption preserves original identity, creation times, native
// fields, body bytes and locations while adopting one explicitly selected set.
func TestWorkflowPlanAdoption(t *testing.T) {
	store, opts := workflowPlanAdoptionFixture(t)

	before, err := store.Snapshot(t.Context())
	if err != nil || len(before.Records) != 0 {
		t.Fatal("ordinary reads adopted legacy work")
	}

	result, err := store.AdoptPlanWorkflow(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if result.Set.Path != opts.PlanPath || len(result.Tasks) != 6 {
		t.Fatal("plan identity or selected tasks replaced")
	}

	checkAdoptedWorkflowTasks(t, store, result, opts)

	again, err := store.AdoptPlanWorkflow(t.Context(), opts)
	if err != nil || again.Set.ID != result.Set.ID {
		t.Fatalf("idempotent adoption = %v", err)
	}

	checkMaterializationRetry(t, store, WorkflowMaterializationOptions{Plan: opts.Plan, ReserveTicketKey: func(context.Context) (string, error) { return "", fmt.Errorf("retry allocated a key") }}, result)
}

func checkAdoptedWorkflowTasks(t *testing.T, store *WorkflowStore, result WorkflowMaterializationResult, opts WorkflowPlanAdoptionOptions) {
	t.Helper()

	for i, record := range result.Tasks {
		if record.Path != opts.Tasks[i].Path || record.ID != record.Task.Key || record.CreatedAt.Year() != 2025 {
			t.Fatal("legacy identity, location or creation time changed")
		}

		if string(record.Body) != opts.Plan.Slices[i/3].Tasks[i%3].Body {
			t.Fatal("original body changed")
		}

		data, err := os.ReadFile(filepath.Join(store.layout.Root(), record.Path))
		if err != nil || !strings.Contains(string(data), "custom: |\n  Preserve original metadata") {
			t.Fatal("compatible native fields lost")
		}
	}
}

// TestWorkflowPlanAdoptionRejects blocks ambiguous identity, foreign state,
// unobserved ownership and stale source images before any record is adopted.
func TestWorkflowPlanAdoptionRejects(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, *WorkflowStore, *WorkflowPlanAdoptionOptions)
	}{
		{"duplicate selection", func(_ *testing.T, _ *WorkflowStore, o *WorkflowPlanAdoptionOptions) { o.Tasks[1] = o.Tasks[0] }},
		{"stale plan", func(_ *testing.T, _ *WorkflowStore, o *WorkflowPlanAdoptionOptions) { o.SourceHash = "changed" }},
		{"stale ticket", func(_ *testing.T, _ *WorkflowStore, o *WorkflowPlanAdoptionOptions) {
			o.Tasks[0].SourceHash = "changed"
		}},
		{"no owner observation", func(_ *testing.T, _ *WorkflowStore, o *WorkflowPlanAdoptionOptions) { o.CheckQuiescent = nil }},
		{"unfinished owner", func(_ *testing.T, _ *WorkflowStore, o *WorkflowPlanAdoptionOptions) {
			o.CheckQuiescent = func(context.Context, []string) error { return fmt.Errorf("live execution") }
		}},
		{"duplicate native key", duplicateLegacyWorkflowTicket},
		{"active attempt", markLegacyWorkflowAttempt},
		{"foreign use case", markLegacyWorkflowUseCase},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowPlanAdoptionFixture(t)
			tc.change(t, store, &opts)

			if _, err := store.AdoptPlanWorkflow(t.Context(), opts); err == nil {
				t.Fatal("unsafe adoption accepted")
			}

			snapshot, err := store.Snapshot(t.Context())
			if err != nil || len(snapshot.Records) != 0 {
				t.Fatalf("partial adoption: %d, %v", len(snapshot.Records), err)
			}
		})
	}
}

func duplicateLegacyWorkflowTicket(t *testing.T, store *WorkflowStore, opts *WorkflowPlanAdoptionOptions) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(store.layout.Root(), opts.Tasks[0].Path))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(store.layout.Root(), "ticket", "duplicate.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func markLegacyWorkflowAttempt(t *testing.T, store *WorkflowStore, opts *WorkflowPlanAdoptionOptions) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), opts.Tasks[0].Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	data = []byte(strings.Replace(string(data), "---\n\n", "active_attempt_id: \"live-attempt\"\n---\n\n", 1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	opts.Tasks[0].SourceHash = contenthash.Sum(data)
}

// TestWorkflowPlanAdoptionOwned blocks an actual held legacy operation lock and
// then recovers matching partial adoption using the existing transition journal.
func TestWorkflowPlanAdoptionOwned(t *testing.T) {
	store, opts := workflowPlanAdoptionFixture(t)

	lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "fixture"})
	if err != nil {
		t.Fatal(err)
	}

	_, blocked := store.AdoptPlanWorkflow(t.Context(), opts)
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}

	if blocked == nil {
		t.Fatal("live legacy operation did not block adoption")
	}

	store.afterReplace = func(int) error { return fmt.Errorf("fixture interrupted") }
	if _, err := store.AdoptPlanWorkflow(t.Context(), opts); err == nil {
		t.Fatal("interruption was not reached")
	}

	store.afterReplace = nil

	result, err := store.AdoptPlanWorkflow(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tasks) != 6 {
		t.Fatal("partial adoption was duplicated or omitted")
	}

	if _, err := os.Stat(filepath.Join(store.layout.Root(), workflowstate.JournalPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unrecovered journal: %v", err)
	}
}

// TestWorkflowPlanAdoptionProcess exits only the owned adoption fixture process
// after the first replacement, leaving its legacy exclusion locks behind.
func TestWorkflowPlanAdoptionProcess(t *testing.T) {
	dir := os.Getenv("CHR_ADOPTION_DIR")
	if dir == "" {
		return
	}

	data, err := os.ReadFile(filepath.Join(dir, "adoption-input.json"))
	if err != nil {
		t.Fatal(err)
	}

	var opts WorkflowPlanAdoptionOptions
	if err := json.Unmarshal(data, &opts); err != nil {
		t.Fatal(err)
	}

	opts.CheckQuiescent = func(context.Context, []string) error { return nil }

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	store.afterReplace = func(int) error { os.Exit(17); return nil }
	if _, err := store.AdoptPlanWorkflow(t.Context(), opts); err != nil {
		t.Fatal(err)
	}

	t.Fatal("owned crash was not reached")
}

// TestWorkflowPlanAdoptionRestart proves adoption is resumable after process
// death, including reclamation of only this operation's dead legacy locks.
func TestWorkflowPlanAdoptionRestart(t *testing.T) {
	store, opts := workflowPlanAdoptionFixture(t)

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(store.layout.ProjectDir(), "adoption-input.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestWorkflowPlanAdoptionProcess$")

	cmd.Env = append(os.Environ(), "CHR_ADOPTION_DIR="+store.layout.ProjectDir())
	output, err := cmd.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned adoption crash = %v: %s", err, output)
	}

	result, err := store.AdoptPlanWorkflow(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	checkAdoptedWorkflowTasks(t, store, result, opts)
}

func markLegacyWorkflowUseCase(t *testing.T, store *WorkflowStore, opts *WorkflowPlanAdoptionOptions) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), opts.Tasks[0].Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	data = []byte(strings.Replace(string(data), `use_case: ""`, `use_case: "other-case"`, 1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	opts.Tasks[0].SourceHash = contenthash.Sum(data)
}
