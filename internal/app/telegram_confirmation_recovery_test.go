package app

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramPresentationControlQuestions keeps a short question-led view while
// exporting every original option, question and candidate binding without cuts.
func TestTelegramPresentationControlQuestions(t *testing.T) {
	questions := []workflowstate.MergerDecision{
		{ID: "question-one", Issue: "Which option should be used? " + strings.Repeat("🚀 literal `value` ", 100), Options: []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}, Authority: "maintainer"},
		{ID: "question-two", Issue: "Which target should be used?", Options: []string{"target one", "target two"}, Authority: "maintainer"},
	}
	status := WorkflowWorkStatus{Identity: workflowstate.Identity{ID: "execution-one", ProjectID: "example-project", Namespace: "default"}, Integration: &WorkflowIntegrationStatus{EntryID: "integration-one", EntryRevision: 3, CandidateHash: strings.Repeat("c", 64), Decision: &WorkflowIntegrationDecisionStatus{ID: "decision-one", Revision: 2, Hash: strings.Repeat("d", 64), State: "waiting", Requests: questions}}}
	target := telegramcore.OperationTarget{ProjectID: "example-project", Namespace: "default", ProjectBindingDigest: strings.Repeat("a", 64)}

	response, err := telegramQuestionDocument(target, status)
	if err != nil {
		t.Fatal(err)
	}

	view := tgmessage.RenderDocument(response.Message)
	if !strings.HasPrefix(view.PlainFallback, "Which option should be used?") || len(utf16.Encode([]rune(view.PlainFallback))) > 900 {
		t.Fatal("question was replaced by metadata or exceeded the caption bound", view)
	}

	var exported WorkflowIntegrationStatus
	if err := json.Unmarshal(response.Document.Payload, &exported); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exported.Decision.Requests, questions) || exported.CandidateHash != status.Integration.CandidateHash || exported.EntryRevision != 3 {
		t.Fatal("short view discarded complete questions or their binding", exported)
	}
}

// This narrow fault exposes the real owner's bounded first page as incomplete.
// Source reads, policy, persistence and admission continue to use the real service.
type telegramPartialWorkPreview struct{ service }

func (s telegramPartialWorkPreview) PreviewTelegramWork(ctx context.Context, target telegramcore.OperationTarget, selector workflowstate.Selector) (WorkflowBoundPreview, error) {
	preview, err := s.service.PreviewTelegramWork(ctx, target, selector)
	if err == nil {
		preview.View.Preview.Total = len(preview.View.Preview.Members) + 1
	}

	return preview, err
}

// TestTelegramPresentationConfirmationIncomplete refuses a first-page-only
// scope before sending a keyboard or admitting any native work.
func TestTelegramPresentationConfirmationIncomplete(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)
	f.receiver.operations = telegramPartialWorkPreview{service: operations}

	notice := f.send(t, "/work ticket "+record.Task.Key, 0)
	if !strings.Contains(notice.Text, "complete confirmation scope is unavailable") || !strings.Contains(notice.Text, "No action was offered") {
		t.Fatal("incomplete selection exposed confirmation authority", notice)
	}

	checkTelegramWorkRunCount(t, operations, 0)

	select {
	case upload := <-f.documents:
		t.Fatal("incomplete scope was uploaded", upload.Filename)
	default:
	}
}
