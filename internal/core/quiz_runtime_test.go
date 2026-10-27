package core

import (
	"fmt"
	"testing"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func TestValidateQuizRuntimeStateAcceptsProjectWideState(t *testing.T) {
	state := validQuizRuntimeState()

	err := quizcore.ValidateRuntimeState(state)
	if err != nil {
		t.Fatalf("quizcore.ValidateQuizRuntimeState returned error: %v", err)
	}
}

func TestValidateQuizRuntimeStateAcceptsUseCaseScopedState(t *testing.T) {
	state := validQuizRuntimeState()
	state.ID = "quiz-blog"
	state.CaseSlug = "blog"
	state.RawPath = "ops/default/quiz/00000-blog-quiz-raw/artifact.md"
	state.CuratedPath = "ops/default/quiz/00000-blog-quiz/artifact.md"

	err := quizcore.ValidateRuntimeState(state)
	if err != nil {
		t.Fatalf("quizcore.ValidateQuizRuntimeState returned error: %v", err)
	}
}

func TestValidateQuizRuntimeStateRejectsInvalidState(t *testing.T) {
	for _, test := range invalidQuizRuntimeStateTests() {
		t.Run(test.name, func(t *testing.T) {
			err := quizcore.ValidateRuntimeState(test.state)
			if err == nil {
				t.Fatal("quizcore.ValidateQuizRuntimeState returned nil error")
			}
		})
	}
}

func TestValidateQuizQuestionRejectsInvalidDepth(t *testing.T) {
	for _, depth := range []int{0, 4} {
		t.Run(fmt.Sprintf("depth-%d", depth), func(t *testing.T) {
			question := validQuizQuestion()
			question.Depth = depth

			err := quizcore.ValidateQuestion(question)
			if err == nil {
				t.Fatal("quizcore.ValidateQuizQuestion returned nil error")
			}
		})
	}
}

func TestValidateQuizEntryRejectsInvalidAnswerState(t *testing.T) {
	for _, test := range invalidQuizEntryTests() {
		t.Run(test.name, func(t *testing.T) {
			err := quizcore.ValidateEntry(test.entry)
			if err == nil {
				t.Fatal("quizcore.ValidateQuizEntry returned nil error")
			}
		})
	}
}

func invalidQuizRuntimeStateTests() []struct {
	name  string
	state quizcore.RuntimeState
} {
	return []struct {
		name  string
		state quizcore.RuntimeState
	}{
		{name: "status", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.Status = quizcore.RuntimeStatus("unknown")
		})},
		{name: "mode", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.Mode = quizcore.Mode("unknown")
		})},
		{name: "budget", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.QuestionBudget = 0
		})},
		{name: "budget overflow", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.QuestionBudget = 1
			state.QuestionsAsked = 2
		})},
		{name: "outcome overflow", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.QuestionsAsked = 1
			state.AnswersRecorded = 1
			state.SkipsRecorded = 1
		})},
		{name: "active question missing", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.ActiveQuestion = nil
		})},
		{name: "inactive active question", state: quizRuntimeStateWith(func(state *quizcore.RuntimeState) {
			state.Status = quizcore.RuntimeStatusDone
		})},
	}
}

func invalidQuizEntryTests() []struct {
	name  string
	entry quizcore.Entry
} {
	return []struct {
		name  string
		entry quizcore.Entry
	}{
		{name: "answered empty answer", entry: quizEntryWith(func(entry *quizcore.Entry) {
			entry.Answer = ""
		})},
		{name: "skipped answer", entry: quizEntryWith(func(entry *quizcore.Entry) {
			entry.Status = quizcore.EntryStatusSkipped
			entry.Answer = "This should not be present."
		})},
		{name: "invalid mode", entry: quizEntryWith(func(entry *quizcore.Entry) {
			entry.Mode = quizcore.Mode("unknown")
		})},
		{name: "invalid status", entry: quizEntryWith(func(entry *quizcore.Entry) {
			entry.Status = quizcore.EntryStatus("unknown")
		})},
		{name: "invalid depth", entry: quizEntryWith(func(entry *quizcore.Entry) {
			entry.Depth = 4
		})},
	}
}

func validQuizRuntimeState() quizcore.RuntimeState {
	now := quizRuntimeTestTime()
	question := validQuizQuestion()

	return quizcore.RuntimeState{
		ID:              "quiz-project",
		ProjectID:       "project-1",
		Status:          quizcore.RuntimeStatusActive,
		Mode:            quizcore.ModeDeep,
		QuestionBudget:  quizcore.DefaultQuestionBudget,
		QuestionsAsked:  1,
		AnswersRecorded: 0,
		SkipsRecorded:   0,
		ActiveTopic:     "workflow",
		Coverage: []quizcore.CoverageTopic{
			{
				ID:            "workflow",
				Label:         "Workflow",
				Status:        quizcore.CoverageStatusActive,
				QuestionCount: 1,
			},
		},
		ActiveQuestion: &question,
		Trail: []quizcore.TrailNode{
			{
				QuestionID: question.ID,
				Prompt:     question.Prompt,
				Topic:      question.Topic,
				Depth:      question.Depth,
			},
		},
		Candidates: []quizcore.Candidate{
			validQuizCandidate(),
		},
		RawPath:     "ops/default/quiz/00000-project-quiz-raw/artifact.md",
		CuratedPath: "ops/default/quiz/00000-project-quiz/artifact.md",
		UpdatedAt:   now,
	}
}

func quizRuntimeStateWith(update func(*quizcore.RuntimeState)) quizcore.RuntimeState {
	state := validQuizRuntimeState()
	update(&state)

	return state
}

func validQuizQuestion() quizcore.Question {
	now := quizRuntimeTestTime()

	return quizcore.Question{
		ID:        "question-1",
		Prompt:    "What should this workflow preserve?",
		Required:  true,
		Topic:     "workflow",
		Depth:     1,
		TrunkID:   "question-1",
		Rationale: "The runtime needs the primary workflow invariant.",
		CreatedAt: now,
	}
}

func validQuizEntry() quizcore.Entry {
	now := quizRuntimeTestTime()

	return quizcore.Entry{
		ID:           "entry-1",
		QuestionID:   "question-1",
		Question:     "What should this workflow preserve?",
		Answer:       "It should preserve one visible question at a time.",
		Status:       quizcore.EntryStatusAnswered,
		Topic:        "workflow",
		Depth:        1,
		Mode:         quizcore.ModeDeep,
		Tags:         []string{"workflow"},
		PromptIntent: `{"kind":"quiz_question","metadata":{"follow_up_depth":0}}`,
		CreatedAt:    now,
		AnsweredAt:   now.Add(time.Minute),
	}
}

func quizEntryWith(update func(*quizcore.Entry)) quizcore.Entry {
	entry := validQuizEntry()
	update(&entry)

	return entry
}

func validQuizCandidate() quizcore.Candidate {
	now := quizRuntimeTestTime()

	return quizcore.Candidate{
		ID:        "candidate-1",
		Prompt:    "Which follow-up would clarify the workflow?",
		Topic:     "workflow",
		Depth:     1,
		RouteID:   "workflow-follow-up",
		Hints:     []string{"workflow"},
		CreatedAt: now,
	}
}

func quizRuntimeTestTime() time.Time {
	return time.Date(2026, 5, 24, 15, 0, 0, 0, time.UTC)
}
