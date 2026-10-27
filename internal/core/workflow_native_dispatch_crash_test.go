package core

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestNativeFaultProcess is the owned host subprocess for publication crashes.
func TestNativeFaultProcess(t *testing.T) {
	root := os.Getenv("CHR_NATIVE_FAULT_ROOT")
	if root == "" {
		return
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: root, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, os.Getenv("CHR_NATIVE_FAULT_PROJECT"))
	id := os.Getenv("CHR_NATIVE_FAULT_ID")

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	candidate, err := store.readWorkflowPreparation(run)
	if err != nil {
		t.Fatal(err)
	}

	provider := &nativePreparationProviderFixture{
		workflowPreparationProviderFixture: &workflowPreparationProviderFixture{
			workflowReviewProviderFixture: &workflowReviewProviderFixture{
				workflowPRProviderFixture: &workflowPRProviderFixture{pr: candidate.Input.Source.PullRequest},
				target:                    candidate.Native.ValidatedBase,
			}, guarded: true, revision: "guard-v1",
		}, policyRevision: "policy-v1",
	}
	stage := os.Getenv("CHR_NATIVE_FAULT_STAGE")
	configureNativeFaultCrash(store, stage)

	transport := NewWorkflowGitPublisher(candidate.Native.SourceRepository, os.Getenv("CHR_NATIVE_FAULT_REMOTE"))

	publisher := nativeFaultPublisher(t, transport, stage)
	if _, _, err := store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, store, id), provider, publisher); err != nil {
		t.Fatal(err)
	}
}

// TestNativeFaultCrashes retains one intent through actual host exits around
// intent, dispatch, send, server application and response retention.
func TestNativeFaultCrashes(t *testing.T) {
	for _, stage := range []string{"intent", "dispatch", "send", "applied", "response"} {
		t.Run(stage, func(t *testing.T) {
			checkNativeFaultCrash(t, stage)
		})
	}
}

// TestNativeFaultLease keeps an actual host alive after the old task lease
// expires. Replacement stays excluded until that same host finishes once.
func TestNativeFaultLease(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	dir := t.TempDir()
	marker, release := filepath.Join(dir, "ready"), filepath.Join(dir, "release")
	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNativeFaultProcess$")

	command.Env = append(os.Environ(), "CHR_NATIVE_FAULT_ROOT="+f.store.layout.ProjectDir(), "CHR_NATIVE_FAULT_PROJECT="+f.store.projectID, "CHR_NATIVE_FAULT_ID="+f.entry.InvocationID, "CHR_NATIVE_FAULT_STAGE=lease", "CHR_NATIVE_FAULT_REMOTE="+remote, "CHR_NATIVE_FAULT_MARKER="+marker, "CHR_NATIVE_FAULT_RELEASE="+release)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	cleanNativeFaultHost(t, command, release)
	waitWorkflowClaimFile(t, marker)

	expireNativeFaultLease(t, f)

	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)

	submission, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	if !errors.Is(err, ErrWorkflowObservationRequired) || publisher.calls.Load() != 0 {
		t.Fatal("expired lease replaced a live native dispatcher", err)
	}

	if alive, err := workflowProcessAlive(*submission.Native.DispatchIdentity); err != nil || !alive {
		t.Fatal("retained host was not actually alive", err)
	}

	finishNativeFaultLease(t, f, command, release, submission)
}

func configureNativeFaultCrash(store *WorkflowStore, stage string) {
	writes := 0
	wanted := map[string]int{"intent": 1, "dispatch": 2, "send": 3, "response": 4}[stage]
	store.afterReplace = func(index int) error {
		if index == -1 {
			writes++
			if wanted != 0 && writes == wanted {
				os.Exit(17)
			}
		}

		return nil
	}
}

func nativeFaultPublisher(t *testing.T, transport OSWorkflowGitPublisher, stage string) WorkflowGitPublisher {
	t.Helper()

	return nativeOutcomePublisherFixture{publish: func(ctx context.Context, request WorkflowGitPublication) (WorkflowGitPublicationResult, error) {
		if stage == "lease" {
			if err := os.WriteFile(os.Getenv("CHR_NATIVE_FAULT_MARKER"), []byte("ready"), 0o600); err != nil {
				t.Fatal(err)
			}

			waitWorkflowClaimFile(t, os.Getenv("CHR_NATIVE_FAULT_RELEASE"))
		}

		result, err := transport.PublishWorkflowGit(ctx, request)
		if stage == "applied" && err == nil {
			os.Exit(17)
		}

		return result, err
	}}
}

func checkNativeFaultCrash(t *testing.T, stage string) {
	t.Helper()

	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNativeFaultProcess$")

	command.Env = append(os.Environ(), "CHR_NATIVE_FAULT_ROOT="+f.store.layout.ProjectDir(), "CHR_NATIVE_FAULT_PROJECT="+f.store.projectID, "CHR_NATIVE_FAULT_ID="+f.entry.InvocationID, "CHR_NATIVE_FAULT_STAGE="+stage, "CHR_NATIVE_FAULT_REMOTE="+remote)
	output, err := command.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("host fault: %v: %s", err, output)
	}

	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)

	submission, _, err := f.store.PublishWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	if !errors.Is(err, ErrWorkflowObservationRequired) || publisher.calls.Load() != 0 || submission.ID == "" {
		t.Fatal("restart dispatched another publication", err)
	}

	checkNativeFaultRetained(t, f, submission)

	checkNativeFaultApplication(t, f, submission, remote, stage)
}

func checkNativeFaultRetained(t *testing.T, f nativeSubmissionFixture, submission WorkflowSubmission) {
	t.Helper()

	if submission.Native.DispatchIdentity != nil {
		if alive, err := workflowProcessAlive(*submission.Native.DispatchIdentity); err != nil || alive {
			t.Fatal("crashed host identity remained live", err)
		}
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("restart refilled submission budget", err)
	}

	nativeSubmissionPendingExcluded(t, f, submission)
}

func cleanNativeFaultHost(t *testing.T, command *exec.Cmd, release string) {
	t.Helper()
	t.Cleanup(func() {
		if command.ProcessState == nil {
			if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
				t.Error(err)
			}

			if err := command.Wait(); err != nil {
				t.Error(err)
			}
		}
	})
}

func expireNativeFaultLease(t *testing.T, f nativeSubmissionFixture) {
	t.Helper()

	err := f.store.withLock(t.Context(), func() error {
		run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
		if err != nil {
			return err
		}

		run.Attempts[len(run.Attempts)-1].LeaseUntil = run.Attempts[len(run.Attempts)-1].StartedAt.Add(time.Nanosecond)

		return f.store.commitWorkflowState(&run)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func checkNativeFaultApplication(t *testing.T, f nativeSubmissionFixture, submission WorkflowSubmission, remote, stage string) {
	t.Helper()

	if stage == "applied" || stage == "response" {
		b := f.candidate.Native

		observed, err := f.store.ReconcileWorkflowNativeGitSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), NewWorkflowGitPublisher(b.SourceRepository, remote), NewWorkflowIntegrationGit(b.SourceRepository, remote))
		if err != nil || observed.Native.ApplicationState != "git_applied" || observed.ID != submission.ID {
			t.Fatal("actual server application was not recovered", err)
		}
	}
}

func finishNativeFaultLease(t *testing.T, f nativeSubmissionFixture, command *exec.Cmd, release string, submission WorkflowSubmission) {
	t.Helper()

	if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := command.Wait(); err != nil {
		t.Fatal(err)
	}

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID)
	if err != nil || retained.ID != submission.ID || retained.Native.DispatchState != "quiescent" {
		t.Fatal("original host did not finish its same intent", err)
	}

	nativeSubmissionPendingExcluded(t, f, retained)
}
