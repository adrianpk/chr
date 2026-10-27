package app

import (
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramOperationTestService(t *testing.T) service {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git record-owner integration is selected explicitly")
	}

	project := newTestWorkspace(t)
	if output, err := exec.Command("git", "init", "-q", project).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}

	writeCLIWorkspaceConfig(t, project, "tickets:\n  key_prefix: PRJ\n")

	return newTestService(t, project)
}

func telegramRecordText(value string) *string { return &value }

func telegramApplyRecord(t *testing.T, s service, id string, request RecordMutationRequest) RecordMutationResult {
	t.Helper()

	prepared, err := s.PrepareRecordMutation(t.Context(), "default", workflowstate.MutationBinding{RequestID: id, Operator: "maintainer"}, request)
	if err != nil {
		t.Fatal(err)
	}
	// Persisted arguments, rather than a newly observed record, drive resumed dispatch.
	encoded, err := json.Marshal(prepared)
	if err != nil {
		t.Fatal(err)
	}

	var retained PreparedRecordMutation
	if err := json.Unmarshal(encoded, &retained); err != nil {
		t.Fatal(err)
	}

	result, err := s.ApplyPreparedRecordMutation(t.Context(), retained)
	if err != nil {
		t.Fatal(err)
	}

	again, err := s.ApplyPreparedRecordMutation(t.Context(), retained)
	if err != nil || !reflect.DeepEqual(again, result) {
		t.Fatal("prepared record replay changed result", again, result, err)
	}

	return result
}

// TestTelegramOperationTicketRecords creates, observes and updates through one retained owner request.
func TestTelegramOperationTicketRecords(t *testing.T) {
	operations := telegramOperationTestService(t)

	first := telegramApplyRecord(t, operations, "create-ticket", RecordMutationRequest{Kind: RecordTicket, Action: "create", Title: telegramRecordText("A task"), Synopsis: telegramRecordText("A useful outcome"), Body: telegramRecordText("Retain these requirements")})
	if first.Record.Identity != "PRJ-1" || first.Record.Phase != "pending" {
		t.Fatal("ticket creation changed native identity or lifecycle", first)
	}

	updated := telegramApplyRecord(t, operations, "update-ticket", RecordMutationRequest{Kind: RecordTicket, Action: "update", Identity: first.Record.Identity, ExpectedRevision: first.Record.Revision, ExpectedHash: first.Record.SourceHash, Title: telegramRecordText("Updated title")})

	page, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket, Identity: first.Record.Identity})
	if err != nil || len(page.Records) != 1 || page.Records[0].Title != "Updated title" || updated.Record.Revision != first.Record.Revision+1 {
		t.Fatal("ticket update was not observable", page, updated, err)
	}

	if !strings.Contains(page.Records[0].Body, "Retain these requirements") || !strings.Contains(page.Records[0].Body, `title: "Updated title"`) {
		t.Fatal("metadata update lost requirements")
	}

	_, err = operations.PrepareRecordMutation(t.Context(), "default", workflowstate.MutationBinding{RequestID: "stale", Operator: "maintainer"}, RecordMutationRequest{Kind: RecordTicket, Action: "update", Identity: first.Record.Identity, ExpectedRevision: first.Record.Revision, ExpectedHash: first.Record.SourceHash, Title: telegramRecordText("stale")})
	if !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale ticket observation accepted", err)
	}
}

func telegramCreateRecordSet(t *testing.T, operations service) RecordMutationResult {
	t.Helper()

	body := "# Existing plan\n\nApproved contract reference.\n"
	writeServiceWorkspaceFile(t, filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "plan", "source.md"), body)

	validation := workflowstate.Validation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	return telegramApplyRecord(t, operations, "create-set", RecordMutationRequest{Kind: RecordSliceSet, Action: "create", Title: telegramRecordText("Prepared set"), PlanRef: "plan/source.md", PlanHash: recordDigest([]byte(body)), Validation: &validation})
}

// TestTelegramOperationSliceRecords preserves parent revision authority and pending proposed-record lifecycle.
func TestTelegramOperationSliceRecords(t *testing.T) {
	operations := telegramOperationTestService(t)

	set := telegramCreateRecordSet(t, operations)
	if set.Record.Phase != "proposed" {
		t.Fatal("record creation admitted a set", set)
	}

	contract := `{"Contract":"A bounded slice","Validation":{"required":true,"commands":[["go","test","./..."]],"timeout_seconds":60}}`
	writeServiceWorkspaceFile(t, filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "plan", "slice.json"), contract)

	descriptor := workflowstate.WorkflowSlice{ID: "slice-1", Order: 1, Title: "First slice", ContractRef: "plan/slice.json", ContractHash: recordDigest([]byte(contract))}

	created := telegramApplyRecord(t, operations, "create-slice", RecordMutationRequest{Kind: RecordSlice, Action: "create", ParentID: set.Record.Identity, ExpectedRevision: set.Record.Revision, ExpectedHash: set.Record.SourceHash, Slice: &descriptor})
	if created.Record.Identity != "slice-1" || created.Record.ParentID != set.Record.Identity || created.Record.Phase != "pending" {
		t.Fatal("slice target was not retained", created)
	}

	updated := telegramApplyRecord(t, operations, "update-slice", RecordMutationRequest{Kind: RecordSlice, Action: "update", Identity: "slice-1", ParentID: set.Record.Identity, ExpectedRevision: created.Record.Revision, ExpectedHash: created.Record.SourceHash, Title: telegramRecordText("Updated slice")})
	if updated.Record.Title != "Updated slice" {
		t.Fatal("slice display update missing", updated)
	}

	final := telegramApplyRecord(t, operations, "update-set", RecordMutationRequest{Kind: RecordSliceSet, Action: "update", Identity: set.Record.Identity, ExpectedRevision: updated.Record.Revision, ExpectedHash: updated.Record.SourceHash, Title: telegramRecordText("Updated set")})
	if final.Record.Title != "Updated set" {
		t.Fatal("set display update missing", final)
	}
}

// TestTelegramOperationCaseRecords uses the case metadata owner and keeps its namespace binding.
func TestTelegramOperationCaseRecords(t *testing.T) {
	operations := telegramOperationTestService(t)

	prepared, err := operations.PrepareRecordMutation(t.Context(), "team", workflowstate.MutationBinding{RequestID: "case", Operator: "maintainer"}, RecordMutationRequest{Kind: RecordUseCase, Action: "create", Title: telegramRecordText("Example case"), Synopsis: telegramRecordText("A description"), Context: telegramRecordText("Initial context")})
	if err != nil || prepared.Case.Namespace != "team" {
		t.Fatal("case namespace was retargeted", prepared, err)
	}

	result, err := operations.ApplyPreparedRecordMutation(t.Context(), prepared)
	if err != nil {
		t.Fatal(err)
	}

	updated := telegramApplyRecord(t, operations, "update-case", RecordMutationRequest{Kind: RecordUseCase, Action: "update", Identity: result.Record.Identity, ExpectedHash: result.Record.SourceHash, Synopsis: telegramRecordText("Updated description")})
	if !strings.Contains(updated.Record.Body, "Updated description") || !strings.Contains(updated.Record.Body, "Initial context") {
		t.Fatal("case update lost context", updated)
	}

	page, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordUseCase})
	if err != nil || len(page.Records) != 1 {
		t.Fatal("case list lost native metadata", page, err)
	}
}

// TestTelegramOperationRecordGuards rejects forged lifecycle input and changed plan contracts.
func TestTelegramOperationRecordGuards(t *testing.T) {
	operations := telegramOperationTestService(t)
	for _, limit := range []int{-1, 51} {
		if _, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket, Limit: limit}); err == nil {
			t.Fatal("invalid list limit accepted", limit)
		}
	}

	request := RecordMutationRequest{Kind: RecordTicket, Action: "pause", Title: telegramRecordText("forged")}
	if _, err := operations.PrepareRecordMutation(t.Context(), "default", workflowstate.MutationBinding{RequestID: "unsupported", Operator: "maintainer"}, request); !errors.Is(err, ErrRecordUnsupported) {
		t.Fatal("unsupported lifecycle action accepted", err)
	}

	set := telegramCreateRecordSet(t, operations)
	contract := `{"Contract":"original"}`
	writeServiceWorkspaceFile(t, filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "plan", "contract.json"), contract)

	descriptor := workflowstate.WorkflowSlice{ID: "slice", Order: 1, ContractRef: "plan/contract.json", ContractHash: recordDigest([]byte(contract))}

	prepared, err := operations.PrepareRecordMutation(t.Context(), "default", workflowstate.MutationBinding{RequestID: "changed-contract", Operator: "maintainer"}, RecordMutationRequest{Kind: RecordSlice, Action: "create", ParentID: set.Record.Identity, ExpectedRevision: set.Record.Revision, ExpectedHash: set.Record.SourceHash, Slice: &descriptor})
	if err != nil {
		t.Fatal(err)
	}

	writeServiceWorkspaceFile(t, filepath.Join(operations.projectContext.ProjectDir, "ops", "default", "plan", "contract.json"), "changed")

	if _, err := operations.ApplyPreparedRecordMutation(t.Context(), prepared); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("changed contract accepted", err)
	}
}

// TestTelegramOperationRecordPages returns bounded deterministic pages without changing the observed records.
func TestTelegramOperationRecordPages(t *testing.T) {
	operations := telegramOperationTestService(t)
	for index, title := range []string{"First", "Second"} {
		telegramApplyRecord(t, operations, string(rune('a'+index)), RecordMutationRequest{Kind: RecordTicket, Action: "create", Title: telegramRecordText(title), Synopsis: telegramRecordText("A bounded result")})
	}

	first, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket, Limit: 1})
	if err != nil || len(first.Records) != 1 || !first.HasMore || first.NextOffset != 1 {
		t.Fatal("invalid first page", first, err)
	}

	checkTelegramOperationNextPage(t, operations, first)
}

func checkTelegramOperationNextPage(t *testing.T, operations service, first RecordPage) {
	t.Helper()

	second, err := operations.ReadRecords(t.Context(), "default", RecordQuery{Kind: RecordTicket, Offset: first.NextOffset, Limit: 1})
	if err != nil || len(second.Records) != 1 || second.HasMore || first.Records[0].Identity == second.Records[0].Identity {
		t.Fatal("invalid next page", second, err)
	}

	if first.Records[0].Body != "" || second.Records[0].Body != "" {
		t.Fatal("list expanded full record bodies")
	}
}
