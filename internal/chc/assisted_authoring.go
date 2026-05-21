package chc

import (
	"fmt"
	"time"
)

type AssistedAuthoringIntent string

const (
	AssistedAuthoringIntentDiscover   AssistedAuthoringIntent = "discover"
	AssistedAuthoringIntentClarify    AssistedAuthoringIntent = "clarify"
	AssistedAuthoringIntentCurate     AssistedAuthoringIntent = "curate"
	AssistedAuthoringIntentSynthesize AssistedAuthoringIntent = "synthesize"
)

type AssistedAuthoringTurnStatus string

const (
	AssistedAuthoringTurnPending   AssistedAuthoringTurnStatus = "pending"
	AssistedAuthoringTurnAsked     AssistedAuthoringTurnStatus = "asked"
	AssistedAuthoringTurnAnswered  AssistedAuthoringTurnStatus = "answered"
	AssistedAuthoringTurnCompleted AssistedAuthoringTurnStatus = "completed"
	AssistedAuthoringTurnBlocked   AssistedAuthoringTurnStatus = "blocked"
)

type AssistedAuthoringCompletionReason string

const (
	AssistedAuthoringCompletionEnoughContext AssistedAuthoringCompletionReason = "enough_context"
	AssistedAuthoringCompletionUserStopped   AssistedAuthoringCompletionReason = "user_stopped"
	AssistedAuthoringCompletionTurnLimit     AssistedAuthoringCompletionReason = "turn_limit"
)

type AssistedAuthoringBlockedReason string

const (
	AssistedAuthoringBlockedMissingContext AssistedAuthoringBlockedReason = "missing_context"
	AssistedAuthoringBlockedModelFailure   AssistedAuthoringBlockedReason = "model_failure"
	AssistedAuthoringBlockedInvalidAnswer  AssistedAuthoringBlockedReason = "invalid_answer"
)

type AssistedAuthoringConfig struct {
	SessionID        string
	ProjectID        string
	Target           AuthoringTarget
	Intent           AssistedAuthoringIntent
	MaxTurns         int
	MaxFollowUps     int
	RequireArtifacts bool
}

type AssistedAuthoringTurn struct {
	ID        string
	SessionID string
	Index     int
	Status    AssistedAuthoringTurnStatus
	Intent    AssistedAuthoringIntent
	Question  AssistedAuthoringQuestion
	Answer    AssistedAuthoringAnswer
	Result    AssistedAuthoringTurnResult
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssistedAuthoringQuestion struct {
	ID        string
	Prompt    string
	Rationale string
	Required  bool
}

type AssistedAuthoringAnswer struct {
	QuestionID string
	Text       string
	AnsweredAt time.Time
}

type AssistedAuthoringTurnResult struct {
	CompletionReason AssistedAuthoringCompletionReason
	BlockedReason    AssistedAuthoringBlockedReason
	Summary          string
	NextIntent       AssistedAuthoringIntent
}

func ValidateAssistedAuthoringConfig(config AssistedAuthoringConfig) error {
	if config.SessionID == "" {
		return fmt.Errorf("assisted authoring session id is required")
	}

	if config.ProjectID == "" {
		return fmt.Errorf("assisted authoring project id is required")
	}

	if !config.Target.Valid() {
		return fmt.Errorf("assisted authoring target is invalid: %s", config.Target)
	}

	if !config.Intent.Valid() {
		return fmt.Errorf("assisted authoring intent is invalid: %s", config.Intent)
	}

	if config.MaxTurns <= 0 {
		return fmt.Errorf("assisted authoring max turns must be greater than zero")
	}

	if config.MaxFollowUps < 0 {
		return fmt.Errorf("assisted authoring max follow ups cannot be negative")
	}

	return nil
}

func ValidateAssistedAuthoringTurn(turn AssistedAuthoringTurn) error {
	if turn.ID == "" {
		return fmt.Errorf("assisted authoring turn id is required")
	}

	if turn.SessionID == "" {
		return fmt.Errorf("assisted authoring turn session id is required")
	}

	if turn.Index <= 0 {
		return fmt.Errorf("assisted authoring turn index must be greater than zero")
	}

	if !turn.Status.Valid() {
		return fmt.Errorf("assisted authoring turn status is invalid: %s", turn.Status)
	}

	if !turn.Intent.Valid() {
		return fmt.Errorf("assisted authoring turn intent is invalid: %s", turn.Intent)
	}

	err := validateAssistedAuthoringTurnPayload(turn)
	if err != nil {
		return err
	}

	return validateAssistedAuthoringTurnTimes(turn)
}

func (i AssistedAuthoringIntent) Valid() bool {
	switch i {
	case AssistedAuthoringIntentDiscover,
		AssistedAuthoringIntentClarify,
		AssistedAuthoringIntentCurate,
		AssistedAuthoringIntentSynthesize:
		return true
	default:
		return false
	}
}

func (s AssistedAuthoringTurnStatus) Valid() bool {
	switch s {
	case AssistedAuthoringTurnPending,
		AssistedAuthoringTurnAsked,
		AssistedAuthoringTurnAnswered,
		AssistedAuthoringTurnCompleted,
		AssistedAuthoringTurnBlocked:
		return true
	default:
		return false
	}
}

func (r AssistedAuthoringCompletionReason) Valid() bool {
	switch r {
	case AssistedAuthoringCompletionEnoughContext,
		AssistedAuthoringCompletionUserStopped,
		AssistedAuthoringCompletionTurnLimit:
		return true
	default:
		return false
	}
}

func (r AssistedAuthoringBlockedReason) Valid() bool {
	switch r {
	case AssistedAuthoringBlockedMissingContext,
		AssistedAuthoringBlockedModelFailure,
		AssistedAuthoringBlockedInvalidAnswer:
		return true
	default:
		return false
	}
}

func validateAssistedAuthoringTurnPayload(turn AssistedAuthoringTurn) error {
	if turn.Status == AssistedAuthoringTurnPending {
		return nil
	}

	err := validateAssistedAuthoringQuestion(turn.Question)
	if err != nil {
		return err
	}

	if turn.Status == AssistedAuthoringTurnAnswered {
		return validateAssistedAuthoringAnswer(turn.Question, turn.Answer)
	}

	if turn.Status == AssistedAuthoringTurnCompleted || turn.Status == AssistedAuthoringTurnBlocked {
		return validateAssistedAuthoringTurnResult(turn.Status, turn.Result)
	}

	return nil
}

func validateAssistedAuthoringQuestion(question AssistedAuthoringQuestion) error {
	if question.ID == "" {
		return fmt.Errorf("assisted authoring question id is required")
	}

	if question.Prompt == "" {
		return fmt.Errorf("assisted authoring question prompt is required")
	}

	return nil
}

func validateAssistedAuthoringAnswer(
	question AssistedAuthoringQuestion,
	answer AssistedAuthoringAnswer,
) error {
	if answer.QuestionID == "" {
		return fmt.Errorf("assisted authoring answer question id is required")
	}

	if answer.QuestionID != question.ID {
		return fmt.Errorf("assisted authoring answer question id does not match question")
	}

	if answer.Text == "" {
		return fmt.Errorf("assisted authoring answer text is required")
	}

	if answer.AnsweredAt.IsZero() {
		return fmt.Errorf("assisted authoring answer time is required")
	}

	return nil
}

func validateAssistedAuthoringTurnResult(
	status AssistedAuthoringTurnStatus,
	result AssistedAuthoringTurnResult,
) error {
	if status == AssistedAuthoringTurnCompleted {
		if !result.CompletionReason.Valid() {
			return fmt.Errorf("assisted authoring completion reason is invalid: %s", result.CompletionReason)
		}
	}

	if status == AssistedAuthoringTurnBlocked {
		if !result.BlockedReason.Valid() {
			return fmt.Errorf("assisted authoring blocked reason is invalid: %s", result.BlockedReason)
		}
	}

	if result.NextIntent != "" && !result.NextIntent.Valid() {
		return fmt.Errorf("assisted authoring next intent is invalid: %s", result.NextIntent)
	}

	return nil
}

func validateAssistedAuthoringTurnTimes(turn AssistedAuthoringTurn) error {
	if turn.CreatedAt.IsZero() {
		return fmt.Errorf("assisted authoring turn created time is required")
	}

	if turn.UpdatedAt.IsZero() {
		return fmt.Errorf("assisted authoring turn updated time is required")
	}

	if turn.UpdatedAt.Before(turn.CreatedAt) {
		return fmt.Errorf("assisted authoring turn updated before it was created")
	}

	return nil
}
