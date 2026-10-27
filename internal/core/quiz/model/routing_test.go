package model

import (
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func TestQuizModelRoutingStateCarriesTopicConductorSignals(t *testing.T) {
	request := routingTestRequest()
	request.Payload.RecentEntries = []EntryPayload{
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("storage"),
		quizModelEntryPayloadWithTags("identity"),
		quizModelEntryPayloadWithTags("identity"),
	}

	state := routingState(request, QuestionPayload{
		Question:      "What identifier should appear in public paths?",
		Tags:          []string{"identity"},
		FollowUpDepth: 1,
	})

	coverage, ok := state["topic_coverage"].(map[string]int)
	if !ok {
		t.Fatalf("topic_coverage = %#v, want map[string]int", state["topic_coverage"])
	}

	if coverage["identity"] != 4 {
		t.Fatalf("identity coverage = %d, want 4", coverage["identity"])
	}

	streak, ok := state["recent_topic_streak"].(map[string]any)
	if !ok {
		t.Fatalf("recent_topic_streak = %#v, want map[string]any", state["recent_topic_streak"])
	}

	if streak["topic"] != "identity" {
		t.Fatalf("recent topic = %#v, want identity", streak["topic"])
	}

	if streak["count"] != 2 {
		t.Fatalf("recent topic count = %#v, want 2", streak["count"])
	}
}

func TestQuizModelRoutingStateCarriesAcceptanceReadinessPromptIntent(t *testing.T) {
	request := routingTestRequest()
	request.Payload.AcceptanceReadiness = &generationcore.AcceptanceReadiness{
		Decision:   generationcore.AcceptanceReadinessNeedsMoreInfo,
		Confidence: 0.66,
		Reasons:    []string{"Need actor details."},
		Missing:    []generationcore.AcceptanceReadinessSlot{generationcore.AcceptanceReadinessSlotActor},
	}

	state := routingState(request, QuestionPayload{
		Question:              "Who starts this workflow?",
		Tags:                  []string{"actor"},
		FollowUpDepth:         1,
		CompletionRecommended: false,
	})

	readiness, ok := state["acceptance_readiness"].(*generationcore.AcceptanceReadiness)
	if !ok {
		t.Fatalf("acceptance_readiness = %#v, want readiness payload", state["acceptance_readiness"])
	}

	if readiness.Decision != generationcore.AcceptanceReadinessNeedsMoreInfo {
		t.Fatalf("decision = %q, want needs_more_info", readiness.Decision)
	}

	intent, ok := state["acceptance_prompt_intent"].(generationcore.AcceptanceReadinessPromptIntent)
	if !ok {
		t.Fatalf("acceptance_prompt_intent = %#v, want prompt intent", state["acceptance_prompt_intent"])
	}

	if intent.Slot != string(generationcore.AcceptanceReadinessSlotActor) {
		t.Fatalf("slot = %q, want actor", intent.Slot)
	}
}

func quizModelEntryPayloadWithTags(tags ...string) EntryPayload {
	return EntryPayload{
		Question: "Question?",
		Answer:   "Answer.",
		Status:   string(quizcore.EntryStatusAnswered),
		Tags:     tags,
	}
}

func TestQuizModelRuntimeDepthForcesTrunkAfterBranchReturn(t *testing.T) {
	state := quizcore.RuntimeState{}
	state.ActiveQuestion = nil
	state.Trail = nil
	state.BranchReturnReason = quizcore.BranchReturnReasonSkipped

	depth := runtimeDepth(state, quizcore.MaxFollowUpDepth)
	if depth != 1 {
		t.Fatalf("depth = %d, want trunk depth after branch return", depth)
	}
}

func routingTestRequest() Request {
	return Request{Strategy: "deep", Payload: Payload{Quiz: QuizPayload{EntryCount: 1, Mode: "deep"}}}
}
