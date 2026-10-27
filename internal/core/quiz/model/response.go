package model

import (
	"encoding/json"
	"fmt"

	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func ResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	switch request.Action {
	case QuizModelActionNextQuestion:
		return quizModelQuestionResponseFromOutput(request, output)
	case QuizModelActionDraftSpec, QuizModelActionGenerateSpec:
		return quizModelDraftSpecResponseFromOutput(request, output)
	case QuizModelActionGenerateAcceptanceFeature:
		return quizModelFeatureResponseFromOutput(request, output)
	case QuizModelActionGeneratePlan:
		return quizModelPlanResponseFromOutput(request, output)
	case QuizModelActionGeneratePlanOutline:
		return quizModelPlanOutlineResponseFromOutput(request, output)
	case QuizModelActionGeneratePlanSliceDetail:
		return quizModelPlanSliceDetailResponseFromOutput(request, output)
	case QuizModelActionMaterializePlanTracking:
		return quizModelPlanTrackingResponseFromOutput(request, output)
	default:
		return QuizModelResponse{}, fmt.Errorf("quiz model action is unsupported: %s", request.Action)
	}
}

func quizModelQuestionResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	if request.Mode == quizcore.ModeFast {
		return quizModelFastQuestionResponseFromOutput(request, output)
	}

	payload, err := DecodeQuizModelQuestionPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		Question:  quizModelQuestionFromPayload(request, payload, 1),
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelFastQuestionResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelFastQuestionPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	questions := payload.Questions
	if len(questions) > QuizModelFastBatchSize {
		questions = questions[:QuizModelFastBatchSize]
	}

	return QuizModelResponse{
		ID:         quizModelResponseID(request),
		RequestID:  request.ID,
		Provider:   QuizModelProviderOpenAI,
		Status:     QuizModelResponseReady,
		Question:   quizModelQuestionFromPayload(request, questions[0], 1),
		Candidates: quizModelCandidatesFromPayloads(request, questions[1:]),
		RawText:    string(output),
		CreatedAt:  request.CreatedAt.UTC(),
	}, nil
}

func quizModelDraftSpecResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelDraftSpecPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
		DraftSpec: generationcore.QuizDraftSpec{
			Scope:     request.Scope,
			Title:     payload.Title,
			Body:      payload.Body,
			Feature:   quizModelFeatureGeneration(payload),
			CreatedAt: request.CreatedAt.UTC(),
		},
		RawText:   string(output),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func quizModelFeatureGeneration(payload QuizModelDraftSpecPayload) generationcore.AcceptanceFeatureGeneration {
	return generationcore.AcceptanceFeatureGeneration{}
}

func quizModelFeatureResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelFeaturePayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  "openai",
		Status:    QuizModelResponseReady,
		Feature:   generationcore.AcceptanceFeatureGeneration(payload),
		RawText:   strings.TrimSpace(string(output)),
		CreatedAt: request.CreatedAt,
	}, nil
}

func quizModelPlanResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
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

func quizModelPlanOutlineResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanOutlinePayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	outline := generatedPlanOutlineFromPayload(request.Scope, payload)
	outline.CreatedAt = request.CreatedAt.UTC()

	return QuizModelResponse{
		ID:          quizModelResponseID(request),
		RequestID:   request.ID,
		Provider:    QuizModelProviderOpenAI,
		Status:      QuizModelResponseReady,
		PlanOutline: outline,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func quizModelPlanSliceDetailResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	payload, err := DecodeQuizModelPlanSliceDetailPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
	}

	detail := generatedPlanSliceDetailFromPayload(request.Scope, payload)
	detail.CreatedAt = request.CreatedAt.UTC()

	return QuizModelResponse{
		ID:          quizModelResponseID(request),
		RequestID:   request.ID,
		Provider:    QuizModelProviderOpenAI,
		Status:      QuizModelResponseReady,
		SliceDetail: detail,
		RawText:     string(output),
		CreatedAt:   request.CreatedAt.UTC(),
	}, nil
}

func quizModelPlanTrackingResponseFromOutput(request QuizModelRequest, output []byte) (QuizModelResponse, error) {
	if request.Strategy == QuizModelStrategyPlanDelivery {
		return quizModelPlanDeliveryResponse(request, output)
	}

	payload, err := DecodeQuizModelPlanTrackingPayload(modelcore.ExtractJSONObject(output))
	if err != nil {
		return QuizModelResponse{}, err
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

	return QuizModelResponse{
		ID:        quizModelResponseID(request),
		RequestID: request.ID,
		Provider:  QuizModelProviderOpenAI,
		Status:    QuizModelResponseReady,
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

func quizModelQuestionFromPayload(request QuizModelRequest, payload QuizModelQuestionPayload, offset int) quizcore.Question {
	depth := quizModelRuntimeDepth(request.RuntimeState, payload.FollowUpDepth)
	id := quizModelQuestionID(request, offset)
	parentID, trunkID := quizModelRoute(request.RuntimeState, depth, id)

	return quizcore.Question{
		ID:        id,
		Prompt:    payload.Question,
		Required:  !payload.CompletionRecommended,
		Topic:     quizModelTopic(payload.Tags),
		Depth:     depth,
		ParentID:  parentID,
		TrunkID:   trunkID,
		RouteID:   trunkID,
		Rationale: quizModelPromptIntentRationale(payload, request),
		CreatedAt: request.CreatedAt.UTC(),
	}
}

func quizModelCandidatesFromPayloads(request QuizModelRequest, payloads []QuizModelQuestionPayload) []quizcore.Candidate {
	candidates := make([]quizcore.Candidate, 0, len(payloads))

	for index, payload := range payloads {
		offset := index + 2
		depth := quizModelRuntimeDepth(request.RuntimeState, payload.FollowUpDepth)
		id := quizModelQuestionID(request, offset)
		parentID, trunkID := quizModelRoute(request.RuntimeState, depth, id)

		candidates = append(candidates, quizcore.Candidate{
			ID:        id,
			Prompt:    payload.Question,
			Topic:     quizModelTopic(payload.Tags),
			Depth:     depth,
			ParentID:  parentID,
			RouteID:   trunkID,
			Hints:     append([]string(nil), payload.Tags...),
			CreatedAt: request.CreatedAt.UTC(),
		})
	}

	return candidates
}

func quizModelRuntimeDepth(state quizcore.RuntimeState, followUpDepth int) int {
	if state.ActiveQuestion == nil && len(state.Trail) == 0 {
		return 1
	}

	if followUpDepth < 1 {
		return 1
	}

	return followUpDepth
}

func quizModelRoute(state quizcore.RuntimeState, depth int, fallbackTrunkID string) (string, string) {
	if depth <= 1 {
		return "", fallbackTrunkID
	}

	if state.ActiveQuestion != nil {
		return state.ActiveQuestion.ID, quizModelTrunkID(state, fallbackTrunkID)
	}

	if len(state.Trail) > 0 {
		parent := state.Trail[len(state.Trail)-1]

		return parent.QuestionID, quizModelTrunkID(state, fallbackTrunkID)
	}

	return "", fallbackTrunkID
}

func quizModelTrunkID(state quizcore.RuntimeState, fallback string) string {
	if state.ActiveQuestion != nil && strings.TrimSpace(state.ActiveQuestion.TrunkID) != "" {
		return state.ActiveQuestion.TrunkID
	}

	if len(state.Trail) > 0 {
		return state.Trail[0].QuestionID
	}

	return fallback
}

func quizModelQuestionID(request QuizModelRequest, offset int) string {
	return fmt.Sprintf("question-%03d", request.RuntimeState.QuestionsAsked+offset)
}

func quizModelResponseID(request QuizModelRequest) string {
	return "response-" + request.ID
}

func quizModelTopic(tags []string) string {
	if len(tags) == 0 {
		return "project-context"
	}

	return tags[0]
}

func quizModelPromptIntentRationale(payload QuizModelQuestionPayload, request QuizModelRequest) string {
	intent := NewQuizModelPromptIntent(payload, quizModelRoutingState(request, payload))

	content, err := json.Marshal(intent)
	if err != nil {
		return ""
	}

	return string(content)
}

func quizModelRoutingState(request QuizModelRequest, payload QuizModelQuestionPayload) map[string]any {
	entryCount := request.Payload.Quiz.EntryCount

	state := map[string]any{
		"phase":                    quizModelPhase(payload),
		"entry_count":              entryCount,
		"remaining_questions":      quizModelMaxInt(quizcore.DefaultQuestionBudget-entryCount, 0),
		"follow_up_depth":          payload.FollowUpDepth,
		"max_follow_up_depth":      quizcore.MaxFollowUpDepth,
		"trunk_return_recommended": payload.FollowUpDepth >= quizcore.MaxFollowUpDepth,
		"covered_tags":             quizModelCoveredTags(request.Payload.RecentEntries),
		"topic_coverage":           quizModelTopicCoverage(request.Payload.RecentEntries),
		"recent_topic_streak":      quizModelRecentTopicStreak(request.Payload.RecentEntries),
		"mode":                     request.Payload.Quiz.Mode,
		"generation_strategy":      request.Strategy,
		"fallback_reason":          "",
		"candidate_count":          len(request.RuntimeState.Candidates),
	}
	if request.Payload.AcceptanceReadiness != nil {
		state["acceptance_readiness"] = request.Payload.AcceptanceReadiness

		promptIntent, ok := quizModelAcceptanceReadinessPromptIntent(*request.Payload.AcceptanceReadiness)
		if ok {
			state["acceptance_prompt_intent"] = promptIntent
		}
	}

	return state
}

func quizModelAcceptanceReadinessPromptIntent(readiness generationcore.AcceptanceReadiness) (generationcore.AcceptanceReadinessPromptIntent, bool) {
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

func quizModelMaxInt(left, right int) int {
	if left > right {
		return left
	}

	return right
}

func quizModelPhase(payload QuizModelQuestionPayload) string {
	if payload.FollowUpDepth > 1 {
		return "follow_up"
	}

	return "trunk"
}

func quizModelCoveredTags(entries []QuizModelEntryPayload) []string {
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

func quizModelTopicCoverage(entries []QuizModelEntryPayload) map[string]int {
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

func quizModelRecentTopicStreak(entries []QuizModelEntryPayload) map[string]any {
	if len(entries) == 0 {
		return map[string]any{
			"topic": "",
			"count": 0,
		}
	}

	lastTopic := quizModelTopic(entries[len(entries)-1].Tags)
	count := 0

	for index := len(entries) - 1; index >= 0; index-- {
		if quizModelTopic(entries[index].Tags) != lastTopic {
			break
		}

		count++
	}

	return map[string]any{
		"topic": lastTopic,
		"count": count,
	}
}
