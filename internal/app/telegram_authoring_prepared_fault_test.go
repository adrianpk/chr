package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramPreparedWorkLostAck struct {
	service
	lost *atomic.Bool
}

func (s telegramPreparedWorkLostAck) ApplyTelegramAuthoringWork(ctx context.Context, target telegramcore.OperationTarget, prepared PreparedAuthoringWork, guard authoringcore.AuthoringCommitGuard) (AuthoringWorkResult, error) {
	result, err := s.service.ApplyTelegramAuthoringWork(ctx, target, prepared, guard)
	if err == nil && s.lost.CompareAndSwap(false, true) {
		return result, errTelegramAuthoringLostAck
	}

	return result, err
}

// TestTelegramAuthoringPreparedLostAck reconciles the native receipt after a
// successful write loses acknowledgement, for both initial authority and a bound decision.
func TestTelegramAuthoringPreparedLostAck(t *testing.T) {
	for _, mode := range []string{"initial", "decision"} {
		t.Run(mode, func(t *testing.T) {
			telegramAuthoringPreparedRecover(t, mode)
		})
	}
}

func telegramAuthoringPreparedRecover(t *testing.T, mode string) {
	t.Helper()

	start := "/develop retry ceiling"
	if mode == "decision" {
		start = "Develop bounded retries."
	}

	f, id, calls := telegramAuthoringPreparedFixture(t, start)

	command := "/prepare-work " + id
	if mode == "decision" {
		action, _ := telegramAuthoringPreparedOffer(t, f, id)
		command = "/answer " + action.Token + " develop"
	}

	operations := telegramOperationConcreteService(t, f.receiver)
	f.receiver.operations = telegramPreparedWorkLostAck{service: operations, lost: &atomic.Bool{}}

	input := telegramOperationRetainInput(t, f, command)
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramAuthoringLostAck) {
		t.Fatal("owner acknowledgement was not interrupted", err)
	}

	f.receiver.operations = operations
	telegramAuthoringPreparedIndependentEdit(t, f, id)
	telegramAuthoringReopen(t, f)

	f.receiver.operations = operations
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 || len(f.notices) != 1 {
		t.Fatal("recovery repeated generation/delivery", calls.Load(), len(f.notices))
	}

	telegramAuthoringPreparedCheck(t, f, id, true, 1)
	telegramAuthoringPreparedCheckEdit(t, f)
}

func telegramAuthoringPreparedIndependentEdit(t *testing.T, f *telegramConversationAppFixture, id string) {
	t.Helper()

	operations, records := telegramAuthoringPreparedStore(t, f)
	if len(records) != 1 {
		t.Fatal("native owner did not commit before interruption")
	}

	request := telegramAuthoringPreparedLatest(t, f, id)

	prepared, err := telegramRetainedPreparedWork(request)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := saveObservedDump(t, operations, DumpSaveRequest{RelativePath: prepared.Scope.DescriptionRef, Content: "Independent later source edit"}); err != nil {
		t.Fatal(err)
	}

	title := "Independent native title"

	patch, err := operations.PrepareRecordMutation(t.Context(), "original", workflowstate.MutationBinding{RequestID: "independent-ticket-edit", Operator: request.Operator}, RecordMutationRequest{Kind: RecordTicket, Action: RecordUpdate, Identity: nativeRecordView(records[0], RecordTicket).Identity, ExpectedRevision: records[0].Revision, ExpectedHash: records[0].SourceHash, Title: &title})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ApplyPreparedRecordMutation(t.Context(), patch); err != nil {
		t.Fatal(err)
	}
}

func telegramAuthoringPreparedCheckEdit(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	operations, records := telegramAuthoringPreparedStore(t, f)
	if nativeRecordView(records[0], RecordTicket).Title != "Independent native title" {
		t.Fatal("replay replaced later native edit")
	}

	var scope authoringcore.AuthoringWorkScope
	if err := json.Unmarshal(records[0].Metadata["authoring_scope"], &scope); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(operations.projectContext.ProjectDir, scope.DescriptionRef))
	if err != nil || !strings.Contains(string(body), "Independent later source edit") {
		t.Fatal("replay replaced later source edit", err)
	}
}

type telegramPreparedWorkChange struct {
	service
	change func()
}

func (s telegramPreparedWorkChange) ApplyTelegramAuthoringWork(ctx context.Context, target telegramcore.OperationTarget, prepared PreparedAuthoringWork, guard authoringcore.AuthoringCommitGuard) (AuthoringWorkResult, error) {
	s.change()

	return s.service.ApplyTelegramAuthoringWork(ctx, target, prepared, guard)
}

// TestTelegramAuthoringPreparedAuthority rejects root/operator changes between
// native preparation and dispatch, before any canonical record effect.
func TestTelegramAuthoringPreparedAuthority(t *testing.T) {
	for _, change := range []string{"operator", "root"} {
		t.Run(change, func(t *testing.T) {
			f, id, _ := telegramAuthoringPreparedFixture(t, "/develop retry ceiling")
			operations := telegramOperationConcreteService(t, f.receiver)
			scoped, _ := telegramAuthoringPreparedStore(t, f)
			f.receiver.operations = telegramPreparedWorkChange{service: operations, change: func() { telegramAuthoringChangeAuthority(t, f, change) }}
			f.send(t, "/prepare-work "+id, 0)

			store, err := scoped.scopedRecordStore("original")
			if err != nil {
				t.Fatal(err)
			}

			snapshot, err := store.Snapshot(t.Context())
			if err != nil || len(snapshot.Records) != 0 {
				t.Fatal("changed authority created native work", snapshot, err)
			}

			telegramAuthoringWorkNoAdmission(t, scoped)
		})
	}
}
