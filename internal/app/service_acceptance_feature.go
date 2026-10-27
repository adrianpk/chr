package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// AcceptanceFeatureGenerateRequest selects the scope used to generate an acceptance feature.
type AcceptanceFeatureGenerateRequest struct {
	OperationID string
	Scope       quizcore.Scope
}

// AcceptanceFeatureOperationResult is the service-facing result for feature generation.
type AcceptanceFeatureOperationResult struct {
	Message         string
	Status          string
	Feature         string
	FeatureRound    string
	RejectionReason string
	Check           core.AcceptanceFeatureCheckResult
}

//nolint:wsl_v5 // Keeps result capture adjacent to the guarded runtime call.
func (s service) GenerateAcceptanceFeature(ctx context.Context, request AcceptanceFeatureGenerateRequest) (AcceptanceFeatureOperationResult, error) {
	scope := s.quizScope(request.Scope)
	var result AcceptanceFeatureOperationResult

	err := s.withGenerationRuntime(ctx, "acceptance_feature", scope, func(ctx context.Context) error {
		bound, ctx, err := s.bindGeneration(ctx, "acceptance_feature-"+dumpRequestID(request.OperationID), string(quizmodel.ActionGenerateAcceptanceFeature))
		if err != nil {
			return err
		}

		result, err = bound.generateAcceptanceFeature(ctx, scope)

		return err
	})

	return result, err
}

//nolint:funlen,wsl_v5 // Generation flow is intentionally linear for logging and failure attribution.
func (s service) generateAcceptanceFeature(ctx context.Context, scope quizcore.Scope) (AcceptanceFeatureOperationResult, error) {
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

	specNode, err := s.latestSpecNodeForFeatureScope(scope)
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

	response, err := quizmodel.Invoke(ctx, s.quizModel, modelRequest)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_model_invocation_failed", map[string]string{
			"case_slug":  scope.CaseSlug,
			"request_id": modelRequest.ID,
			"error":      err.Error(),
		})

		return AcceptanceFeatureOperationResult{}, err
	}

	feature, err := generationcore.ValidateAcceptanceFeatureGeneration(response.Feature)
	if err != nil {
		return AcceptanceFeatureOperationResult{}, err
	}

	if feature.Status != generationcore.AcceptanceFeatureStatusGenerated {
		return s.acceptanceFeatureNonGeneratedResult(ctx, scope, modelRequest, response, feature)
	}

	artifact := core.AcceptanceFeatureArtifact{
		Scope:            scope,
		SpecSlug:         specNodeSlug(specNode),
		SpecPath:         specNode.RelativePath,
		UseCaseSlug:      scope.CaseSlug,
		Body:             feature.Body,
		SourceRequestID:  modelRequest.ID,
		SourceResponseID: response.ID,
		CreatedAt:        response.CreatedAt.UTC(),
	}

	paths, check, err := core.WriteAcceptanceFeature(ctx, s.projectContext.ProjectDir, artifact)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_rejected", map[string]string{
			"case_slug": scope.CaseSlug,
			"findings":  fmt.Sprintf("%d", len(check.Findings)),
			"error":     err.Error(),
		})

		return rejectedAcceptanceFeatureResult(check, err)
	}

	s.appendPlanGenerationLog("generate_acceptance_feature_written", map[string]string{
		"case_slug":     scope.CaseSlug,
		"request_id":    modelRequest.ID,
		"response_id":   response.ID,
		"relative_path": paths.RelativePath,
		"path":          paths.Path,
	})

	return AcceptanceFeatureOperationResult{
		Message: acceptanceFeatureResultMessage(generationcore.AcceptanceFeatureStatusGenerated),
		Status:  generationcore.AcceptanceFeatureStatusGenerated,
		Feature: paths.RelativePath,
		Check:   check,
	}, nil
}

func (s service) acceptanceFeatureNonGeneratedResult(ctx context.Context, scope quizcore.Scope, request quizmodel.Request, response quizmodel.Response, feature generationcore.AcceptanceFeatureGeneration) (AcceptanceFeatureOperationResult, error) {
	s.appendPlanGenerationLog("generate_acceptance_feature_status", map[string]string{
		"case_slug": scope.CaseSlug,
		"status":    feature.Status,
		"reason":    feature.RejectionReason,
	})

	featureRoundPath, err := s.acceptanceFeatureRoundPath(ctx, scope, request, response, feature)
	if err != nil {
		return AcceptanceFeatureOperationResult{}, err
	}

	return AcceptanceFeatureOperationResult{
		Message:         acceptanceFeatureResultMessage(feature.Status),
		Status:          feature.Status,
		FeatureRound:    featureRoundPath,
		RejectionReason: feature.RejectionReason,
	}, nil
}

func (s service) acceptanceFeatureRoundPath(ctx context.Context, scope quizcore.Scope, request quizmodel.Request, response quizmodel.Response, feature generationcore.AcceptanceFeatureGeneration) (string, error) {
	if feature.Status != generationcore.AcceptanceFeatureStatusNeedsMoreInfo {
		return "", nil
	}

	paths, err := s.writeAcceptanceFeatureRoundNotes(ctx, scope, request, response, feature)
	if err != nil {
		return "", err
	}

	return paths.RelativePath, nil
}

func (s service) latestSpecNodeForFeatureScope(scope quizcore.Scope) (core.WorkspaceNode, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return core.WorkspaceNode{}, err
	}

	node, ok := s.latestSpecNodeForScope(scope, index)
	if !ok {
		return core.WorkspaceNode{}, fmt.Errorf("acceptance feature generation requires a generated spec")
	}

	return node, nil
}

func (s service) newAcceptanceFeatureModelRequest(state quizcore.RuntimeState) (quizmodel.Request, error) {
	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, quizmodel.ActionGenerateAcceptanceFeature)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err = s.appendAcceptanceFeatureGenerationContextSections(sections, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	return quizmodel.Request{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("acceptance-feature-generation-%s", specScopeID(scope))),
		Action:          quizmodel.ActionGenerateAcceptanceFeature,
		Strategy:        "acceptance_feature_generation",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: quizmodel.OutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) appendAcceptanceFeatureGenerationContextSections(sections []quizmodel.ContextSection, scope quizcore.Scope) ([]quizmodel.ContextSection, error) {
	sections, err := s.appendGenerationContextSections(sections, scope, sortedAcceptanceFeatureContextNodeIDs)
	if err != nil {
		return nil, err
	}

	return s.appendAcceptanceFeatureRoundContextSection(sections, scope)
}

func (s service) appendAcceptanceFeatureRoundContextSection(sections []quizmodel.ContextSection, scope quizcore.Scope) ([]quizmodel.ContextSection, error) {
	paths, err := core.AcceptanceFeatureRoundLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(paths.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return sections, nil
		}

		return nil, fmt.Errorf("read acceptance feature round notes: %w", err)
	}

	sections = append(sections, quizmodel.ContextSection{
		Name:      "acceptance_feature_round",
		Layer:     "acceptance",
		Authority: "user",
		Content:   string(content),
		Metadata: map[string]string{
			"path": paths.RelativePath,
		},
	})

	return sections, nil
}

func (s service) writeAcceptanceFeatureRoundNotes(ctx context.Context, scope quizcore.Scope, request quizmodel.Request, response quizmodel.Response, feature generationcore.AcceptanceFeatureGeneration) (core.AcceptanceFeatureRoundPaths, error) {
	notes := core.AcceptanceFeatureRoundNotes{
		Scope:            scope,
		Status:           feature.Status,
		RejectionReason:  feature.RejectionReason,
		Prompts:          feature.ClarificationPrompts,
		SourceRequestID:  request.ID,
		SourceResponseID: response.ID,
		CreatedAt:        response.CreatedAt.UTC(),
	}

	existing, err := s.existingAcceptanceFeatureRoundNotes(ctx, scope)
	if err != nil {
		return core.AcceptanceFeatureRoundPaths{}, err
	}

	notes = core.MergeAcceptanceFeatureRoundAnswers(notes, existing)

	paths, err := core.WriteAcceptanceFeatureRoundNotes(ctx, s.projectContext.ProjectDir, notes)
	if err != nil {
		s.appendPlanGenerationLog("generate_acceptance_feature_round_notes_failed", map[string]string{
			"case_slug": scope.CaseSlug,
			"error":     err.Error(),
		})

		return core.AcceptanceFeatureRoundPaths{}, err
	}

	s.appendPlanGenerationLog("generate_acceptance_feature_round_notes_written", map[string]string{
		"case_slug":     scope.CaseSlug,
		"relative_path": paths.RelativePath,
		"path":          paths.Path,
	})

	return paths, nil
}

func (s service) existingAcceptanceFeatureRoundNotes(ctx context.Context, scope quizcore.Scope) (string, error) {
	err := ctx.Err()
	if err != nil {
		return "", err
	}

	paths, err := core.AcceptanceFeatureRoundLayout(s.projectContext.ProjectDir, scope)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(paths.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("read existing acceptance feature round notes: %w", err)
	}

	return string(content), nil
}

func sortedAcceptanceFeatureContextNodeIDs(index core.WorkspaceIndex) []string {
	return sortedGenerationContextNodeIDs(index, []core.WorkspaceNodeKind{
		core.WorkspaceNodeDump,
		core.WorkspaceNodeQuiz,
		core.WorkspaceNodeSpec,
	})
}

func specNodeSlug(node core.WorkspaceNode) string {
	name := filepath.Base(node.RelativePath)

	return name[:len(name)-len(filepath.Ext(name))]
}

func rejectedAcceptanceFeatureResult(check core.AcceptanceFeatureCheckResult, err error) (AcceptanceFeatureOperationResult, error) {
	return AcceptanceFeatureOperationResult{
		Message:         acceptanceFeatureResultMessage(generationcore.AcceptanceFeatureStatusRejected),
		Status:          generationcore.AcceptanceFeatureStatusRejected,
		RejectionReason: err.Error(),
		Check:           check,
	}, nil
}

func acceptanceFeatureResultMessage(status string) string {
	switch status {
	case generationcore.AcceptanceFeatureStatusGenerated:
		return "Acceptance feature generated."
	case generationcore.AcceptanceFeatureStatusNeedsMoreInfo:
		return "Acceptance feature needs more information."
	case generationcore.AcceptanceFeatureStatusNotApplicable:
		return "Acceptance feature not applicable."
	case generationcore.AcceptanceFeatureStatusRejected:
		return "Acceptance feature rejected."
	default:
		return "Acceptance feature generation completed."
	}
}
