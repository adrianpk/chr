package authoring

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func validCodexQuizModelRequest(action quizmodel.Action) quizmodel.Request {
	request := validQuizModelRequest()
	request.Action = action
	request.Strategy = string(action)
	request.ID = string(action) + "-blog"

	request.Payload.OutputContract = quizmodel.NewOutputContract(action, request.Mode)
	request.OutputContract = quizmodel.OutputContract{
		RequiredFields: request.Payload.OutputContract.RequiredFields,
	}

	if action == quizmodel.ActionGeneratePlan {
		request.OutputContract = quizmodel.OutputContract{RequiredFields: []string{"title", "body", "status"}}
	}

	return request
}

func validQuizModelRequest() quizmodel.Request {
	state := validQuizRuntimeState()
	payload := quizmodel.Payload{
		Project: quizmodel.ProjectPayload{
			ID:          "project-1",
			Name:        "Project",
			Description: "Project context.",
		},
		Quiz: quizmodel.QuizPayload{
			ID:         state.ID,
			EntryCount: 1,
			Mode:       string(quizcore.ModeDeep),
		},
		RecentEntries: []quizmodel.EntryPayload{
			{
				Question: "What should this workflow preserve?",
				Answer:   "The workflow should preserve review.",
				Status:   string(quizcore.EntryStatusAnswered),
				Tags:     []string{"workflow"},
			},
		},
		CurrentQuestion: quizmodel.CurrentQuestionPayload{
			Question: "What should this workflow preserve?",
			Tags:     []string{"workflow"},
			PromptIntent: map[string]any{
				"metadata": map[string]any{"follow_up_depth": 0},
			},
		},
		OutputContract: quizmodel.NewOutputContract(quizmodel.ActionNextQuestion, quizcore.ModeDeep),
	}

	return quizmodel.Request{
		ID:           "quiz-model-request-1",
		Action:       quizmodel.ActionNextQuestion,
		Strategy:     "deep",
		Scope:        quizcore.Scope{ProjectID: "project-1"},
		Mode:         quizcore.ModeDeep,
		RuntimeState: state,
		Payload:      payload,
		ContextSections: []quizmodel.ContextSection{
			{Name: "project", Layer: "project", Authority: "project_metadata", Content: "Project context."},
		},
		OutputContract: quizmodel.OutputContract{
			RequiredFields: []string{"id", "prompt", "topic", "depth"},
		},
		CreatedAt: quizRuntimeTestTime(),
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
