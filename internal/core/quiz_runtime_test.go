package core

import (
	"fmt"
	"testing"
	"time"
)

func TestValidateQuizRuntimeStateAcceptsProjectWideState(t *testing.T) {
	state := validQuizRuntimeState()

	err := ValidateQuizRuntimeState(state)
	if err != nil {
		t.Fatalf("ValidateQuizRuntimeState returned error: %v", err)
	}
}

func TestValidateQuizRuntimeStateAcceptsUseCaseScopedState(t *testing.T) {
	state := validQuizRuntimeState()
	state.ID = "quiz-blog"
	state.CaseSlug = "blog"
	state.RawPath = ".chr/user/quizzes/blog-quiz-raw.md"
	state.CuratedPath = ".chr/user/quizzes/blog-quiz.md"

	err := ValidateQuizRuntimeState(state)
	if err != nil {
		t.Fatalf("ValidateQuizRuntimeState returned error: %v", err)
	}
}

func TestValidateQuizRuntimeStateRejectsInvalidState(t *testing.T) {
	for _, test := range invalidQuizRuntimeStateTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateQuizRuntimeState(test.state)
			if err == nil {
				t.Fatal("ValidateQuizRuntimeState returned nil error")
			}
		})
	}
}

func TestValidateQuizQuestionRejectsInvalidDepth(t *testing.T) {
	for _, depth := range []int{0, 4} {
		t.Run(fmt.Sprintf("depth-%d", depth), func(t *testing.T) {
			question := validQuizQuestion()
			question.Depth = depth

			err := ValidateQuizQuestion(question)
			if err == nil {
				t.Fatal("ValidateQuizQuestion returned nil error")
			}
		})
	}
}

func TestValidateQuizEntryRejectsInvalidAnswerState(t *testing.T) {
	for _, test := range invalidQuizEntryTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateQuizEntry(test.entry)
			if err == nil {
				t.Fatal("ValidateQuizEntry returned nil error")
			}
		})
	}
}

func invalidQuizRuntimeStateTests() []struct {
	name  string
	state QuizRuntimeState
} {
	return []struct {
		name  string
		state QuizRuntimeState
	}{
		{name: "status", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.Status = QuizRuntimeStatus("unknown")
		})},
		{name: "mode", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.Mode = QuizMode("unknown")
		})},
		{name: "budget", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.QuestionBudget = 0
		})},
		{name: "budget overflow", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.QuestionBudget = 1
			state.QuestionsAsked = 2
		})},
		{name: "outcome overflow", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.QuestionsAsked = 1
			state.AnswersRecorded = 1
			state.SkipsRecorded = 1
		})},
		{name: "active question missing", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.ActiveQuestion = nil
		})},
		{name: "inactive active question", state: quizRuntimeStateWith(func(state *QuizRuntimeState) {
			state.Status = QuizRuntimeStatusDone
		})},
	}
}

func invalidQuizEntryTests() []struct {
	name  string
	entry QuizEntry
} {
	return []struct {
		name  string
		entry QuizEntry
	}{
		{name: "answered empty answer", entry: quizEntryWith(func(entry *QuizEntry) {
			entry.Answer = ""
		})},
		{name: "skipped answer", entry: quizEntryWith(func(entry *QuizEntry) {
			entry.Status = QuizEntryStatusSkipped
			entry.Answer = "This should not be present."
		})},
		{name: "invalid mode", entry: quizEntryWith(func(entry *QuizEntry) {
			entry.Mode = QuizMode("unknown")
		})},
		{name: "invalid status", entry: quizEntryWith(func(entry *QuizEntry) {
			entry.Status = QuizEntryStatus("unknown")
		})},
		{name: "invalid depth", entry: quizEntryWith(func(entry *QuizEntry) {
			entry.Depth = 4
		})},
	}
}

func validQuizRuntimeState() QuizRuntimeState {
	now := quizRuntimeTestTime()
	question := validQuizQuestion()

	return QuizRuntimeState{
		ID:              "quiz-project",
		ProjectID:       "project-1",
		Status:          QuizRuntimeStatusActive,
		Mode:            QuizModeDeep,
		QuestionBudget:  DefaultQuizQuestionBudget,
		QuestionsAsked:  1,
		AnswersRecorded: 0,
		SkipsRecorded:   0,
		ActiveTopic:     "workflow",
		Coverage: []QuizCoverageTopic{
			{
				ID:            "workflow",
				Label:         "Workflow",
				Status:        QuizCoverageStatusActive,
				QuestionCount: 1,
			},
		},
		ActiveQuestion: &question,
		Trail: []QuizTrailNode{
			{
				QuestionID: question.ID,
				Prompt:     question.Prompt,
				Topic:      question.Topic,
				Depth:      question.Depth,
			},
		},
		Candidates: []QuizCandidate{
			validQuizCandidate(),
		},
		RawPath:     ".chr/user/quizzes/project-quiz-raw.md",
		CuratedPath: ".chr/user/quizzes/project-quiz.md",
		UpdatedAt:   now,
	}
}

func quizRuntimeStateWith(update func(*QuizRuntimeState)) QuizRuntimeState {
	state := validQuizRuntimeState()
	update(&state)

	return state
}

func validQuizQuestion() QuizQuestion {
	now := quizRuntimeTestTime()

	return QuizQuestion{
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

func validQuizEntry() QuizEntry {
	now := quizRuntimeTestTime()

	return QuizEntry{
		ID:           "entry-1",
		QuestionID:   "question-1",
		Question:     "What should this workflow preserve?",
		Answer:       "It should preserve one visible question at a time.",
		Status:       QuizEntryStatusAnswered,
		Topic:        "workflow",
		Depth:        1,
		Mode:         QuizModeDeep,
		Tags:         []string{"workflow"},
		PromptIntent: `{"kind":"quiz_question","metadata":{"follow_up_depth":0}}`,
		CreatedAt:    now,
		AnsweredAt:   now.Add(time.Minute),
	}
}

func quizEntryWith(update func(*QuizEntry)) QuizEntry {
	entry := validQuizEntry()
	update(&entry)

	return entry
}

func validQuizCandidate() QuizCandidate {
	now := quizRuntimeTestTime()

	return QuizCandidate{
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
