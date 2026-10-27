package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) newPlanOutlineGenerationModelRequest(state quizcore.RuntimeState, sources core.PlanSourceObservation) (quizmodel.Request, error) {
	return s.newPlanningModelRequest(
		state, sources,
		quizmodel.ActionGeneratePlanOutline,
		"plan_outline_generation",
		"plan-outline-generation",
	)
}

func (s service) newPlanGenerationModelRequest(state quizcore.RuntimeState, sources core.PlanSourceObservation) (quizmodel.Request, error) {
	return s.newPlanningModelRequest(
		state, sources,
		quizmodel.ActionGeneratePlan,
		"plan_generation",
		"plan-generation",
	)
}

func (s service) newPlanSliceDetailGenerationModelRequest(state quizcore.RuntimeState, sources core.PlanSourceObservation, outline generationcore.GeneratedPlanOutline, targetSliceID string) (quizmodel.Request, error) {
	if err := generationcore.ValidateGeneratedPlanOutline(outline); err != nil {
		return quizmodel.Request{}, err
	}

	target, ok := outline.Slice(targetSliceID)
	if !ok {
		return quizmodel.Request{}, fmt.Errorf("plan outline slice is unknown: %s", targetSliceID)
	}

	request, err := s.newPlanningModelRequest(state, sources, quizmodel.ActionGeneratePlanSliceDetail, "plan_slice_detail_generation", "plan-slice-detail-generation")
	if err != nil {
		return request, err
	}

	request.ID += "-" + planSliceIDSlug(targetSliceID)
	request.ContextSections, err = appendPlanSliceDetailContextSections(request.ContextSections, outline, target)

	return request, err
}

func (s service) newPlanningModelRequest(state quizcore.RuntimeState, sources core.PlanSourceObservation, action quizmodel.Action, strategy, idPrefix string) (quizmodel.Request, error) {
	scope := sources.Scope
	payload := quizmodel.Payload{
		Project: quizmodel.ProjectPayload{ID: scope.ProjectID, Name: sources.Project.Name, Description: sources.Project.Description},
		Quiz:    quizModelQuizPayload(state), RecentEntries: quizModelRecentEntryPayloads(sources.Quiz.Entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, sources.Quiz.Entries), OutputContract: quizmodel.NewOutputContract(action, state.Mode),
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	for _, input := range sources.Inputs {
		sections = append(sections, quizmodel.ContextSection{Name: input.Kind, Layer: input.Kind, Authority: input.Reference, Content: input.Content, Metadata: map[string]string{"source_revision": input.Revision, "source_observation_hash": sources.Hash}})
	}

	return quizmodel.Request{
		ID: core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("%s-%s", idPrefix, specScopeID(scope))), Action: action, Strategy: strategy, Scope: scope, Mode: state.Mode, RuntimeState: state, Payload: payload, ContextSections: sections,
		OutputContract: quizmodel.OutputContract{RequiredFields: payload.OutputContract.RequiredFields}, CreatedAt: s.utcNow(),
	}, nil
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
