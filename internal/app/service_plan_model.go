package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) newPlanOutlineGenerationModelRequest(ctx context.Context, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (quizmodel.Request, error) {
	return s.newPlanningModelRequest(ctx,
		state, sources,
		documentModelOperation{action: quizmodel.ActionGeneratePlanOutline, strategy: "plan_outline_generation", requestIDPrefix: "plan-outline-generation"},
	)
}

func (s service) newPlanGenerationModelRequest(ctx context.Context, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (quizmodel.Request, error) {
	return s.newPlanningModelRequest(ctx,
		state, sources,
		documentModelOperation{action: quizmodel.ActionGeneratePlan, strategy: "plan_generation", requestIDPrefix: "plan-generation"},
	)
}

func (s service) newPlanSliceDetailGenerationModelRequest(ctx context.Context, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation, outline generationcore.GeneratedPlanOutline, targetSliceID string) (quizmodel.Request, error) {
	if err := generationcore.ValidateGeneratedPlanOutline(outline); err != nil {
		return quizmodel.Request{}, err
	}

	target, ok := outline.Slice(targetSliceID)
	if !ok {
		return quizmodel.Request{}, fmt.Errorf("plan outline slice is unknown: %s", targetSliceID)
	}

	request, err := s.newPlanningModelRequest(ctx, state, sources, documentModelOperation{action: quizmodel.ActionGeneratePlanSliceDetail, strategy: "plan_slice_detail_generation", requestIDPrefix: "plan-slice-detail-generation"})
	if err != nil {
		return request, err
	}

	request.ID += "-" + planSliceIDSlug(targetSliceID)
	request.ContextSections, err = appendPlanSliceDetailContextSections(request.ContextSections, outline, target)

	return request, err
}

func (s service) newPlanningModelRequest(ctx context.Context, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation, operation documentModelOperation) (quizmodel.Request, error) {
	scope := sources.Scope
	payload := quizmodel.Payload{
		Project: quizmodel.ProjectPayload{ID: scope.ProjectID, Name: sources.Project.Name, Description: sources.Project.Description},
		Quiz:    quizModelQuizPayload(state), RecentEntries: quizModelRecentEntryPayloads(sources.Quiz.Entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, sources.Quiz.Entries), OutputContract: quizmodel.NewOutputContract(operation.action, state.Mode),
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	for _, input := range sources.Inputs {
		sections = append(sections, quizmodel.ContextSection{Name: input.Kind, Layer: input.Kind, Authority: input.Reference, Content: input.Content, Metadata: map[string]string{"source_revision": input.Revision, "source_observation_hash": sources.Hash}})
	}

	request := s.documentGenerationModelRequest(scope, state, payload, sections, operation)
	request.ID = harnesshost.HarnessGenerationRequestID(ctx, request.ID)

	return request, nil
}

func appendPlanSliceDetailContextSections(sections []quizmodel.ContextSection, outline generationcore.GeneratedPlanOutline, targetSlice generationcore.GeneratedPlanOutlineSlice) ([]quizmodel.ContextSection, error) {
	outlineContent, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return nil, err
	}

	targetContent, err := json.MarshalIndent(targetSlice, "", "  ")
	if err != nil {
		return nil, err
	}

	sections = append(sections, quizmodel.ContextSection{
		Name:      "plan_outline",
		Layer:     "plan",
		Authority: "generated_plan_outline",
		Content:   string(outlineContent),
		Metadata: map[string]string{
			"title":  outline.Title,
			"status": outline.Status,
		},
	})
	sections = append(sections, quizmodel.ContextSection{
		Name:      "target_slice",
		Layer:     "plan",
		Authority: targetSlice.ID,
		Content:   string(targetContent),
		Metadata: map[string]string{
			"slice_id":    targetSlice.ID,
			"short_name":  targetSlice.ShortName,
			"branch_name": targetSlice.BranchName,
		},
	})

	return sections, nil
}

func planSliceIDSlug(sliceID string) string {
	slug := strings.ToLower(strings.TrimSpace(sliceID))
	slug = strings.ReplaceAll(slug, " ", "-")

	return slug
}
