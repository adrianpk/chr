package model

import (
	"fmt"
	"strings"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// ValidatePlanPayload validates the corresponding package boundary value.
func ValidatePlanPayload(payload PlanPayload) (PlanPayload, error) {
	normalized := PlanPayload{
		Order:  payload.Order,
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return PlanPayload{}, fmt.Errorf("generated plan title is required")
	}

	if normalized.Body == "" {
		return PlanPayload{}, fmt.Errorf("generated plan body is required")
	}

	if normalized.Status != "planned" {
		return PlanPayload{}, fmt.Errorf("generated plan status must be planned")
	}

	if normalized.Order != nil {
		if err := generationcore.ValidatePlannedOrder(normalized.Order); err != nil {
			return PlanPayload{}, err
		}
	}

	return normalized, nil
}

// ValidatePlanOutlinePayload validates the corresponding package boundary value.
func ValidatePlanOutlinePayload(payload PlanOutlinePayload) (PlanOutlinePayload, error) {
	normalized := normalizePlanOutlinePayload(payload)
	outline := generatedPlanOutlineFromPayload(quizcore.Scope{ProjectID: "payload-validation"}, normalized)

	if err := generationcore.ValidateGeneratedPlanOutline(outline); err != nil {
		return PlanOutlinePayload{}, err
	}

	return normalized, nil
}

func normalizePlanOutlinePayload(payload PlanOutlinePayload) PlanOutlinePayload {
	normalized := PlanOutlinePayload{
		Title:              strings.TrimSpace(payload.Title),
		Status:             strings.TrimSpace(payload.Status),
		Summary:            strings.TrimSpace(payload.Summary),
		Assumptions:        normalizePlanTrackingDependencies(payload.Assumptions),
		OpenQuestions:      normalizePlanTrackingDependencies(payload.OpenQuestions),
		Risks:              normalizePlanTrackingDependencies(payload.Risks),
		ValidationStrategy: strings.TrimSpace(payload.ValidationStrategy),
		Slices:             make([]PlanOutlineSlicePayload, 0, len(payload.Slices)),
		Dependencies:       make([]PlanOutlineDependencyPayload, 0, len(payload.Dependencies)),
	}

	for _, slice := range payload.Slices {
		normalized.Slices = append(normalized.Slices, normalizePlanOutlineSlicePayload(slice))
	}

	for _, dependency := range payload.Dependencies {
		normalized.Dependencies = append(
			normalized.Dependencies,
			normalizePlanOutlineDependencyPayload(dependency),
		)
	}

	return normalized
}

func normalizePlanOutlineSlicePayload(slice PlanOutlineSlicePayload) PlanOutlineSlicePayload {
	return PlanOutlineSlicePayload{
		ID:                     strings.TrimSpace(slice.ID),
		ShortName:              strings.TrimSpace(slice.ShortName),
		Objective:              strings.TrimSpace(slice.Objective),
		Rationale:              strings.TrimSpace(slice.Rationale),
		Dependencies:           normalizePlanTrackingDependencies(slice.Dependencies),
		BranchName:             strings.TrimSpace(slice.BranchName),
		PRTitle:                strings.TrimSpace(slice.PRTitle),
		ReportPath:             strings.TrimSpace(slice.ReportPath),
		Complexity:             strings.TrimSpace(slice.Complexity),
		ImplementationBoundary: strings.TrimSpace(slice.ImplementationBoundary),
	}
}

func normalizePlanOutlineDependencyPayload(dependency PlanOutlineDependencyPayload) PlanOutlineDependencyPayload {
	return PlanOutlineDependencyPayload{
		FromSliceID: strings.TrimSpace(dependency.FromSliceID),
		ToSliceID:   strings.TrimSpace(dependency.ToSliceID),
		Reason:      strings.TrimSpace(dependency.Reason),
	}
}

func generatedPlanOutlineFromPayload(scope quizcore.Scope, payload PlanOutlinePayload) generationcore.GeneratedPlanOutline {
	outline := generationcore.GeneratedPlanOutline{
		Scope:              scope,
		Title:              payload.Title,
		Status:             payload.Status,
		Summary:            payload.Summary,
		Assumptions:        payload.Assumptions,
		OpenQuestions:      payload.OpenQuestions,
		Risks:              payload.Risks,
		ValidationStrategy: payload.ValidationStrategy,
		Slices:             make([]generationcore.GeneratedPlanOutlineSlice, 0, len(payload.Slices)),
		Dependencies:       make([]generationcore.GeneratedPlanOutlineDependency, 0, len(payload.Dependencies)),
	}

	for _, slice := range payload.Slices {
		outline.Slices = append(outline.Slices, generationcore.GeneratedPlanOutlineSlice(slice))
	}

	for _, dependency := range payload.Dependencies {
		outline.Dependencies = append(outline.Dependencies, generationcore.GeneratedPlanOutlineDependency(dependency))
	}

	return outline
}

// ValidatePlanSliceDetailPayload validates the corresponding package boundary value.
func ValidatePlanSliceDetailPayload(payload PlanSliceDetailPayload) (PlanSliceDetailPayload, error) {
	normalized := normalizePlanSliceDetailPayload(payload)
	detail := generatedPlanSliceDetailFromPayload(quizcore.Scope{ProjectID: "payload-validation"}, normalized)

	if err := generationcore.ValidateGeneratedPlanSliceDetailContent(detail); err != nil {
		return PlanSliceDetailPayload{}, err
	}

	return normalized, nil
}

func normalizePlanSliceDetailPayload(payload PlanSliceDetailPayload) PlanSliceDetailPayload {
	normalized := PlanSliceDetailPayload{
		SliceID:                strings.TrimSpace(payload.SliceID),
		ShortName:              strings.TrimSpace(payload.ShortName),
		Objective:              strings.TrimSpace(payload.Objective),
		ImplementationNotes:    normalizePlanTrackingDependencies(payload.ImplementationNotes),
		Tasks:                  make([]PlanSliceTaskPayload, 0, len(payload.Tasks)),
		AcceptanceCriteria:     normalizePlanTrackingDependencies(payload.AcceptanceCriteria),
		TicketMetadata:         normalizePlanSliceTicketMetadataPayload(payload.TicketMetadata),
		BranchMetadata:         normalizePlanSliceBranchMetadataPayload(payload.BranchMetadata),
		PRExpectations:         normalizePlanTrackingDependencies(payload.PRExpectations),
		ValidationExpectations: normalizePlanTrackingDependencies(payload.ValidationExpectations),
		Dependencies:           normalizePlanTrackingDependencies(payload.Dependencies),
		NonGoals:               normalizePlanTrackingDependencies(payload.NonGoals),
	}

	for _, task := range payload.Tasks {
		normalized.Tasks = append(normalized.Tasks, normalizePlanSliceTaskPayload(task))
	}

	return normalized
}

func normalizePlanSliceTaskPayload(task PlanSliceTaskPayload) PlanSliceTaskPayload {
	return PlanSliceTaskPayload{
		ID:         strings.TrimSpace(task.ID),
		Summary:    strings.TrimSpace(task.Summary),
		CommitHint: strings.TrimSpace(task.CommitHint),
	}
}

func normalizePlanSliceTicketMetadataPayload(metadata PlanSliceTicketMetadataPayload) PlanSliceTicketMetadataPayload {
	return PlanSliceTicketMetadataPayload{
		Title:    strings.TrimSpace(metadata.Title),
		Synopsis: strings.TrimSpace(metadata.Synopsis),
		Priority: strings.TrimSpace(metadata.Priority),
		Severity: strings.TrimSpace(metadata.Severity),
		Type:     strings.TrimSpace(metadata.Type),
		Tags:     normalizePlanTrackingDependencies(metadata.Tags),
	}
}

func normalizePlanSliceBranchMetadataPayload(metadata PlanSliceBranchMetadataPayload) PlanSliceBranchMetadataPayload {
	return PlanSliceBranchMetadataPayload{
		BranchName: strings.TrimSpace(metadata.BranchName),
	}
}

func generatedPlanSliceDetailFromPayload(scope quizcore.Scope, payload PlanSliceDetailPayload) generationcore.GeneratedPlanSliceDetail {
	detail := generationcore.GeneratedPlanSliceDetail{
		Scope:                  scope,
		SliceID:                payload.SliceID,
		ShortName:              payload.ShortName,
		Objective:              payload.Objective,
		ImplementationNotes:    payload.ImplementationNotes,
		Tasks:                  make([]generationcore.GeneratedPlanSliceTask, 0, len(payload.Tasks)),
		AcceptanceCriteria:     payload.AcceptanceCriteria,
		TicketMetadata:         generationcore.GeneratedPlanSliceTicketMetadata(payload.TicketMetadata),
		BranchMetadata:         generationcore.GeneratedPlanSliceBranchMetadata(payload.BranchMetadata),
		PRExpectations:         payload.PRExpectations,
		ValidationExpectations: payload.ValidationExpectations,
		Dependencies:           payload.Dependencies,
		NonGoals:               payload.NonGoals,
	}

	for _, task := range payload.Tasks {
		detail.Tasks = append(detail.Tasks, generationcore.GeneratedPlanSliceTask(task))
	}

	return detail
}
