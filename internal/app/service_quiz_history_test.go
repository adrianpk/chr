package app

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Case discovery inherits project decisions while retaining its own pending
// answer; the project history is loaded from real persisted quiz artifacts.
func TestServiceQuizInheritedHistory(t *testing.T) {
	operations, model := newRetainedQuizFixture(t)
	final := serviceQuizModelResponse("final", "final", serviceQuizClock()())
	final.Question.Required = false
	model.responses = []core.QuizModelResponse{serviceQuizModelResponse("first", "first", serviceQuizClock()()), final}

	if _, err := operations.StartQuiz(QuizStartRequest{}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "GENERAL_DECISION"}); err != nil {
		t.Fatal(err)
	}

	entries := []quizcore.Entry{{Question: "Case constraint?", Answer: "CASE_DECISION", Status: quizcore.EntryStatusAnswered}}

	section, err := operations.quizHistoryContext(operations.quizScope(quizcore.Scope{CaseSlug: "blog"}), entries)
	if err != nil || !strings.Contains(section.Content, "GENERAL_DECISION") || !strings.Contains(section.Content, "CASE_DECISION") {
		t.Fatal("case discovery lost prior decisions", section, err)
	}
}

// History bounds reject oversized context explicitly instead of silently
// forgetting old answers and asking the same requirements again.
func TestServiceQuizHistoryBounds(t *testing.T) {
	operations, _ := newRetainedQuizFixture(t)
	for _, name := range []string{"entries", "bytes"} {
		t.Run(name, func(t *testing.T) {
			entries := []quizcore.Entry{{Question: "Retained decision?", Answer: strings.Repeat("x", quizHistoryByteLimit), Status: quizcore.EntryStatusAnswered}}
			if name == "entries" {
				entries = make([]quizcore.Entry, quizHistoryEntryLimit+1)
			}

			_, err := operations.quizHistoryContext(operations.quizScope(quizcore.Scope{}), entries)
			if err == nil || !strings.Contains(err.Error(), "quiz discovery history exceeds") {
				t.Fatal("oversized history was truncated or accepted", err)
			}
		})
	}
}
