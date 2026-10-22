package app

import (
	"context"
	"encoding/json"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) guidedPlanSources(session core.AuthoringSession) (core.PlanSourceObservation, core.QuizRuntimeState, error) {
	sections, state, entries, _, err := s.guidedSpecInputs(session)
	if err != nil {
		return core.PlanSourceObservation{}, state, err
	}

	specification, err := s.guidedCommittedSections(session, "specification")
	if err != nil {
		return core.PlanSourceObservation{}, state, err
	}

	scope := core.QuizScope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}

	sources := core.PlanSourceObservation{Scope: scope, Project: core.ProjectMetadata{Name: session.Guided.CaseName}, Quiz: core.QuizObservation{Exists: true, State: state, Entries: entries}, HasDiscovery: true}
	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage == "specification" {
			sources.Hash, err = core.GuidedCheckpointDigest(checkpoint)
			if err != nil {
				return sources, state, err
			}
		}
	}

	for _, section := range append(sections, specification...) {
		kind := "committed_discovery"

		status := ""
		if section.Authority == "committed" && section.Name == session.Guided.SpecApproval.Reference {
			kind, status = "spec", "approved"
		}

		sources.Inputs = append(sources.Inputs, core.PlanSourceInput{Kind: kind, Reference: section.Name, Revision: section.Metadata["commit"] + ":" + section.Metadata["hash"], Content: section.Content, Status: status})
	}

	return sources, state, nil
}

func (s service) generateGuidedContracts(ctx context.Context, session core.AuthoringSession) (core.GeneratedWorkflowPlan, error) {
	scope := session.Guided.Work.PublicationOperation.Scope

	input, err := core.ObservePlanWorkflow(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return input, err
	}

	input.UseCase = session.Guided.CaseSlug

	bound, ctx, err := s.bindGeneration(ctx, input.MaterializationID, string(core.QuizModelActionMaterializePlanTracking))
	if err != nil {
		return input, err
	}

	request, err := bound.guidedContractsRequest(session, input)
	if err != nil {
		return input, err
	}

	response, err := core.InvokeQuizModel(ctx, bound.quizModel, request)
	if err != nil {
		return input, err
	}

	if response.PlanTracking.Workflow == nil {
		return input, errors.New("the provider did not return native ticket contracts")
	}

	generated := *response.PlanTracking.Workflow
	if err := core.ValidateGuidedContracts(input, generated); err != nil {
		return input, err
	}

	return generated, nil
}

func (s service) guidedContractsRequest(session core.AuthoringSession, input core.GeneratedWorkflowPlan) (core.QuizModelRequest, error) {
	sources, state, err := s.guidedPlanSources(session)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	request, err := s.newPlanningModelRequest(state, sources, core.QuizModelActionMaterializePlanTracking, core.QuizModelStrategyPlanDelivery, "guided-contracts")
	if err != nil {
		return request, err
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return request, err
	}

	body := core.RenderGeneratedPlan(session.Guided.PendingOperation.PreparedPlan.Plan)
	if core.WorkflowContentHash([]byte(body)) != input.PlanHash {
		return request, core.ErrWorkflowConflict
	}

	request.ContextSections = append(request.ContextSections, core.QuizModelContextSection{Name: "planned_workflow", Layer: "plan", Authority: input.PlanRef, Content: string(encoded)}, core.QuizModelContextSection{Name: "implementation_plan", Layer: "plan", Authority: input.PlanRef, Content: body})

	return request, nil
}
