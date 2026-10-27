package app

import (
	"context"
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) generateQuizDraftSpec(scope quizcore.Scope) (QuizOperationResult, error) {
	bound, err := s.bindGeneration(s.ctx, "", string(quizmodel.ActionDraftSpec))
	if err != nil {
		return QuizOperationResult{}, err
	}

	ctx := bound.context(s.ctx)
	s = bound.service
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	paths, err := s.writeQuizCuratedNotes(ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	err = s.writeQuizDraftSpec(ctx, state, paths)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Spec generated.", State: state}, nil
}

func (s service) writeQuizCuratedNotes(ctx context.Context, scope quizcore.Scope) (authoringcore.QuizArtifactPaths, error) {
	repository := s.quizRepository()

	entries, err := repository.LoadRawCaptureEntries(ctx, scope)
	if err != nil {
		return authoringcore.QuizArtifactPaths{}, err
	}

	return repository.WriteCuratedNotes(ctx, quizcore.CurateEntries(scope, entries, s.utcNow()))
}

func (s service) writeQuizDraftSpec(ctx context.Context, state quizcore.RuntimeState, paths authoringcore.QuizArtifactPaths) error {
	target, err := authoringcore.QuizDraftSpecLayout(s.projectContext.ProjectDir, quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug})
	if err != nil {
		return err
	}

	revision, err := authoringcore.SpecTargetRevision(ctx, s.projectContext.ProjectDir, target.RelativePath)
	if err != nil {
		return err
	}

	request, err := s.newQuizDraftSpecModelRequest(ctx, state, paths)
	if err != nil {
		return err
	}

	response, err := quizmodel.Invoke(ctx, s.quizModel, request)
	if err != nil {
		return err
	}

	spec := response.DraftSpec
	spec.ExpectedRevision = revision
	spec.Scope = quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
	spec.SourceQuizPath = paths.CuratedRelativePath
	spec.SourceRawPath = paths.RawRelativePath
	spec.SourceDumpPaths = quizDumpSourcePaths(request.ContextSections)
	spec.SourceRequestID = request.ID
	spec.SourceResponseID = response.ID
	spec.CreatedAt = response.CreatedAt.UTC()

	_, err = authoringcore.WriteQuizDraftSpec(ctx, s.projectContext.ProjectDir, spec)

	return err
}

func (s service) newQuizDraftSpecModelRequest(ctx context.Context, state quizcore.RuntimeState, paths authoringcore.QuizArtifactPaths) (quizmodel.Request, error) {
	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(ctx, scope, state, quizmodel.ActionDraftSpec)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err = s.appendQuizSupplementalContextSections(ctx, sections, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections = appendQuizFileContextSection(
		sections,
		s.projectContext.ProjectDir,
		"curated_quiz",
		"quiz",
		paths.CuratedRelativePath,
	)

	return quizmodel.Request{
		ID:              harnesshost.GenerationRequestID(ctx, fmt.Sprintf("quiz-draft-spec-%s", paths.ScopeID)),
		Action:          quizmodel.ActionDraftSpec,
		Strategy:        "draft_spec",
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

func quizDumpSourcePaths(sections []quizmodel.ContextSection) []string {
	paths := []string{}

	for _, section := range sections {
		if section.Layer == "dump" {
			paths = append(paths, section.Authority)
		}
	}

	return paths
}
