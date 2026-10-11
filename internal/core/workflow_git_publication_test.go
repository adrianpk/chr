package core

import (
	"errors"
	"path/filepath"
	"testing"
)

type workflowGitFixture struct {
	request WorkflowGitPublication
	remote  string
}

func newWorkflowGitFixture(t *testing.T, method string) workflowGitFixture {
	t.Helper()

	root := t.TempDir()
	actor, remote := filepath.Join(root, "actor"), filepath.Join(root, "canonical.git")
	workflowIntegrationCommand(t, root, "init", "--quiet", "--initial-branch=dev", actor)
	workflowIntegrationCommand(t, actor, "config", "user.name", "Fixture")
	workflowIntegrationCommand(t, actor, "config", "user.email", "fixture@example.invalid")

	origin := workflowIntegrationCommit(t, actor, "initial.txt", "Initial fixture")
	workflowIntegrationCommand(t, actor, "checkout", "-b", "source")

	source := workflowIntegrationCommit(t, actor, "source.txt", "Source fixture")
	workflowIntegrationCommand(t, actor, "checkout", "dev")

	base := workflowIntegrationCommit(t, actor, "target.txt", "Target fixture")
	workflowFixtureIntegrate(t, actor, method)

	result := workflowIntegrationCommand(t, actor, "rev-parse", "HEAD")
	workflowIntegrationCommand(t, root, "init", "--bare", "--quiet", remote)
	workflowIntegrationCommand(t, actor, "push", remote, source+":refs/heads/chr/workflow/fixture", base+":refs/heads/dev")

	query := WorkflowIntegrationQuery{Repository: "https://forge.example.invalid/team/project", TargetBranch: "dev", Method: method, SourceCommit: source, SourceTree: workflowIntegrationCommand(t, actor, "rev-parse", source+"^{tree}"), OriginCommit: origin, ResultCommit: result}
	proof := WorkflowIntegrationProof{WorkflowIntegrationQuery: query, BaseCommit: base, ObservedTarget: base}

	reader := workflowIntegrationReader{dir: actor}
	if err := reader.verifyMethod(t.Context(), &proof); err != nil {
		t.Fatal(err)
	}

	var err error

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		t.Fatal(err)
	}

	request := WorkflowGitPublication{Repository: query.Repository, SourceRepository: query.Repository, Directory: actor, SourceRef: "refs/heads/chr/workflow/fixture", TargetRef: "refs/heads/dev", SourceRewriteAuthorized: true, OwnershipRevision: "fixture-ownership/v1", Git: proof}
	sealWorkflowGitFixture(t, &request)

	return workflowGitFixture{request: request, remote: remote}
}

func sealWorkflowGitFixture(t *testing.T, request *WorkflowGitPublication) {
	t.Helper()

	var err error

	request.CandidateRevision, err = WorkflowGitPublicationRevision(*request)
	if err != nil {
		t.Fatal(err)
	}
}

func checkWorkflowGitRefs(t *testing.T, fixture workflowGitFixture, source, target string) {
	t.Helper()

	for _, expected := range []struct{ ref, commit string }{{fixture.request.SourceRef, source}, {fixture.request.TargetRef, target}} {
		if actual := workflowIntegrationCommand(t, fixture.remote, "rev-parse", expected.ref); actual != expected.commit {
			t.Fatalf("%s = %s, want %s", expected.ref, actual, expected.commit)
		}
	}
}

// TestWorkflowGitAtomic publishes the same exact prepared object to both refs
// through a real receive-pack, preserving each configured history method.
func TestWorkflowGitAtomic(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			fixture := newWorkflowGitFixture(t, method)
			publisher := NewWorkflowGitPublisher(fixture.request.Repository, fixture.remote)

			result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
			if err != nil || result.State != "applied" || !result.CommandsSent {
				t.Fatalf("publication: %+v, %v", result, err)
			}

			checkWorkflowGitRefs(t, fixture, fixture.request.Git.ResultCommit, fixture.request.Git.ResultCommit)
		})
	}
}

// TestWorkflowGitAdvertisement checks actual remote values rather than local
// tracking refs, and refuses servers without atomic support before commands.
func TestWorkflowGitAdvertisement(t *testing.T) {
	for _, change := range []string{"source", "target", "atomic"} {
		t.Run(change, func(t *testing.T) {
			fixture := newWorkflowGitFixture(t, "merge")
			source, target := fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit
			expected := ErrWorkflowConflict

			switch change {
			case "source":
				source = target
				workflowIntegrationCommand(t, fixture.remote, "update-ref", fixture.request.SourceRef, source)
			case "target":
				target = source
				workflowIntegrationCommand(t, fixture.remote, "update-ref", fixture.request.TargetRef, target)
			case "atomic":
				expected = ErrWorkflowAtomicGuard

				workflowIntegrationCommand(t, fixture.remote, "config", "receive.advertiseAtomic", "false")
			}

			publisher := NewWorkflowGitPublisher(fixture.request.Repository, fixture.remote)

			result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
			if !errors.Is(err, expected) || result.CommandsSent || result.State != "not_sent" {
				t.Fatalf("publication: %+v, %v", result, err)
			}

			checkWorkflowGitRefs(t, fixture, source, target)
		})
	}
}
