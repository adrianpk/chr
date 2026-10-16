package app

import (
	"context"
	"encoding/json"
	"errors"

	"forge.adrianpk.com/chr/internal/core"
)

// PreparedAuthoringExplanation freezes the question and canonical context before
// provider dispatch. Retrying this image never silently reads different sources.
type PreparedAuthoringExplanation struct {
	ProjectID string                      `json:"project_id"`
	Request   AuthoringExplanationRequest `json:"request"`
	Sources   []authoringModelSource      `json:"sources"`
}

// PrepareAuthoringExplanation reads bounded sources without invoking a provider
// or writing authoring artifacts. Callers may retain the complete returned image.
func (s service) PrepareAuthoringExplanation(ctx context.Context, request AuthoringExplanationRequest) (PreparedAuthoringExplanation, error) {
	result := PreparedAuthoringExplanation{ProjectID: s.projectContext.ProjectID, Request: request}
	if err := validateAuthoringExplanation(request, s.projectContext); err != nil {
		return result, err
	}

	sources, err := s.authoringExplanationSources(ctx, request.SessionID)
	result.Sources = sources

	return result, err
}

// ExplainPreparedAuthoring invokes the configured provider with a retained source
// image. It cannot create artifacts, native descriptions or execution authority.
func (s service) ExplainPreparedAuthoring(ctx context.Context, prepared PreparedAuthoringExplanation) (AuthoringExplanation, error) {
	var result AuthoringExplanation
	if err := validateAuthoringExplanation(prepared.Request, s.projectContext); err != nil {
		return result, err
	}

	if prepared.ProjectID != s.projectContext.ProjectID {
		return result, core.ErrWorkflowConflict
	}

	if s.conversationModel == nil {
		return result, errors.New("configured authoring provider is unavailable")
	}

	payload, err := json.Marshal(struct {
		Question string                 `json:"question"`
		Sources  []authoringModelSource `json:"sources"`
	}{prepared.Request.Question, prepared.Sources})
	if err != nil {
		return result, err
	}

	schema := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"explanation"}, "properties": map[string]any{"explanation": map[string]any{"type": "string", "maxLength": 12000}}}

	output, err := s.conversationModel.InvokeStructuredModel(ctx, core.StructuredModelRequest{ID: prepared.Request.ID, Purpose: "authoring_explanation", ProjectDir: s.projectContext.ProjectDir, Input: payload, Schema: schema, Instructions: authoringExplanationInstructions})
	if err != nil {
		return result, err
	}

	result.Text, err = decodeAuthoringExplanation(output)
	if err != nil {
		return result, err
	}

	result.SourceRefs = []string{}
	for _, source := range prepared.Sources {
		result.SourceRefs = append(result.SourceRefs, source.Reference)
	}

	return result, nil
}
