package chc

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type QuizModelAction string

const (
	QuizModelActionNextQuestion QuizModelAction = "quiz_next_question"
	QuizModelActionDraftSpec    QuizModelAction = "quiz_to_draft_spec"
	QuizModelActionGenerateSpec QuizModelAction = "spec_generation"
)

type QuizModelResponseStatus string

const (
	QuizModelResponseReady   QuizModelResponseStatus = "ready"
	QuizModelResponseBlocked QuizModelResponseStatus = "blocked"
)

type QuizModelRequest struct {
	ID              string
	Action          QuizModelAction
	Strategy        string
	Scope           QuizScope
	Mode            QuizMode
	RuntimeState    QuizRuntimeState
	Payload         QuizModelPayload
	ContextSections []QuizModelContextSection
	OutputContract  QuizModelOutputContract
	CreatedAt       time.Time
}

type QuizModelContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
	Metadata  map[string]string
}

type QuizModelOutputContract struct {
	RequiredFields []string
}

type QuizModelResponse struct {
	ID         string
	RequestID  string
	Provider   string
	Status     QuizModelResponseStatus
	Question   QuizQuestion
	Candidates []QuizCandidate
	DraftSpec  QuizDraftSpec
	RawText    string
	CreatedAt  time.Time
}

type QuizModelAdapter interface {
	InvokeQuizModel(ctx context.Context, request QuizModelRequest) (QuizModelResponse, error)
}

func InvokeQuizModel(
	ctx context.Context,
	adapter QuizModelAdapter,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	err := ValidateQuizModelRequest(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	if adapter == nil {
		return QuizModelResponse{}, fmt.Errorf("quiz model adapter is required")
	}

	response, err := adapter.InvokeQuizModel(ctx, request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	err = ValidateQuizModelResponse(response)
	if err != nil {
		return QuizModelResponse{}, err
	}

	return response, nil
}

func ValidateQuizModelRequest(request QuizModelRequest) error {
	if strings.TrimSpace(request.ID) == "" {
		return fmt.Errorf("quiz model request id is required")
	}

	if !request.Action.Valid() {
		return fmt.Errorf("quiz model request action is invalid: %s", request.Action)
	}

	if strings.TrimSpace(request.Strategy) == "" {
		return fmt.Errorf("quiz model request strategy is required")
	}

	if !request.Mode.Valid() {
		return fmt.Errorf("quiz model request mode is invalid: %s", request.Mode)
	}

	err := ValidateQuizScope(request.Scope)
	if err != nil {
		return err
	}

	err = ValidateQuizRuntimeState(request.RuntimeState)
	if err != nil {
		return err
	}

	return validateQuizModelRequestPayload(request)
}

func validateQuizModelRequestPayload(request QuizModelRequest) error {
	err := ValidateQuizModelPayload(request.Payload)
	if err != nil {
		return err
	}

	if len(request.ContextSections) == 0 {
		return fmt.Errorf("quiz model request context sections are required")
	}

	if len(request.OutputContract.RequiredFields) == 0 {
		return fmt.Errorf("quiz model request output contract is required")
	}

	if request.CreatedAt.IsZero() {
		return fmt.Errorf("quiz model request created time is required")
	}

	return nil
}

func ValidateQuizModelResponse(response QuizModelResponse) error {
	if strings.TrimSpace(response.ID) == "" {
		return fmt.Errorf("quiz model response id is required")
	}

	if strings.TrimSpace(response.RequestID) == "" {
		return fmt.Errorf("quiz model response request id is required")
	}

	if strings.TrimSpace(response.Provider) == "" {
		return fmt.Errorf("quiz model response provider is required")
	}

	if !response.Status.Valid() {
		return fmt.Errorf("quiz model response status is invalid: %s", response.Status)
	}

	if response.Status == QuizModelResponseReady {
		err := validateQuizModelReadyResponse(response)
		if err != nil {
			return err
		}
	}

	for index, candidate := range response.Candidates {
		err := ValidateQuizCandidate(candidate)
		if err != nil {
			return fmt.Errorf("quiz model response candidate %d: %w", index+1, err)
		}
	}

	if response.CreatedAt.IsZero() {
		return fmt.Errorf("quiz model response created time is required")
	}

	return nil
}

func (a QuizModelAction) Valid() bool {
	switch a {
	case QuizModelActionNextQuestion, QuizModelActionDraftSpec, QuizModelActionGenerateSpec:
		return true
	default:
		return false
	}
}

func validateQuizModelReadyResponse(response QuizModelResponse) error {
	if response.DraftSpec.HasContent() {
		return ValidateQuizDraftSpecModelContent(response.DraftSpec)
	}

	return ValidateQuizQuestion(response.Question)
}

func (s QuizModelResponseStatus) Valid() bool {
	switch s {
	case QuizModelResponseReady, QuizModelResponseBlocked:
		return true
	default:
		return false
	}
}
