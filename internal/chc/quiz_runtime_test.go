package chc

import (
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
	state.RawPath = ".chc/quizzes/blog-quiz-raw.md"
	state.CuratedPath = ".chc/quizzes/blog-quiz.md"

	err := ValidateQuizRuntimeState(state)
	if err != nil {
		t.Fatalf("ValidateQuizRuntimeState returned error: %v", err)
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
		RawPath:     ".chc/quizzes/project-quiz-raw.md",
		CuratedPath: ".chc/quizzes/project-quiz.md",
		UpdatedAt:   now,
	}
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
