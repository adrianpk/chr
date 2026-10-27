package app

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func telegramWorkFixture(t *testing.T, plans ...core.GeneratedWorkflowPlan) (*telegramConversationAppFixture, service, core.WorkflowRecord) {
	t.Helper()

	operations, record, _, _ := serviceWorkNativeFixture(t, plans...)
	telegramWorkBindCase(t, operations)

	f := newTelegramConversationAppFixture(t)

	registry, err := projectcore.NewRegistry(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	project := projectcore.Project{ID: operations.projectContext.ProjectID, Name: "Workflow project", Path: operations.projectContext.ProjectDir, Enabled: true, RegisteredAt: time.Now()}
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
	if err := telegramcore.SaveTelegramReceiverConfig(f.receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	operations.env = f.receiver.env
	f.receiver.operations = operations

	return f, operations, record
}

func telegramWorkBindCase(t *testing.T, operations service) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.InspectWorkflowRecords(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Set != nil {
			record.Set.UseCase = "checkout"
			if _, err := store.Write(t.Context(), record, record.Revision); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func telegramWorkSelector(t *testing.T, operations service, independent core.WorkflowRecord, extent string) core.WorkflowSelector {
	t.Helper()

	selector := core.WorkflowSelector{Extent: extent}
	if extent == "ticket" {
		selector.Identity = independent.Task.Key
	}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.InspectWorkflowRecords(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Set != nil {
			switch extent {
			case "slice":
				selector.Identity = record.Set.Slices[0].ID
			case "slice_set":
				selector.Identity = record.ID
			case "use_case":
				selector.Identity = record.Set.UseCase
			}
		}
	}

	return selector
}

func telegramWorkOffer(t *testing.T, f *telegramConversationAppFixture, command string) (telegramcore.TelegramAction, telegramConversationNotice) {
	t.Helper()

	prompt := f.send(t, command, 0)
	if !strings.Contains(prompt.Text, "Start ") || !strings.Contains(prompt.Text, "Publication:") {
		t.Fatal("missing exact selection/policy preview", prompt)
	}

	parts := strings.Split(prompt.Text, "/answer ")
	if len(parts) != 2 {
		t.Fatal("missing exact action token", prompt)
	}

	token := strings.Fields(parts[1])[0]

	input, err := f.receiver.store.PendingUpdates(t.Context(), f.receiver.botID)
	if err != nil || len(input) != 0 {
		t.Fatal("preview source was not acknowledged", input, err)
	}

	return telegramcore.TelegramAction{Token: token}, prompt
}

// TestTelegramWorkExtents carries all five selectors through real HTTP, SQLite,
// shared preview/admission and canonical status without launching a worker.
func TestTelegramWorkExtents(t *testing.T) {
	for _, extent := range []string{"ticket", "slice", "slice_set", "use_case", "loose"} {
		t.Run(extent, func(t *testing.T) {
			f, operations, record := telegramWorkFixture(t)
			selector := telegramWorkSelector(t, operations, record, extent)
			command := strings.TrimSpace("/work " + strings.ReplaceAll(extent, "_", "-") + " " + selector.Identity)
			action, prompt := telegramWorkOffer(t, f, command)
			checkTelegramWorkRunCount(t, operations, 0)

			notice := f.send(t, "start", prompt.ID)
			if !strings.Contains(notice.Text, "Work admitted:") {
				t.Fatal("exact decision did not admit native work", notice, action)
			}

			checkTelegramWorkNative(t, operations, selector)

			if text := f.send(t, "/answer "+action.Token+" start", 0).Text; !strings.Contains(text, "already confirmed") {
				t.Fatal("duplicate decision did not reuse receipt", text)
			}

			checkTelegramWorkRunCount(t, operations, 1)
		})
	}
}

func checkTelegramWorkRunCount(t *testing.T, operations service, count int) core.WorkflowInvocation {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	runs, _, err := store.InspectWorkflowPage(t.Context(), 0, 10)
	if err != nil || len(runs) != count {
		t.Fatal("unexpected native invocations", runs, err)
	}

	if len(runs) == 0 {
		return core.WorkflowInvocation{}
	}

	return runs[0]
}

// TestTelegramWorkCommandBounds rejects missing, ambiguous and extra targets
// before any native work can be previewed or admitted.
func TestTelegramWorkCommandBounds(t *testing.T) {
	for _, test := range []struct {
		text  string
		valid bool
	}{
		{"/work ticket PRJ-1", true},
		{"/work loose", true},
		{"/work ticket", false},
		{"/work loose PRJ-1", false},
		{"/work slice one two", false},
		{"/work --project p --project q ticket PRJ-1", false},
		{"/work all", false},
	} {
		t.Run(test.text, func(t *testing.T) {
			_, recognized, err := parseTelegramWork(telegramcore.TelegramReceivedUpdate{Kind: "message", Text: test.text})
			if !recognized || (err == nil) != test.valid {
				t.Fatal("unexpected command validation", recognized, err)
			}
		})
	}
}

// TestTelegramWorkNatural uses the same typed admission path after configured
// HTTP intent resolution; proposing work alone never starts an invocation.
func TestTelegramWorkNatural(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "work", Arguments: ConversationIntentArguments{Action: "ticket", Text: record.Task.Key}})
	})
	telegramWorkOffer(t, f, "Work on the independent ticket.")
	checkTelegramWorkRunCount(t, operations, 0)

	if calls.Load() != 1 {
		t.Fatal("intent was reinterpreted", calls.Load())
	}
}

func telegramWorkDecodeResult(t *testing.T, request telegramcore.TelegramOperationRequest) telegramWorkAdmissionResult {
	t.Helper()

	var result telegramWorkAdmissionResult
	if err := json.Unmarshal(request.Result, &result); err != nil {
		t.Fatal(err)
	}

	return result
}

func checkTelegramWorkNative(t *testing.T, operations service, selector core.WorkflowSelector) {
	t.Helper()

	run := checkTelegramWorkRunCount(t, operations, 1)
	if run.Selection.Selector != selector || run.Admission == nil || run.Admission.Policy.Revision == "" || len(run.Attempts) != 0 {
		t.Fatal("native admission lost scope/policy or ran a worker", run)
	}

	status, err := operations.WorkflowStatus(t.Context(), run.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ID != run.ID || status.Total != len(run.Members) || status.ConfirmedTickets != 0 || status.ValidatedTasks != 0 {
		t.Fatal("another adapter cannot observe exact admission", status, err)
	}
}
