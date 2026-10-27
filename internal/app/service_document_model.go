package app

import (
	"context"
	"fmt"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type documentModelOperation struct {
	action          quizmodel.Action
	strategy        string
	requestIDPrefix string
}

func (s service) newDocumentGenerationModelRequest(ctx context.Context, state quizcore.RuntimeState, operation documentModelOperation, appendContext func(context.Context, []quizmodel.ContextSection, quizcore.Scope) ([]quizmodel.ContextSection, error)) (quizmodel.Request, error) {
	scope := quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}

	payload, err := s.newQuizModelPayload(ctx, scope, state, operation.action)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err = appendContext(ctx, sections, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	request := s.documentGenerationModelRequest(scope, state, payload, sections, operation)
	request.ID = harnesshost.GenerationRequestID(ctx, request.ID)

	return request, nil
}

func (s service) documentGenerationModelRequest(scope quizcore.Scope, state quizcore.RuntimeState, payload quizmodel.Payload, sections []quizmodel.ContextSection, operation documentModelOperation) quizmodel.Request {
	return quizmodel.Request{
		ID:              fmt.Sprintf("%s-%s", operation.requestIDPrefix, specScopeID(scope)),
		Action:          operation.action,
		Strategy:        operation.strategy,
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract:  quizmodel.OutputContract{RequiredFields: payload.OutputContract.RequiredFields},
		CreatedAt:       s.utcNow(),
	}
}
