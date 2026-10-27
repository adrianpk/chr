package app

import (
	"fmt"
	"strings"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramControlDecisionFixture(t *testing.T, action string) (*telegramConversationAppFixture, service, string) {
	t.Helper()

	var (
		operations service
		id         string
	)
	if action == "answer_decision" {
		operations, id = serviceIntegrationDecisionReady(t)
	} else {
		operations, id = serviceIntegrationReady(t)
	}

	f := newTelegramConversationAppFixture(t)

	registry, err := projectcore.NewRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	project := projectcore.Project{ID: operations.projectContext.ProjectID, Name: "Control project", Path: operations.projectContext.ProjectDir, Enabled: true, RegisteredAt: time.Now()}
	if err := registry.AddProject(t.Context(), project); err != nil {
		t.Fatal(err)
	}

	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].ProjectIDs = []string{project.ID}
	if err := telegramcore.SaveReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	operations.env = f.receiver.env
	f.receiver.operations = operations

	return f, operations, id
}

func telegramControlDecisionCommand(t *testing.T, operations service, id, action string) string {
	t.Helper()

	switch action {
	case "approve":
		return "/approve " + id
	case "grant":
		return "/grant " + id + " submission 1 One additional bounded submission"
	case "answer_decision":
		status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
		if err != nil || status.Integration == nil || status.Integration.Decision == nil {
			t.Fatal("question unavailable", err)
		}

		return fmt.Sprintf("/answer-decision %s %s resolve 1 Preserve both requirements", id, status.Integration.Decision.Requests[0].ID)
	default:
		t.Fatal("unknown control", action)

		return ""
	}
}

// TestTelegramControlDecisions confirms exact review, grant and question controls
// over real HTTP/SQLite. Each journey loses its acknowledgement, reopens SQLite
// and preserves a later operator pause while retaining native receipts and actors.
func TestTelegramControlDecisions(t *testing.T) {
	telegramControlParallel(t)

	for _, action := range []string{"approve", "grant", "answer_decision"} {
		t.Run(action, func(t *testing.T) { checkTelegramControlDecision(t, action) })
	}
}

func checkTelegramControlDecision(t *testing.T, action string) {
	t.Helper()

	f, operations, id := telegramControlDecisionFixture(t, action)

	before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	token := telegramControlDecisionOffer(t, f, operations, id, action)
	input := telegramControlInterruptAck(t, f, operations, token, action)

	after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	checkTelegramControlDecisionEffect(t, action, before, after)

	current := telegramControlRecoverAck(t, f, operations, id, input)
	checkTelegramControlDecisionConversation(t, f, current)

	select {
	case notice := <-f.notices:
		if !strings.Contains(notice.Text, "The original control receipt is retained.") {
			t.Fatal("native receipt unavailable", notice)
		}
	default:
		t.Fatal("recovered confirmation was not delivered")
	}

	checkTelegramControlConsumed(t, f, operations, current, token, action)
}

func checkTelegramControlConsumed(t *testing.T, f *telegramConversationAppFixture, operations service, after WorkflowWorkStatus, token, action string) {
	t.Helper()

	if text := f.send(t, "/answer "+token+" "+action, 0).Text; !strings.Contains(text, "already confirmed") {
		t.Fatal("consumed control was repeated", text)
	}

	again, err := operations.WorkflowStatus(t.Context(), after.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || again.Revision != after.Revision {
		t.Fatal("duplicate control changed native state", err)
	}

	if refs, err := f.receiver.store.WorkReferences(t.Context(), f.receiver.botID, "", 16); err != nil || len(refs) != 0 {
		t.Fatal("control-only conversation granted automatic execution", refs, err)
	}
}

func checkTelegramControlDecisionEffect(t *testing.T, action string, before, after WorkflowWorkStatus) {
	t.Helper()

	if after.IntegratedUnits != 0 || after.ConfirmedTickets != 0 || after.Integration.Budget.Consumed != before.Integration.Budget.Consumed {
		t.Fatal("control fabricated delivery or changed consumed history")
	}

	switch action {
	case "approve":
		if after.ReviewState != "authorized" || after.ApprovalHash != before.ApprovalHash {
			t.Fatal("approval lost the exact review binding")
		}
	case "grant":
		if after.Integration.Budget.Limits.Submission != before.Integration.Budget.Limits.Submission+1 {
			t.Fatal("bounded grant missing")
		}
	case "answer_decision":
		checkWorkflowIntegrationAnswer(t, before, after, "resolved", nil)

		checkTelegramControlQuestionActor(t, after)
	}
}

func checkTelegramControlDecisionConversation(t *testing.T, f *telegramConversationAppFixture, after WorkflowWorkStatus) {
	t.Helper()

	c := f.sessions(t)[0]
	if c.InvocationID != after.ID || c.LastObservedRevision != after.Revision || c.AuthorizationRef == "" {
		t.Fatal("control conversation lost native identity", c)
	}
}

func checkTelegramControlQuestionActor(t *testing.T, after WorkflowWorkStatus) {
	t.Helper()

	if after.Integration.Decision.Answers[0].Actor != "maintainer" {
		t.Fatal("answer did not use the configured actor", after.Integration.Decision.Answers)
	}
}

func telegramControlDecisionOffer(t *testing.T, f *telegramConversationAppFixture, operations service, id, action string) string {
	t.Helper()

	command := telegramControlDecisionCommand(t, operations, id, action)
	if action != "grant" {
		return telegramControlOffer(t, f, command)
	}

	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "control", Arguments: ConversationIntentArguments{Action: "grant", Text: id, RecordJSON: `{"kind":"submission","amount":1,"reason":"One bounded submission"}`}})
	})

	token := telegramControlOffer(t, f, "Grant one additional submission for this work.")
	if calls.Load() != 1 {
		t.Fatal("grant intent was repeated", calls.Load())
	}

	return token
}
