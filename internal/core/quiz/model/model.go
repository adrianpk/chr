package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizModelAction defines a package boundary type.
type QuizModelAction string

const (
	// QuizModelActionNextQuestion defines a package boundary constant.
	QuizModelActionNextQuestion QuizModelAction = "quiz_next_question"
	// QuizModelActionDraftSpec defines a package boundary constant.
	QuizModelActionDraftSpec QuizModelAction = "quiz_to_draft_spec"
	// QuizModelActionGenerateSpec defines a package boundary constant.
	QuizModelActionGenerateSpec QuizModelAction = "spec_generation"
	// QuizModelActionGenerateAcceptanceFeature defines a package boundary constant.
	QuizModelActionGenerateAcceptanceFeature QuizModelAction = "acceptance_feature_generation"
	// QuizModelActionGeneratePlan defines a package boundary constant.
	QuizModelActionGeneratePlan QuizModelAction = "plan_generation"
	// QuizModelActionGeneratePlanOutline defines a package boundary constant.
	QuizModelActionGeneratePlanOutline QuizModelAction = "plan_outline_generation"
	// QuizModelActionGeneratePlanSliceDetail defines a package boundary constant.
	QuizModelActionGeneratePlanSliceDetail QuizModelAction = "plan_slice_detail_generation"
	// QuizModelActionMaterializePlanTracking defines a package boundary constant.
	QuizModelActionMaterializePlanTracking QuizModelAction = "plan_tracking_materialization"
)

// QuizModelResponseStatus defines a package boundary type.
type QuizModelResponseStatus string

const (
	// QuizModelResponseReady defines a package boundary constant.
	QuizModelResponseReady QuizModelResponseStatus = "ready"
	// QuizModelResponseBlocked defines a package boundary constant.
	QuizModelResponseBlocked QuizModelResponseStatus = "blocked"
)

// QuizModelRequest describes a package boundary value.
type QuizModelRequest struct {
	ID              string
	Action          QuizModelAction
	Strategy        string
	Scope           quizcore.Scope
	Mode            quizcore.Mode
	RuntimeState    quizcore.RuntimeState
	Payload         QuizModelPayload
	ContextSections []QuizModelContextSection
	OutputContract  QuizModelOutputContract
	CreatedAt       time.Time
}

// QuizModelContextSection describes a package boundary value.
type QuizModelContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
	Metadata  map[string]string
}

// QuizModelOutputContract describes a package boundary value.
type QuizModelOutputContract struct {
	RequiredFields []string
}

// QuizModelResponse describes a package boundary value.
type QuizModelResponse struct {
	ID           string
	RequestID    string
	Provider     string
	Status       QuizModelResponseStatus
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

// QuizModelAdapter defines a package boundary interface.
type QuizModelAdapter interface {
	InvokeQuizModel(ctx context.Context, request QuizModelRequest) (QuizModelResponse, error)
}

// InvokeQuizModel executes the corresponding package operation.
func InvokeQuizModel(ctx context.Context, adapter QuizModelAdapter, request QuizModelRequest) (QuizModelResponse, error) {
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

// ValidateQuizModelRequest validates the corresponding package boundary value.
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

	err := quizcore.ValidateScope(request.Scope)
	if err != nil {
		return err
	}

	err = quizcore.ValidateRuntimeState(request.RuntimeState)
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

// ValidateQuizModelResponse validates the corresponding package boundary value.
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
func (a QuizModelAction) Valid() bool {
	switch a {
	case QuizModelActionNextQuestion,
		QuizModelActionDraftSpec,
		QuizModelActionGenerateSpec,
		QuizModelActionGenerateAcceptanceFeature,
		QuizModelActionGeneratePlan,
		QuizModelActionGeneratePlanOutline,
		QuizModelActionGeneratePlanSliceDetail,
		QuizModelActionMaterializePlanTracking:
		return true
	default:
		return false
	}
}

func validateQuizModelReadyResponse(response QuizModelResponse) error {
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
func (s QuizModelResponseStatus) Valid() bool {
	switch s {
	case QuizModelResponseReady, QuizModelResponseBlocked:
		return true
	default:
		return false
	}
}
