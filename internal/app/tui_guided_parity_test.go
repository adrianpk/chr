package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

// TestGuidedUseCaseParityNative continues one Web run through actual Telegram
// and terminal inputs, preserving native review/limits and verified delivery.
func TestGuidedUseCaseParityNative(t *testing.T) {
	t.Parallel()

	f, operations, provider, host, publication := guidedTelegramSharedFixture(t)
	original := readWebGuidedSession(t, operations, "guided-quiz")
	f.send(t, "Show use case Guided clarification in project Shared project", 0)

	operations.guidedWorkHost = host
	executor, model := guidedTUIModel(operations)
	openGuidedTUI(t, model)
	runGuidedTUIKey(t, model, "enter", true)

	if current := readWebGuidedSession(t, operations, original.ID); current.Revision != original.Revision {
		t.Fatal("terminal navigation advanced the shared run")
	}

	lifecycle := &telegramGuidedHost{receiver: f.receiver, shared: host}
	pumpGuidedTelegram(t, lifecycle)
	host.wait.Wait()
	runGuidedTUIKey(t, model, "r", true)

	if !strings.Contains(model.View(), "Native review requires approval") {
		t.Fatal("terminal hid the real native wait", model.View())
	}

	started := readWebGuidedSession(t, operations, original.ID)
	native := assertGuidedNativeAdmission(t, operations, started)
	assertGuidedTUIExport(t, executor, started, false)
	runGuidedTUIKey(t, model, "p", true)
	host.wait.Wait()

	paused := readWebGuidedSession(t, operations, original.ID)
	if paused.Guided.Status != "paused" {
		t.Fatal("terminal pause did not reach the shared owner")
	}

	assertGuidedTUIStaleScope(t, executor, operations, started)
	f.send(t, "Show current step", 0)
	f.send(t, "Resume use case", 0)
	host.wait.Wait()
	runGuidedTUIKey(t, model, "r", true)
	runGuidedTUIKey(t, model, "i", false)
	runGuidedTUIKey(t, model, "Keep an export for a later review.", false)
	runGuidedTUIKey(t, model, "enter", true)

	current := readWebGuidedSession(t, operations, original.ID)

	retained := assertGuidedNativeAdmission(t, operations, current)
	if retained.ID != native.ID || retained.Policy != native.Policy || serviceWorkTaskAttempts(retained) != 1 || len(current.Guided.LaterIdeas) != 1 {
		t.Fatal("interface continuation changed invocation, attempts or later ideas")
	}

	completeGuidedWebPublication(t, provider, host, publication, current)

	completed := readWebGuidedSession(t, operations, original.ID)
	assertGuidedTUIExport(t, executor, completed, true)
	assertGuidedTUIFinalParity(t, executor, f, operations, provider, completed)
}

func assertGuidedTUIStaleScope(t *testing.T, executor tuiGuidedExecutor, operations service, old authoringcore.AuthoringSession) {
	t.Helper()

	for _, foreign := range []string{"revision", "project", "namespace"} {
		command := guidedQuizCommand(old, "stale-terminal-"+foreign, "cancel", "")
		if foreign == "project" {
			command.ProjectID = "another-project"
		}

		if foreign == "namespace" {
			command.Namespace = "another-namespace"
		}

		if _, err := executor.ExecuteGuided(tui.GuidedRequest{Action: "input", Command: command}); !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal("stale or foreign terminal command was accepted", foreign, err)
		}
	}

	if current := readWebGuidedSession(t, operations, old.ID); current.Guided.Status != "paused" {
		t.Fatal("foreign command changed the shared stop state")
	}
}

func assertGuidedTUIExport(t *testing.T, executor tuiGuidedExecutor, session authoringcore.AuthoringSession, complete bool) {
	t.Helper()

	view, err := executor.ExecuteGuided(tui.GuidedRequest{Action: "document", Document: "results", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}}})
	if err != nil || view.Document == nil {
		t.Fatal("verified terminal result export unavailable", err)
	}

	var result guidedResultExport
	if err := json.Unmarshal([]byte(view.Document.Content), &result); err != nil {
		t.Fatal(err)
	}

	if result.Complete != complete {
		t.Fatal("terminal confused native work with confirmed delivery")
	}

	if complete {
		assertGuidedTUIEvidence(t, session, result)
	}
}

func assertGuidedTUIFinalParity(t *testing.T, executor tuiGuidedExecutor, f *telegramConversationAppFixture, operations service, provider webViewProvider, session authoringcore.AuthoringSession) {
	t.Helper()

	request := tui.GuidedRequest{Action: "inspect", Command: authoringcore.GuidedUseCaseCommand{GuidedUseCaseScope: authoringcore.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}}}

	terminal, err := executor.ExecuteGuided(request)
	if err != nil || terminal.Status != "completed" || !strings.Contains(strings.Join(terminal.Progress, "\n"), "Code:") {
		t.Fatal("terminal lacks verified code locations", err)
	}

	web, err := provider.GuidedUseCaseView(t.Context(), session.ID, session.Namespace)
	if err != nil || !guidedParityWebComplete(web) {
		t.Fatal("Web lost the same verified results", err)
	}

	f.send(t, "Show current step", 0)

	telegram := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if !guidedTUIRunMatches(terminal, telegram, session.ID) {
		t.Fatal("interfaces exchanged canonical run, actor or later ideas")
	}

	for _, kind := range []string{"plan", "tracker"} {
		assertGuidedTUIPreparedOriginal(t, executor, operations, request, kind)
	}
}

func guidedTUIRunMatches(terminal tui.GuidedView, telegram authoringcore.AuthoringSession, runID string) bool {
	return terminal.Session.ID == telegram.ID && telegram.ID == runID && terminal.Session.Guided.Actor == telegram.Guided.Actor && len(telegram.Guided.LaterIdeas) == 1
}

func guidedParityWebComplete(view webui.GuidedUseCaseView) bool {
	return view.Result != nil && view.Result.Complete && len(view.Result.Deliveries) != 0
}

func assertGuidedTUIPreparedOriginal(t *testing.T, executor tuiGuidedExecutor, operations service, request tui.GuidedRequest, kind string) {
	t.Helper()

	original, err := operations.ReadGuidedUseCaseOriginal(t.Context(), request.Command.RunID, request.Command.Namespace, kind)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(operations.projectContext.ProjectDir, original.Reference), []byte("Independent later replacement.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	request.Action, request.Document = "document", kind

	view, err := executor.ExecuteGuided(request)
	if err != nil || view.Document == nil || view.Document.Content != original.Content || view.Document.Commit != original.Commit {
		t.Fatal("terminal followed a later prepared-input replacement", kind, err)
	}
}
