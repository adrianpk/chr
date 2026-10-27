package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowReviewChangedInputs rejects approval reuse for changed policy,
// target, canonical PR, host evidence or authoritative task content.
func TestWorkflowReviewChangedInputs(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, *WorkflowStore, WorkflowInvocation, *workflowReviewProviderFixture)
	}{
		{"policy", func(t *testing.T, s *WorkflowStore, _ WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewConfig(t, s, workflowstate.ReviewTrustedAutoMerge)
		}},
		{"method", func(t *testing.T, s *WorkflowStore, _ WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewWrite(t, UserPath(s.layout.ProjectDir(), "config.yml"), []byte("review:\n  mode: human_review\nwork:\n  merge_strategy: squash\n"))
		}},
		{"target", func(_ *testing.T, _ *WorkflowStore, _ WorkflowInvocation, p *workflowReviewProviderFixture) {
			p.target = strings.Repeat("b", 40)
			p.pr.BaseCommit = p.target
		}},
		{"head", func(_ *testing.T, _ *WorkflowStore, _ WorkflowInvocation, p *workflowReviewProviderFixture) {
			p.pr.HeadCommit = strings.Repeat("b", 40)
		}},
		{"fork", func(_ *testing.T, _ *WorkflowStore, _ WorkflowInvocation, p *workflowReviewProviderFixture) {
			p.pr.HeadRepository = "https://forge.example/other/repo"
		}},
		{"closed", func(_ *testing.T, _ *WorkflowStore, _ WorkflowInvocation, p *workflowReviewProviderFixture) {
			p.pr.State = "closed"
		}},
		{"body", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewWrite(t, filepath.Join(s.layout.Root(), filepath.Dir(r.Members[0].Admission.Path), "artifact.md"), []byte("Changed task body\n"))
		}},
		{"candidate", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewWrite(t, filepath.Join(s.layout.Root(), r.CandidateRef), []byte("{}\n"))
		}},
		{"grant", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewWrite(t, filepath.Join(s.layout.Root(), r.Review.InputRef), []byte("{}\n"))
		}},
		{"dirty lane", func(t *testing.T, _ *WorkflowStore, r WorkflowInvocation, _ *workflowReviewProviderFixture) {
			workflowReviewWrite(t, filepath.Join(r.Lane.Worktree, "result.txt"), []byte("Changed Git tree\n"))
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store, _, run, provider := workflowReviewFixture(t, workflowstate.ReviewHuman)

			binding, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil)
			if err != nil {
				t.Fatal(err)
			}

			retained := workflowTestRead(t, store, run.ID)
			test.mutate(t, store, retained, provider)

			if _, err := store.ApproveWorkflowSlice(t.Context(), run.ID, binding.InputHash, provider); err == nil {
				t.Fatal("changed inputs reused explicit approval")
			}
		})
	}
}

func workflowReviewWrite(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowReviewPolicy resolves the existing defaults and only restrictive
// invocation overrides; invalid values cannot relax a project requirement.
func TestWorkflowReviewPolicy(t *testing.T) {
	for _, test := range []struct {
		name, body string
		mode       workflowstate.ReviewMode
		source     string
		valid      bool
	}{
		{"legacy", "", workflowstate.ReviewHuman, "legacy_default", true},
		{"project", "review:\n  mode: trusted_auto_merge\n", workflowstate.ReviewTrustedAutoMerge, "project", true},
		{"invalid", "review:\n  mode: automatic\n", "", "", false},
		{"method", "work:\n  merge_strategy: fast\n", "", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := parseWorkflowReviewPolicy([]byte(test.body))
			if (err == nil) != test.valid {
				t.Fatalf("policy: %+v, %v", p, err)
			}

			if test.valid && (p.Mode != test.mode || p.Source != test.source) {
				t.Fatal("effective policy differs")
			}
		})
	}
}

// TestWorkflowReviewRestriction retains admission policy without allowing a
// caller to relax the project mode on a retry.
func TestWorkflowReviewRestriction(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey:        "restricted",
		ReviewRestriction: workflowstate.ReviewHuman,
		Selection:         workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}
	run := workflowTestStart(t, store, request)
	workflowReviewConfig(t, store, workflowstate.ReviewTrustedAutoMerge)

	file, err := store.readWorkflowInvocation(run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	policy, err := store.workflowReviewPolicy(file)
	if err != nil || policy.Mode != workflowstate.ReviewHuman || policy.Source != "invocation_restriction" {
		t.Fatal("invocation relaxed project policy")
	}

	request.ReviewRestriction = workflowstate.ReviewTrustedAutoMerge
	if _, err := store.StartWorkflow(t.Context(), request); err == nil {
		t.Fatal("relaxing invocation override accepted")
	}

	data, err := json.Marshal(run.Invocation)
	if err != nil || !strings.Contains(string(data), "review_restriction") {
		t.Fatal("restriction was not retained")
	}
}
