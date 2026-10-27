package model

import (
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
)

const (
	// DeepAction defines a package boundary constant.
	DeepAction = "quiz_question_generation"
	// FastAction defines a package boundary constant.
	FastAction = "quiz_fast_question_generation"

	// DeepStrategy defines a package boundary constant.
	DeepStrategy = "deep"
	// FastStrategy defines a package boundary constant.
	FastStrategy = "batch"

	// ContextLayerProject defines a package boundary constant.
	ContextLayerProject = "project"
	// ContextLayerQuiz defines a package boundary constant.
	ContextLayerQuiz = "quizzes"

	// MaxTags defines a package boundary constant.
	MaxTags = 8
)

// QuestionRequiredFields defines a package boundary variable.
var QuestionRequiredFields = []string{
	"question",
	"tags",
	"follow_up_depth",
	"completion_recommended",
}

// Payload describes a package boundary value.
type Payload struct {
	Project             ProjectPayload                      `json:"project"`
	Quiz                QuizPayload                         `json:"quiz"`
	RecentEntries       []EntryPayload                      `json:"recent_entries"`
	CurrentQuestion     CurrentQuestionPayload              `json:"current_question"`
	AcceptanceReadiness *generationcore.AcceptanceReadiness `json:"acceptance_readiness,omitempty"`
	OutputContract      OutputContractPayload               `json:"output_contract"`
}

// ProjectPayload describes a package boundary value.
type ProjectPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StackSlug   string `json:"stack_slug,omitempty"`
}

// QuizPayload describes a package boundary value.
type QuizPayload struct {
	ID         string `json:"id"`
	UseCaseID  string `json:"use_case_id,omitempty"`
	EntryCount int    `json:"entry_count"`
	Mode       string `json:"mode"`
}

// CurrentQuestionPayload describes a package boundary value.
type CurrentQuestionPayload struct {
	Question     string         `json:"question"`
	Tags         []string       `json:"tags"`
	PromptIntent map[string]any `json:"prompt_intent"`
}

// EntryPayload describes a package boundary value.
type EntryPayload struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags"`
}

// OutputContractPayload describes a package boundary value.
type OutputContractPayload struct {
	RequiredFields    []string `json:"required_fields"`
	RenderedTextField string   `json:"rendered_text_field"`
}

// QuestionPayload describes a package boundary value.
type QuestionPayload struct {
	Question              string   `json:"question"`
	Tags                  []string `json:"tags"`
	FollowUpDepth         int      `json:"follow_up_depth"`
	CompletionRecommended bool     `json:"completion_recommended"`
}

// FastQuestionPayload describes a package boundary value.
type FastQuestionPayload struct {
	Questions []QuestionPayload `json:"questions"`
}

// DraftSpecPayload describes a package boundary value.
type DraftSpecPayload struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

// FeaturePayload describes optional generated feature output.
type FeaturePayload struct {
	Status               string                                   `json:"status"`
	Body                 string                                   `json:"body,omitempty"`
	RejectionReason      string                                   `json:"rejection_reason,omitempty"`
	ClarificationPrompts []generationcore.AcceptanceFeaturePrompt `json:"clarification_prompts,omitempty"`
}

// PlanPayload describes a package boundary value.
type PlanPayload struct {
	Order  []generationcore.PlanOrderSlice `json:"order,omitempty"`
	Title  string                          `json:"title"`
	Body   string                          `json:"body"`
	Status string                          `json:"status"`
}

// PlanOutlinePayload describes a package boundary value.
type PlanOutlinePayload struct {
	Title              string                         `json:"title"`
	Status             string                         `json:"status"`
	Summary            string                         `json:"summary"`
	Assumptions        []string                       `json:"assumptions"`
	OpenQuestions      []string                       `json:"open_questions"`
	Risks              []string                       `json:"risks"`
	ValidationStrategy string                         `json:"validation_strategy"`
	Slices             []PlanOutlineSlicePayload      `json:"slices"`
	Dependencies       []PlanOutlineDependencyPayload `json:"dependencies"`
}

// PlanOutlineSlicePayload describes a package boundary value.
type PlanOutlineSlicePayload struct {
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

// PlanOutlineDependencyPayload describes a package boundary value.
type PlanOutlineDependencyPayload struct {
	FromSliceID string `json:"from_slice_id"`
	ToSliceID   string `json:"to_slice_id"`
	Reason      string `json:"reason"`
}

// PlanSliceDetailPayload describes a package boundary value.
type PlanSliceDetailPayload struct {
	SliceID                string                         `json:"slice_id"`
	ShortName              string                         `json:"short_name"`
	Objective              string                         `json:"objective"`
	ImplementationNotes    []string                       `json:"implementation_notes"`
	Tasks                  []PlanSliceTaskPayload         `json:"tasks"`
	AcceptanceCriteria     []string                       `json:"acceptance_criteria"`
	TicketMetadata         PlanSliceTicketMetadataPayload `json:"ticket_metadata"`
	BranchMetadata         PlanSliceBranchMetadataPayload `json:"branch_metadata"`
	PRExpectations         []string                       `json:"pr_expectations"`
	ValidationExpectations []string                       `json:"validation_expectations"`
	Dependencies           []string                       `json:"dependencies"`
	NonGoals               []string                       `json:"non_goals"`
}

// PlanSliceTaskPayload describes a package boundary value.
type PlanSliceTaskPayload struct {
	ID         string `json:"id"`
	Summary    string `json:"summary"`
	CommitHint string `json:"commit_hint"`
}

// PlanSliceTicketMetadataPayload describes a package boundary value.
type PlanSliceTicketMetadataPayload struct {
	Title    string   `json:"title"`
	Synopsis string   `json:"synopsis"`
	Priority string   `json:"priority"`
	Severity string   `json:"severity"`
	Type     string   `json:"type"`
	Tags     []string `json:"tags"`
}

// PlanSliceBranchMetadataPayload describes a package boundary value.
type PlanSliceBranchMetadataPayload struct {
	BranchName string `json:"branch_name"`
}

// PlanTrackingPayload describes a package boundary value.
type PlanTrackingPayload struct {
	Title   string                      `json:"title"`
	Body    string                      `json:"body"`
	Status  string                      `json:"status"`
	Tickets []PlanTrackingTicketPayload `json:"tickets"`
}

// PlanTrackingTicketPayload describes a package boundary value.
type PlanTrackingTicketPayload struct {
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

// PromptIntentPayload describes a package boundary value.
type PromptIntentPayload struct {
	Kind               string         `json:"kind"`
	Controls           []string       `json:"controls"`
	ExpectedInputShape string         `json:"expected_input_shape"`
	Locale             string         `json:"locale"`
	LLMAllowed         bool           `json:"llm_allowed"`
	Metadata           map[string]any `json:"metadata"`
}
