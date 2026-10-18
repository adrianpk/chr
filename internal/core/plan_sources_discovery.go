package core

import (
	"context"
	"encoding/json"
	"strings"
)

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

	var answers []QuizEntry

	for _, entry := range quiz.Entries {
		if entry.Status == QuizEntryStatusAnswered && strings.TrimSpace(entry.Answer) != "" {
			answers = append(answers, entry)
		}
	}

	content, err := json.Marshal(answers)
	if err != nil {
		return err
	}

	observed.Inputs = append(observed.Inputs, PlanSourceInput{"quiz", paths.RawRelativePath, quiz.Revision, string(content)})
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

		observed.Inputs = append(observed.Inputs, PlanSourceInput{"spec", document.Reference, document.Revision, string(document.images[0].After)})
	}

	return nil
}
