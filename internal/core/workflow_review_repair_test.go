package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// TestWorkflowReviewRepairs carries the three-repair budget across real candidate
// changes and fresh store instances, retaining the validated prefix and one PR.
func TestWorkflowReviewRepairs(t *testing.T) {
	store, hierarchy, run, provider := workflowReviewFixture(t, workflowstate.ReviewAgentic)

	key := hierarchy.Tasks[2].Task.Key
	for count := 0; count < 3; count++ {
		checkWorkflowReviewRepairStep(t, store, hierarchy, run.ID, key, provider, count)
	}

	checkWorkflowReviewExhausted(t, store, run.ID, key, provider)

	if provider.creates != 1 {
		t.Fatal("review repair created another slice PR")
	}
}

func checkWorkflowReviewRepairStep(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, id, key string, provider *workflowReviewProviderFixture, count int) {
	t.Helper()

	caller := NewWorkflowStore(store.layout, store.projectID)

	binding, err := caller.ReviewWorkflowSlice(t.Context(), id, provider, workflowReviewWorker("needs_changes", key, ""))
	if err != nil {
		t.Fatal(err)
	}

	retained := workflowTestRead(t, store, id)
	checkWorkflowReviewRepairState(t, retained, binding, count)
	checkWorkflowReviewRepairPrefix(t, store, hierarchy.Tasks[:2])

	candidate, err := caller.ExecuteWorkflowSlice(t.Context(), id, workflowPRWorker())
	if err != nil {
		t.Fatal(err)
	}

	provider.pr.HeadCommit = candidate.Result.Commit
	if _, err := caller.EnsureWorkflowSlicePullRequest(t.Context(), id, workflowPRContent(), provider); err != nil {
		t.Fatal(err)
	}
}

func checkWorkflowReviewRepairState(t *testing.T, retained WorkflowInvocation, binding WorkflowReviewBinding, count int) {
	t.Helper()

	if binding.State != "needs_repair" || retained.Cursor != 2 || retained.Status != "pending" || retained.CandidateRef != "" {
		t.Fatal("repair did not rewind the affected suffix")
	}

	if len(retained.ReviewRepairs) != 1 || retained.ReviewRepairs[0].Consumed != count+1 {
		t.Fatal("candidate revision reset review repair consumption")
	}
}

func checkWorkflowReviewRepairPrefix(t *testing.T, store *WorkflowStore, tasks []WorkflowRecord) {
	t.Helper()

	for _, task := range tasks {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil || len(record.Task.Evidence) == 0 || record.Task.Phase != "validated" {
			t.Fatal("repair discarded the unaffected prefix")
		}
	}
}

func checkWorkflowReviewExhausted(t *testing.T, store *WorkflowStore, id, key string, provider *workflowReviewProviderFixture) {
	t.Helper()

	caller := NewWorkflowStore(store.layout, store.projectID)
	_, err := caller.ReviewWorkflowSlice(t.Context(), id, provider, workflowReviewWorker("needs_changes", key, ""))

	retained := workflowTestRead(t, store, id)
	if !errors.Is(err, ErrWorkflowReviewRepairBudget) || retained.Review.State != "repair_limit_exceeded" || retained.Blocker == nil {
		t.Fatalf("repair budget: %+v, %v", retained.Review, err)
	}

	if _, err := caller.ReviewWorkflowSlice(t.Context(), id, provider, []string{"false"}); !errors.Is(err, ErrWorkflowReviewRepairBudget) {
		t.Fatal(err)
	}

	if workflowTestRead(t, store, id).Revision != retained.Revision {
		t.Fatal("exhausted review retry reset the budget")
	}
}

// TestWorkflowReviewDuringIO refuses a role result after policy or Git changes,
// retaining the source evidence and a paused review checkpoint for reconciliation.
func TestWorkflowReviewDuringIO(t *testing.T) {
	cases := []struct{ name, before string }{
		{"policy", "import os\np=os.path.join(os.path.dirname(i['lane']['worktree']),'unused')\nopen(sys.argv[3],'w').write('review:\\n  mode: human_review\\n')"},
		{"lane", "open('result.txt','a').write('Unapproved change\\n')"},
		{"wrong binding", "i['review_hash']='another binding'"},
		{"invalid pass", ""},
		{"oversized", ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			testWorkflowReviewDuringIO(t, test.name, test.before)
		})
	}
}

// TestWorkflowReviewStop supervises durable pause/cancel at a running review
// checkpoint, keeps validated tasks and frees capacity only after process exit.
func TestWorkflowReviewStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) {
			testWorkflowReviewStop(t, intent)
		})
	}
}

func testWorkflowReviewDuringIO(t *testing.T, name, before string) {
	t.Helper()

	store, _, run, provider := workflowReviewFixture(t, workflowstate.ReviewAgentic)

	worker := workflowReviewWorker("pass", "", before)
	if name == "policy" {
		worker = append(worker, workspacecore.UserPath(store.layout.ProjectDir(), "config.yml"))
	}

	if name == "invalid pass" {
		worker = []string{"sh", "-c", "printf '{}\\n'"}
	}

	if name == "oversized" {
		worker[2] += "\nprint(' '*20000)"
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, worker); err == nil {
		t.Fatal("role returned an unauthorized grant")
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.Status != "paused" || retained.ResumeStatus != "waiting_review" || retained.CandidateRef != run.CandidateRef || retained.Review.State != "stale" {
		t.Fatal("review failure lost retained candidate or resume boundary")
	}

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), run.ID, provider); err == nil {
		t.Fatal("failed review still authorizes publication")
	}
}

func testWorkflowReviewStop(t *testing.T, intent string) {
	t.Helper()

	store, _, run, provider := workflowReviewFixture(t, workflowstate.ReviewAgentic)
	marker := filepath.Join(t.TempDir(), "started")
	worker := []string{"sh", "-c", "printf started > \"$1\"; sleep 20", "review", marker}
	done := make(chan error, 1)

	go func() { _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, worker); done <- err }()

	waitWorkflowClaimFile(t, marker)

	if _, err := store.StopWorkflow(t.Context(), run.ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatalf("review stop: %v", err)
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.Status != workflowstate.StopStatus(intent) || retained.ResumeStatus != "waiting_review" || retained.CandidateRef != run.CandidateRef {
		t.Fatal("stop rewound validated slice work")
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, []string{"false"}); err == nil {
		t.Fatal("stopped review restarted")
	}

	checkWorkflowReviewResume(t, store, run.ID, provider, intent)
	checkWorkflowReviewCapacity(t, store, marker)
}

func checkWorkflowReviewResume(t *testing.T, store *WorkflowStore, id string, provider *workflowReviewProviderFixture, intent string) {
	t.Helper()

	if intent == "pause" {
		if _, err := store.ResumeWorkflow(t.Context(), id); err != nil {
			t.Fatal(err)
		}

		if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, workflowReviewWorker("pass", "", "")); err != nil {
			t.Fatal(err)
		}
	}
}

func checkWorkflowReviewCapacity(t *testing.T, store *WorkflowStore, marker string) {
	t.Helper()

	capacity, err := acquireWorkflowCapacity(t.Context(), store.layout.ProjectDir())
	if err != nil {
		t.Fatal("review left owned capacity live", err)
	}

	if err := capacity.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatal(err)
	}
}
