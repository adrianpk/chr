package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// Discovery retains answers outside the recent-five window, including the
// answer being committed, so a new provider process can avoid resolved topics.
func TestServiceQuizFullHistory(t *testing.T) {
	operations, _ := newRetainedQuizFixture(t)

	state, err := operations.newQuizRuntimeState(operations.quizScope(quizcore.Scope{}), quizcore.ModeDeep)
	if err != nil {
		t.Fatal(err)
	}

	entries := make([]quizcore.Entry, 8)
	for index := range entries {
		entries[index] = quizcore.Entry{Question: fmt.Sprintf("Decision %d?", index), Answer: fmt.Sprintf("Answer %d.", index), Status: quizcore.EntryStatusAnswered}
	}

	request, err := operations.newQuizModelRequestWithEntries(state, entries)
	if err != nil {
		t.Fatal(err)
	}

	for _, section := range request.ContextSections {
		if section.Name != "quiz_history" {
			continue
		}

		var history []quizmodel.EntryPayload
		if err := json.Unmarshal([]byte(section.Content), &history); err != nil {
			t.Fatal(err)
		}

		if len(history) != len(entries) || history[0].Answer != entries[0].Answer || history[7].Answer != entries[7].Answer {
			t.Fatal("discovery omitted confirmed answers", history)
		}

		return
	}

	t.Fatal("quiz request has no complete discovery history")
}

// Provider completion and budget exhaustion stop generation without losing the
// final answer; a repeated resolved question blocks instead of being presented.
func TestServiceQuizNaturalStop(t *testing.T) {
	for _, name := range []string{"recommendation", "budget", "duplicate"} {
		t.Run(name, func(t *testing.T) { checkQuizNaturalStop(t, name) })
	}
}

func checkQuizNaturalStop(t *testing.T, name string) {
	t.Helper()

	operations, model := newRetainedQuizFixture(t)
	first := serviceQuizModelResponse("first", "first", serviceQuizClock()())
	second := serviceQuizModelResponse("second", "second", serviceQuizClock()())

	second.Question.Required = name != "recommendation"
	if name == "duplicate" {
		second.Question.Prompt = "  " + strings.ToUpper(first.Question.Prompt) + "  "
	}

	model.responses = []quizmodel.Response{first, second}

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeDeep})
	if err != nil {
		t.Fatal(err)
	}

	if name == "budget" {
		started.State.QuestionBudget = 1
		if _, err := operations.quizRepository().SaveRuntimeState(t.Context(), started.State); err != nil {
			t.Fatal(err)
		}
	}

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Keep the saved decision."})
	if err != nil {
		t.Fatal(err)
	}

	want := quizcore.RuntimeStatusDone
	if name == "duplicate" {
		want = quizcore.RuntimeStatusBlocked
	}

	checkQuizFinalAnswer(t, operations, answered.State, want)

	if name == "duplicate" {
		checkQuizDuplicateRecovery(t, operations, model, answered.State)
	}
}

func checkQuizDuplicateRecovery(t *testing.T, operations service, model *fakeServiceQuizModel, state quizcore.RuntimeState) {
	t.Helper()

	if state.GenerationOperationID != "" {
		t.Fatal("resume would replay the rejected provider output")
	}

	model.responses = []quizmodel.Response{serviceQuizModelResponse("third", "third", serviceQuizClock()())}

	resumed, err := operations.ResumeQuiz(QuizResumeRequest{})
	if err != nil || resumed.State.ActiveQuestion == nil || resumed.State.AnswersRecorded != 1 || resumed.State.ContinuationPending {
		t.Fatal("new question could not recover rejected continuation", resumed, err)
	}
}

func checkQuizFinalAnswer(t *testing.T, operations service, state quizcore.RuntimeState, want quizcore.RuntimeStatus) {
	t.Helper()

	if state.Status != want || state.ActiveQuestion != nil || state.QuestionsAsked != 1 || state.AnswersRecorded != 1 {
		t.Fatal("quiz did not stop safely", state)
	}

	entries, err := operations.quizRepository().LoadRawCaptureEntries(t.Context(), operations.quizScope(quizcore.Scope{}))
	if err != nil || len(entries) != 1 || entries[0].Answer != "Keep the saved decision." {
		t.Fatal("final answer was lost", entries, err)
	}
}
