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
	ContextSections []QuizModelContextSection
	OutputContract  QuizModelOutputContract
	CreatedAt       time.Time
}

type QuizModelContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
}

type QuizModelOutputContract struct {
	RequiredFields []string
}

type QuizModelResponse struct {
	ID        string
	RequestID string
	Provider  string
	Status    QuizModelResponseStatus
	Question  QuizQuestion
	RawText   string
	CreatedAt time.Time
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
		err := ValidateQuizQuestion(response.Question)
		if err != nil {
			return err
		}
	}

	if response.CreatedAt.IsZero() {
		return fmt.Errorf("quiz model response created time is required")
	}

	return nil
}

func (a QuizModelAction) Valid() bool {
	switch a {
	case QuizModelActionNextQuestion:
		return true
	default:
		return false
	}
}

func (s QuizModelResponseStatus) Valid() bool {
	switch s {
	case QuizModelResponseReady, QuizModelResponseBlocked:
		return true
	default:
		return false
	}
}
