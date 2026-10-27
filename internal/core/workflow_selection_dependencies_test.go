package core

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowSelectionGit(t *testing.T, store *WorkflowStore, args ...string) string {
	t.Helper()

	argv := append([]string{"-C", store.layout.ProjectDir()}, args...)

	output, err := exec.Command("git", argv...).CombinedOutput()
	if err != nil {
		t.Fatalf("git: %v: %s", err, output)
	}

	return strings.TrimSpace(string(output))
}

func workflowSelectionCommit(t *testing.T, store *WorkflowStore) (string, string) {
	t.Helper()
	// Publish only the coordinator marker for an otherwise empty target fixture.
	// Task files remain host-owned inputs synchronized into the isolated lane.
	if _, err := os.Stat(filepath.Join(store.layout.Root(), workflowCoordinatorPath)); err == nil {
		workflowSelectionGit(t, store, "add", filepath.Join(OpsDirectoryName, store.layout.Namespace(), workflowCoordinatorPath))
	}

	workflowSelectionGit(t, store, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Fixture")

	return workflowSelectionGit(t, store, "rev-parse", "HEAD"), workflowSelectionGit(t, store, "rev-parse", "HEAD^{tree}")
}

func writeSelectionProof(t *testing.T, store *WorkflowStore, record WorkflowRecord, proof WorkflowTaskValidation) WorkflowRecord {
	t.Helper()

	relative := "run/fixture/evidence/" + record.Task.Key + ".json"

	path := filepath.Join(store.layout.Root(), relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	record.Task.Evidence, record.Task.Phase, record.Task.Status = []string{relative}, "validated", TicketStatusInProgress

	return workflowTestWrite(t, store, record)
}

// TestWorkflowSelectionPrerequisites checks exact tickets cannot pull in siblings or bypass lane evidence.
func TestWorkflowSelectionPrerequisites(t *testing.T) {
	store, result := workflowSelectionFixture(t)

	preview := workflowTestPreview(t, store, "ticket", result.Tasks[1].Task.Key)
	if len(preview.Members) != 1 || preview.Members[0].Blocker == nil || !strings.HasPrefix(preview.Members[0].Blocker.Reason, "outside_selection:") {
		t.Fatalf("unmet exact prerequisite: %+v", preview)
	}

	commit, tree := workflowSelectionCommit(t, store)
	branch := workflowSelectionGit(t, store, "symbolic-ref", "HEAD")
	proof := WorkflowTaskValidation{Schema: "chr-task-validation/v1", Key: result.Tasks[0].Task.Key, ContractHash: result.Tasks[0].Task.ContractHash, Branch: branch, Commit: commit, Tree: tree, Successful: true}
	writeSelectionProof(t, store, result.Tasks[0], proof)

	request := WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "ticket", Identity: result.Tasks[1].Task.Key}, Limit: 1, CumulativeBranch: branch, CumulativeCommit: commit}

	accepted, err := store.Preview(t.Context(), request)
	if err != nil || accepted.Members[0].Blocker != nil {
		t.Fatalf("valid cumulative proof blocked: %+v, %v", accepted, err)
	}

	request.CumulativeCommit = strings.Repeat("a", 40)

	rejected, err := store.Preview(t.Context(), request)
	if err != nil || rejected.Members[0].Blocker == nil {
		t.Fatal("foreign cumulative lane accepted")
	}
}

// TestWorkflowSelectionInvalidDependencies rejects forward, missing and cyclic work before admission.
func TestWorkflowSelectionInvalidDependencies(t *testing.T) {
	cases := []struct {
		name                string
		refs                []workflowstate.Ref
		failMaterialization bool
	}{
		{"forward", []workflowstate.Ref{{Kind: "ticket", ID: "task-1-3"}}, false},
		{"missing", []workflowstate.Ref{{Kind: "ticket", ID: "missing"}}, true},
		{"self", []workflowstate.Ref{{Kind: "ticket", ID: "task-1-1"}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, opts := workflowMaterializationFixture(t)
			opts.Plan.Slices[0].Tasks[0].DependsOn = tc.refs

			result, err := store.MaterializePlan(t.Context(), opts)
			if tc.failMaterialization {
				if err == nil {
					t.Fatal("invalid dependency persisted")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if _, err := store.Preview(t.Context(), WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "slice_set", Identity: result.Set.ID}, Limit: 128}); err == nil {
				t.Fatal("forward dependency admitted")
			}
		})
	}
}

func selectionReceipt(t *testing.T, store *WorkflowStore, unit workflowstate.Ref, commit, tree string) WorkflowRecord {
	t.Helper()

	identity := WorkflowIdentity{ID: "receipt-1", ProjectID: store.projectID, Namespace: store.layout.Namespace(), Revision: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	record := WorkflowRecord{WorkflowIdentity: identity, Kind: "receipt", Path: "run/fixture/evidence/receipt.json", Receipt: &WorkflowReceipt{Schema: WorkflowSchema, Unit: unit, CandidateID: "candidate", SubmissionID: "submission", Provider: "forgejo", Repository: "https://forge.example/test/repo", PullRequest: "https://forge.example/test/repo/pulls/1", Method: "rebase", SourceCommit: commit, BaseCommit: commit, PreparedTree: tree, HistoryProof: "host-verification", ResultCommit: commit, VerifiedTree: tree, ObservedTarget: commit, VerificationRef: "host-proof", ValidationRefs: []string{"host-validation"}, ConfirmedAt: time.Now().UTC()}}

	receipt, err := store.CreateReceipt(t.Context(), record)
	if err != nil {
		t.Fatal(err)
	}

	return receipt
}

// TestWorkflowSelectionTargetProof requires target reachability and canonical delivery facts for earlier slices.
func TestWorkflowSelectionTargetProof(t *testing.T) {
	store, result := workflowSelectionFixture(t)

	preview := workflowTestPreview(t, store, "slice", result.Set.Set.Slices[1].ID)
	if preview.Members[0].Blocker == nil || preview.Members[0].Blocker.Dependency.Kind != "slice" {
		t.Fatal("earlier slice integration bypassed")
	}

	commit, tree := workflowSelectionCommit(t, store)
	receipt := selectionReceipt(t, store, workflowstate.Ref{Kind: "slice", ID: result.Set.Set.Slices[0].ID}, commit, tree)
	set := result.Set
	set.Set.Slices[0].Phase, set.Set.Slices[0].DeliveryRef = "integrated", receipt.Path
	workflowTestWrite(t, store, set)

	request := WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "slice", Identity: result.Set.Set.Slices[1].ID}, Limit: 1, Target: WorkflowPreviewTarget{Repository: receipt.Receipt.Repository, Commit: commit}}

	confirmed, err := store.Preview(t.Context(), request)
	if err != nil || confirmed.Members[0].Blocker != nil {
		t.Fatalf("delivered prerequisite blocked: %+v, %v", confirmed, err)
	}

	unrelated := workflowSelectionGit(t, store, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit-tree", tree, "-m", "Unrelated lineage")
	request.Target.Commit = unrelated

	unreachable, err := store.Preview(t.Context(), request)
	if err != nil || unreachable.Members[0].Blocker == nil {
		t.Fatal("unreachable target delivery accepted")
	}

	request.Target.Commit = commit
	request.Target.Repository = "https://forge.example/foreign/repo"

	foreign, err := store.Preview(t.Context(), request)
	if err != nil || foreign.Members[0].Blocker == nil {
		t.Fatal("foreign target delivery accepted")
	}
}

// TestWorkflowSelectionBadProof rejects labels, failed validation, changed contracts and wrong trees.
func TestWorkflowSelectionBadProof(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*WorkflowTaskValidation)
	}{
		{"failed", func(p *WorkflowTaskValidation) { p.Successful = false }},
		{"contract", func(p *WorkflowTaskValidation) { p.ContractHash = "changed" }},
		{"ticket", func(p *WorkflowTaskValidation) { p.Key = "foreign" }},
		{"tree", func(p *WorkflowTaskValidation) { p.Tree = strings.Repeat("a", 40) }},
		{"branch", func(p *WorkflowTaskValidation) { p.Branch = "refs/heads/foreign" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, result := workflowSelectionFixture(t)
			commit, tree := workflowSelectionCommit(t, store)
			branch := workflowSelectionGit(t, store, "symbolic-ref", "HEAD")
			proof := WorkflowTaskValidation{Schema: "chr-task-validation/v1", Key: result.Tasks[0].Task.Key, ContractHash: result.Tasks[0].Task.ContractHash, Branch: branch, Commit: commit, Tree: tree, Successful: true}
			tc.mutate(&proof)
			writeSelectionProof(t, store, result.Tasks[0], proof)

			request := WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "ticket", Identity: result.Tasks[1].Task.Key}, Limit: 1, CumulativeBranch: branch, CumulativeCommit: commit}

			preview, err := store.Preview(t.Context(), request)
			if err != nil || preview.Members[0].Blocker == nil {
				t.Fatalf("invalid proof admitted: %+v, %v", preview, err)
			}
		})
	}
}

// TestWorkflowSelectionEvidencePages retains one evidence snapshot across pages and rejects changed proof bytes.
func TestWorkflowSelectionEvidencePages(t *testing.T) {
	store, result := workflowSelectionFixture(t)
	commit, tree := workflowSelectionCommit(t, store)
	branch := workflowSelectionGit(t, store, "symbolic-ref", "HEAD")
	proof := WorkflowTaskValidation{Schema: "chr-task-validation/v1", Key: result.Tasks[0].Task.Key, ContractHash: result.Tasks[0].Task.ContractHash, Branch: branch, Commit: commit, Tree: tree, Successful: true}
	task := writeSelectionProof(t, store, result.Tasks[0], proof)
	request := WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "slice_set", Identity: result.Set.ID}, Limit: 1, CumulativeBranch: branch, CumulativeCommit: commit}

	first, err := store.Preview(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	request.Offset, request.SnapshotHash = 1, first.SnapshotHash

	next, err := store.Preview(t.Context(), request)
	if err != nil || next.Members[0].Blocker != nil {
		t.Fatalf("unchanged proof pagination failed: %+v, %v", next, err)
	}

	path := filepath.Join(store.layout.Root(), task.Task.Evidence[0])
	if err := os.WriteFile(path, []byte(`{"schema":"changed"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := store.Preview(t.Context(), request); err == nil {
		t.Fatal("changed proof bytes kept preview snapshot")
	}
}

// TestWorkflowSelectionEarlierTasks prevents validation of the immediate predecessor from hiding an earlier missing proof.
func TestWorkflowSelectionEarlierTasks(t *testing.T) {
	store, result := workflowSelectionFixture(t)
	commit, tree := workflowSelectionCommit(t, store)
	branch := workflowSelectionGit(t, store, "symbolic-ref", "HEAD")
	task := result.Tasks[1]
	proof := WorkflowTaskValidation{Schema: "chr-task-validation/v1", Key: task.Task.Key, ContractHash: task.Task.ContractHash, Branch: branch, Commit: commit, Tree: tree, Successful: true}
	writeSelectionProof(t, store, task, proof)

	request := WorkflowPreviewRequest{Selector: WorkflowSelector{Extent: "ticket", Identity: result.Tasks[2].Task.Key}, Limit: 1, CumulativeBranch: branch, CumulativeCommit: commit}

	preview, err := store.Preview(t.Context(), request)
	if err != nil || preview.Members[0].Blocker == nil || preview.Members[0].Blocker.Dependency.ID != result.Tasks[0].Task.Key {
		t.Fatalf("earlier task proof bypassed: %+v, %v", preview, err)
	}
}
