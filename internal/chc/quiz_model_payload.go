package chc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	QuizModelDeepAction = "quiz_question_generation"
	QuizModelFastAction = "quiz_fast_question_generation"

	QuizModelDeepStrategy = "deep"
	QuizModelFastStrategy = "batch"

	QuizModelContextLayerProject = "project"
	QuizModelContextLayerQuiz    = "quizzes"

	QuizModelMaxTags = 8
)

var QuizModelQuestionRequiredFields = []string{
	"question",
	"tags",
	"follow_up_depth",
	"completion_recommended",
}

type QuizModelPayload struct {
	Project         QuizModelProjectPayload         `json:"project"`
	Quiz            QuizModelQuizPayload            `json:"quiz"`
	RecentEntries   []QuizModelEntryPayload         `json:"recent_entries"`
	CurrentQuestion QuizModelCurrentQuestionPayload `json:"current_question"`
	OutputContract  QuizModelOutputContractPayload  `json:"output_contract"`
}

type QuizModelProjectPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StackSlug   string `json:"stack_slug,omitempty"`
}

type QuizModelQuizPayload struct {
	ID         string `json:"id"`
	UseCaseID  string `json:"use_case_id,omitempty"`
	EntryCount int    `json:"entry_count"`
	Mode       string `json:"mode"`
}

type QuizModelCurrentQuestionPayload struct {
	Question     string         `json:"question"`
	Tags         []string       `json:"tags"`
	PromptIntent map[string]any `json:"prompt_intent"`
}

type QuizModelEntryPayload struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags"`
}

type QuizModelOutputContractPayload struct {
	RequiredFields    []string `json:"required_fields"`
	RenderedTextField string   `json:"rendered_text_field"`
}

type QuizModelQuestionPayload struct {
	Question              string   `json:"question"`
	Tags                  []string `json:"tags"`
	FollowUpDepth         int      `json:"follow_up_depth"`
	CompletionRecommended bool     `json:"completion_recommended"`
}

type QuizModelFastQuestionPayload struct {
	Questions []QuizModelQuestionPayload `json:"questions"`
}

type QuizModelDraftSpecPayload struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

type QuizModelPromptIntentPayload struct {
	Kind               string         `json:"kind"`
	Controls           []string       `json:"controls"`
	ExpectedInputShape string         `json:"expected_input_shape"`
	Locale             string         `json:"locale"`
	LLMAllowed         bool           `json:"llm_allowed"`
	Metadata           map[string]any `json:"metadata"`
}

func NewQuizModelOutputContract(action QuizModelAction, mode QuizMode) QuizModelOutputContractPayload {
	if action == QuizModelActionDraftSpec || action == QuizModelActionGenerateSpec {
		return QuizModelOutputContractPayload{
			RequiredFields:    []string{"title", "body", "status"},
			RenderedTextField: "body",
		}
	}

	if mode == QuizModeFast {
		return QuizModelOutputContractPayload{
			RequiredFields:    []string{"questions"},
			RenderedTextField: "questions",
		}
	}

	return QuizModelOutputContractPayload{
		RequiredFields:    append([]string(nil), QuizModelQuestionRequiredFields...),
		RenderedTextField: "question",
	}
}

func ValidateQuizModelQuestionPayload(payload QuizModelQuestionPayload) (QuizModelQuestionPayload, error) {
	normalized := QuizModelQuestionPayload{
		Question:              strings.TrimSpace(payload.Question),
		Tags:                  normalizeQuizModelTags(payload.Tags),
		FollowUpDepth:         normalizeQuizModelFollowUpDepth(payload.FollowUpDepth),
		CompletionRecommended: payload.CompletionRecommended,
	}

	err := validateQuizModelQuestionValues(normalized)
	if err != nil {
		return QuizModelQuestionPayload{}, err
	}

	return normalized, nil
}

func DecodeQuizModelQuestionPayload(content []byte) (QuizModelQuestionPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelQuestionPayload{}, err
	}

	err = validateQuizModelExactFields(fields, QuizModelQuestionRequiredFields)
	if err != nil {
		return QuizModelQuestionPayload{}, err
	}

	var payload QuizModelQuestionPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelQuestionPayload{}, err
	}

	return ValidateQuizModelQuestionPayload(payload)
}

func DecodeQuizModelFastQuestionPayload(content []byte) (QuizModelFastQuestionPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelFastQuestionPayload{}, err
	}

	err = validateQuizModelExactFields(fields, []string{"questions"})
	if err != nil {
		return QuizModelFastQuestionPayload{}, err
	}

	var payload QuizModelFastQuestionPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelFastQuestionPayload{}, err
	}

	return ValidateQuizModelFastQuestionPayload(payload)
}

func DecodeQuizModelDraftSpecPayload(content []byte) (QuizModelDraftSpecPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelDraftSpecPayload{}, err
	}

	err = validateQuizModelExactFields(fields, []string{"title", "body", "status"})
	if err != nil {
		return QuizModelDraftSpecPayload{}, err
	}

	var payload QuizModelDraftSpecPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelDraftSpecPayload{}, err
	}

	return ValidateQuizModelDraftSpecPayload(payload)
}

func ValidateQuizModelFastQuestionPayload(payload QuizModelFastQuestionPayload) (QuizModelFastQuestionPayload, error) {
	if len(payload.Questions) == 0 {
		return QuizModelFastQuestionPayload{}, fmt.Errorf("quiz model fast response questions must be an array with at least one question")
	}

	questions := make([]QuizModelQuestionPayload, 0, len(payload.Questions))
	for index, question := range payload.Questions {
		normalized, err := ValidateQuizModelQuestionPayload(question)
		if err != nil {
			return QuizModelFastQuestionPayload{}, fmt.Errorf("quiz model fast response question %d: %w", index+1, err)
		}

		questions = append(questions, normalized)
	}

	return QuizModelFastQuestionPayload{Questions: questions}, nil
}

func ValidateQuizModelDraftSpecPayload(payload QuizModelDraftSpecPayload) (QuizModelDraftSpecPayload, error) {
	normalized := QuizModelDraftSpecPayload{
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return QuizModelDraftSpecPayload{}, fmt.Errorf("quiz draft spec title is required")
	}

	if normalized.Body == "" {
		return QuizModelDraftSpecPayload{}, fmt.Errorf("quiz draft spec body is required")
	}

	if normalized.Status != "draft" {
		return QuizModelDraftSpecPayload{}, fmt.Errorf("quiz draft spec status must be draft")
	}

	return normalized, nil
}

func NewQuizModelPromptIntent(payload QuizModelQuestionPayload, routingState map[string]any) QuizModelPromptIntentPayload {
	return QuizModelPromptIntentPayload{
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

func QuizModelContextSectionsForPayload(payload QuizModelPayload) ([]QuizModelContextSection, error) {
	err := ValidateQuizModelPayload(payload)
	if err != nil {
		return nil, err
	}

	project, err := quizModelJSONContent(payload.Project)
	if err != nil {
		return nil, err
	}

	quiz, err := quizModelJSONContent(payload.Quiz)
	if err != nil {
		return nil, err
	}

	recentEntries, err := quizModelJSONContent(payload.RecentEntries)
	if err != nil {
		return nil, err
	}

	currentQuestion, err := quizModelJSONContent(payload.CurrentQuestion)
	if err != nil {
		return nil, err
	}

	return []QuizModelContextSection{
		{
			Name:      "project",
			Layer:     QuizModelContextLayerProject,
			Authority: "project_context",
			Content:   project,
			Metadata:  map[string]string{"context_layer": QuizModelContextLayerProject},
		},
		{
			Name:      "quiz",
			Layer:     QuizModelContextLayerQuiz,
			Authority: "request_context",
			Content:   quiz,
			Metadata:  map[string]string{"context_layer": QuizModelContextLayerQuiz},
		},
		{
			Name:      "recent_entries",
			Layer:     QuizModelContextLayerQuiz,
			Authority: "quiz_context",
			Content:   recentEntries,
			Metadata:  map[string]string{"context_layer": QuizModelContextLayerQuiz},
		},
		{
			Name:      "current_question",
			Layer:     QuizModelContextLayerQuiz,
			Authority: "request_context",
			Content:   currentQuestion,
			Metadata:  map[string]string{"context_layer": QuizModelContextLayerQuiz},
		},
	}, nil
}

func ValidateQuizModelPayload(payload QuizModelPayload) error {
	if strings.TrimSpace(payload.Project.ID) == "" {
		return fmt.Errorf("quiz model payload project id is required")
	}

	if strings.TrimSpace(payload.Quiz.ID) == "" {
		return fmt.Errorf("quiz model payload quiz id is required")
	}

	if payload.Quiz.EntryCount < 0 {
		return fmt.Errorf("quiz model payload quiz entry count cannot be negative")
	}

	if strings.TrimSpace(payload.Quiz.Mode) == "" {
		return fmt.Errorf("quiz model payload quiz mode is required")
	}

	if len(payload.OutputContract.RequiredFields) == 0 {
		return fmt.Errorf("quiz model payload output contract is required")
	}

	for index, entry := range payload.RecentEntries {
		if strings.TrimSpace(entry.Question) == "" {
			return fmt.Errorf("quiz model payload recent entry %d question is required", index+1)
		}

		if strings.TrimSpace(entry.Status) == "" {
			return fmt.Errorf("quiz model payload recent entry %d status is required", index+1)
		}
	}

	return nil
}

func quizModelJSONContent(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode quiz model context section: %w", err)
	}

	return string(content), nil
}

func quizModelRawJSONFields(content []byte) ([]string, error) {
	fields := map[string]json.RawMessage{}

	err := json.Unmarshal(content, &fields)
	if err != nil {
		return nil, fmt.Errorf("decode quiz model payload fields: %w", err)
	}

	names := make([]string, 0, len(fields))
	for name, raw := range fields {
		if string(raw) == "null" {
			continue
		}

		names = append(names, name)
	}

	return names, nil
}

func decodeStrictQuizModelJSON(content []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("decode quiz model payload: %w", err)
	}

	return nil
}

func validateQuizModelExactFields(fields, required []string) error {
	seen := map[string]bool{}
	for _, field := range fields {
		seen[field] = true
	}

	for _, field := range required {
		if !seen[field] {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	allowed := map[string]bool{}
	for _, field := range required {
		allowed[field] = true
	}

	for _, field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unexpected field: %s", field)
		}
	}

	return nil
}

func validateQuizModelQuestionValues(payload QuizModelQuestionPayload) error {
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
		if quizModelContainsSensitivePattern(value) {
			return fmt.Errorf("question generation must not ask for secrets, credentials, tokens, passwords, private keys, or PII")
		}
	}

	return nil
}

func normalizeQuizModelFollowUpDepth(depth int) int {
	maxFollowUpDepth := MaxQuizFollowUpDepth
	if depth > maxFollowUpDepth {
		return maxFollowUpDepth
	}

	return depth
}

func normalizeQuizModelTags(tags []string) []string {
	normalized := []string{}
	seen := map[string]bool{}

	for _, tag := range tags {
		value := strings.ToLower(slugify(tag))
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true

		normalized = append(normalized, value)
		if len(normalized) == QuizModelMaxTags {
			break
		}
	}

	return normalized
}

func quizModelContainsSensitivePattern(value string) bool {
	normalized := strings.ToLower(value)
	patterns := []string{
		"api key",
		"api-key",
		"apikey",
		"credential",
		"password",
		"private key",
		"private-key",
		"secret",
		"ssn",
		"token",
	}

	for _, pattern := range patterns {
		if strings.Contains(normalized, pattern) {
			return true
		}
	}

	return false
}
