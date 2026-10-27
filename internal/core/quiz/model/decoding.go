package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// DecodeQuestionPayload decodes the corresponding package payload.
func DecodeQuestionPayload(content []byte) (QuestionPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return QuestionPayload{}, err
	}

	err = validateExactFields(fields, QuestionRequiredFields)
	if err != nil {
		return QuestionPayload{}, err
	}

	var payload QuestionPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return QuestionPayload{}, err
	}

	return ValidateQuestionPayload(payload)
}

// DecodeFastQuestionPayload decodes the corresponding package payload.
func DecodeFastQuestionPayload(content []byte) (FastQuestionPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return FastQuestionPayload{}, err
	}

	err = validateExactFields(fields, []string{"questions"})
	if err != nil {
		return FastQuestionPayload{}, err
	}

	var payload FastQuestionPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return FastQuestionPayload{}, err
	}

	return ValidateFastQuestionPayload(payload)
}

// DecodeDraftSpecPayload decodes the corresponding package payload.
func DecodeDraftSpecPayload(content []byte) (DraftSpecPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return DraftSpecPayload{}, err
	}

	err = validateExactFields(fields, []string{"title", "body", "status"})
	if err != nil {
		return DraftSpecPayload{}, err
	}

	var payload DraftSpecPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return DraftSpecPayload{}, err
	}

	return ValidateDraftSpecPayload(payload)
}

// DecodeFeaturePayload decodes the corresponding package payload.
func DecodeFeaturePayload(content []byte) (FeaturePayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return FeaturePayload{}, err
	}

	err = validateFields(fields, []string{"status"}, []string{"body", "rejection_reason", "clarification_prompts"})
	if err != nil {
		return FeaturePayload{}, err
	}

	var payload FeaturePayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return FeaturePayload{}, err
	}

	return ValidateFeaturePayload(payload)
}

// DecodePlanPayload decodes the corresponding package payload.
func DecodePlanPayload(content []byte) (PlanPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return PlanPayload{}, err
	}

	required := []string{"title", "body", "status"}
	if slices.Contains(fields, "order") {
		required = append(required, "order")
	}

	err = validateExactFields(fields, required)
	if err != nil {
		return PlanPayload{}, err
	}

	var payload PlanPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return PlanPayload{}, err
	}

	return ValidatePlanPayload(payload)
}

// DecodePlanOutlinePayload decodes the corresponding package payload.
func DecodePlanOutlinePayload(content []byte) (PlanOutlinePayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return PlanOutlinePayload{}, err
	}

	required := planOutlineRequiredFields()

	err = validateExactFields(fields, required)
	if err != nil {
		return PlanOutlinePayload{}, err
	}

	var payload PlanOutlinePayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return PlanOutlinePayload{}, err
	}

	return ValidatePlanOutlinePayload(payload)
}

// DecodePlanSliceDetailPayload decodes the corresponding package payload.
func DecodePlanSliceDetailPayload(content []byte) (PlanSliceDetailPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return PlanSliceDetailPayload{}, err
	}

	required := planSliceDetailRequiredFields()

	err = validateExactFields(fields, required)
	if err != nil {
		return PlanSliceDetailPayload{}, err
	}

	var payload PlanSliceDetailPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return PlanSliceDetailPayload{}, err
	}

	return ValidatePlanSliceDetailPayload(payload)
}

// DecodePlanTrackingPayload decodes the corresponding package payload.
func DecodePlanTrackingPayload(content []byte) (PlanTrackingPayload, error) {
	fields, err := rawJSONFields(content)
	if err != nil {
		return PlanTrackingPayload{}, err
	}

	err = validateExactFields(fields, []string{"title", "body", "status", "tickets"})
	if err != nil {
		return PlanTrackingPayload{}, err
	}

	var payload PlanTrackingPayload

	err = decodeStrictJSON(content, &payload)
	if err != nil {
		return PlanTrackingPayload{}, err
	}

	return ValidatePlanTrackingPayload(payload)
}

func rawJSONFields(content []byte) ([]string, error) {
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

func decodeStrictJSON(content []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("decode quiz model payload: %w", err)
	}

	return nil
}

func validateExactFields(fields, required []string) error {
	return validateFields(fields, required, nil)
}

func validateFields(fields, required, optional []string) error {
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

	for _, field := range optional {
		allowed[field] = true
	}

	for _, field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unexpected field: %s", field)
		}
	}

	return nil
}
