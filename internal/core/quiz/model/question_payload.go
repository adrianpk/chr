package model

import (
	"fmt"
	"strings"

	slugcore "forge.adrianpk.com/chr/internal/core/ops/slug"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// ValidateQuestionPayload validates the corresponding package boundary value.
func ValidateQuestionPayload(payload QuestionPayload) (QuestionPayload, error) {
	normalized := QuestionPayload{
		Question:              strings.TrimSpace(payload.Question),
		Tags:                  normalizeTags(payload.Tags),
		FollowUpDepth:         normalizeFollowUpDepth(payload.FollowUpDepth),
		CompletionRecommended: payload.CompletionRecommended,
	}

	err := validateQuestionValues(normalized)
	if err != nil {
		return QuestionPayload{}, err
	}

	return normalized, nil
}

// ValidateFastQuestionPayload validates the corresponding package boundary value.
func ValidateFastQuestionPayload(payload FastQuestionPayload) (FastQuestionPayload, error) {
	if len(payload.Questions) == 0 {
		return FastQuestionPayload{}, fmt.Errorf("quiz model fast response questions must be an array with at least one question")
	}

	questions := make([]QuestionPayload, 0, len(payload.Questions))
	for index, question := range payload.Questions {
		normalized, err := ValidateQuestionPayload(question)
		if err != nil {
			return FastQuestionPayload{}, fmt.Errorf("quiz model fast response question %d: %w", index+1, err)
		}

		questions = append(questions, normalized)
	}

	return FastQuestionPayload{Questions: questions}, nil
}

// NewPromptIntent constructs the corresponding package service or value.
func NewPromptIntent(payload QuestionPayload, routingState map[string]any) PromptIntentPayload {
	return PromptIntentPayload{
		Kind:               "quiz_question",
		Controls:           []string{"answer", "skip", "done", "cancel"},
		ExpectedInputShape: "free_text_or_control",
		Locale:             "auto",
		LLMAllowed:         true,
		Metadata: map[string]any{
			"question_text":          payload.Question,
			"tags":                   payload.Tags,
			"follow_up_depth":        payload.FollowUpDepth,
			"completion_recommended": payload.CompletionRecommended,
			"routing_state":          routingState,
		},
	}
}

func validateQuestionValues(payload QuestionPayload) error {
	if payload.Question == "" {
		return fmt.Errorf("question can't be blank")
	}

	if len(payload.Tags) == 0 {
		return fmt.Errorf("tags must include at least one normalized tag")
	}

	if payload.FollowUpDepth < 0 {
		return fmt.Errorf("follow_up_depth must be greater than or equal to 0")
	}

	values := append([]string{payload.Question}, payload.Tags...)
	for _, value := range values {
		if containsSensitivePattern(value) {
			return fmt.Errorf("question generation must not ask for secrets, credentials, tokens, passwords, private keys, or PII")
		}
	}

	return nil
}

func normalizeFollowUpDepth(depth int) int {
	maxFollowUpDepth := quizcore.MaxFollowUpDepth
	if depth > maxFollowUpDepth {
		return maxFollowUpDepth
	}

	return depth
}

func normalizeTags(tags []string) []string {
	normalized := []string{}
	seen := map[string]bool{}

	for _, tag := range tags {
		value := strings.ToLower(slugcore.Normalize(tag))
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true

		normalized = append(normalized, value)
		if len(normalized) == MaxTags {
			break
		}
	}

	return normalized
}
