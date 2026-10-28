package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeOutcomePublisherFixture struct {
	publish func(context.Context, GitPublication) (GitPublicationResult, error)
}

func (p nativeOutcomePublisherFixture) PublishWorkflowGit(ctx context.Context, request GitPublication) (GitPublicationResult, error) {
	return p.publish(ctx, request)
}

// TestNativeGitOutcomes uses actual atomic receive-pack and independently fetched
// history. Lost responses never retry; definitive rejection observes both refs.
func TestNativeGitOutcomes(t *testing.T) {
	cases := []struct{ name, application, state string }{
		{"applied", "git_applied", "awaiting_observation"},
		{"lost response", "git_applied", "awaiting_observation"},
		{"permissions", "not_applied", "awaiting_observation"},
		{"source race", "not_applied", "unpublished"},
		{"target race", "not_applied", "unpublished"},
		{"before send", "not_applied", "unpublished"},
		{"unknown unchanged", "unknown", "awaiting_observation"},
		{"unknown moved", "unknown", "awaiting_observation"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			nativeGitOutcome(t, test.name, test.application, test.state)
		})
	}
}

func nativeGitOutcome(t *testing.T, name, application, state string) {
	t.Helper()

	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	transport := NewWorkflowGitPublisher(f.candidate.Native.SourceRepository, remote)
	publisher.publisher = nativeOutcomePublisher(t, f, remote, transport, name)

	nativePublication, publishErr := f.store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	submission, result := nativePublication.Submission, nativePublication.Transport

	if (publishErr == nil) != (name == "applied") {
		t.Fatal("unexpected dispatch outcome", publishErr)
	}

	if submission.Native.TransportResponseRef == nil || submission.Native.DispatchCompletionRef == nil {
		t.Fatal("returned dispatch lost its durable outcome", publishErr)
	}

	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)
	git := NewWorkflowIntegrationGit(f.candidate.Native.SourceRepository, remote)

	retained, err := f.store.ReconcileWorkflowNativeGitSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), transport, git)
	if err != nil || retained.ID != submission.ID || retained.Native.ApplicationState != application || retained.State != state {
		t.Fatal("actual refs and prepared history were not classified", err)
	}

	checkNativeGitOutcome(t, f, publisher, remote, name, result, retained)
}

func nativeOutcomePublisher(t *testing.T, f nativeSubmissionFixture, remote string, transport OSWorkflowGitPublisher, name string) GitPublisher {
	t.Helper()

	if name == "permissions" {
		if err := os.WriteFile(filepath.Join(remote, "hooks", "pre-receive"), []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	if name == "lost response" {
		transport = nativeLostResponsePublisher(t, f.candidate.Native.SourceRepository, remote)
	}

	return nativeOutcomePublisherFixture{publish: func(ctx context.Context, request GitPublication) (GitPublicationResult, error) {
		if name == "unknown unchanged" || name == "unknown moved" {
			result, err := nativeUnknownExchange(ctx, request, transport)
			if name == "unknown moved" {
				moveNativeOutcomeRef(t, ctx, f, remote, "target race")
			}

			return result, err
		}

		before := request.BeforeSend

		request.BeforeSend = func(ctx context.Context, advertisement GitPublicationAdvertisement) error {
			if err := before(ctx, advertisement); err != nil {
				return err
			}

			moveNativeOutcomeRef(t, ctx, f, remote, name)

			return nil
		}
		if name == "before send" {
			moveNativeOutcomeRef(t, ctx, f, remote, "target race")
		}

		return transport.PublishWorkflowGit(ctx, request)
	}}
}

func moveNativeOutcomeRef(t *testing.T, ctx context.Context, f nativeSubmissionFixture, remote, name string) {
	t.Helper()

	b := f.candidate.Native
	ref, value := "", ""

	switch name {
	case "source race":
		ref, value = b.SourceRef, b.ValidatedBase
	case "target race":
		ref, value = b.TargetRef, b.OriginalSource
	default:
		return
	}

	if _, err := (workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: f.candidate.Lane.Worktree}}).Output(ctx, "--git-dir="+remote, "update-ref", ref, value); err != nil {
		t.Fatal(err)
	}
}

func nativeLostResponsePublisher(t *testing.T, repository, remote string) OSWorkflowGitPublisher {
	t.Helper()

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
	force := "env CHR_NATIVE_LOST_REPORT=" + quote(remote) + " " + quote(executable) + " -test.run=^TestWorkflowGitReceiveProxy$"
	endpoint, config := workflowSSHFixtureCommand(t, force)

	return NewWorkflowGitPublisher(repository, endpoint+remote).WithSSHConfig(config)
}

func nativeUnknownExchange(ctx context.Context, request GitPublication, transport OSWorkflowGitPublisher) (GitPublicationResult, error) {
	advertisement, err := transport.InspectWorkflowGitPublication(ctx, request)
	if err != nil {
		return GitPublicationResult{}, err
	}

	if err := request.BeforeSend(ctx, advertisement); err != nil {
		return GitPublicationResult{}, err
	}

	result := GitPublicationResult{Schema: "chr-git-publication-result/v1", State: "unknown", CommandsSent: true, OriginalSourceCommit: request.Git.SourceCommit, AdvertisedSourceCommit: advertisement.SourceCommit, AdvertisedTargetCommit: advertisement.TargetCommit, ResultCommit: request.Git.ResultCommit, CandidateRevision: request.CandidateRevision}

	return result, context.DeadlineExceeded
}

func checkNativeGitOutcome(t *testing.T, f nativeSubmissionFixture, publisher *nativeDispatchPublisherFixture, remote, name string, result GitPublicationResult, retained workflowstate.Submission) {
	t.Helper()

	if retained.Native.ReceiptRef != nil || retained.State == "confirmed" {
		t.Fatal("Git observation manufactured forge delivery")
	}

	if name == "lost response" {
		var response GitPublicationResult
		if err := f.store.readWorkflowNativeEvidence(retained.InvocationID, *retained.Native.TransportResponseRef, &response); err != nil || response.State != "unknown" || result.State != "unknown" {
			t.Fatal("observation invented a successful response", err)
		}
	}

	checkNativeGitOutcomeRefs(t, f, remote, name)

	if retained.State == "unpublished" {
		guard := nativeBudgetGuard(t, f.store, retained.InvocationID)
		if _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), guard, f.provider, publisher); !errors.Is(err, workflowstate.ErrConflict) || publisher.calls.Load() != 1 {
			t.Fatal("sealed rejection reused stale authorization", err)
		}
	} else {
		nativeDispatchNoRepeat(t, f, publisher, retained)
		nativeSubmissionPendingExcluded(t, f, retained)
	}
}

func checkNativeGitOutcomeRefs(t *testing.T, f nativeSubmissionFixture, remote, name string) {
	t.Helper()

	if name == "applied" || name == "lost response" {
		return
	}

	b := f.candidate.Native
	for _, ref := range []string{b.SourceRef, b.TargetRef} {
		if got := workflowIntegrationCommand(t, f.candidate.Lane.Worktree, "--git-dir="+remote, "rev-parse", ref); got == b.PreparedResult {
			t.Fatal("rejected transaction partially applied", ref)
		}
	}
}
