package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func (s service) newQuizModelRequestWithEntries(ctx context.Context, state quizcore.RuntimeState, entries []quizcore.Entry) (quizmodel.Request, error) {
	if s.guidedContext != nil {
		return s.guidedQuizModelRequest(ctx, state, entries)
	}

	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload := s.quizModelPayloadWithEntries(state, quizmodel.ActionNextQuestion, entries)

	coreSections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	supplementalSections, err := s.appendQuizSupplementalContextSections(ctx, nil, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	history, err := s.quizHistoryContext(scope, entries)
	if err != nil {
		return quizmodel.Request{}, err
	}

	supplementalSections = append(supplementalSections, history)

	readiness := quizAcceptanceReadinessFromContext(state, append(coreSections, supplementalSections...))
	payload.AcceptanceReadiness = &readiness

	coreSections, err = quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	return quizmodel.Request{
		ID:              harnesshost.HarnessGenerationRequestID(ctx, fmt.Sprintf("quiz-request-%d", state.QuestionsAsked+1)),
		Action:          quizmodel.ActionNextQuestion,
		Strategy:        quizModelStrategy(state.Mode),
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: append(coreSections, supplementalSections...),
		OutputContract: quizmodel.OutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) newQuizModelPayload(ctx context.Context, scope quizcore.Scope, state quizcore.RuntimeState, action quizmodel.Action) (quizmodel.Payload, error) {
	entries, err := s.quizRepository().LoadRawCaptureEntries(ctx, scope)
	if err != nil {
		return quizmodel.Payload{}, err
	}

	return s.quizModelPayloadWithEntries(state, action, entries), nil
}

func (s service) quizModelPayloadWithEntries(state quizcore.RuntimeState, action quizmodel.Action, entries []quizcore.Entry) quizmodel.Payload {
	return quizmodel.Payload{
		Project:         s.quizModelProjectPayload(),
		Quiz:            quizModelQuizPayload(state),
		RecentEntries:   quizModelRecentEntryPayloads(entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, entries),
		OutputContract:  quizmodel.NewOutputContract(action, state.Mode),
	}
}

func (s service) quizModelProjectPayload() quizmodel.ProjectPayload {
	payload := quizmodel.ProjectPayload{ID: s.projectContext.ProjectID}

	metadata, err := authoringcore.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return payload
	}

	payload.Name = metadata.Name
	payload.Description = metadata.Description

	return payload
}

func (s service) appendQuizSupplementalContextSections(ctx context.Context, sections []quizmodel.ContextSection, scope quizcore.Scope) ([]quizmodel.ContextSection, error) {
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project_document", "project", authoringcore.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case_document",
			"use_case",
			authoringcore.CaseMetadataPath("", scope.CaseSlug),
		)
	}

	index, err := s.workspaceIndex(ctx)
	if err != nil {
		return nil, err
	}

	for _, nodeID := range index.ByKind[core.WorkspaceNodeDump] {
		node := index.NodesByID[nodeID]
		if core.IsRawDumpNode(node) {
			continue
		}

		if node.CaseSlug != "" && node.CaseSlug != scope.CaseSlug {
			continue
		}

		sections = appendQuizNodeContextSection(sections, s.projectContext.ProjectDir, node)
	}

	return sections, nil
}

func quizModelQuizPayload(state quizcore.RuntimeState) quizmodel.QuizPayload {
	return quizmodel.QuizPayload{
		ID:         state.ID,
		UseCaseID:  state.CaseSlug,
		EntryCount: state.AnswersRecorded + state.SkipsRecorded,
		Mode:       string(state.Mode),
	}
}

func quizModelRecentEntryPayloads(entries []quizcore.Entry, limit int) []quizmodel.EntryPayload {
	if limit <= 0 {
		return []quizmodel.EntryPayload{}
	}

	start := len(entries) - limit
	if start < 0 {
		start = 0
	}

	payloads := make([]quizmodel.EntryPayload, 0, len(entries[start:]))
	for _, entry := range entries[start:] {
		payloads = append(payloads, quizmodel.EntryPayload{
			Question: entry.Question,
			Answer:   entry.Answer,
			Status:   string(entry.Status),
			Tags:     entry.Tags,
		})
	}

	return payloads
}

func quizModelCurrentQuestionPayload(state quizcore.RuntimeState, entries []quizcore.Entry) quizmodel.CurrentQuestionPayload {
	if state.ActiveQuestion != nil {
		return quizModelQuestionPayloadFromQuestion(*state.ActiveQuestion)
	}

	if len(entries) == 0 {
		return quizmodel.CurrentQuestionPayload{PromptIntent: map[string]any{}}
	}

	entry := entries[len(entries)-1]
	promptIntent := quizModelPromptIntentFromJSON(entry.PromptIntent)

	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth": entry.Depth - 1,
				"status":          string(entry.Status),
			},
		}
	}

	return quizmodel.CurrentQuestionPayload{
		Question:     entry.Question,
		Tags:         quizcore.EntryTags(entry),
		PromptIntent: promptIntent,
	}
}

func quizModelQuestionPayloadFromQuestion(question quizcore.Question) quizmodel.CurrentQuestionPayload {
	promptIntent := quizModelPromptIntentFromJSON(question.Rationale)
	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth":        question.Depth - 1,
				"completion_recommended": !question.Required,
			},
		}
	}

	return quizmodel.CurrentQuestionPayload{
		Question:     question.Prompt,
		Tags:         quizcore.QuestionTags(question),
		PromptIntent: promptIntent,
	}
}

func quizModelPromptIntentFromJSON(value string) map[string]any {
	value = strings.TrimSpace(value)
	if value == "" {
		return map[string]any{}
	}

	var promptIntent map[string]any

	err := json.Unmarshal([]byte(value), &promptIntent)
	if err != nil {
		return map[string]any{}
	}

	return promptIntent
}

func quizModelStrategy(mode quizcore.Mode) string {
	if mode == quizcore.ModeFast {
		return quizmodel.FastStrategy
	}

	return quizmodel.DeepStrategy
}
