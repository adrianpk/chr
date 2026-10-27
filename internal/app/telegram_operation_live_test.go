package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func telegramOperationAppFixture(t *testing.T) *telegramConversationAppFixture {
	t.Helper()

	if testing.Short() {
		t.Skip("real Git receiver-to-owner integration is selected explicitly")
	}

	f := newTelegramConversationAppFixture(t)

	projects, err := f.receiver.operations.ListRegisteredProjects([]string{"example-project"})
	if err != nil || len(projects) != 1 {
		t.Fatal("project fixture unavailable", projects, err)
	}

	telegramOperationInitProject(t, projects[0].Path)

	return f
}

// TestTelegramOperationLiveRecords carries typed commands through real HTTP, SQLite and shared record owners.
func TestTelegramOperationLiveRecords(t *testing.T) {
	f := telegramOperationAppFixture(t)

	notice := f.send(t, `/tickets create {"title":"Remote task","synopsis":"One useful outcome"}`, 0)
	if !strings.Contains(notice.Text, "Record saved") || !strings.Contains(notice.Text, "PRJ-1") {
		t.Fatal("ticket mutation unavailable", notice)
	}

	page := telegramOperationReadPage(t, f, `/tickets show {"identity":"PRJ-1"}`)
	if len(page.Records) != 1 || page.Records[0].Title != "Remote task" {
		t.Fatal("record read lost confirmed result", page)
	}

	first := page.Records[0]

	update := fmt.Sprintf(`/ticket update {"identity":"PRJ-1","revision":%d,"source_hash":%q,"title":"Updated remotely"}`, first.Revision, first.SourceHash)
	if notice := f.send(t, update, 0); !strings.Contains(notice.Text, "Record saved") {
		t.Fatal("guarded update failed", notice)
	}

	if notice := f.send(t, update, 0); !strings.Contains(notice.Text, "Record conflict") {
		t.Fatal("stale revision accepted", notice)
	}

	if notice := f.send(t, `/use-cases create {"title":"Remote case","synopsis":"Retained case"}`, 0); !strings.Contains(notice.Text, "Record saved") {
		t.Fatal("case mutation failed", notice)
	}

	checkTelegramNoWork(t, f.receiver.env)
}

func telegramOperationReadPage(t *testing.T, f *telegramConversationAppFixture, command string) RecordPage {
	t.Helper()

	for len(f.documents) > 0 {
		<-f.documents
	}

	notice := f.send(t, command, 0)
	if strings.Contains(notice.Text, "No records matched") {
		return RecordPage{}
	}

	document := telegramLatestDocument(t, f, "records-")

	var page RecordPage
	if err := json.Unmarshal(document.Payload, &page); err != nil {
		t.Fatal("records-page export unavailable", err)
	}

	return page
}

type telegramLostRecordAck struct {
	service
	lost *bool
}

var errTelegramLostRecordAck = errors.New("injected lost owner acknowledgement")

func (s telegramLostRecordAck) ApplyTelegramRecord(ctx context.Context, target core.TelegramOperationTarget, prepared PreparedRecordMutation) (RecordMutationResult, error) {
	result, err := s.service.ApplyTelegramRecord(ctx, target, prepared)
	if err == nil && !*s.lost {
		*s.lost = true

		return result, errTelegramLostRecordAck
	}

	return result, err
}

// TestTelegramOperationLostAcknowledgement reopens SQLite after a successful write and observes the same owner receipt.
func TestTelegramOperationLostAcknowledgement(t *testing.T) {
	f := telegramOperationAppFixture(t)
	operations := telegramOperationConcreteService(t, f.receiver)
	lost := false

	f.receiver.operations = telegramLostRecordAck{service: operations, lost: &lost}
	f.updates <- json.RawMessage(`[{"update_id":1,"message":{"message_id":1,"from":{"id":100001},"chat":{"id":100001,"type":"private"},"text":"/tickets create {\"title\":\"One task\",\"synopsis\":\"One effect\"}"}}]`)

	if _, err := f.receiver.step(t.Context()); !errors.Is(err, errTelegramLostRecordAck) {
		t.Fatal("owner acknowledgement was not interrupted", err)
	}

	if err := f.receiver.store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := core.NewTelegramReceiverStore(t.Context(), f.receiver.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	f.receiver.store, f.receiver.operations = store, operations

	t.Setenv(core.OpsUserEnv, "changed-after-prepare")

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	checkTelegramOperationRecoveredRecord(t, f)
}

func checkTelegramOperationRecoveredRecord(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	operations := telegramOperationConcreteService(t, f.receiver)
	target := core.TelegramOperationTarget{ProjectID: "example-project", Namespace: "default"}

	page, err := operations.ReadTelegramRecords(t.Context(), target, RecordQuery{Kind: RecordTicket})
	if err != nil || len(page.Records) != 1 || page.Records[0].Identity != "PRJ-1" {
		t.Fatal("lost acknowledgement duplicated or moved artifact", page, err)
	}

	if len(f.notices) != 1 {
		t.Fatal("lost acknowledgement duplicated delivery", len(f.notices))
	}

	notice := <-f.notices
	if !strings.Contains(notice.Text, "Record saved") {
		t.Fatal("owner result was not confirmed", notice)
	}

	pending, err := f.receiver.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 0 {
		t.Fatal("confirmed source remains pending", pending, err)
	}
}

func telegramOperationConcreteService(t *testing.T, receiver telegramReceiver) service {
	t.Helper()

	operations, ok := receiver.operations.(service)
	if !ok {
		t.Fatal("concrete operation service unavailable")
	}

	return operations
}

type telegramMoveRecordRoot struct {
	service
	replacement string
}

func (s telegramMoveRecordRoot) PrepareTelegramRecord(ctx context.Context, target core.TelegramOperationTarget, binding core.WorkflowMutationBinding, request RecordMutationRequest) (PreparedRecordMutation, error) {
	prepared, err := s.service.PrepareTelegramRecord(ctx, target, binding, request)
	if err != nil {
		return prepared, err
	}

	registry, err := projectcore.NewRegistry(ctx, s.env, s.assetsFS)
	if err != nil {
		return prepared, err
	}

	_, moveErr := registry.UpdateProjectPath(ctx, "example-project", s.replacement)

	return prepared, errors.Join(moveErr, registry.Close())
}

// TestTelegramOperationChangedRoot blocks retained input instead of applying it to a replacement registry root.
func TestTelegramOperationChangedRoot(t *testing.T) {
	f := telegramOperationAppFixture(t)
	operations := telegramOperationConcreteService(t, f.receiver)
	replacement := t.TempDir()
	telegramOperationInitProject(t, replacement)

	f.receiver.operations = telegramMoveRecordRoot{service: operations, replacement: replacement}

	notice := f.send(t, `/tickets create {"title":"Bound task","synopsis":"Original root only"}`, 0)
	if !strings.Contains(notice.Text, "Record conflict") {
		t.Fatal("changed registry root was not blocked", notice)
	}

	pending, err := f.receiver.store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 0 {
		t.Fatal("blocked source was not acknowledged", pending, err)
	}

	f.receiver.operations = operations

	page, err := operations.ReadTelegramRecords(t.Context(), core.TelegramOperationTarget{ProjectID: "example-project", Namespace: "default"}, RecordQuery{Kind: RecordTicket})
	if err != nil || len(page.Records) != 0 {
		t.Fatal("replacement root received a mutation", page, err)
	}
}

func telegramOperationInitProject(t *testing.T, root string) {
	t.Helper()

	if output, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}

	writeCLIWorkspaceConfig(t, root, "tickets:\n  key_prefix: PRJ\n")
}
