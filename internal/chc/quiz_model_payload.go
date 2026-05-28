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

type QuizModelPlanPayload struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

type QuizModelPlanOutlinePayload struct {
	Title              string                                  `json:"title"`
	Status             string                                  `json:"status"`
	Summary            string                                  `json:"summary"`
	Assumptions        []string                                `json:"assumptions"`
	OpenQuestions      []string                                `json:"open_questions"`
	Risks              []string                                `json:"risks"`
	ValidationStrategy string                                  `json:"validation_strategy"`
	Slices             []QuizModelPlanOutlineSlicePayload      `json:"slices"`
	Dependencies       []QuizModelPlanOutlineDependencyPayload `json:"dependencies"`
}

type QuizModelPlanOutlineSlicePayload struct {
	ID                     string   `json:"id"`
	ShortName              string   `json:"short_name"`
	Objective              string   `json:"objective"`
	Rationale              string   `json:"rationale"`
	Dependencies           []string `json:"dependencies"`
	BranchName             string   `json:"branch_name"`
	PRTitle                string   `json:"pr_title"`
	ReportPath             string   `json:"report_path"`
	Complexity             string   `json:"complexity"`
	ImplementationBoundary string   `json:"implementation_boundary"`
}

type QuizModelPlanOutlineDependencyPayload struct {
	FromSliceID string `json:"from_slice_id"`
	ToSliceID   string `json:"to_slice_id"`
	Reason      string `json:"reason"`
}

type QuizModelPlanSliceDetailPayload struct {
	SliceID                string                                  `json:"slice_id"`
	ShortName              string                                  `json:"short_name"`
	Objective              string                                  `json:"objective"`
	ImplementationNotes    []string                                `json:"implementation_notes"`
	Tasks                  []QuizModelPlanSliceTaskPayload         `json:"tasks"`
	AcceptanceCriteria     []string                                `json:"acceptance_criteria"`
	TicketMetadata         QuizModelPlanSliceTicketMetadataPayload `json:"ticket_metadata"`
	BranchMetadata         QuizModelPlanSliceBranchMetadataPayload `json:"branch_metadata"`
	PRExpectations         []string                                `json:"pr_expectations"`
	ValidationExpectations []string                                `json:"validation_expectations"`
	Dependencies           []string                                `json:"dependencies"`
	NonGoals               []string                                `json:"non_goals"`
}

type QuizModelPlanSliceTaskPayload struct {
	ID         string `json:"id"`
	Summary    string `json:"summary"`
	CommitHint string `json:"commit_hint"`
}

type QuizModelPlanSliceTicketMetadataPayload struct {
	Title    string   `json:"title"`
	Synopsis string   `json:"synopsis"`
	Priority string   `json:"priority"`
	Severity string   `json:"severity"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
}

type QuizModelPlanSliceBranchMetadataPayload struct {
	BranchName string `json:"branch_name"`
}

type QuizModelPlanTrackingPayload struct {
	Title   string                               `json:"title"`
	Body    string                               `json:"body"`
	Status  string                               `json:"status"`
	Tickets []QuizModelPlanTrackingTicketPayload `json:"tickets"`
}

type QuizModelPlanTrackingTicketPayload struct {
	SliceID      string   `json:"slice_id"`
	Title        string   `json:"title"`
	Synopsis     string   `json:"synopsis"`
	Summary      string   `json:"summary"`
	Body         string   `json:"body"`
	Priority     string   `json:"priority"`
	Severity     string   `json:"severity"`
	Type         string   `json:"type"`
	Tags         []string `json:"tags"`
	Dependencies []string `json:"dependencies"`
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
	if action == QuizModelActionGeneratePlanSliceDetail {
		return QuizModelOutputContractPayload{
			RequiredFields:    quizModelPlanSliceDetailRequiredFields(),
			RenderedTextField: "objective",
		}
	}

	if action == QuizModelActionGeneratePlanOutline {
		return QuizModelOutputContractPayload{
			RequiredFields:    quizModelPlanOutlineRequiredFields(),
			RenderedTextField: "summary",
		}
	}

	if action == QuizModelActionDraftSpec || action == QuizModelActionGenerateSpec || action == QuizModelActionGeneratePlan {
		return QuizModelOutputContractPayload{
			RequiredFields:    []string{"title", "body", "status"},
			RenderedTextField: "body",
		}
	}

	if action == QuizModelActionMaterializePlanTracking {
		return QuizModelOutputContractPayload{
			RequiredFields:    []string{"title", "body", "status", "tickets"},
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

func quizModelPlanOutlineRequiredFields() []string {
	return []string{
		"title",
		"status",
		"summary",
		"assumptions",
		"open_questions",
		"risks",
		"validation_strategy",
		"slices",
		"dependencies",
	}
}

func quizModelPlanSliceDetailRequiredFields() []string {
	return []string{
		"slice_id",
		"short_name",
		"objective",
		"implementation_notes",
		"tasks",
		"acceptance_criteria",
		"ticket_metadata",
		"branch_metadata",
		"pr_expectations",
		"validation_expectations",
		"dependencies",
		"non_goals",
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

func DecodeQuizModelPlanPayload(content []byte) (QuizModelPlanPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelPlanPayload{}, err
	}

	err = validateQuizModelExactFields(fields, []string{"title", "body", "status"})
	if err != nil {
		return QuizModelPlanPayload{}, err
	}

	var payload QuizModelPlanPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelPlanPayload{}, err
	}

	return ValidateQuizModelPlanPayload(payload)
}

func DecodeQuizModelPlanOutlinePayload(content []byte) (QuizModelPlanOutlinePayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelPlanOutlinePayload{}, err
	}

	required := quizModelPlanOutlineRequiredFields()

	err = validateQuizModelExactFields(fields, required)
	if err != nil {
		return QuizModelPlanOutlinePayload{}, err
	}

	var payload QuizModelPlanOutlinePayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelPlanOutlinePayload{}, err
	}

	return ValidateQuizModelPlanOutlinePayload(payload)
}

func DecodeQuizModelPlanSliceDetailPayload(content []byte) (QuizModelPlanSliceDetailPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelPlanSliceDetailPayload{}, err
	}

	required := quizModelPlanSliceDetailRequiredFields()

	err = validateQuizModelExactFields(fields, required)
	if err != nil {
		return QuizModelPlanSliceDetailPayload{}, err
	}

	var payload QuizModelPlanSliceDetailPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelPlanSliceDetailPayload{}, err
	}

	return ValidateQuizModelPlanSliceDetailPayload(payload)
}

func DecodeQuizModelPlanTrackingPayload(content []byte) (QuizModelPlanTrackingPayload, error) {
	fields, err := quizModelRawJSONFields(content)
	if err != nil {
		return QuizModelPlanTrackingPayload{}, err
	}

	err = validateQuizModelExactFields(fields, []string{"title", "body", "status", "tickets"})
	if err != nil {
		return QuizModelPlanTrackingPayload{}, err
	}

	var payload QuizModelPlanTrackingPayload

	err = decodeStrictQuizModelJSON(content, &payload)
	if err != nil {
		return QuizModelPlanTrackingPayload{}, err
	}

	return ValidateQuizModelPlanTrackingPayload(payload)
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

func ValidateQuizModelPlanPayload(payload QuizModelPlanPayload) (QuizModelPlanPayload, error) {
	normalized := QuizModelPlanPayload{
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return QuizModelPlanPayload{}, fmt.Errorf("generated plan title is required")
	}

	if normalized.Body == "" {
		return QuizModelPlanPayload{}, fmt.Errorf("generated plan body is required")
	}

	if normalized.Status != "planned" {
		return QuizModelPlanPayload{}, fmt.Errorf("generated plan status must be planned")
	}

	return normalized, nil
}

func ValidateQuizModelPlanOutlinePayload(payload QuizModelPlanOutlinePayload) (QuizModelPlanOutlinePayload, error) {
	normalized := normalizeQuizModelPlanOutlinePayload(payload)
	outline := generatedPlanOutlineFromPayload(QuizScope{ProjectID: "payload-validation"}, normalized)

	if err := ValidateGeneratedPlanOutline(outline); err != nil {
		return QuizModelPlanOutlinePayload{}, err
	}

	return normalized, nil
}

func normalizeQuizModelPlanOutlinePayload(payload QuizModelPlanOutlinePayload) QuizModelPlanOutlinePayload {
	normalized := QuizModelPlanOutlinePayload{
		Title:              strings.TrimSpace(payload.Title),
		Status:             strings.TrimSpace(payload.Status),
		Summary:            strings.TrimSpace(payload.Summary),
		Assumptions:        normalizePlanTrackingDependencies(payload.Assumptions),
		OpenQuestions:      normalizePlanTrackingDependencies(payload.OpenQuestions),
		Risks:              normalizePlanTrackingDependencies(payload.Risks),
		ValidationStrategy: strings.TrimSpace(payload.ValidationStrategy),
		Slices:             make([]QuizModelPlanOutlineSlicePayload, 0, len(payload.Slices)),
		Dependencies:       make([]QuizModelPlanOutlineDependencyPayload, 0, len(payload.Dependencies)),
	}

	for _, slice := range payload.Slices {
		normalized.Slices = append(normalized.Slices, normalizeQuizModelPlanOutlineSlicePayload(slice))
	}

	for _, dependency := range payload.Dependencies {
		normalized.Dependencies = append(
			normalized.Dependencies,
			normalizeQuizModelPlanOutlineDependencyPayload(dependency),
		)
	}

	return normalized
}

func normalizeQuizModelPlanOutlineSlicePayload(
	slice QuizModelPlanOutlineSlicePayload,
) QuizModelPlanOutlineSlicePayload {
	return QuizModelPlanOutlineSlicePayload{
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

func normalizeQuizModelPlanOutlineDependencyPayload(
	dependency QuizModelPlanOutlineDependencyPayload,
) QuizModelPlanOutlineDependencyPayload {
	return QuizModelPlanOutlineDependencyPayload{
		FromSliceID: strings.TrimSpace(dependency.FromSliceID),
		ToSliceID:   strings.TrimSpace(dependency.ToSliceID),
		Reason:      strings.TrimSpace(dependency.Reason),
	}
}

func generatedPlanOutlineFromPayload(
	scope QuizScope,
	payload QuizModelPlanOutlinePayload,
) GeneratedPlanOutline {
	outline := GeneratedPlanOutline{
		Scope:              scope,
		Title:              payload.Title,
		Status:             payload.Status,
		Summary:            payload.Summary,
		Assumptions:        payload.Assumptions,
		OpenQuestions:      payload.OpenQuestions,
		Risks:              payload.Risks,
		ValidationStrategy: payload.ValidationStrategy,
		Slices:             make([]GeneratedPlanOutlineSlice, 0, len(payload.Slices)),
		Dependencies:       make([]GeneratedPlanOutlineDependency, 0, len(payload.Dependencies)),
	}

	for _, slice := range payload.Slices {
		outline.Slices = append(outline.Slices, GeneratedPlanOutlineSlice(slice))
	}

	for _, dependency := range payload.Dependencies {
		outline.Dependencies = append(outline.Dependencies, GeneratedPlanOutlineDependency(dependency))
	}

	return outline
}

func ValidateQuizModelPlanSliceDetailPayload(
	payload QuizModelPlanSliceDetailPayload,
) (QuizModelPlanSliceDetailPayload, error) {
	normalized := normalizeQuizModelPlanSliceDetailPayload(payload)
	detail := generatedPlanSliceDetailFromPayload(QuizScope{ProjectID: "payload-validation"}, normalized)

	if err := ValidateGeneratedPlanSliceDetailContent(detail); err != nil {
		return QuizModelPlanSliceDetailPayload{}, err
	}

	return normalized, nil
}

func normalizeQuizModelPlanSliceDetailPayload(
	payload QuizModelPlanSliceDetailPayload,
) QuizModelPlanSliceDetailPayload {
	normalized := QuizModelPlanSliceDetailPayload{
		SliceID:                strings.TrimSpace(payload.SliceID),
		ShortName:              strings.TrimSpace(payload.ShortName),
		Objective:              strings.TrimSpace(payload.Objective),
		ImplementationNotes:    normalizePlanTrackingDependencies(payload.ImplementationNotes),
		Tasks:                  make([]QuizModelPlanSliceTaskPayload, 0, len(payload.Tasks)),
		AcceptanceCriteria:     normalizePlanTrackingDependencies(payload.AcceptanceCriteria),
		TicketMetadata:         normalizeQuizModelPlanSliceTicketMetadataPayload(payload.TicketMetadata),
		BranchMetadata:         normalizeQuizModelPlanSliceBranchMetadataPayload(payload.BranchMetadata),
		PRExpectations:         normalizePlanTrackingDependencies(payload.PRExpectations),
		ValidationExpectations: normalizePlanTrackingDependencies(payload.ValidationExpectations),
		Dependencies:           normalizePlanTrackingDependencies(payload.Dependencies),
		NonGoals:               normalizePlanTrackingDependencies(payload.NonGoals),
	}

	for _, task := range payload.Tasks {
		normalized.Tasks = append(normalized.Tasks, normalizeQuizModelPlanSliceTaskPayload(task))
	}

	return normalized
}

func normalizeQuizModelPlanSliceTaskPayload(
	task QuizModelPlanSliceTaskPayload,
) QuizModelPlanSliceTaskPayload {
	return QuizModelPlanSliceTaskPayload{
		ID:         strings.TrimSpace(task.ID),
		Summary:    strings.TrimSpace(task.Summary),
		CommitHint: strings.TrimSpace(task.CommitHint),
	}
}

func normalizeQuizModelPlanSliceTicketMetadataPayload(
	metadata QuizModelPlanSliceTicketMetadataPayload,
) QuizModelPlanSliceTicketMetadataPayload {
	return QuizModelPlanSliceTicketMetadataPayload{
		Title:    strings.TrimSpace(metadata.Title),
		Synopsis: strings.TrimSpace(metadata.Synopsis),
		Priority: strings.TrimSpace(metadata.Priority),
		Severity: strings.TrimSpace(metadata.Severity),
		Type:     strings.TrimSpace(metadata.Type),
		Tags:     normalizePlanTrackingDependencies(metadata.Tags),
	}
}

func normalizeQuizModelPlanSliceBranchMetadataPayload(
	metadata QuizModelPlanSliceBranchMetadataPayload,
) QuizModelPlanSliceBranchMetadataPayload {
	return QuizModelPlanSliceBranchMetadataPayload{
		BranchName: strings.TrimSpace(metadata.BranchName),
	}
}

func generatedPlanSliceDetailFromPayload(
	scope QuizScope,
	payload QuizModelPlanSliceDetailPayload,
) GeneratedPlanSliceDetail {
	detail := GeneratedPlanSliceDetail{
		Scope:                  scope,
		SliceID:                payload.SliceID,
		ShortName:              payload.ShortName,
		Objective:              payload.Objective,
		ImplementationNotes:    payload.ImplementationNotes,
		Tasks:                  make([]GeneratedPlanSliceTask, 0, len(payload.Tasks)),
		AcceptanceCriteria:     payload.AcceptanceCriteria,
		TicketMetadata:         GeneratedPlanSliceTicketMetadata(payload.TicketMetadata),
		BranchMetadata:         GeneratedPlanSliceBranchMetadata(payload.BranchMetadata),
		PRExpectations:         payload.PRExpectations,
		ValidationExpectations: payload.ValidationExpectations,
		Dependencies:           payload.Dependencies,
		NonGoals:               payload.NonGoals,
	}

	for _, task := range payload.Tasks {
		detail.Tasks = append(detail.Tasks, GeneratedPlanSliceTask(task))
	}

	return detail
}

func ValidateQuizModelPlanTrackingPayload(
	payload QuizModelPlanTrackingPayload,
) (QuizModelPlanTrackingPayload, error) {
	normalized := QuizModelPlanTrackingPayload{
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return QuizModelPlanTrackingPayload{}, fmt.Errorf("plan tracking title is required")
	}

	if normalized.Body == "" {
		return QuizModelPlanTrackingPayload{}, fmt.Errorf("plan tracking body is required")
	}

	if normalized.Status != "tracking" {
		return QuizModelPlanTrackingPayload{}, fmt.Errorf("plan tracking status must be tracking")
	}

	if len(payload.Tickets) == 0 {
		return QuizModelPlanTrackingPayload{}, fmt.Errorf("plan tracking tickets are required")
	}

	normalized.Tickets = make([]QuizModelPlanTrackingTicketPayload, 0, len(payload.Tickets))
	for index, ticket := range payload.Tickets {
		normalizedTicket, err := ValidateQuizModelPlanTrackingTicketPayload(ticket)
		if err != nil {
			return QuizModelPlanTrackingPayload{}, fmt.Errorf("plan tracking ticket %d: %w", index+1, err)
		}

		normalized.Tickets = append(normalized.Tickets, normalizedTicket)
	}

	return normalized, nil
}

func ValidateQuizModelPlanTrackingTicketPayload(
	payload QuizModelPlanTrackingTicketPayload,
) (QuizModelPlanTrackingTicketPayload, error) {
	normalized := QuizModelPlanTrackingTicketPayload{
		SliceID:      strings.TrimSpace(payload.SliceID),
		Title:        strings.TrimSpace(payload.Title),
		Synopsis:     strings.TrimSpace(payload.Synopsis),
		Summary:      strings.TrimSpace(payload.Summary),
		Body:         strings.TrimSpace(payload.Body),
		Priority:     normalizePlanTrackingPriority(payload.Priority),
		Severity:     normalizePlanTrackingSeverity(payload.Severity),
		Type:         normalizePlanTrackingType(payload.Type),
		Tags:         normalizeQuizModelTags(payload.Tags),
		Dependencies: normalizePlanTrackingDependencies(payload.Dependencies),
	}

	if normalized.SliceID == "" {
		return QuizModelPlanTrackingTicketPayload{}, fmt.Errorf("slice id is required")
	}

	if normalized.Title == "" {
		return QuizModelPlanTrackingTicketPayload{}, fmt.Errorf("ticket title is required")
	}

	if normalized.Summary == "" {
		return QuizModelPlanTrackingTicketPayload{}, fmt.Errorf("ticket summary is required")
	}

	if normalized.Synopsis == "" {
		return QuizModelPlanTrackingTicketPayload{}, fmt.Errorf("ticket synopsis is required")
	}

	if normalized.Body == "" {
		return QuizModelPlanTrackingTicketPayload{}, fmt.Errorf("ticket body is required")
	}

	return normalized, nil
}

func normalizePlanTrackingPriority(priority string) string {
	switch strings.TrimSpace(priority) {
	case "low", "normal", "high", "critical":
		return strings.TrimSpace(priority)
	default:
		return "normal"
	}
}

func normalizePlanTrackingSeverity(severity string) string {
	switch strings.TrimSpace(severity) {
	case "low", "normal", "high", "critical":
		return strings.TrimSpace(severity)
	default:
		return "normal"
	}
}

func normalizePlanTrackingType(ticketType string) string {
	switch strings.TrimSpace(ticketType) {
	case "task", "bug", "feature", "cleanup", "research", "docs":
		return strings.TrimSpace(ticketType)
	default:
		return "task"
	}
}

//nolint:wsl_v5 // Small normalization loop is clearer inline.
func normalizePlanTrackingDependencies(dependencies []string) []string {
	normalized := []string{}
	for _, dependency := range dependencies {
		dependency = strings.TrimSpace(dependency)
		if dependency != "" {
			normalized = append(normalized, dependency)
		}
	}

	return normalized
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
