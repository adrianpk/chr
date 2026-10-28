package workflow

import (
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeObservationFixture(t *testing.T, fixture nativeCapabilityFixture) workflowstate.NativePublicationBinding {
	t.Helper()

	capability, err := fixture.adapter.WorkflowNativeIntegrationCapability(t.Context(), fixture.request, fixture.policy, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	publication, err := NewWorkflowGitPublisher(fixture.request.Publication.Repository, fixture.remote).PublishWorkflowGit(t.Context(), fixture.request.Publication)
	if err != nil {
		t.Fatal(err)
	}

	fixture.runner.outputs[fixture.prefix+"/branches/chr%2Fworkflow%2Ffixture"] = []byte(`{"protected":false,"commit":{"sha":"` + publication.ResultCommit + `"}}`)

	return workflowstate.NativePublicationBinding{Capability: capability, PullRequest: fixture.request.PullRequest, Git: fixture.request.Publication.Git, Publication: publication, SourceDirectory: fixture.request.Publication.Directory, ObservationDeadline: time.Now().Add(time.Minute)}
}

// TestNativeObservation proves actual target history alongside the original PR,
// tolerates S/P metadata lag and reports close, retarget and result races explicitly.
func TestNativeObservation(t *testing.T) {
	for _, forge := range []string{"forgejo", "github"} {
		t.Run(forge, func(t *testing.T) {
			fixture := newNativeCapabilityFixture(t, forge)
			binding := nativeObservationFixture(t, fixture)
			checkNativeObservationCases(t, fixture, binding)
		})
	}
}

func checkNativeObservationCases(t *testing.T, fixture nativeCapabilityFixture, binding workflowstate.NativePublicationBinding) {
	t.Helper()

	cases := []struct {
		name, state, condition string
		change                 func(*workflowstate.PullRequestSnapshot)
	}{
		{"original_head_lag", "observing", "", func(_ *workflowstate.PullRequestSnapshot) {}},
		{"published_head_lag", "observing", "", func(pr *workflowstate.PullRequestSnapshot) { pr.HeadCommit = binding.Git.ResultCommit }},
		{"closed", "reconciliation_required", "canonical_pr_closed_unmerged", func(pr *workflowstate.PullRequestSnapshot) { pr.State = "closed" }},
		{"retargeted", "reconciliation_required", "canonical_pr_retargeted", func(pr *workflowstate.PullRequestSnapshot) { pr.BaseBranch = "main" }},
		{"different_pr", "reconciliation_required", "canonical_pr_identity_changed", func(pr *workflowstate.PullRequestSnapshot) { pr.URL += "0" }},
		{"foreign_head", "reconciliation_required", "canonical_source_changed", func(pr *workflowstate.PullRequestSnapshot) { pr.HeadCommit = strings.Repeat("d", 40) }},
		{"wrong_result", "reconciliation_required", "canonical_pr_result_mismatch", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, binding.Git.SourceCommit
		}},
		{"merged_original_head", "merged", "", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, binding.Git.ResultCommit
		}},
		{"merged_published_head", "merged", "", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit, pr.HeadCommit = "closed", true, binding.Git.ResultCommit, binding.Git.ResultCommit
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			pr := binding.PullRequest
			test.change(&pr)
			setNativePRFixture(t, fixture, pr)

			observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote))
			if err != nil || observed.State != test.state || observed.Condition != test.condition || observed.OriginalSourceCommit != binding.Git.SourceCommit || observed.PublishedSourceCommit != binding.Git.ResultCommit {
				t.Fatalf("native observation: %+v, %v", observed, err)
			}

			checkNativeObservedHistory(t, observed, binding)
		})
	}
}

func checkNativeObservedHistory(t *testing.T, observed workflowstate.NativePublicationObservation, binding workflowstate.NativePublicationBinding) {
	t.Helper()

	if observed.State == "merged" && (observed.Git == nil || observed.Git.HistoryHash != binding.Git.HistoryHash || observed.Git.ObservedTarget != binding.Git.ResultCommit) {
		t.Fatal("merged metadata lacked the actual exact target history")
	}
}

// TestNativeLaterSource retains the original transaction when later source and
// target descendants arrive, while requiring stronger attribution for unknown IO.
func TestNativeLaterSource(t *testing.T) {
	fixture := newNativeCapabilityFixture(t, "github")
	binding := nativeObservationFixture(t, fixture)
	git := NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote)

	prior, err := workflowforge.VerifyNativeTarget(t.Context(), binding, git, workflowNativePublicationProofs{})
	if err != nil {
		t.Fatal(err)
	}

	later := workflowIntegrationCommit(t, binding.SourceDirectory, "later.txt", "Later external work")
	workflowIntegrationCommand(t, binding.SourceDirectory, "push", fixture.remote, later+":"+binding.Capability.SourceRef, later+":"+binding.Capability.TargetRef)

	pr := binding.PullRequest
	pr.HeadCommit, pr.BaseCommit, pr.ResultCommit, pr.State, pr.Merged = later, later, binding.Git.ResultCommit, "closed", true
	setNativePRFixture(t, fixture, pr)

	fixture.runner.outputs[fixture.prefix+"/branches/chr%2Fworkflow%2Ffixture"] = []byte(`{"commit":{"sha":"` + later + `"}}`)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "merged" || observed.LaterSourceCommit != later || observed.PublishedSourceCommit != binding.Git.ResultCommit {
		t.Fatalf("later source replaced original published proof: %+v, %v", observed, err)
	}

	binding.Publication.State, binding.Publication.PublishedSourceCommit = "unknown", ""

	observed, err = fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "reconciliation_required" {
		t.Fatal("unknown transaction attributed a later source without proof")
	}

	checkNativeLaterAttribution(t, fixture, binding, prior, later)
}

func checkNativeLaterAttribution(t *testing.T, fixture nativeCapabilityFixture, binding workflowstate.NativePublicationBinding, prior workflowstate.IntegrationProof, later string) {
	t.Helper()

	git := NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote)

	binding.GitApplication = &workflowstate.NativeGitObservation{SchemaVersion: 1, CandidateRevision: binding.Publication.CandidateRevision, SourceCommit: binding.Git.ResultCommit, TargetCommit: prior.ObservedTarget, Condition: "result_verified", Git: &prior}

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "merged" || observed.LaterSourceCommit != later {
		t.Fatal("retained actual application lost attribution after source advanced", err)
	}

	prior.HistoryHash = contenthash.Sum([]byte("different history"))

	if _, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git); err == nil {
		t.Fatal("foreign history attributed a later source")
	}
}

// TestNativeObservationProof rejects a mismatched transaction or absent target
// and exposes observation deadlines without retrying publication or finalizing.
func TestNativeObservationProof(t *testing.T) {
	fixture := newNativeCapabilityFixture(t, "forgejo")
	binding := nativeObservationFixture(t, fixture)
	git := NewWorkflowIntegrationGit(binding.Git.Repository, fixture.remote)
	binding.ObservationDeadline = time.Now().Add(-time.Second)

	observed, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git)
	if err != nil || observed.State != "reconciliation_required" || observed.Condition != "canonical_pr_observation_deadline" {
		t.Fatal("expired metadata observation did not retain reconciliation")
	}

	binding.Publication.OriginalSourceCommit = binding.Git.ResultCommit
	if _, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git); err == nil {
		t.Fatal("rewritten P replaced retained original source S")
	}

	binding.Publication.OriginalSourceCommit = binding.Git.SourceCommit
	workflowIntegrationCommand(t, fixture.remote, "update-ref", binding.Capability.TargetRef, binding.Git.BaseCommit, binding.Git.ResultCommit)

	if _, err := fixture.adapter.ObserveWorkflowNativePublication(t.Context(), binding, git); err == nil {
		t.Fatal("metadata without actual result reachability yielded proof")
	}
}
