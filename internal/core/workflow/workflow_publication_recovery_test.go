package workflow

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPublicationCrashProcess exits while the shared delivery fan-out
// retains a confirmed common-directory intent and its exact submission identity.
func TestWorkflowPublicationCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_PUBLICATION_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id := os.Getenv("CHR_PUBLICATION_CRASH_ID")

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	journal, err := store.readWorkflowSubmissionJournal(t.Context(), run)
	if err != nil {
		t.Fatal(err)
	}

	index, err := strconv.Atoi(os.Getenv("CHR_PUBLICATION_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(replaced int) error {
		if replaced == index && workflowFinalizationJournal(t, store) {
			os.Exit(17)
		}

		return nil
	}

	pr := journal.Attempts[len(journal.Attempts)-1].Guard.PullRequest
	provider := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: &workflowPreparationProviderFixture{workflowReviewProviderFixture: &workflowReviewProviderFixture{workflowPRProviderFixture: &workflowPRProviderFixture{pr: pr}}}, remote: os.Getenv("CHR_PUBLICATION_CRASH_REMOTE")}
	ports := PublicationPorts{Provider: provider, Observer: NewWorkflowIntegrationGit(pr.Repository, provider.remote)}
	_, err = store.AdvanceWorkflowSliceIntegration(t.Context(), id, ports)
	t.Fatalf("owned publication writer did not exit: %v", err)
}

// TestWorkflowPublicationCrash proves restart replay after receipt introduction,
// task fan-out and final checkpoint. Canonical delivery is never submitted twice.
func TestWorkflowPublicationCrash(t *testing.T) {
	for _, index := range []int{0, 5, 9} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			f, p := newWorkflowSubmissionFixture(t, "rebase")
			if _, err := f.store.SubmitWorkflowSliceIntegration(t.Context(), f.run.ID, p, workflowSubmissionObserver(f)); err != nil {
				t.Fatal(err)
			}

			crashWorkflowPublication(t, f, index)

			restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

			progress, err := restarted.AdvanceWorkflowSliceIntegration(t.Context(), f.run.ID, workflowPublicationPorts(f, p))
			if err != nil || progress.Receipt == nil || progress.Invocation.Submission.State != workflowstate.SubmissionConfirmed || p.calls != 1 {
				t.Fatal("publication recovery lost verified intent or repeated mutation", err)
			}

			checkWorkflowPublicationFinalized(t, f, progress)
			checkWorkflowPublicationReuse(t, restarted, f.run.ID, workflowPublicationPorts(f, p), progress)
		})
	}
}

func crashWorkflowPublication(t *testing.T, f workflowPreparationFixture, index int) {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowPublicationCrashProcess$")

	command.Env = append(os.Environ(), "CHR_PUBLICATION_CRASH_DIR="+f.store.layout.ProjectDir(), "CHR_PUBLICATION_CRASH_ID="+f.run.ID, "CHR_PUBLICATION_CRASH_INDEX="+strconv.Itoa(index), "CHR_PUBLICATION_CRASH_REMOTE="+f.remote)
	output, err := command.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("owned publication crash: %v: %s", err, output)
	}
}
