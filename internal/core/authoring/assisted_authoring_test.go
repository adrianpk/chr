package authoring

import (
	"testing"
	"time"
)

func TestValidateAssistedAuthoringConfigAcceptsValidConfig(t *testing.T) {
	config := validAssistedAuthoringConfig()

	err := ValidateAssistedAuthoringConfig(config)
	if err != nil {
		t.Fatalf("ValidateAssistedAuthoringConfig returned error: %v", err)
	}
}

func TestValidateAssistedAuthoringConfigRejectsInvalidConfig(t *testing.T) {
	for _, test := range invalidAssistedAuthoringConfigTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAssistedAuthoringConfig(test.config)
			if err == nil {
				t.Fatal("ValidateAssistedAuthoringConfig returned nil error")
			}
		})
	}
}

func invalidAssistedAuthoringConfigTests() []struct {
	name   string
	config AssistedAuthoringConfig
} {
	return []struct {
		name   string
		config AssistedAuthoringConfig
	}{
		{name: "session id", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.SessionID = "" })},
		{name: "project id", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.ProjectID = "" })},
		{name: "target", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.Target = Target("unknown") })},
		{name: "intent", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.Intent = AssistedAuthoringIntent("unknown") })},
		{name: "max turns", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.MaxTurns = 0 })},
		{name: "max follow ups", config: assistedAuthoringConfigWith(func(config *AssistedAuthoringConfig) { config.MaxFollowUps = -1 })},
	}
}

func TestValidateAssistedAuthoringTurnAcceptsValidStates(t *testing.T) {
	for _, turn := range []AssistedAuthoringTurn{
		validAssistedAuthoringTurn(AssistedAuthoringTurnPending),
		validAssistedAuthoringTurn(AssistedAuthoringTurnAsked),
		validAssistedAuthoringTurn(AssistedAuthoringTurnAnswered),
		validAssistedAuthoringTurn(AssistedAuthoringTurnCompleted),
		validAssistedAuthoringTurn(AssistedAuthoringTurnBlocked),
	} {
		t.Run(string(turn.Status), func(t *testing.T) {
			err := ValidateAssistedAuthoringTurn(turn)
			if err != nil {
				t.Fatalf("ValidateAssistedAuthoringTurn returned error: %v", err)
			}
		})
	}
}

func TestValidateAssistedAuthoringTurnRejectsInvalidTurn(t *testing.T) {
	for _, test := range invalidAssistedAuthoringTurnTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAssistedAuthoringTurn(test.turn)
			if err == nil {
				t.Fatal("ValidateAssistedAuthoringTurn returned nil error")
			}
		})
	}
}

func invalidAssistedAuthoringTurnTests() []struct {
	name string
	turn AssistedAuthoringTurn
} {
	return []struct {
		name string
		turn AssistedAuthoringTurn
	}{
		{name: "id", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.ID = "" })},
		{name: "session id", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.SessionID = "" })},
		{name: "index", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.Index = 0 })},
		{name: "status", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.Status = AssistedAuthoringTurnStatus("unknown") })},
		{name: "intent", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.Intent = AssistedAuthoringIntent("unknown") })},
		{name: "question", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.Question.Prompt = "" })},
		{name: "answer", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.Answer.Text = "" })},
		{name: "completion reason", turn: assistedAuthoringCompletedTurnWith(func(turn *AssistedAuthoringTurn) {
			turn.Result.CompletionReason = AssistedAuthoringCompletionReason("unknown")
		})},
		{name: "blocked reason", turn: assistedAuthoringBlockedTurnWith(func(turn *AssistedAuthoringTurn) {
			turn.Result.BlockedReason = AssistedAuthoringBlockedReason("unknown")
		})},
		{name: "created at", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.CreatedAt = time.Time{} })},
		{name: "updated at", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.UpdatedAt = time.Time{} })},
		{name: "time order", turn: assistedAuthoringTurnWith(func(turn *AssistedAuthoringTurn) { turn.UpdatedAt = turn.CreatedAt.Add(-time.Minute) })},
	}
}

func validAssistedAuthoringConfig() AssistedAuthoringConfig {
	return AssistedAuthoringConfig{
		SessionID:    "auth-1",
		ProjectID:    "project-1",
		Target:       TargetSpec,
		Intent:       AssistedAuthoringIntentDiscover,
		MaxTurns:     6,
		MaxFollowUps: 2,
	}
}

func assistedAuthoringConfigWith(update func(*AssistedAuthoringConfig)) AssistedAuthoringConfig {
	config := validAssistedAuthoringConfig()
	update(&config)

	return config
}

func validAssistedAuthoringTurn(status AssistedAuthoringTurnStatus) AssistedAuthoringTurn {
	now := time.Date(2026, 5, 21, 16, 30, 0, 0, time.UTC)
	turn := AssistedAuthoringTurn{
		ID:        "turn-1",
		SessionID: "auth-1",
		Index:     1,
		Status:    status,
		Intent:    AssistedAuthoringIntentDiscover,
		Question: AssistedAuthoringQuestion{
			ID:        "q-1",
			Prompt:    "What should this spec capture?",
			Rationale: "The assistant needs the core problem statement.",
			Required:  true,
		},
		Answer: AssistedAuthoringAnswer{
			QuestionID: "q-1",
			Text:       "It should capture the retry behavior.",
			AnsweredAt: now,
		},
		Result: AssistedAuthoringTurnResult{
			CompletionReason: AssistedAuthoringCompletionEnoughContext,
			BlockedReason:    AssistedAuthoringBlockedMissingContext,
			Summary:          "Retry behavior is the central topic.",
			NextIntent:       AssistedAuthoringIntentClarify,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if status == AssistedAuthoringTurnPending {
		turn.Question = AssistedAuthoringQuestion{}
		turn.Answer = AssistedAuthoringAnswer{}
		turn.Result = AssistedAuthoringTurnResult{}
	}

	return turn
}

func assistedAuthoringTurnWith(update func(*AssistedAuthoringTurn)) AssistedAuthoringTurn {
	turn := validAssistedAuthoringTurn(AssistedAuthoringTurnAnswered)
	update(&turn)

	return turn
}

func assistedAuthoringCompletedTurnWith(update func(*AssistedAuthoringTurn)) AssistedAuthoringTurn {
	turn := validAssistedAuthoringTurn(AssistedAuthoringTurnCompleted)
	update(&turn)

	return turn
}

func assistedAuthoringBlockedTurnWith(update func(*AssistedAuthoringTurn)) AssistedAuthoringTurn {
	turn := validAssistedAuthoringTurn(AssistedAuthoringTurnBlocked)
	update(&turn)

	return turn
}
