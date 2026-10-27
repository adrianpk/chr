package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Action defines a package boundary type.
type Action string

const (
	// ActionNextQuestion defines a package boundary constant.
	ActionNextQuestion Action = "quiz_next_question"
	// ActionDraftSpec defines a package boundary constant.
	ActionDraftSpec Action = "quiz_to_draft_spec"
	// ActionGenerateSpec defines a package boundary constant.
	ActionGenerateSpec Action = "spec_generation"
	// ActionGenerateAcceptanceFeature defines a package boundary constant.
	ActionGenerateAcceptanceFeature Action = "acceptance_feature_generation"
	// ActionGeneratePlan defines a package boundary constant.
	ActionGeneratePlan Action = "plan_generation"
	// ActionGeneratePlanOutline defines a package boundary constant.
	ActionGeneratePlanOutline Action = "plan_outline_generation"
	// ActionGeneratePlanSliceDetail defines a package boundary constant.
	ActionGeneratePlanSliceDetail Action = "plan_slice_detail_generation"
	// ActionMaterializePlanTracking defines a package boundary constant.
	ActionMaterializePlanTracking Action = "plan_tracking_materialization"
)

// ResponseStatus defines a package boundary type.
type ResponseStatus string

const (
	// ResponseReady defines a package boundary constant.
	ResponseReady ResponseStatus = "ready"
	// ResponseBlocked defines a package boundary constant.
	ResponseBlocked ResponseStatus = "blocked"
)

// Request describes a package boundary value.
type Request struct {
	ID              string
	Action          Action
	Strategy        string
	Scope           quizcore.Scope
	Mode            quizcore.Mode
	RuntimeState    quizcore.RuntimeState
	Payload         Payload
	ContextSections []ContextSection
	OutputContract  OutputContract
	CreatedAt       time.Time
}

// ContextSection describes a package boundary value.
type ContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
	Metadata  map[string]string
}

// OutputContract describes a package boundary value.
type OutputContract struct {
	RequiredFields []string
}

// Response describes a package boundary value.
type Response struct {
	ID           string
	RequestID    string
	Provider     string
	Status       ResponseStatus
	Question     quizcore.Question
	Candidates   []quizcore.Candidate
	DraftSpec    generationcore.QuizDraftSpec
	Feature      generationcore.AcceptanceFeatureGeneration
	Plan         generationcore.GeneratedPlan
	PlanOutline  generationcore.GeneratedPlanOutline
	SliceDetail  generationcore.GeneratedPlanSliceDetail
	PlanTracking generationcore.GeneratedPlanTracking
	RawText      string
	CreatedAt    time.Time
}

// Adapter defines a package boundary interface.
type Adapter interface {
	Invoke(ctx context.Context, request Request) (Response, error)
}

// Invoke executes the corresponding package operation.
func Invoke(ctx context.Context, adapter Adapter, request Request) (Response, error) {
	err := ValidateRequest(request)
	if err != nil {
		return Response{}, err
	}

	if adapter == nil {
		return Response{}, fmt.Errorf("quiz model adapter is required")
	}

	response, err := adapter.Invoke(ctx, request)
	if err != nil {
		return Response{}, err
	}

	err = ValidateResponse(response)
	if err != nil {
		return Response{}, err
	}

	return response, nil
}

// ValidateRequest validates the corresponding package boundary value.
func ValidateRequest(request Request) error {
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

	err := quizcore.ValidateScope(request.Scope)
	if err != nil {
		return err
	}

	err = quizcore.ValidateRuntimeState(request.RuntimeState)
	if err != nil {
		return err
	}

	return validateRequestPayload(request)
}

func validateRequestPayload(request Request) error {
	err := ValidatePayload(request.Payload)
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

// ValidateResponse validates the corresponding package boundary value.
func ValidateResponse(response Response) error {
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

	if response.Status == ResponseReady {
		err := validateReadyResponse(response)
		if err != nil {
			return err
		}
	}

	for index, candidate := range response.Candidates {
		err := quizcore.ValidateCandidate(candidate)
		if err != nil {
			return fmt.Errorf("quiz model response candidate %d: %w", index+1, err)
		}
	}

	if response.CreatedAt.IsZero() {
		return fmt.Errorf("quiz model response created time is required")
	}

	return nil
}

// Valid reports whether the value is one of the supported constants.
func (a Action) Valid() bool {
	switch a {
	case ActionNextQuestion,
		ActionDraftSpec,
		ActionGenerateSpec,
		ActionGenerateAcceptanceFeature,
		ActionGeneratePlan,
		ActionGeneratePlanOutline,
		ActionGeneratePlanSliceDetail,
		ActionMaterializePlanTracking:
		return true
	default:
		return false
	}
}

func validateReadyResponse(response Response) error {
	if response.DraftSpec.HasContent() {
		return generationcore.ValidateQuizDraftSpecModelContent(response.DraftSpec)
	}

	if response.Feature.HasContent() {
		_, err := generationcore.ValidateAcceptanceFeatureGeneration(response.Feature)

		return err
	}

	if response.Plan.HasContent() {
		return generationcore.ValidateGeneratedPlanModelContent(response.Plan)
	}

	if response.PlanOutline.HasContent() {
		return generationcore.ValidateGeneratedPlanOutline(response.PlanOutline)
	}

	if response.SliceDetail.HasContent() {
		return generationcore.ValidateGeneratedPlanSliceDetailContent(response.SliceDetail)
	}

	if response.PlanTracking.HasContent() {
		return generationcore.ValidateGeneratedPlanTrackingModelContent(response.PlanTracking)
	}

	return quizcore.ValidateQuestion(response.Question)
}

// Valid reports whether the value is one of the supported constants.
func (s ResponseStatus) Valid() bool {
	switch s {
	case ResponseReady, ResponseBlocked:
		return true
	default:
		return false
	}
}
