package authoring

import (
	"context"
	"encoding/json"
	"strings"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Planning consumes discovery content; routing controls remain in the retained
// quiz observation. The source reference and revision bind the complete history.
type planQuizAnswer struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
}

func appendPlanQuiz(ctx context.Context, projectDir string, observed *PlanSourceObservation) error {
	repository := NewQuizArtifactRepository(projectDir)

	quiz, err := repository.Observe(ctx, observed.Scope)
	if err != nil {
		return err
	}

	paths, err := repository.Paths(observed.Scope)
	if err != nil {
		return err
	}

	observed.Quiz = quiz

	answers := make([]planQuizAnswer, 0, len(quiz.Entries))

	for _, entry := range quiz.Entries {
		if entry.Status == quizcore.EntryStatusAnswered && strings.TrimSpace(entry.Answer) != "" {
			answers = append(answers, planQuizAnswer{ID: entry.ID, QuestionID: entry.QuestionID, Question: entry.Question, Answer: entry.Answer})
		}
	}

	content, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	observed.Inputs = append(observed.Inputs, PlanSourceInput{Kind: "quiz", Reference: paths.RawRelativePath, Revision: quiz.Revision, Content: string(content)})
	observed.HasDiscovery = observed.HasDiscovery || len(answers) > 0

	return nil
}

func appendPlanSpecs(ctx context.Context, projectDir string, observed *PlanSourceObservation) error {
	documents, err := ListSpecDocuments(ctx, projectDir)
	if err != nil {
		return err
	}

	for _, document := range documents {
		if document.CaseSlug != observed.Scope.CaseSlug || strings.Contains(document.front, "archived_from:") {
			continue
		}

		observed.Inputs = append(observed.Inputs, PlanSourceInput{Kind: "spec", Reference: document.Reference, Revision: document.Revision, Content: string(document.images[0].After), Status: document.Status})
	}

	return nil
}
