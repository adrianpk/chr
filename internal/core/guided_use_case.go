package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// GuidedUseCaseState extends a canonical authoring session without adopting manual sessions.
type GuidedUseCaseState struct {
	LaterIdeas          []GuidedIdeaReference        `json:"later_ideas,omitempty"`
	Result              *GuidedResultReference       `json:"result,omitempty"`
	Control             *GuidedControlState          `json:"control,omitempty"`
	Work                *GuidedWorkState             `json:"work,omitempty"`
	SpecDraft           *GuidedSpecDraftReference    `json:"spec_draft,omitempty"`
	SpecApproval        *GuidedSpecApprovalReference `json:"spec_approval,omitempty"`
	Quiz                *GuidedQuizState             `json:"quiz,omitempty"`
	Blocker             string                       `json:"blocker,omitempty"`
	CaseHash            string                       `json:"case_hash,omitempty"`
	PendingCheckpoint   *GuidedStageCheckpoint       `json:"pending_checkpoint,omitempty"`
	SchemaVersion       int                          `json:"schema_version"`
	Actor               string                       `json:"actor"`
	ProgressionIntent   bool                         `json:"progression_intent"`
	CaseSlug            string                       `json:"case_slug,omitempty"`
	CaseName            string                       `json:"case_name,omitempty"`
	Stage               string                       `json:"stage"`
	Status              string                       `json:"status"`
	Capture             string                       `json:"capture,omitempty"`
	PendingOperation    *GuidedUseCaseOperation      `json:"pending_operation,omitempty"`
	CompletedOperations []GuidedUseCaseOperation     `json:"completed_operations,omitempty"`
	Checkpoints         []GuidedStageCheckpoint      `json:"checkpoints,omitempty"`
}

// GuidedUseCaseOperation binds retries to their original arguments and revision.
type GuidedUseCaseOperation struct {
	NativeRevision   uint64                 `json:"native_revision,omitempty"`
	PreparedPlan     *GuidedPreparedPlan    `json:"prepared_plan,omitempty"`
	PreparedWorkflow *GeneratedWorkflowPlan `json:"prepared_workflow,omitempty"`
	PreparedSpec     *GuidedPreparedSpec    `json:"prepared_spec,omitempty"`
	QuizRequest      *QuizActionRequest     `json:"quiz_request,omitempty"`
	QuestionID       string                 `json:"question_id,omitempty"`
	InputDigest      string                 `json:"input_digest,omitempty"`
	ExpectedRevision int64                  `json:"expected_revision,omitempty"`
	ID               string                 `json:"id"`
	Action           string                 `json:"action"`
	Digest           string                 `json:"digest"`
	Content          string                 `json:"content,omitempty"`
}

// GuidedStageCheckpoint retains immutable input bytes independently of a code target.
type GuidedStageCheckpoint struct {
	BaseCommit  string                    `json:"base_commit"`
	OperationID string                    `json:"operation_id"`
	Stage       string                    `json:"stage"`
	Repository  string                    `json:"repository"`
	Commit      string                    `json:"commit"`
	Ref         string                    `json:"ref"`
	Artifacts   []GuidedCommittedArtifact `json:"artifacts"`
	Sources     []GuidedCommittedArtifact `json:"sources,omitempty"`
}

// GuidedCommittedArtifact binds a repository-relative path to its committed content.
type GuidedCommittedArtifact struct {
	Repository string `json:"repository,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Path       string `json:"path"`
	Hash       string `json:"hash"`
}

// GuidedUseCaseScope prevents a callback from addressing another project or namespace.
type GuidedUseCaseScope struct {
	RunID     string
	ProjectID string
	Namespace string
}

// GuidedUseCaseStart records the actor's intent to proceed through development.
type GuidedUseCaseStart struct {
	GuidedUseCaseScope
	Actor             string
	ProgressionIntent bool
}

// GuidedUseCaseCommand is a revision-bound mutation of the retained journey.
type GuidedUseCaseCommand struct {
	NativeRevision uint64 `json:",omitempty"`
	QuestionID     string `json:",omitempty"`
	InputDigest    string `json:",omitempty"`
	GuidedUseCaseScope
	OperationID      string
	ExpectedRevision int64
	Action           string
	Content          string
}

func validateGuidedUseCase(session AuthoringSession) error {
	state := session.Guided
	if state == nil {
		return nil
	}

	if !validGuidedIntent(session) {
		return errors.New("invalid guided use-case schema, scope or progression intent")
	}

	if !boundedGuidedState(state) {
		return errors.New("guided use-case state exceeds its bounds")
	}

	if !guidedStageValid(state.Stage) || !guidedStatusValid(state.Status) {
		return errors.New("invalid guided use-case stage or status")
	}

	if state.Stage != "name" && (state.CaseSlug == "" || state.CaseName == "") {
		return errors.New("guided use case requires a canonical name")
	}

	return nil
}

func guidedStageValid(value string) bool {
	switch value {
	case "name", "dump", "quiz", "specification", "preparation", "execution", "results":
		return true
	default:
		return false
	}
}

func guidedStatusValid(value string) bool {
	switch value {
	case "waiting_input", "running", "blocked", "paused", "completed", "cancelled", "failed":
		return true
	default:
		return false
	}
}

func parseGuidedUseCase(value string) (*GuidedUseCaseState, error) {
	if value == "" {
		return nil, nil
	}

	var state GuidedUseCaseState
	if err := decodeAuthoringState([]byte(unquoteMetadataString(value)), &state); err != nil {
		return nil, fmt.Errorf("parse guided state: %w", err)
	}

	return &state, nil
}

func guidedOperation(command GuidedUseCaseCommand) (GuidedUseCaseOperation, error) {
	if err := validateAuthoringSessionID(command.OperationID); err != nil {
		return GuidedUseCaseOperation{}, err
	}

	if len(command.Content) > 65536 || command.ExpectedRevision < 1 {
		return GuidedUseCaseOperation{}, errors.New("guided command requires bounded content and a positive revision")
	}

	encoded, err := json.Marshal(command)

	return GuidedUseCaseOperation{ID: command.OperationID, Action: command.Action, NativeRevision: command.NativeRevision, QuestionID: command.QuestionID, InputDigest: command.InputDigest, ExpectedRevision: command.ExpectedRevision, Digest: workflowHash(encoded), Content: command.Content}, err
}

func renderGuidedState(output *bytes.Buffer, state *GuidedUseCaseState) error {
	if state == nil {
		return nil
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "guided_state: %s\n", yamlString(string(encoded)))

	return err
}

func validGuidedIntent(session AuthoringSession) bool {
	state := session.Guided

	return state.SchemaVersion == 1 && session.Revision > 0 && IsValidOpsNamespace(session.Namespace) && strings.TrimSpace(state.Actor) != "" && len(state.Actor) <= 128 && state.ProgressionIntent
}

func boundedGuidedState(state *GuidedUseCaseState) bool {
	return len(state.Capture) <= 65536 && len(state.CompletedOperations) <= 256 && len(state.Checkpoints) <= 8 && len(state.LaterIdeas) <= 64 && boundedGuidedControls(state.Control)
}

// GuidedQuizState retains the canonical quiz owner binding and its committed input digest.
type GuidedQuizState struct {
	SessionID   string `json:"session_id"`
	Revision    string `json:"revision"`
	QuestionID  string `json:"question_id,omitempty"`
	InputDigest string `json:"input_digest"`
}

// GuidedPreparedSpec retains validated generation before canonical artifact effects.
type GuidedPreparedSpec struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	InputDigest string `json:"input_digest"`
	RequestID   string `json:"request_id"`
	ResponseID  string `json:"response_id"`
}

// GuidedSpecDraftReference identifies exactly the generated owner revision.
type GuidedSpecDraftReference struct {
	Reference   string `json:"reference"`
	Revision    string `json:"revision"`
	ContentHash string `json:"content_hash"`
}

// GuidedSpecApproval binds default approval to the initiating intent and generated bytes.
type GuidedSpecApproval struct {
	RunID             string `json:"run_id"`
	Actor             string `json:"actor"`
	Policy            string `json:"policy"`
	GeneratedRevision string `json:"generated_revision"`
	ContentHash       string `json:"content_hash"`
	InputDigest       string `json:"input_digest"`
}

// GuidedSpecApprovalReference retains the canonical approval result.
type GuidedSpecApprovalReference struct {
	GuidedSpecApproval
	Reference string `json:"reference"`
	Revision  string `json:"revision"`
}

// GuidedSpecApprovalPolicy is authorized only by a guided run's retained progression intent.
const GuidedSpecApprovalPolicy = "guided-default/v1"
