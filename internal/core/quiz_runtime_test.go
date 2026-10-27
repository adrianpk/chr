package core

import (
	"fmt"
	"testing"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func TestValidateQuizRuntimeStateAcceptsProjectWideState(t *testing.T) {
	state := validQuizRuntimeState()

	err := quizcore.ValidateQuizRuntimeState(state)
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

	err := quizcore.ValidateQuizRuntimeState(state)
	if err != nil {
		t.Fatalf("quizcore.ValidateQuizRuntimeState returned error: %v", err)
	}
}

func TestValidateQuizRuntimeStateRejectsInvalidState(t *testing.T) {
	for _, test := range invalidQuizRuntimeStateTests() {
		t.Run(test.name, func(t *testing.T) {
			err := quizcore.ValidateQuizRuntimeState(test.state)
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

			err := quizcore.ValidateQuizQuestion(question)
			if err == nil {
				t.Fatal("quizcore.ValidateQuizQuestion returned nil error")
			}
		})
	}
}

func TestValidateQuizEntryRejectsInvalidAnswerState(t *testing.T) {
	for _, test := range invalidQuizEntryTests() {
		t.Run(test.name, func(t *testing.T) {
			err := quizcore.ValidateQuizEntry(test.entry)
			if err == nil {
				t.Fatal("quizcore.ValidateQuizEntry returned nil error")
			}
		})
	}
}

func invalidQuizRuntimeStateTests() []struct {
	name  string
	state quizcore.QuizRuntimeState
} {
	return []struct {
		name  string
		state quizcore.QuizRuntimeState
	}{
		{name: "status", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.Status = quizcore.QuizRuntimeStatus("unknown")
		})},
		{name: "mode", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.Mode = quizcore.QuizMode("unknown")
		})},
		{name: "budget", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.QuestionBudget = 0
		})},
		{name: "budget overflow", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.QuestionBudget = 1
			state.QuestionsAsked = 2
		})},
		{name: "outcome overflow", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.QuestionsAsked = 1
			state.AnswersRecorded = 1
			state.SkipsRecorded = 1
		})},
		{name: "active question missing", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.ActiveQuestion = nil
		})},
		{name: "inactive active question", state: quizRuntimeStateWith(func(state *quizcore.QuizRuntimeState) {
			state.Status = quizcore.QuizRuntimeStatusDone
		})},
	}
}

func invalidQuizEntryTests() []struct {
	name  string
	entry quizcore.QuizEntry
} {
	return []struct {
		name  string
		entry quizcore.QuizEntry
	}{
		{name: "answered empty answer", entry: quizEntryWith(func(entry *quizcore.QuizEntry) {
			entry.Answer = ""
		})},
		{name: "skipped answer", entry: quizEntryWith(func(entry *quizcore.QuizEntry) {
			entry.Status = quizcore.QuizEntryStatusSkipped
			entry.Answer = "This should not be present."
		})},
		{name: "invalid mode", entry: quizEntryWith(func(entry *quizcore.QuizEntry) {
			entry.Mode = quizcore.QuizMode("unknown")
		})},
		{name: "invalid status", entry: quizEntryWith(func(entry *quizcore.QuizEntry) {
			entry.Status = quizcore.QuizEntryStatus("unknown")
		})},
		{name: "invalid depth", entry: quizEntryWith(func(entry *quizcore.QuizEntry) {
			entry.Depth = 4
		})},
	}
}

func validQuizRuntimeState() quizcore.QuizRuntimeState {
	now := quizRuntimeTestTime()
	question := validQuizQuestion()

	return quizcore.QuizRuntimeState{
		ID:              "quiz-project",
		ProjectID:       "project-1",
		Status:          quizcore.QuizRuntimeStatusActive,
		Mode:            quizcore.QuizModeDeep,
		QuestionBudget:  quizcore.DefaultQuizQuestionBudget,
		QuestionsAsked:  1,
		AnswersRecorded: 0,
		SkipsRecorded:   0,
		ActiveTopic:     "workflow",
		Coverage: []quizcore.QuizCoverageTopic{
			{
				ID:            "workflow",
				Label:         "Workflow",
				Status:        quizcore.QuizCoverageStatusActive,
				QuestionCount: 1,
			},
		},
		ActiveQuestion: &question,
		Trail: []quizcore.QuizTrailNode{
			{
				QuestionID: question.ID,
				Prompt:     question.Prompt,
				Topic:      question.Topic,
				Depth:      question.Depth,
			},
		},
		Candidates: []quizcore.QuizCandidate{
			validQuizCandidate(),
		},
		RawPath:     "ops/default/quiz/00000-project-quiz-raw/artifact.md",
		CuratedPath: "ops/default/quiz/00000-project-quiz/artifact.md",
		UpdatedAt:   now,
	}
}

func quizRuntimeStateWith(update func(*quizcore.QuizRuntimeState)) quizcore.QuizRuntimeState {
	state := validQuizRuntimeState()
	update(&state)

	return state
}

func validQuizQuestion() quizcore.QuizQuestion {
	now := quizRuntimeTestTime()

	return quizcore.QuizQuestion{
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

func validQuizEntry() quizcore.QuizEntry {
	now := quizRuntimeTestTime()

	return quizcore.QuizEntry{
		ID:           "entry-1",
		QuestionID:   "question-1",
		Question:     "What should this workflow preserve?",
		Answer:       "It should preserve one visible question at a time.",
		Status:       quizcore.QuizEntryStatusAnswered,
		Topic:        "workflow",
		Depth:        1,
		Mode:         quizcore.QuizModeDeep,
		Tags:         []string{"workflow"},
		PromptIntent: `{"kind":"quiz_question","metadata":{"follow_up_depth":0}}`,
		CreatedAt:    now,
		AnsweredAt:   now.Add(time.Minute),
	}
}

func quizEntryWith(update func(*quizcore.QuizEntry)) quizcore.QuizEntry {
	entry := validQuizEntry()
	update(&entry)

	return entry
}

func validQuizCandidate() quizcore.QuizCandidate {
	now := quizRuntimeTestTime()

	return quizcore.QuizCandidate{
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
