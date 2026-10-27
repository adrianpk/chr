package app

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramAuthoringWorkFixture(t *testing.T) (service, authoringcore.ContinuationResult) {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git record-owner integration is selected explicitly")
	}

	operations, _ := telegramAuthoringServiceFixture(t, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "## Requirements\n\nRetain the configured retry ceiling.\n\n## Validation\n\nCheck exhaustion and reopening.", "")
	})

	project := operations.projectContext.ProjectDir
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "tickets:\n  key_prefix: PRJ\n")

	created, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceRequest())
	if err != nil {
		t.Fatal(err)
	}

	return operations, created
}

func telegramAuthoringWorkRequest(result authoringcore.ContinuationResult, intent authoringstate.Intent) AuthoringWorkRequest {
	return AuthoringWorkRequest{Binding: workflowstate.MutationBinding{RequestID: "work-description", Operator: "maintainer"}, SessionID: result.Session.ID, ExpectedRevision: result.Session.Revision, ExpectedSourceHash: result.SourceHash, Intent: intent, Instruction: "Retain the retry ceiling and existing defaults."}
}

// TestTelegramAuthoringWorkPreparation materializes a real native backlog ticket
// with canonical provenance and retained intent, without admitting an invocation.
func TestTelegramAuthoringWorkPreparation(t *testing.T) {
	for _, intent := range []authoringstate.Intent{authoringstate.IntentPrepare, authoringstate.IntentDevelop} {
		t.Run(string(intent), func(t *testing.T) {
			operations, created := telegramAuthoringWorkFixture(t)

			prepared, err := operations.PrepareAuthoringWork(t.Context(), telegramAuthoringWorkRequest(created, intent))
			if err != nil {
				t.Fatal(err)
			}

			prepared = telegramAuthoringWorkGrant(t, operations, prepared)

			// Dispatch the persisted preview after reopening, not a newly prepared key.
			encoded, err := json.Marshal(prepared)
			if err != nil {
				t.Fatal(err)
			}

			var retained PreparedAuthoringWork
			if err := json.Unmarshal(encoded, &retained); err != nil {
				t.Fatal(err)
			}

			result, err := operations.ApplyPreparedAuthoringWork(t.Context(), retained)
			if err != nil {
				t.Fatal(err)
			}

			telegramAuthoringWorkCheckRecord(t, operations, result)
			telegramAuthoringWorkCheckReplay(t, operations, retained, result)
			telegramAuthoringWorkNoAdmission(t, operations)
		})
	}
}

func telegramAuthoringWorkGrant(t *testing.T, operations service, prepared PreparedAuthoringWork) PreparedAuthoringWork {
	t.Helper()

	if prepared.Scope.Intent != authoringstate.IntentDevelop {
		return prepared
	}

	if _, err := operations.ApplyPreparedAuthoringWork(t.Context(), prepared); !errors.Is(err, authoringcore.ErrAuthorityRequired) {
		t.Fatal("missing development authority accepted", err)
	}

	grant, err := authoringcore.AuthorizeAuthoringWork(prepared.Scope, workflowstate.MutationBinding{RequestID: "explicit-instruction", Operator: "maintainer"}, prepared.Scope.Hash)
	if err != nil {
		t.Fatal(err)
	}

	prepared.Authorization = &grant

	return prepared
}

func telegramAuthoringWorkCheckRecord(t *testing.T, operations service, result AuthoringWorkResult) {
	t.Helper()

	store, err := operations.scopedRecordStore("original")
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 1 {
		t.Fatalf("native records=%+v err=%v", snapshot, err)
	}

	record := snapshot.Records[0]
	if record.Task.Status != core.TicketStatusBacklog || record.Task.Phase != "pending" || !strings.Contains(string(record.Body), "configured retry ceiling") || result.Scope.Namespace != "original" || result.Result.ReceiptRef == "" {
		t.Fatalf("not a retained proposed description: %+v", result)
	}

	telegramAuthoringWorkCheckProvenance(t, record, result)
}

func telegramAuthoringWorkCheckProvenance(t *testing.T, record core.WorkflowRecord, result AuthoringWorkResult) {
	t.Helper()

	var scope authoringcore.WorkScope
	if err := json.Unmarshal(record.Metadata["authoring_scope"], &scope); err != nil || scope != result.Scope {
		t.Fatal("canonical scope/provenance lost", scope, err)
	}

	var authority *authoringcore.WorkAuthorization
	if err := json.Unmarshal(record.Metadata["authoring_authorization"], &authority); err != nil || !reflect.DeepEqual(authority, result.Authorization) {
		t.Fatal("canonical authority lost", authority, err)
	}
}

func telegramAuthoringWorkCheckReplay(t *testing.T, operations service, prepared PreparedAuthoringWork, original AuthoringWorkResult) {
	t.Helper()

	if _, err := saveObservedDump(t, operations, DumpSaveRequest{RelativePath: prepared.Scope.DescriptionRef, Content: "Independent later context"}); err != nil {
		t.Fatal(err)
	}

	title := "Independent later title"

	patch, err := operations.PrepareRecordMutation(t.Context(), "original", workflowstate.MutationBinding{RequestID: "later-display", Operator: "maintainer"}, RecordMutationRequest{Kind: RecordTicket, Action: RecordUpdate, Identity: original.Result.Record.Identity, ExpectedRevision: original.Result.Record.Revision, ExpectedHash: original.Result.Record.SourceHash, Title: &title})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApplyPreparedRecordMutation(t.Context(), patch); err != nil {
		t.Fatal(err)
	}

	again, err := operations.ApplyPreparedAuthoringWork(t.Context(), prepared)
	if err != nil || !reflect.DeepEqual(again, original) {
		t.Fatal("original owner receipt did not reconcile", again, original, err)
	}

	page, err := operations.ReadRecords(t.Context(), "original", RecordQuery{Kind: RecordTicket})
	if err != nil || len(page.Records) != 1 || page.Records[0].Title != title {
		t.Fatal("lost acknowledgement created another description", page, err)
	}
}

// TestTelegramAuthoringWorkGuards rejects stale sources and modified prepared
// descriptions, while an explanation request cannot prepare a native record.
func TestTelegramAuthoringWorkGuards(t *testing.T) {
	tests := []struct {
		name   string
		change func(*testing.T, service, *PreparedAuthoringWork)
	}{
		{"source", func(t *testing.T, s service, p *PreparedAuthoringWork) {
			if _, err := saveObservedDump(t, s, DumpSaveRequest{RelativePath: p.Scope.DescriptionRef, Content: "Changed requirements"}); err != nil {
				t.Fatal(err)
			}
		}},
		{"payload", func(_ *testing.T, _ service, p *PreparedAuthoringWork) {
			p.Mutation.Native.Record.Body = []byte("Wider unapproved work")
		}},
		{"instruction", func(_ *testing.T, _ service, p *PreparedAuthoringWork) {
			p.Scope.Instruction = "Also replace scheduling"
		}},
		{"project", func(_ *testing.T, _ service, p *PreparedAuthoringWork) { p.Mutation.ProjectID = "other" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations, created := telegramAuthoringWorkFixture(t)

			prepared, err := operations.PrepareAuthoringWork(t.Context(), telegramAuthoringWorkRequest(created, authoringstate.IntentPrepare))
			if err != nil {
				t.Fatal(err)
			}

			tt.change(t, operations, &prepared)

			if _, err := operations.ApplyPreparedAuthoringWork(t.Context(), prepared); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("changed prepared work accepted", err)
			}

			page, err := operations.ReadRecords(t.Context(), "original", RecordQuery{Kind: RecordTicket})
			if err != nil || len(page.Records) != 0 {
				t.Fatal("guard rejection wrote a record", page, err)
			}

			telegramAuthoringWorkNoAdmission(t, operations)
		})
	}

	operations, created := telegramAuthoringWorkFixture(t)
	if _, err := operations.PrepareAuthoringWork(t.Context(), telegramAuthoringWorkRequest(created, authoringstate.IntentExplain)); err == nil {
		t.Fatal("explanation prepared native work")
	}
}

// TestTelegramAuthoringWorkFailedBinding preserves a failed dispatch's original
// arguments even when a caller observes fresh sources and prepares another key.
func TestTelegramAuthoringWorkFailedBinding(t *testing.T) {
	operations, created := telegramAuthoringWorkFixture(t)
	request := telegramAuthoringWorkRequest(created, authoringstate.IntentPrepare)

	prepared, err := operations.PrepareAuthoringWork(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := saveObservedDump(t, operations, DumpSaveRequest{RelativePath: prepared.Scope.DescriptionRef, Content: "Changed canonical description"}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApplyPreparedAuthoringWork(t.Context(), prepared); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale source accepted", err)
	}

	current, err := operations.ObserveAuthoring(t.Context(), request.SessionID)
	if err != nil {
		t.Fatal(err)
	}

	request.ExpectedRevision, request.ExpectedSourceHash = current.Session.Revision, current.SourceHash

	fresh, err := operations.PrepareAuthoringWork(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApplyPreparedAuthoringWork(t.Context(), fresh); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("failed request exchanged its arguments", err)
	}

	telegramAuthoringWorkNoAdmission(t, operations)
}

func telegramAuthoringWorkNoAdmission(t *testing.T, operations service) {
	t.Helper()

	executions, err := core.NewExecutionRepository(t.Context(), operations.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	runs, readErr := executions.ListProjectExecutions(t.Context(), operations.projectContext.ProjectID)
	if err := executions.Close(); err != nil {
		t.Fatal(err)
	}

	if readErr != nil || len(runs) != 0 {
		t.Fatalf("preparation created executions=%+v err=%v", runs, readErr)
	}

	store, err := operations.scopedRecordStore("original")
	if err != nil {
		t.Fatal(err)
	}

	invocations, _, err := store.InspectWorkflowPage(t.Context(), 0, 20)
	if err != nil || len(invocations) != 0 {
		t.Fatalf("preparation admitted invocations=%+v err=%v", invocations, err)
	}
}

// TestTelegramAuthoringExplanation uses the configured HTTP provider with real
// canonical sources and leaves the session and invocation stores unchanged.
func TestTelegramAuthoringExplanation(t *testing.T) {
	var reply atomic.Pointer[string]

	initial := telegramAuthoringServiceOutput(t, "Retries remain bounded.", "")
	reply.Store(&initial)

	operations, calls := telegramAuthoringServiceFixture(t, func(input authoringModelInput) string {
		if len(input.Sources) == 0 || !strings.Contains(input.Sources[0].Content, "retry") {
			t.Error("provider did not receive canonical context")
		}

		return *reply.Load()
	})

	created, err := operations.ContinueAuthoring(t.Context(), telegramAuthoringServiceRequest())
	if err != nil {
		t.Fatal(err)
	}

	explanation := `{"explanation":"Retries retain their configured ceiling; preparation starts no work."}`
	reply.Store(&explanation)

	result, err := operations.ExplainAuthoring(t.Context(), AuthoringExplanationRequest{ID: "explain", Question: "How are retries bounded?", SessionID: created.Session.ID})
	if err != nil || !strings.Contains(result.Text, "ceiling") || len(result.SourceRefs) != 2 || calls.Load() != 2 {
		t.Fatal("explanation", result, err)
	}

	after, err := operations.ObserveAuthoring(t.Context(), created.Session.ID)
	if err != nil || after.SourceHash != created.SourceHash {
		t.Fatal("explanation changed canonical authoring", after, err)
	}

	telegramAuthoringServiceNoWork(t, operations)
	telegramAuthoringExplanationInvalid(t, operations, created.Session.ID, &reply)
}

func telegramAuthoringExplanationInvalid(t *testing.T, operations service, sessionID string, reply *atomic.Pointer[string]) {
	t.Helper()

	for _, invalid := range []string{`{"explanation":null}`, `{"explanation":"answer","tool":"execute"}`, `{"explanation":""}`} {
		reply.Store(&invalid)

		if _, err := operations.ExplainAuthoring(t.Context(), AuthoringExplanationRequest{ID: "invalid-explain", Question: "Explain", SessionID: sessionID}); err == nil {
			t.Fatal("invalid explanation accepted", invalid)
		}
	}
}

// TestTelegramAuthoringExplanationOverview bounds project context and rejects a
// README symlink instead of following it outside the registered root.
func TestTelegramAuthoringExplanationOverview(t *testing.T) {
	operations, _ := telegramAuthoringServiceFixture(t, func(input authoringModelInput) string {
		if len(input.Sources) != 1 || input.Sources[0].Reference != "README.md" {
			t.Error("missing project overview")
		}

		return `{"explanation":"The project retains retry limits."}`
	})

	readme := filepath.Join(operations.projectContext.ProjectDir, "README.md")
	if err := os.WriteFile(readme, []byte("# Project\n\nBounded retry behavior.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	request := AuthoringExplanationRequest{ID: "overview", Question: "Explain the project"}
	if _, err := operations.ExplainAuthoring(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(readme); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(t.TempDir(), "private.md"), readme); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ExplainAuthoring(t.Context(), request); err == nil {
		t.Fatal("symlinked overview accepted")
	}

	telegramAuthoringServiceNoWork(t, operations)
}
