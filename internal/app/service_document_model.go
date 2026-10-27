package app

import (
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

type documentModelOperation struct {
	action          quizmodel.Action
	strategy        string
	requestIDPrefix string
}

func (s service) newDocumentGenerationModelRequest(state quizcore.RuntimeState, operation documentModelOperation, appendContext func([]quizmodel.ContextSection, quizcore.Scope) ([]quizmodel.ContextSection, error)) (quizmodel.Request, error) {
	scope := quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}

	payload, err := s.newQuizModelPayload(scope, state, operation.action)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err = appendContext(sections, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	return quizmodel.Request{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("%s-%s", operation.requestIDPrefix, specScopeID(scope))),
		Action:          operation.action,
		Strategy:        operation.strategy,
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract:  quizmodel.OutputContract{RequiredFields: payload.OutputContract.RequiredFields},
		CreatedAt:       s.utcNow(),
	}, nil
}
