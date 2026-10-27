package app

import (
	"context"
	"encoding/json"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type guidedPlanSourcesResult struct {
	Sources core.PlanSourceObservation
	State   quizcore.RuntimeState
}

func (s service) guidedPlanSources(session core.AuthoringSession) (guidedPlanSourcesResult, error) {
	inputs, err := s.guidedSpecInputs(session)
	if err != nil {
		return guidedPlanSourcesResult{State: inputs.State}, err
	}

	specification, err := s.guidedCommittedSections(session, "specification")
	if err != nil {
		return guidedPlanSourcesResult{State: inputs.State}, err
	}

	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}

	sources := core.PlanSourceObservation{Scope: scope, Project: core.ProjectMetadata{Name: session.Guided.CaseName}, Quiz: quizcore.Observation{Exists: true, State: inputs.State, Entries: inputs.Entries}, HasDiscovery: true}
	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage == "specification" {
			sources.Hash, err = core.GuidedCheckpointDigest(checkpoint)
			if err != nil {
				return guidedPlanSourcesResult{Sources: sources, State: inputs.State}, err
			}
		}
	}

	for _, section := range append(inputs.Sections, specification...) {
		kind := "committed_discovery"

		status := ""
		if section.Authority == "committed" && section.Name == session.Guided.SpecApproval.Reference {
			kind, status = "spec", "approved"
		}

		sources.Inputs = append(sources.Inputs, core.PlanSourceInput{Kind: kind, Reference: section.Name, Revision: section.Metadata["commit"] + ":" + section.Metadata["hash"], Content: section.Content, Status: status})
	}

	return guidedPlanSourcesResult{Sources: sources, State: inputs.State}, nil
}

func (s service) generateGuidedContracts(ctx context.Context, session core.AuthoringSession) (generationcore.GeneratedWorkflowPlan, error) {
	scope := session.Guided.Work.PublicationOperation.Scope

	input, err := core.ObservePlanWorkflow(ctx, s.projectContext.ProjectDir, scope)
	if err != nil {
		return input, err
	}

	input.UseCase = session.Guided.CaseSlug

	bound, ctx, err := s.bindGeneration(ctx, input.MaterializationID, string(quizmodel.ActionMaterializePlanTracking))
	if err != nil {
		return input, err
	}

	request, err := bound.guidedContractsRequest(session, input)
	if err != nil {
		return input, err
	}

	response, err := quizmodel.Invoke(ctx, bound.quizModel, request)
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

func (s service) guidedContractsRequest(session core.AuthoringSession, input generationcore.GeneratedWorkflowPlan) (quizmodel.Request, error) {
	observed, err := s.guidedPlanSources(session)
	if err != nil {
		return quizmodel.Request{}, err
	}

	request, err := s.newPlanningModelRequest(observed.State, observed.Sources, documentModelOperation{action: quizmodel.ActionMaterializePlanTracking, strategy: quizmodel.StrategyPlanDelivery, requestIDPrefix: "guided-contracts"})
	if err != nil {
		return request, err
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return request, err
	}

	body := core.RenderGeneratedPlan(session.Guided.PendingOperation.PreparedPlan.Plan)
	if core.WorkflowContentHash([]byte(body)) != input.PlanHash {
		return request, workflowstate.ErrConflict
	}

	request.ContextSections = append(request.ContextSections, quizmodel.ContextSection{Name: "planned_workflow", Layer: "plan", Authority: input.PlanRef, Content: string(encoded)}, quizmodel.ContextSection{Name: "implementation_plan", Layer: "plan", Authority: input.PlanRef, Content: body})

	return request, nil
}
