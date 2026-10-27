package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) generatePlanWithConfiguredBackend(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources core.PlanSourceObservation) (generationcore.GeneratedPlan, []string, []string, error) {
	if s.singlePassPlan(ctx) {
		return s.generateSinglePassPlan(ctx, scope, state, sources)
	}

	return s.generateStagedPlan(ctx, scope, state, sources)
}

func (s service) generateSinglePassPlan(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources core.PlanSourceObservation) (generationcore.GeneratedPlan, []string, []string, error) {
	modelRequest, err := s.newPlanGenerationModelRequest(state, sources)
	if err != nil {
		return generationcore.GeneratedPlan{}, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_single_pass_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return generationcore.GeneratedPlan{}, nil, nil, err
	}

	plan := response.Plan
	plan.Scope = scope
	plan.SourcePaths = specSourcePaths(modelRequest.ContextSections)
	plan.CreatedAt = response.CreatedAt.UTC()

	s.appendPlanGenerationLog("generate_plan_single_pass_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       plan.Title,
	})

	return plan, []string{modelRequest.ID}, []string{response.ID}, nil
}

func (s service) generateStagedPlan(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources core.PlanSourceObservation) (generationcore.GeneratedPlan, []string, []string, error) {
	outline, sourcePaths, requestIDs, responseIDs, err := s.generatePlanOutline(ctx, scope, state, sources)
	if err != nil {
		return generationcore.GeneratedPlan{}, nil, nil, err
	}

	details := make([]generationcore.GeneratedPlanSliceDetail, 0, len(outline.Slices))
	for _, slice := range outline.Slices {
		detail, requestID, responseID, err := s.generatePlanSliceDetail(ctx, state, sources, outline, slice.ID)
		if err != nil {
			return generationcore.GeneratedPlan{}, nil, nil, err
		}

		details = append(details, detail)
		requestIDs = append(requestIDs, requestID)
		responseIDs = append(responseIDs, responseID)
	}

	body, err := core.AssembleGeneratedPlanBody(outline, details, requestIDs, responseIDs)
	if err != nil {
		return generationcore.GeneratedPlan{}, nil, nil, err
	}

	plan := generationcore.GeneratedPlan{
		Scope:       scope,
		Title:       outline.Title,
		Body:        body,
		Order:       core.PlannedOrderFromDetails(details),
		SourcePaths: sourcePaths,
		CreatedAt:   outline.CreatedAt.UTC(),
	}

	return plan, requestIDs, responseIDs, nil
}

func (s service) generatePlanOutline(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources core.PlanSourceObservation) (generationcore.GeneratedPlanOutline, []string, []string, []string, error) {
	modelRequest, err := s.newPlanOutlineGenerationModelRequest(state, sources)
	if err != nil {
		return generationcore.GeneratedPlanOutline{}, nil, nil, nil, err
	}

	s.appendPlanGenerationLog("generate_plan_outline_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return generationcore.GeneratedPlanOutline{}, nil, nil, nil, err
	}

	outline := response.PlanOutline
	outline.Scope = scope
	outline.CreatedAt = response.CreatedAt.UTC()

	s.appendPlanGenerationLog("generate_plan_outline_finished", map[string]string{
		"case_slug":   scope.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"title":       outline.Title,
		"slices":      fmt.Sprintf("%d", len(outline.Slices)),
	})

	return outline,
		specSourcePaths(modelRequest.ContextSections),
		[]string{modelRequest.ID},
		[]string{response.ID},
		nil
}

func (s service) generatePlanSliceDetail(ctx context.Context, state quizcore.RuntimeState, sources core.PlanSourceObservation, outline generationcore.GeneratedPlanOutline, targetSliceID string) (generationcore.GeneratedPlanSliceDetail, string, string, error) {
	modelRequest, err := s.newPlanSliceDetailGenerationModelRequest(state, sources, outline, targetSliceID)
	if err != nil {
		return generationcore.GeneratedPlanSliceDetail{}, "", "", err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_started", map[string]string{
		"case_slug":  state.CaseSlug,
		"request_id": modelRequest.ID,
		"slice_id":   targetSliceID,
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return generationcore.GeneratedPlanSliceDetail{}, "", "", err
	}

	detail := response.SliceDetail
	detail.Scope = outline.Scope
	detail.CreatedAt = response.CreatedAt.UTC()

	if err := generationcore.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, targetSliceID); err != nil {
		return generationcore.GeneratedPlanSliceDetail{}, "", "", err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_finished", map[string]string{
		"case_slug":   state.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"slice_id":    targetSliceID,
	})

	return detail, modelRequest.ID, response.ID, nil
}
