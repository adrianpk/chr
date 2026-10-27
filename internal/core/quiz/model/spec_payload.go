package model

import (
	"fmt"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

// ValidateDraftSpecPayload validates the corresponding package boundary value.
func ValidateDraftSpecPayload(payload DraftSpecPayload) (DraftSpecPayload, error) {
	normalized := DraftSpecPayload{
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return DraftSpecPayload{}, fmt.Errorf("quiz draft spec title is required")
	}

	if normalized.Body == "" {
		return DraftSpecPayload{}, fmt.Errorf("quiz draft spec body is required")
	}

	if normalized.Status != "draft" {
		return DraftSpecPayload{}, fmt.Errorf("quiz draft spec status must be draft")
	}

	return normalized, nil
}

// ValidateFeaturePayload validates the corresponding package boundary value.
func ValidateFeaturePayload(payload FeaturePayload) (FeaturePayload, error) {
	feature, err := generationcore.ValidateAcceptanceFeatureGeneration(generationcore.AcceptanceFeatureGeneration(payload))
	if err != nil {
		return FeaturePayload{}, err
	}

	return FeaturePayload(feature), nil
}
