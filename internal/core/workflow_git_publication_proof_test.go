package core

import (
	"errors"
	"testing"
)

// TestWorkflowGitCandidate rejects changed or forged proof before publication,
// even when invalid structural input has been assigned a fresh request hash.
func TestWorkflowGitCandidate(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowGitPublication)
		reseal bool
	}{
		{"ownership", func(r *WorkflowGitPublication) { r.SourceRewriteAuthorized = false }, true},
		{"fork", func(r *WorkflowGitPublication) { r.SourceRepository = "https://forge.example.invalid/other/project" }, true},
		{"source_ref", func(r *WorkflowGitPublication) { r.SourceRef = "refs/heads/*" }, true},
		{"target_ref", func(r *WorkflowGitPublication) { r.TargetRef = "refs/heads/dev:other" }, true},
		{"same_ref", func(r *WorkflowGitPublication) { r.SourceRef = r.TargetRef }, true},
		{"source_noop", func(r *WorkflowGitPublication) { r.Git.ResultCommit = r.Git.SourceCommit }, true},
		{"target_noop", func(r *WorkflowGitPublication) { r.Git.ResultCommit = r.Git.BaseCommit }, true},
		{"revision", func(r *WorkflowGitPublication) { r.OwnershipRevision = "changed" }, false},
		{"method", func(r *WorkflowGitPublication) { r.Git.Method = "squash" }, true},
		{"tree", func(r *WorkflowGitPublication) { r.Git.PreparedTree = r.Git.SourceTree }, true},
		{"history", func(r *WorkflowGitPublication) { r.Git.History[0].Parent = r.Git.SourceCommit }, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkflowGitFixture(t, "merge")
			request := fixture.request
			test.change(&request)

			if test.reseal {
				sealWorkflowGitFixture(t, &request)
			}

			result, err := NewWorkflowGitPublisher(request.Repository, fixture.remote).PublishWorkflowGit(t.Context(), request)
			if err == nil || result.CommandsSent || result.State != "not_sent" {
				t.Fatalf("invalid candidate published: %+v, %v", result, err)
			}

			checkWorkflowGitRefs(t, fixture, fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit)
		})
	}
}

// TestWorkflowGitAlreadyPrepared prevents CLI no-op filtering from silently
// removing one immutable guard when an advertised ref already points to P.
func TestWorkflowGitAlreadyPrepared(t *testing.T) {
	for _, ref := range []string{"source", "target"} {
		t.Run(ref, func(t *testing.T) {
			fixture := newWorkflowGitFixture(t, "merge")
			workflowIntegrationCommand(t, fixture.request.Directory, "push", fixture.remote, fixture.request.Git.ResultCommit+":refs/heads/objects-only")

			source, target := fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit

			fullRef := fixture.request.SourceRef
			if ref == "source" {
				source = fixture.request.Git.ResultCommit
			} else {
				fullRef, target = fixture.request.TargetRef, fixture.request.Git.ResultCommit
			}

			workflowIntegrationCommand(t, fixture.remote, "update-ref", fullRef, fixture.request.Git.ResultCommit)

			result, err := NewWorkflowGitPublisher(fixture.request.Repository, fixture.remote).PublishWorkflowGit(t.Context(), fixture.request)
			if !errors.Is(err, ErrWorkflowObservationRequired) || result.CommandsSent {
				t.Fatalf("already-P advertisement was partially published: %+v, %v", result, err)
			}

			checkWorkflowGitRefs(t, fixture, source, target)
		})
	}
}

// TestWorkflowGitFastForward independently blocks a target update that explicit
// old-object leases would otherwise allow to overwrite newer target history.
func TestWorkflowGitFastForward(t *testing.T) {
	fixture := newWorkflowGitFixture(t, "merge")
	newer := workflowIntegrationCommit(t, fixture.request.Directory, "newer.txt", "Newer target")
	workflowIntegrationCommand(t, fixture.request.Directory, "push", fixture.remote, newer+":"+fixture.request.TargetRef)

	fixture.request.Git.BaseCommit = newer
	sealWorkflowGitFixture(t, &fixture.request)

	result, err := NewWorkflowGitPublisher(fixture.request.Repository, fixture.remote).PublishWorkflowGit(t.Context(), fixture.request)
	if !errors.Is(err, ErrWorkflowConflict) || result.CommandsSent {
		t.Fatalf("non-fast-forward target published: %+v, %v", result, err)
	}

	checkWorkflowGitRefs(t, fixture, fixture.request.Git.SourceCommit, newer)
}
