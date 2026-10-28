package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowParallelPublisherSpec struct {
	ProjectDir, Namespace, InvocationID, Remote string
	Mode, Entered, Release, Result              string
	PullRequest                                 workflowstate.PullRequestSnapshot
}

type workflowParallelPublisherResult struct {
	ID, State, Error string
	Owned, Unknown   bool
}

// TestWorkflowParallelPublisher is an owned provider fixture. Its entry barrier
// is reached only after the shared target journal has retained publication intent.
func TestWorkflowParallelPublisher(t *testing.T) {
	value := os.Getenv("CHR_PARALLEL_PUBLISHER")
	if value == "" {
		return
	}

	var spec workflowParallelPublisherSpec
	if err := json.Unmarshal([]byte(value), &spec); err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: spec.ProjectDir, Namespace: spec.Namespace})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "publication-alias")
	provider := workflowParallelPublicationProviderFor(spec.PullRequest, spec.Remote)
	provider.unknown, provider.lost = spec.Mode == "unknown", spec.Mode == "lost"
	provider.before = func(ctx context.Context, submission workflowstate.Submission) {
		workflowParallelPublisherBarrier(t, ctx, store, spec, submission)
	}

	git := NewWorkflowIntegrationGit(spec.PullRequest.Repository, spec.Remote)
	submission, err := store.SubmitWorkflowSliceIntegration(t.Context(), spec.InvocationID, provider, git)
	result := workflowParallelPublisherResult{ID: submission.ID, State: string(submission.State), Owned: errors.Is(err, workflowstate.ErrOwned), Unknown: errors.Is(err, context.DeadlineExceeded)}

	if err != nil {
		result.Error = err.Error()
	}

	writeWorkflowParallelPublisherResult(t, spec.Result, result)
}

func workflowParallelPublisherBarrier(t *testing.T, ctx context.Context, store *Store, spec workflowParallelPublisherSpec, submission workflowstate.Submission) {
	t.Helper()

	retained, err := store.ReadWorkflowSubmission(ctx, spec.InvocationID)
	if err != nil || retained.ID != submission.ID || retained.State != workflowstate.SubmissionAwaitingObservation {
		t.Fatal("provider IO preceded durable target intent", err)
	}

	if err := opsfile.WriteAtomic(spec.Entered, []byte(submission.ID)); err != nil {
		t.Fatal(err)
	}

	if spec.Release != "" {
		waitWorkflowClaimFile(t, spec.Release)
	}
}

func writeWorkflowParallelPublisherResult(t *testing.T, path string, result workflowParallelPublisherResult) {
	t.Helper()

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	if err := opsfile.WriteAtomic(path, data); err != nil {
		t.Fatal(err)
	}
}

func workflowParallelPublisher(t *testing.T, lane workflowParallelPublicationLane, caller *Store, mode, dir, release string) (*exec.Cmd, workflowParallelPublisherSpec) {
	t.Helper()

	spec := workflowParallelPublisherSpec{ProjectDir: caller.layout.ProjectDir(), Namespace: caller.layout.Namespace(), InvocationID: lane.fixture.run.ID, Remote: lane.fixture.remote, Mode: mode, Entered: filepath.Join(dir, "entered"), Result: filepath.Join(dir, "result.json"), Release: release, PullRequest: lane.provider.pr}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowParallelPublisher$")

	command.Env = append(os.Environ(), "CHR_PARALLEL_PUBLISHER="+string(data))

	t.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				t.Error(err)
			}

			var exited *exec.ExitError
			if err := command.Wait(); err != nil && !errors.As(err, &exited) {
				t.Error(err)
			}
		}
	})

	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	return command, spec
}

func workflowParallelPublisherResultFor(t *testing.T, command *exec.Cmd, spec workflowParallelPublisherSpec) workflowParallelPublisherResult {
	t.Helper()

	var result workflowParallelPublisherResult

	waitWorkflowClaimFile(t, spec.Result)

	data, err := os.ReadFile(spec.Result)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if err := command.Wait(); err != nil {
		t.Fatal("owned publication fixture", err)
	}

	return result
}
