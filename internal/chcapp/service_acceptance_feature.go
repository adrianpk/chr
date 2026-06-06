package chcapp

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chc/internal/chc"
)

// AcceptanceFeatureGenerateRequest selects the scope used to generate an acceptance feature.
type AcceptanceFeatureGenerateRequest struct {
	Scope chc.QuizScope
}

// AcceptanceFeatureOperationResult is the service-facing result for feature generation.
type AcceptanceFeatureOperationResult struct {
	Message         string
	Status          string
	Feature         string
	RejectionReason string
	Check           chc.AcceptanceFeatureCheckResult
}

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) GenerateAcceptanceFeature(
	ctx context.Context,
	request AcceptanceFeatureGenerateRequest,
) (AcceptanceFeatureOperationResult, error) {
	scope := s.quizScope(request.Scope)
	var result AcceptanceFeatureOperationResult

	err := s.withGenerationRuntime(ctx, "acceptance_feature", scope, func(ctx context.Context) error {
		var err error

		result, err = s.generateAcceptanceFeature(ctx, scope)

		return err
	})

	return result, err
}

//nolint:funlen,wsl_v5 // Generation flow is intentionally linear for logging and failure attribution.
func (s service) generateAcceptanceFeature(
	ctx context.Context,
	scope chc.QuizScope,
) (AcceptanceFeatureOperationResult, error) {
	s.appendPlanGenerationLog("generate_acceptance_feature_started", map[string]string{
		"case_slug": scope.CaseSlug,
		"timeout":   longGenerationTimeout.String(),
	})

	state, err := s.specGenerationRuntimeState(scope)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_runtime_state_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return AcceptanceFeatureOperationResult{}, err
	}

	_, err = s.latestSpecNodeForFeatureScope(scope)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_spec_missing", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return AcceptanceFeatureOperationResult{}, err
	}

	modelRequest, err := s.newAcceptanceFeatureModelRequest(state)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_request_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return AcceptanceFeatureOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_acceptance_feature_model_invocation_started", map[string]string{
		"case_slug":        scope.CaseSlug,
		"request_id":       modelRequest.ID,
		"context_sections": fmt.Sprintf("%d", len(modelRequest.ContextSections)),
	})

	response, err := chc.InvokeQuizModel(ctx, s.quizModel, modelRequest)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_model_invocation_failed", map[string]string{
			"case_slug":  scope.CaseSlug,
			"request_id": modelRequest.ID,
			"error":      err.Error(),
		})

		return AcceptanceFeatureOperationResult{}, err
	}

	feature, err := chc.ValidateAcceptanceFeatureGeneration(response.Feature)
	if err != nil {
		return AcceptanceFeatureOperationResult{}, err
	}

	s.appendPlanGenerationLog("generate_acceptance_feature_status", map[string]string{
		"case_slug": scope.CaseSlug,
		"status":    feature.Status,
		"reason":    feature.RejectionReason,
	})

	return AcceptanceFeatureOperationResult{
		Message:         acceptanceFeatureResultMessage(feature.Status),
		Status:          feature.Status,
		RejectionReason: feature.RejectionReason,
	}, nil
}

func (s service) latestSpecNodeForFeatureScope(scope chc.QuizScope) (chc.WorkspaceNode, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return chc.WorkspaceNode{}, err
	}

	node, ok := s.latestSpecNodeForScope(scope, index)
	if !ok {
		return chc.WorkspaceNode{}, fmt.Errorf("acceptance feature generation requires a generated spec")
	}

	return node, nil
}

func (s service) newAcceptanceFeatureModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionGenerateAcceptanceFeature)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err := chc.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err = s.appendAcceptanceFeatureGenerationContextSections(sections, scope)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("acceptance-feature-generation-%s", specScopeID(scope)),
		Action:          chc.QuizModelActionGenerateAcceptanceFeature,
		Strategy:        "acceptance_feature_generation",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: chc.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) appendAcceptanceFeatureGenerationContextSections(
	sections []chc.QuizModelContextSection,
	scope chc.QuizScope,
) ([]chc.QuizModelContextSection, error) {
	return s.appendGenerationContextSections(sections, scope, sortedAcceptanceFeatureContextNodeIDs)
}

func sortedAcceptanceFeatureContextNodeIDs(index chc.WorkspaceIndex) []string {
	return sortedGenerationContextNodeIDs(index, []chc.WorkspaceNodeKind{
		chc.WorkspaceNodeDump,
		chc.WorkspaceNodeQuiz,
		chc.WorkspaceNodeSpec,
	})
}

func acceptanceFeatureResultMessage(status string) string {
	switch status {
	case chc.AcceptanceFeatureStatusGenerated:
		return "Acceptance feature generated."
	case chc.AcceptanceFeatureStatusNeedsMoreInfo:
		return "Acceptance feature needs more information."
	case chc.AcceptanceFeatureStatusNotApplicable:
		return "Acceptance feature not applicable."
	case chc.AcceptanceFeatureStatusRejected:
		return "Acceptance feature rejected."
	default:
		return "Acceptance feature generation completed."
	}
}
