package app

import (
	"context"
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramJourney follows one conversation through actual Telegram/model
// HTTP, clarification, canonical preparation, restart, exact admission and
// unattended native integration. Fixture publication is not live bot evidence.
func TestTelegramJourney(t *testing.T) {
	t.Setenv(core.OpsUserEnv, "default")

	f, operations, _ := telegramWorkAdmissionFixture(t, "ticket")
	provider := telegramJourneyProvider(t, operations)
	operations.workflowPullRequests = provider
	f.receiver.operations = operations
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "tickets:\n  key_prefix: JRN\nruntime:\n  validation_command: git show HEAD:result.txt\nwork:\n  state_authority: tickets\nreview:\n  mode: trusted_auto_merge\n")

	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string { return telegramAuthoringBridgeGeneration(t, input) }, telegramAuthoringBridgeCapture)

	conversation, key := prepareTelegramJourney(t, f)
	if calls.Load() != 4 {
		t.Fatal("canonical preparation bypassed configured model", calls.Load())
	}

	telegramAuthoringReopen(t, f)

	operations = telegramOperationConcreteService(t, f.receiver)

	action, prompt := telegramWorkOffer(t, f, "/work --conversation "+conversation+" ticket "+key)
	if !strings.Contains(prompt.Text, "Confirmation also makes "+key) || !strings.Contains(prompt.Text, "git show HEAD:result.txt") {
		t.Fatal("preview hid the readiness transition or validation command", prompt)
	}

	checkTelegramWorkRunCount(t, operations, 0)

	notice := f.send(t, "/answer "+action.Token+" start", 0)
	if !strings.Contains(notice.Text, "Work admitted:") {
		t.Fatal("prepared conversation could not enter native work", notice.Text)
	}

	admission := append(json.RawMessage(nil), f.last...)
	run := checkTelegramWorkRunCount(t, operations, 1)
	stop := runTelegramJourneyReceiver(t, f.receiver)
	telegramWorkWaitTerminal(t, f)
	stop()
	checkTelegramJourneyReceipt(t, f, operations, provider, run.ID, key)

	completed := checkTelegramWorkRunCount(t, operations, 1)

	f.updates <- admission

	if _, err := f.receiver.step(t.Context()); err != nil {
		t.Fatal(err)
	}

	if current := checkTelegramWorkRunCount(t, operations, 1); !reflect.DeepEqual(current.Attempts, completed.Attempts) || provider.creates != 1 || len(provider.submissions) != 1 {
		t.Fatal("replayed admission repeated work", len(completed.Attempts), len(current.Attempts), provider.creates, len(provider.submissions))
	}
}

func telegramJourneyProvider(t *testing.T, operations service) *telegramWorkAutomaticProvider {
	t.Helper()

	base, ok := operations.workflowPullRequests.(*serviceWorkEntryProvider)
	if !ok {
		t.Fatal("native publication fixture unavailable")
	}

	return &telegramWorkAutomaticProvider{base}
}

func prepareTelegramJourney(t *testing.T, f *telegramConversationAppFixture) (string, string) {
	t.Helper()
	f.send(t, "/develop Retain the bounded retry ceiling", 0)

	conversation := f.sessions(t)[0].ID
	f.send(t, "/done "+conversation, 0)

	question := f.send(t, "/quiz "+conversation, 0)
	if !strings.Contains(question.Text, "Which retry ceiling") {
		t.Fatal("missing actual clarification", question)
	}

	f.send(t, "Five attempts.", question.ID)
	f.send(t, "/spec "+conversation, 0)
	f.send(t, "/plan "+conversation, 0)

	notice := f.send(t, "/prepare-work "+conversation, 0)
	if !strings.Contains(notice.Text, "Ticket prepared in backlog:") {
		t.Fatal("native work was not prepared", notice)
	}

	request := telegramAuthoringPreparedLatest(t, f, conversation)

	var result AuthoringWorkResult
	if err := json.Unmarshal(request.Result, &result); err != nil {
		t.Fatal(err)
	}

	if request.State != "confirmed" || result.Result.ReceiptRef == "" || result.Authorization == nil {
		t.Fatal("preparation lost its canonical receipt or scope", request)
	}

	return conversation, result.Result.Record.Identity
}

func runTelegramJourneyReceiver(t *testing.T, receiver telegramReceiver) func() {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	finished := make(chan error, 1)
	go func() { finished <- receiver.run(ctx) }()

	stop := sync.OnceFunc(func() {
		cancel()

		select {
		case err := <-finished:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(30 * time.Second):
			t.Error("journey receiver did not quiesce")
		}
	})
	t.Cleanup(stop)

	return stop
}

func checkTelegramJourneyReceipt(t *testing.T, f *telegramConversationAppFixture, operations service, provider *telegramWorkAutomaticProvider, id, key string) {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Status != "completed" || status.ConfirmedTickets != 1 || status.Remaining != 0 || status.DeliveryRef == "" || status.DeliveryHash == "" {
		t.Fatal("journey lacks confirmed native integration", status, err)
	}

	checkTelegramWorkAutomaticPolicy(t, provider, status)
	checkTelegramJourneyResult(t, operations, provider, key)

	conversation := f.sessions(t)[0]
	if conversation.InvocationID != id || conversation.Phase != "completed" {
		t.Fatal("conversation lost its native evidence", conversation)
	}
}

func checkTelegramJourneyResult(t *testing.T, operations service, provider *telegramWorkAutomaticProvider, key string) {
	t.Helper()

	output, err := exec.CommandContext(t.Context(), "git", "--git-dir", provider.remote, "show", "refs/heads/dev:result.txt").CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Delivered") {
		t.Fatal("target does not contain worker result", output)
	}

	page, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket, Identity: key})
	if err != nil || len(page.Records) != 1 || !strings.HasPrefix(page.Records[0].Reference, "ticket/") {
		t.Fatal("canonical prepared ticket unavailable", page, err)
	}
}
