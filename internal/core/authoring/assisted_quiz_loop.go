package authoring

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AssistedQuizLoopStopReason defines a package boundary type.
type AssistedQuizLoopStopReason string

const (
	// AssistedQuizLoopStopNone defines a package boundary constant.
	AssistedQuizLoopStopNone AssistedQuizLoopStopReason = ""
	// AssistedQuizLoopStopCompleted defines a package boundary constant.
	AssistedQuizLoopStopCompleted AssistedQuizLoopStopReason = "completed"
	// AssistedQuizLoopStopUser defines a package boundary constant.
	AssistedQuizLoopStopUser AssistedQuizLoopStopReason = "user_stopped"
	// AssistedQuizLoopStopTurnLimit defines a package boundary constant.
	AssistedQuizLoopStopTurnLimit AssistedQuizLoopStopReason = "turn_limit"
	// AssistedQuizLoopStopBlocked defines a package boundary constant.
	AssistedQuizLoopStopBlocked AssistedQuizLoopStopReason = "blocked"
	// AssistedQuizLoopStopModelError defines a package boundary constant.
	AssistedQuizLoopStopModelError AssistedQuizLoopStopReason = "model_error"
)

// AssistedQuizLoopOptions describes a package boundary value.
type AssistedQuizLoopOptions struct {
	Config           AssistedAuthoringConfig
	Session          Session
	SessionPaths     SessionPaths
	ContextPackage   AssistedAuthoringContextPackage
	TranscriptEvents []AssistedTranscriptEvent
	Repository       SessionRepository
	Adapter          AssistedModelAdapter
	Now              func() time.Time
}

// AssistedQuizAskResult describes a package boundary value.
type AssistedQuizAskResult struct {
	Turn            AssistedAuthoringTurn
	Request         AssistedModelRequest
	Response        AssistedModelResponse
	InvocationPaths AssistedModelInvocationPaths
	StopReason      AssistedQuizLoopStopReason
}

// AssistedQuizAnswerOptions describes a package boundary value.
type AssistedQuizAnswerOptions struct {
	SessionID  string
	TurnID     string
	TurnIndex  int
	Intent     AssistedAuthoringIntent
	Question   AssistedAuthoringQuestion
	AnswerText string
	UserStop   bool
	Repository SessionRepository
	Now        func() time.Time
}

// AssistedQuizAnswerResult describes a package boundary value.
type AssistedQuizAnswerResult struct {
	Turn       AssistedAuthoringTurn
	StopReason AssistedQuizLoopStopReason
}

// AskNextAssistedQuizQuestion executes the corresponding package operation.
func AskNextAssistedQuizQuestion(
	ctx context.Context,
	opts AssistedQuizLoopOptions,
) (AssistedQuizAskResult, error) {
	err := validateAssistedQuizLoopOptions(opts)
	if err != nil {
		return AssistedQuizAskResult{}, err
	}

	now := assistedQuizNow(opts.Now)
	summary := summarizeAssistedQuizLoopEvents(opts.TranscriptEvents)
	stopReason := assistedQuizLoopStopReason(opts.Config, summary)
	nextIndex := summary.QuestionCount + 1

	if stopReason != AssistedQuizLoopStopNone {
		turn, err := assistedQuizStopTurn(ctx, opts, nextIndex, stopReason, now)
		if err != nil {
			return AssistedQuizAskResult{}, err
		}

		return AssistedQuizAskResult{Turn: turn, StopReason: stopReason}, nil
	}

	turnID := assistedQuizTurnID(nextIndex)
	request := assistedQuizModelRequest(opts, summary, turnID, nextIndex, now)

	invocation, err := invokeAndRecordAssistedQuizModel(ctx, opts, request)
	if err != nil {
		return AssistedQuizAskResult{}, err
	}

	turn := assistedQuizTurnFromModelResponse(opts, turnID, nextIndex, invocation.Response, now)

	err = appendAssistedQuizTurn(ctx, opts.Repository, turn)
	if err != nil {
		return AssistedQuizAskResult{}, err
	}

	return AssistedQuizAskResult{
		Turn:            turn,
		Request:         request,
		Response:        invocation.Response,
		InvocationPaths: invocation.Paths,
		StopReason:      assistedQuizResponseStopReason(invocation.Response),
	}, nil
}

// AnswerAssistedQuizQuestion executes the corresponding package operation.
func AnswerAssistedQuizQuestion(
	ctx context.Context,
	opts AssistedQuizAnswerOptions,
) (AssistedQuizAnswerResult, error) {
	err := validateAssistedQuizAnswerOptions(opts)
	if err != nil {
		return AssistedQuizAnswerResult{}, err
	}

	now := assistedQuizNow(opts.Now)
	turn := AssistedAuthoringTurn{
		ID:        opts.TurnID,
		SessionID: opts.SessionID,
		Index:     opts.TurnIndex,
		Status:    AssistedAuthoringTurnAnswered,
		Intent:    opts.Intent,
		Question:  opts.Question,
		Answer: AssistedAuthoringAnswer{
			QuestionID: opts.Question.ID,
			Text:       strings.TrimSpace(opts.AnswerText),
			AnsweredAt: now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = appendAssistedQuizTurn(ctx, opts.Repository, turn)
	if err != nil {
		return AssistedQuizAnswerResult{}, err
	}

	if opts.UserStop {
		return AssistedQuizAnswerResult{Turn: turn, StopReason: AssistedQuizLoopStopUser}, nil
	}

	return AssistedQuizAnswerResult{Turn: turn}, nil
}

func validateAssistedQuizLoopOptions(opts AssistedQuizLoopOptions) error {
	err := ValidateAssistedAuthoringConfig(opts.Config)
	if err != nil {
		return err
	}

	err = ValidateSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.Config.SessionID != opts.Session.ID {
		return fmt.Errorf("assisted quiz config session id does not match session")
	}

	if opts.Config.ProjectID != opts.Session.ProjectID {
		return fmt.Errorf("assisted quiz config project id does not match session")
	}

	if opts.Repository == nil {
		return fmt.Errorf("assisted quiz repository is required")
	}

	if opts.Adapter == nil {
		return fmt.Errorf("assisted quiz model adapter is required")
	}

	return nil
}

func assistedQuizModelRequest(
	opts AssistedQuizLoopOptions,
	summary AssistedTranscriptSummary,
	turnID string,
	nextIndex int,
	now time.Time,
) AssistedModelRequest {
	return AssistedModelRequest{
		ID:        assistedQuizRequestID(nextIndex),
		SessionID: opts.Config.SessionID,
		TurnID:    turnID,
		Intent:    opts.Config.Intent,
		Target:    opts.Config.Target,
		Context:   RenderAssistedAuthoringContextPackage(opts.ContextPackage),
		Prompt:    assistedQuizQuestionPrompt(opts.Config, summary),
		CreatedAt: now,
	}
}

type assistedQuizInvocation struct {
	Response AssistedModelResponse
	Paths    AssistedModelInvocationPaths
}

func invokeAndRecordAssistedQuizModel(ctx context.Context, opts AssistedQuizLoopOptions, request AssistedModelRequest) (assistedQuizInvocation, error) {
	response, err := InvokeAssistedModel(ctx, opts.Adapter, request)
	if err != nil {
		return assistedQuizInvocation{}, err
	}

	paths, err := PlanAssistedModelInvocationArtifacts(opts.SessionPaths, request)
	if err != nil {
		return assistedQuizInvocation{}, err
	}

	err = RecordAssistedModelInvocation(ctx, AssistedModelInvocationArtifacts{
		Paths:    paths,
		Request:  request,
		Response: response,
	})
	if err != nil {
		return assistedQuizInvocation{}, err
	}

	return assistedQuizInvocation{Response: response, Paths: paths}, nil
}

func validateAssistedQuizAnswerOptions(opts AssistedQuizAnswerOptions) error {
	if strings.TrimSpace(opts.SessionID) == "" {
		return fmt.Errorf("assisted quiz answer session id is required")
	}

	if strings.TrimSpace(opts.TurnID) == "" {
		return fmt.Errorf("assisted quiz answer turn id is required")
	}

	if opts.TurnIndex <= 0 {
		return fmt.Errorf("assisted quiz answer turn index must be greater than zero")
	}

	if !opts.Intent.Valid() {
		return fmt.Errorf("assisted quiz answer intent is invalid: %s", opts.Intent)
	}

	err := validateAssistedAuthoringQuestion(opts.Question)
	if err != nil {
		return err
	}

	if strings.TrimSpace(opts.AnswerText) == "" {
		return fmt.Errorf("assisted quiz answer text is required")
	}

	if opts.Repository == nil {
		return fmt.Errorf("assisted quiz repository is required")
	}

	return nil
}

func summarizeAssistedQuizLoopEvents(events []AssistedTranscriptEvent) AssistedTranscriptSummary {
	summary := AssistedTranscriptSummary{}
	questions := map[string]AssistedTranscriptQuestionGap{}
	answered := map[string]struct{}{}

	for _, event := range events {
		switch event.Type {
		case AssistedTranscriptEventQuestion:
			summary.QuestionCount++
			questions[event.Question.ID] = AssistedTranscriptQuestionGap{
				TurnID:     event.TurnID,
				TurnIndex:  event.TurnIndex,
				QuestionID: event.Question.ID,
				Prompt:     event.Question.Prompt,
			}
		case AssistedTranscriptEventAnswer:
			summary.AnswerCount++
			answered[event.Answer.QuestionID] = struct{}{}
		case AssistedTranscriptEventResult:
			summary.ResultCount++
			if event.Result.Summary != "" {
				summary.LastResultSummary = event.Result.Summary
			}
		case AssistedTranscriptEventState:
			summary.StateCount++
		}
	}

	for questionID, question := range questions {
		_, ok := answered[questionID]
		if !ok {
			summary.UnansweredQuestions = append(summary.UnansweredQuestions, question)
		}
	}

	return summary
}

func assistedQuizLoopStopReason(
	config AssistedAuthoringConfig,
	summary AssistedTranscriptSummary,
) AssistedQuizLoopStopReason {
	if summary.QuestionCount >= config.MaxTurns {
		return AssistedQuizLoopStopTurnLimit
	}

	followUps := summary.QuestionCount - 1
	if followUps >= config.MaxFollowUps && summary.QuestionCount > 0 {
		return AssistedQuizLoopStopTurnLimit
	}

	return AssistedQuizLoopStopNone
}

func assistedQuizStopTurn(
	ctx context.Context,
	opts AssistedQuizLoopOptions,
	index int,
	reason AssistedQuizLoopStopReason,
	now time.Time,
) (AssistedAuthoringTurn, error) {
	turn := AssistedAuthoringTurn{
		ID:        assistedQuizTurnID(index),
		SessionID: opts.Config.SessionID,
		Index:     index,
		Status:    AssistedAuthoringTurnCompleted,
		Intent:    opts.Config.Intent,
		Question: AssistedAuthoringQuestion{
			ID:       fmt.Sprintf("%s-stop", assistedQuizTurnID(index)),
			Prompt:   "The assisted quiz loop stopped before asking another question.",
			Required: false,
		},
		Result: AssistedAuthoringTurnResult{
			CompletionReason: AssistedAuthoringCompletionTurnLimit,
			Summary:          string(reason),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := appendAssistedQuizTurn(ctx, opts.Repository, turn)
	if err != nil {
		return AssistedAuthoringTurn{}, err
	}

	return turn, nil
}

func assistedQuizTurnFromModelResponse(
	opts AssistedQuizLoopOptions,
	turnID string,
	index int,
	response AssistedModelResponse,
	now time.Time,
) AssistedAuthoringTurn {
	turn := AssistedAuthoringTurn{
		ID:        turnID,
		SessionID: opts.Config.SessionID,
		Index:     index,
		Status:    AssistedAuthoringTurnAsked,
		Intent:    opts.Config.Intent,
		Question:  response.Question,
		CreatedAt: now,
		UpdatedAt: response.CreatedAt,
	}

	if response.Status == AssistedModelResponseBlocked {
		turn.Status = AssistedAuthoringTurnBlocked
		turn.Question = AssistedAuthoringQuestion{
			ID:       fmt.Sprintf("%s-blocked", turnID),
			Prompt:   "The assisted quiz loop was blocked before asking another question.",
			Required: false,
		}
		turn.Result = response.Result
	}

	return turn
}

func appendAssistedQuizTurn(
	ctx context.Context,
	repository SessionRepository,
	turn AssistedAuthoringTurn,
) error {
	events, err := AssistedTranscriptEventsFromTurn(turn)
	if err != nil {
		return err
	}

	return repository.AppendTranscriptEvents(ctx, turn.SessionID, events)
}

func assistedQuizResponseStopReason(response AssistedModelResponse) AssistedQuizLoopStopReason {
	if response.Status == AssistedModelResponseBlocked {
		return AssistedQuizLoopStopBlocked
	}

	return AssistedQuizLoopStopNone
}

func assistedQuizQuestionPrompt(
	config AssistedAuthoringConfig,
	summary AssistedTranscriptSummary,
) string {
	return fmt.Sprintf(
		"Ask the next bounded authoring question for target %s and intent %s. Questions already asked: %d. Answers received: %d.",
		config.Target,
		config.Intent,
		summary.QuestionCount,
		summary.AnswerCount,
	)
}

func assistedQuizNow(now func() time.Time) time.Time {
	if now == nil {
		return time.Now().UTC()
	}

	return now().UTC()
}

func assistedQuizTurnID(index int) string {
	return fmt.Sprintf("turn-%03d", index)
}

func assistedQuizRequestID(index int) string {
	return fmt.Sprintf("request-%03d", index)
}
