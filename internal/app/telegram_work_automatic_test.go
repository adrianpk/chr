package app

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkAutomaticProvider struct{ *serviceWorkEntryProvider }

func (p *telegramWorkAutomaticProvider) SubmitWorkflowIntegration(ctx context.Context, submission core.WorkflowSubmission) (core.WorkflowSubmissionResponse, error) {
	proof := submission.Guard.Git
	branch := "refs/heads/" + p.pr.BaseBranch

	command := exec.CommandContext(ctx, "git", "-C", submission.PreparedLane.Worktree, "push", "--force-with-lease="+branch+":"+proof.BaseCommit, p.remote, proof.ResultCommit+":"+branch)
	if output, err := command.CombinedOutput(); err != nil {
		return core.WorkflowSubmissionResponse{}, fmt.Errorf("fixture publication: %w: %s", err, output)
	}

	p.submissions = append(p.submissions, submission.ID)
	p.pr.State, p.pr.Merged, p.pr.ResultCommit = "closed", true, proof.ResultCommit

	return core.WorkflowSubmissionResponse{Accepted: true, ResultCommit: proof.ResultCommit}, nil
}

// TestTelegramWorkAutomatic sends only admission and its confirmation. The real
// receiver then reaches trusted review and receipt finalization on a bare target.
func TestTelegramWorkAutomatic(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)

	base, ok := operations.workflowPullRequests.(*serviceWorkEntryProvider)
	if !ok {
		t.Fatal("native publication fixture unavailable")
	}

	provider := &telegramWorkAutomaticProvider{base}
	operations.workflowPullRequests = provider
	f.receiver.operations = operations
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\nreview:\n  mode: trusted_auto_merge\n")

	action, _ := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
	f.send(t, "/answer "+action.Token+" start", 0)

	run := checkTelegramWorkRunCount(t, operations, 1)
	telegramWorkRunReceiver(t, f.receiver, filepath.Join(t.TempDir(), "release"))
	telegramWorkWaitTerminal(t, f)

	status, err := operations.WorkflowStatus(t.Context(), run.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Status != "completed" || status.ConfirmedTickets != 1 || status.Remaining != 0 || status.DeliveryRef == "" || status.DeliveryHash == "" {
		t.Fatal("unattended work lacks native confirmation", status, err)
	}

	checkTelegramWorkAutomaticPolicy(t, provider, status)
}

func checkTelegramWorkAutomaticPolicy(t *testing.T, provider *telegramWorkAutomaticProvider, status WorkflowWorkStatus) {
	t.Helper()

	if provider.creates != 1 || len(provider.submissions) != 1 || status.ReviewState != "authorized" {
		t.Fatal("receiver repeated publication or bypassed policy", provider.creates, len(provider.submissions), status)
	}

	text := telegramWorkProgressText(status)
	if !strings.Contains(text, "Integration confirmed") || !strings.Contains(text, "confirmed tickets: 1") || strings.Contains(text, "0 unit(s)") {
		t.Fatal("terminal projection lost confirmed evidence", text)
	}
}

func telegramWorkWaitTerminal(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if conversations := f.sessions(t); len(conversations) == 1 && conversations[0].Phase == "completed" {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("receiver did not project a terminal native receipt")
}
