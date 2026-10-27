package app

import (
	"context"
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type planGenerationResult struct {
	Plan        generationcore.GeneratedPlan
	RequestIDs  []string
	ResponseIDs []string
}

type planOutlineGenerationResult struct {
	Outline     generationcore.GeneratedPlanOutline
	SourcePaths []string
	RequestIDs  []string
	ResponseIDs []string
}

type planSliceDetailGenerationResult struct {
	Detail     generationcore.GeneratedPlanSliceDetail
	RequestID  string
	ResponseID string
}

func (s service) generatePlanWithConfiguredBackend(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (planGenerationResult, error) {
	if s.singlePassPlan(ctx) {
		return s.generateSinglePassPlan(ctx, scope, state, sources)
	}

	return s.generateStagedPlan(ctx, scope, state, sources)
}

func (s service) generateSinglePassPlan(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (planGenerationResult, error) {
	modelRequest, err := s.newPlanGenerationModelRequest(ctx, state, sources)
	if err != nil {
		return planGenerationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_single_pass_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return planGenerationResult{}, err
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

	return planGenerationResult{Plan: plan, RequestIDs: []string{modelRequest.ID}, ResponseIDs: []string{response.ID}}, nil
}

func (s service) generateStagedPlan(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (planGenerationResult, error) {
	generatedOutline, err := s.generatePlanOutline(ctx, scope, state, sources)
	if err != nil {
		return planGenerationResult{}, err
	}

	outline := generatedOutline.Outline
	requestIDs, responseIDs := generatedOutline.RequestIDs, generatedOutline.ResponseIDs

	details := make([]generationcore.GeneratedPlanSliceDetail, 0, len(outline.Slices))
	for _, slice := range outline.Slices {
		generatedDetail, err := s.generatePlanSliceDetail(ctx, state, sources, outline, slice.ID)
		if err != nil {
			return planGenerationResult{}, err
		}

		details = append(details, generatedDetail.Detail)
		requestIDs = append(requestIDs, generatedDetail.RequestID)
		responseIDs = append(responseIDs, generatedDetail.ResponseID)
	}

	body, err := authoringcore.AssembleGeneratedPlanBody(outline, details, requestIDs, responseIDs)
	if err != nil {
		return planGenerationResult{}, err
	}

	plan := generationcore.GeneratedPlan{
		Scope:       scope,
		Title:       outline.Title,
		Body:        body,
		Order:       authoringcore.PlannedOrderFromDetails(details),
		SourcePaths: generatedOutline.SourcePaths,
		CreatedAt:   outline.CreatedAt.UTC(),
	}

	return planGenerationResult{Plan: plan, RequestIDs: requestIDs, ResponseIDs: responseIDs}, nil
}

func (s service) generatePlanOutline(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation) (planOutlineGenerationResult, error) {
	modelRequest, err := s.newPlanOutlineGenerationModelRequest(ctx, state, sources)
	if err != nil {
		return planOutlineGenerationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_outline_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return planOutlineGenerationResult{}, err
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

	return planOutlineGenerationResult{
		Outline: outline, SourcePaths: specSourcePaths(modelRequest.ContextSections),
		RequestIDs: []string{modelRequest.ID}, ResponseIDs: []string{response.ID},
	}, nil
}

func (s service) generatePlanSliceDetail(ctx context.Context, state quizcore.RuntimeState, sources authoringcore.PlanSourceObservation, outline generationcore.GeneratedPlanOutline, targetSliceID string) (planSliceDetailGenerationResult, error) {
	modelRequest, err := s.newPlanSliceDetailGenerationModelRequest(ctx, state, sources, outline, targetSliceID)
	if err != nil {
		return planSliceDetailGenerationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_started", map[string]string{
		"case_slug":  state.CaseSlug,
		"request_id": modelRequest.ID,
		"slice_id":   targetSliceID,
	})

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		return planSliceDetailGenerationResult{}, err
	}

	detail := response.SliceDetail
	detail.Scope = outline.Scope
	detail.CreatedAt = response.CreatedAt.UTC()

	if err := generationcore.ValidateGeneratedPlanSliceDetailAgainstOutline(detail, outline, targetSliceID); err != nil {
		return planSliceDetailGenerationResult{}, err
	}

	s.appendPlanGenerationLog("generate_plan_slice_detail_finished", map[string]string{
		"case_slug":   state.CaseSlug,
		"request_id":  modelRequest.ID,
		"response_id": response.ID,
		"slice_id":    targetSliceID,
	})

	return planSliceDetailGenerationResult{Detail: detail, RequestID: modelRequest.ID, ResponseID: response.ID}, nil
}
