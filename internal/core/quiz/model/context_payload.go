package model

import (
	"encoding/json"
	"fmt"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

// ContextSectionsForPayload executes the corresponding package operation.
func ContextSectionsForPayload(payload Payload) ([]ContextSection, error) {
	err := ValidatePayload(payload)
	if err != nil {
		return nil, err
	}

	sections, err := baseContextSections(payload)
	if err != nil {
		return nil, err
	}

	return appendAcceptanceReadinessSection(sections, payload.AcceptanceReadiness)
}

func baseContextSections(payload Payload) ([]ContextSection, error) {
	project, err := jsonContent(payload.Project)
	if err != nil {
		return nil, err
	}

	quiz, err := jsonContent(payload.Quiz)
	if err != nil {
		return nil, err
	}

	recentEntries, err := jsonContent(payload.RecentEntries)
	if err != nil {
		return nil, err
	}

	currentQuestion, err := jsonContent(payload.CurrentQuestion)
	if err != nil {
		return nil, err
	}

	return []ContextSection{
		{
			Name:      "project",
			Layer:     ContextLayerProject,
			Authority: "project_context",
			Content:   project,
			Metadata:  map[string]string{"context_layer": ContextLayerProject},
		},
		{
			Name:      "quiz",
			Layer:     ContextLayerQuiz,
			Authority: "request_context",
			Content:   quiz,
			Metadata:  map[string]string{"context_layer": ContextLayerQuiz},
		},
		{
			Name:      "recent_entries",
			Layer:     ContextLayerQuiz,
			Authority: "quiz_context",
			Content:   recentEntries,
			Metadata:  map[string]string{"context_layer": ContextLayerQuiz},
		},
		{
			Name:      "current_question",
			Layer:     ContextLayerQuiz,
			Authority: "request_context",
			Content:   currentQuestion,
			Metadata:  map[string]string{"context_layer": ContextLayerQuiz},
		},
	}, nil
}

func appendAcceptanceReadinessSection(sections []ContextSection, readiness *generationcore.AcceptanceReadiness) ([]ContextSection, error) {
	if readiness == nil {
		return sections, nil
	}

	acceptanceReadiness, err := jsonContent(readiness)
	if err != nil {
		return nil, err
	}

	sections = append(sections, ContextSection{
		Name:      "acceptance_readiness",
		Layer:     ContextLayerQuiz,
		Authority: "request_context",
		Content:   acceptanceReadiness,
		Metadata:  map[string]string{"context_layer": ContextLayerQuiz},
	})

	return sections, nil
}

// ValidatePayload validates the corresponding package boundary value.
func ValidatePayload(payload Payload) error {
	err := validatePayloadHeader(payload)
	if err != nil {
		return err
	}

	return validatePayloadEntries(payload.RecentEntries)
}

func validatePayloadHeader(payload Payload) error {
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

	if payload.AcceptanceReadiness != nil {
		_, err := generationcore.ValidateAcceptanceReadiness(*payload.AcceptanceReadiness)
		if err != nil {
			return fmt.Errorf("quiz model payload acceptance readiness: %w", err)
		}
	}

	return nil
}

func validatePayloadEntries(entries []EntryPayload) error {
	for index, entry := range entries {
		if strings.TrimSpace(entry.Question) == "" {
			return fmt.Errorf("quiz model payload recent entry %d question is required", index+1)
		}

		if strings.TrimSpace(entry.Status) == "" {
			return fmt.Errorf("quiz model payload recent entry %d status is required", index+1)
		}
	}

	return nil
}

func jsonContent(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode quiz model context section: %w", err)
	}

	return string(content), nil
}

func containsSensitivePattern(value string) bool {
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
