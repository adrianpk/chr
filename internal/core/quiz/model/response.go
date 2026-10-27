package model

import (
	"encoding/json"
	"fmt"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// ResponseFromOutput decodes generated output and retains question and artifact identity.
func ResponseFromOutput(request Request, output []byte) (Response, error) {
	switch request.Action {
	case ActionNextQuestion:
		return questionResponseFromOutput(request, output)
	case ActionDraftSpec, ActionGenerateSpec:
		return draftSpecResponseFromOutput(request, output)
	case ActionGenerateAcceptanceFeature:
		return featureResponseFromOutput(request, output)
	case ActionGeneratePlan:
		return planResponseFromOutput(request, output)
	case ActionGeneratePlanOutline:
		return planOutlineResponseFromOutput(request, output)
	case ActionGeneratePlanSliceDetail:
		return planSliceDetailResponseFromOutput(request, output)
	case ActionMaterializePlanTracking:
		return planTrackingResponseFromOutput(request, output)
	default:
		return Response{}, fmt.Errorf("quiz model action is unsupported: %s", request.Action)
	}
}

func questionResponseFromOutput(request Request, output []byte) (Response, error) {
	if request.Mode == quizcore.ModeFast {
		return fastQuestionResponseFromOutput(request, output)
	}

	payload, err := DecodeQuestionPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	return Response{
		ID:        responseID(request),
		RequestID: request.ID,
		Provider:  ProviderOpenAI,
		Status:    ResponseReady,
		Question:  questionFromPayload(request, payload, 1),
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func fastQuestionResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodeFastQuestionPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	questions := payload.Questions
	if len(questions) > FastBatchSize {
		questions = questions[:FastBatchSize]
	}

	return Response{
		ID:         responseID(request),
		RequestID:  request.ID,
		Provider:   ProviderOpenAI,
		Status:     ResponseReady,
		Question:   questionFromPayload(request, questions[0], 1),
		Candidates: candidatesFromPayloads(request, questions[1:]),
		RawText:    string(output),
		CreatedAt:  request.CreatedAt.UTC(),
	}, nil
}

func draftSpecResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodeDraftSpecPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	return Response{
		ID:        responseID(request),
		RequestID: request.ID,
		Provider:  ProviderOpenAI,
		Status:    ResponseReady,
		DraftSpec: generationcore.QuizDraftSpec{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Feature:   featureGeneration(payload),
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func featureGeneration(payload DraftSpecPayload) generationcore.AcceptanceFeatureGeneration {
	return generationcore.AcceptanceFeatureGeneration{}
}

func featureResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodeFeaturePayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	return Response{
		ID:        responseID(request),
		RequestID: request.ID,
		Provider:  "openai",
		Status:    ResponseReady,
		Feature:   generationcore.AcceptanceFeatureGeneration(payload),
		RawText:   strings.TrimSpace(string(output)),
		CreatedAt: request.CreatedAt,
	}, nil
}

func planResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodePlanPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	return Response{
		ID:        responseID(request),
		RequestID: request.ID,
		Provider:  ProviderOpenAI,
		Status:    ResponseReady,
		Plan: generationcore.GeneratedPlan{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Order:     payload.Order,
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func planOutlineResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodePlanOutlinePayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	outline := generatedPlanOutlineFromPayload(request.Scope, payload)
	outline.CreatedAt = request.CreatedAt.UTC()

	return Response{
		ID:          responseID(request),
		RequestID:   request.ID,
		Provider:    ProviderOpenAI,
		Status:      ResponseReady,
		PlanOutline: outline,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func planSliceDetailResponseFromOutput(request Request, output []byte) (Response, error) {
	payload, err := DecodePlanSliceDetailPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	detail := generatedPlanSliceDetailFromPayload(request.Scope, payload)
	detail.CreatedAt = request.CreatedAt.UTC()

	return Response{
		ID:          responseID(request),
		RequestID:   request.ID,
		Provider:    ProviderOpenAI,
		Status:      ResponseReady,
		SliceDetail: detail,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func planTrackingResponseFromOutput(request Request, output []byte) (Response, error) {
	if request.Strategy == StrategyPlanDelivery {
		return planDeliveryResponse(request, output)
	}

	payload, err := DecodePlanTrackingPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return Response{}, err
	}

	tickets := make([]generationcore.GeneratedPlanTrackingTicket, 0, len(payload.Tickets))
	for _, ticket := range payload.Tickets {
		tickets = append(tickets, generationcore.GeneratedPlanTrackingTicket{
			SliceID:      ticket.SliceID,
			Title:        ticket.Title,
			Synopsis:     ticket.Synopsis,
			Summary:      ticket.Summary,
			Body:         ticket.Body,
			Priority:     ticket.Priority,
			Severity:     ticket.Severity,
			Type:         ticket.Type,
			Tags:         ticket.Tags,
			Dependencies: ticket.Dependencies,
		})
	}

	return Response{
		ID:        responseID(request),
		RequestID: request.ID,
		Provider:  ProviderOpenAI,
		Status:    ResponseReady,
		PlanTracking: generationcore.GeneratedPlanTracking{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Tickets:   tickets,
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func questionFromPayload(request Request, payload QuestionPayload, offset int) quizcore.Question {
	depth := runtimeDepth(request.RuntimeState, payload.FollowUpDepth)
	id := questionID(request, offset)
	routing := route(request.RuntimeState, depth, id)

	return quizcore.Question{
		ID:        id,
		Prompt:    payload.Question,
		Required:  !payload.CompletionRecommended,
		Topic:     topic(payload.Tags),
		Depth:     depth,
		ParentID:  routing.ParentID,
		TrunkID:   routing.TrunkID,
		RouteID:   routing.TrunkID,
		Rationale: promptIntentRationale(payload, request),
		CreatedAt: request.CreatedAt.UTC(),
	}
}

func candidatesFromPayloads(request Request, payloads []QuestionPayload) []quizcore.Candidate {
	candidates := make([]quizcore.Candidate, 0, len(payloads))

	for index, payload := range payloads {
		offset := index + 2
		depth := runtimeDepth(request.RuntimeState, payload.FollowUpDepth)
		id := questionID(request, offset)
		routing := route(request.RuntimeState, depth, id)

		candidates = append(candidates, quizcore.Candidate{
			ID:        id,
			Prompt:    payload.Question,
			Topic:     topic(payload.Tags),
			Depth:     depth,
			ParentID:  routing.ParentID,
			RouteID:   routing.TrunkID,
			Hints:     append([]string(nil), payload.Tags...),
			CreatedAt: request.CreatedAt.UTC(),
		})
	}

	return candidates
}

func runtimeDepth(state quizcore.RuntimeState, followUpDepth int) int {
	if state.ActiveQuestion == nil && len(state.Trail) == 0 {
		return 1
	}

	if followUpDepth < 1 {
		return 1
	}

	return followUpDepth
}

type questionRoute struct{ ParentID, TrunkID string }

func route(state quizcore.RuntimeState, depth int, fallbackTrunkID string) questionRoute {
	if depth <= 1 {
		return questionRoute{TrunkID: fallbackTrunkID}
	}

	if state.ActiveQuestion != nil {
		return questionRoute{ParentID: state.ActiveQuestion.ID, TrunkID: trunkID(state, fallbackTrunkID)}
	}

	if len(state.Trail) > 0 {
		parent := state.Trail[len(state.Trail)-1]

		return questionRoute{ParentID: parent.QuestionID, TrunkID: trunkID(state, fallbackTrunkID)}
	}

	return questionRoute{TrunkID: fallbackTrunkID}
}

func trunkID(state quizcore.RuntimeState, fallback string) string {
	if state.ActiveQuestion != nil && strings.TrimSpace(state.ActiveQuestion.TrunkID) != "" {
		return state.ActiveQuestion.TrunkID
	}

	if len(state.Trail) > 0 {
		return state.Trail[0].QuestionID
	}

	return fallback
}

func questionID(request Request, offset int) string {
	return fmt.Sprintf("question-%03d", request.RuntimeState.QuestionsAsked+offset)
}

func responseID(request Request) string {
	return "response-" + request.ID
}

func topic(tags []string) string {
	if len(tags) == 0 {
		return "project-context"
	}

	return tags[0]
}

func promptIntentRationale(payload QuestionPayload, request Request) string {
	intent := NewPromptIntent(payload, routingState(request, payload))

	content, err := json.Marshal(intent)
	if err != nil {
		return ""
	}

	return string(content)
}

func routingState(request Request, payload QuestionPayload) map[string]any {
	entryCount := request.Payload.Quiz.EntryCount

	state := map[string]any{
		"phase":                    phase(payload),
		"entry_count":              entryCount,
		"remaining_questions":      maxInt(quizcore.DefaultQuestionBudget-entryCount, 0),
		"follow_up_depth":          payload.FollowUpDepth,
		"max_follow_up_depth":      quizcore.MaxFollowUpDepth,
		"trunk_return_recommended": payload.FollowUpDepth >= quizcore.MaxFollowUpDepth,
		"covered_tags":             coveredTags(request.Payload.RecentEntries),
		"topic_coverage":           topicCoverage(request.Payload.RecentEntries),
		"recent_topic_streak":      recentTopicStreak(request.Payload.RecentEntries),
		"mode":                     request.Payload.Quiz.Mode,
		"generation_strategy":      request.Strategy,
		"fallback_reason":          "",
		"candidate_count":          len(request.RuntimeState.Candidates),
	}
	if request.Payload.AcceptanceReadiness != nil {
		state["acceptance_readiness"] = request.Payload.AcceptanceReadiness

		promptIntent, ok := acceptanceReadinessPromptIntent(*request.Payload.AcceptanceReadiness)
		if ok {
			state["acceptance_prompt_intent"] = promptIntent
		}
	}

	return state
}

func acceptanceReadinessPromptIntent(readiness generationcore.AcceptanceReadiness) (generationcore.AcceptanceReadinessPromptIntent, bool) {
	if readiness.Decision != generationcore.AcceptanceReadinessNeedsMoreInfo {
		return generationcore.AcceptanceReadinessPromptIntent{}, false
	}

	for _, slot := range readiness.Missing {
		intent, err := generationcore.NewAcceptanceReadinessPromptIntent(slot)
		if err == nil {
			return intent, true
		}
	}

	return generationcore.AcceptanceReadinessPromptIntent{}, false
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}

	return right
}

func phase(payload QuestionPayload) string {
	if payload.FollowUpDepth > 1 {
		return "follow_up"
	}

	return "trunk"
}

func coveredTags(entries []EntryPayload) []string {
	seen := map[string]bool{}
	tags := []string{}

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" || seen[tag] {
				continue
			}

			seen[tag] = true
			tags = append(tags, tag)
		}
	}

	return tags
}

func topicCoverage(entries []EntryPayload) map[string]int {
	coverage := map[string]int{}

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}

			coverage[tag]++
		}
	}

	return coverage
}

func recentTopicStreak(entries []EntryPayload) map[string]any {
	if len(entries) == 0 {
		return map[string]any{
			"topic": "",
			"count": 0,
		}
	}

	lastTopic := topic(entries[len(entries)-1].Tags)
	count := 0

	for index := len(entries) - 1; index >= 0; index-- {
		if topic(entries[index].Tags) != lastTopic {
			break
		}

		count++
	}

	return map[string]any{
		"topic": lastTopic,
		"count": count,
	}
}
