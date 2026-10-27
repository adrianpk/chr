package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// AuthoringWorkRequest proposes a native description from observed authoring
// sources. Instruction is the retained user instruction, never a model summary.
type AuthoringWorkRequest struct {
	Binding            workflowstate.MutationBinding `json:"binding"`
	SessionID          string                        `json:"session_id"`
	ExpectedRevision   int64                         `json:"expected_revision"`
	ExpectedSourceHash string                        `json:"expected_source_hash"`
	Intent             authoringstate.Intent         `json:"intent"`
	Instruction        string                        `json:"instruction"`
}

// PreparedAuthoringWork is a preview to retain before dispatch. Preparing again
// can reserve another ticket identity; callers replay the original retained input.
type PreparedAuthoringWork struct {
	Scope         core.AuthoringWorkScope          `json:"scope"`
	Authorization *core.AuthoringWorkAuthorization `json:"authorization,omitempty"`
	Mutation      PreparedRecordMutation           `json:"mutation"`
}

// AuthoringWorkResult reports the actual native owner receipt and retained scope.
// It never reports admission or execution from preparation success.
type AuthoringWorkResult struct {
	Scope         core.AuthoringWorkScope          `json:"scope"`
	Authorization *core.AuthoringWorkAuthorization `json:"authorization,omitempty"`
	Result        RecordMutationResult             `json:"result"`
}

// PrepareAuthoringWork previews a backlog ticket through the shared record owner.
// Existing real dump/spec content supplies its description without another model
// choice. This neither claims work nor changes its execution/review policy.
func (s service) PrepareAuthoringWork(ctx context.Context, request AuthoringWorkRequest) (result PreparedAuthoringWork, err error) {
	if err := request.Binding.Validate(); err != nil {
		return result, err
	}

	err = core.WithAuthoringMutation(ctx, s.projectContext.ProjectDir, func(ctx context.Context) error {
		snapshot, err := s.ObserveAuthoring(ctx, request.SessionID)
		if err != nil {
			return err
		}

		if err := guardAuthoringWorkSource(snapshot, request, s.projectContext.ProjectID); err != nil {
			return err
		}

		source, err := authoringWorkDescription(snapshot)
		if err != nil {
			return err
		}

		title, synopsis := snapshot.Session.Topic, snapshot.Session.Topic
		body := source.Body + "\n\n## Authoring Source\n\n" + source.Ref.Path + "\n"

		prepared, err := s.PrepareRecordMutation(ctx, snapshot.Session.Namespace, request.Binding, RecordMutationRequest{Kind: RecordTicket, Action: RecordCreate, Title: &title, Synopsis: &synopsis, Body: &body})
		if err != nil {
			return err
		}

		prepared.Native.Record, err = s.authoringWorkContract(prepared.Native.Record, title, request, source)
		if err != nil {
			return err
		}

		image, err := json.Marshal(prepared)
		if err != nil {
			return err
		}

		scope, err := core.SealAuthoringWorkScope(core.AuthoringWorkScope{Binding: request.Binding, Intent: request.Intent, Instruction: request.Instruction, ProjectID: snapshot.Session.ProjectID, Namespace: snapshot.Session.Namespace, SessionID: snapshot.Session.ID, SessionRef: snapshot.SessionRef, SessionRevision: snapshot.Session.Revision, SourceHash: snapshot.SourceHash, DescriptionRef: source.Ref.Path, DescriptionHash: source.SourceHash, ContractHash: prepared.Native.Record.Task.ContractHash, RecordHash: recordDigest(image)})
		result = PreparedAuthoringWork{Scope: scope, Mutation: prepared}

		return err
	})

	return result, err
}

// ApplyPreparedAuthoringWork creates only the previewed proposed native record.
// A confirmed owner receipt takes precedence over later source edits on replay.
func (s service) ApplyPreparedAuthoringWork(ctx context.Context, prepared PreparedAuthoringWork) (AuthoringWorkResult, error) {
	return s.applyPreparedAuthoringWork(ctx, prepared, nil)
}

// ApplyPreparedAuthoringWorkGuarded rechecks caller authority while holding the
// authoring owner, immediately before a new native effect. An existing owner
// receipt reconciles first without acquiring fresh authority for a changed scope.
func (s service) ApplyPreparedAuthoringWorkGuarded(ctx context.Context, prepared PreparedAuthoringWork, guard core.AuthoringCommitGuard) (AuthoringWorkResult, error) {
	if guard == nil {
		return AuthoringWorkResult{}, workflowstate.ErrConflict
	}

	return s.applyPreparedAuthoringWork(ctx, prepared, guard)
}

func (s service) applyPreparedAuthoringWork(ctx context.Context, prepared PreparedAuthoringWork, guard core.AuthoringCommitGuard) (result AuthoringWorkResult, err error) {
	if err := validatePreparedAuthoringWork(prepared, s.projectContext.ProjectID); err != nil {
		return result, err
	}

	mutation, err := authoringWorkMutation(prepared)
	if err != nil {
		return result, err
	}

	err = core.WithAuthoringMutation(ctx, s.projectContext.ProjectDir, func(ctx context.Context) error {
		var applyErr error

		result, applyErr = s.applyAuthoringWork(ctx, prepared, mutation, guard)

		return applyErr
	})

	return result, err
}

func (s service) applyAuthoringWork(ctx context.Context, prepared PreparedAuthoringWork, mutation PreparedRecordMutation, guard core.AuthoringCommitGuard) (AuthoringWorkResult, error) {
	result := AuthoringWorkResult{Scope: prepared.Scope, Authorization: prepared.Authorization}

	store, err := s.scopedRecordStore(prepared.Scope.Namespace)
	if err != nil {
		return result, err
	}

	retained, found, err := store.ReserveRecordMutation(ctx, *mutation.Native)
	if err != nil {
		return result, err
	}

	if found {
		view := nativeRecordView(retained.Record, RecordTicket)
		result.Result = RecordMutationResult{Record: view, ReceiptRef: retained.ReceiptRef, ArgumentDigest: retained.ArgumentDigest}

		return result, nil
	}

	if err := s.guardPreparedAuthoringSource(ctx, prepared.Scope); err != nil {
		return result, err
	}

	if guard != nil {
		if err := guard(ctx); err != nil {
			return result, err
		}
	}

	result.Result, err = s.ApplyPreparedRecordMutation(ctx, mutation)

	return result, err
}

func (s service) guardPreparedAuthoringSource(ctx context.Context, scope core.AuthoringWorkScope) error {
	snapshot, err := s.ObserveAuthoring(ctx, scope.SessionID)
	if err != nil {
		return err
	}

	request := AuthoringWorkRequest{SessionID: scope.SessionID, ExpectedRevision: scope.SessionRevision, ExpectedSourceHash: scope.SourceHash, Intent: scope.Intent, Instruction: scope.Instruction}
	if err := guardAuthoringWorkSource(snapshot, request, s.projectContext.ProjectID); err != nil {
		return err
	}

	source, err := authoringWorkDescription(snapshot)
	if err != nil {
		return err
	}

	if source.Ref.Path != scope.DescriptionRef || source.SourceHash != scope.DescriptionHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func guardAuthoringWorkSource(snapshot core.AuthoringSnapshot, request AuthoringWorkRequest, projectID string) error {
	if request.Intent != authoringstate.IntentPrepare && request.Intent != authoringstate.IntentDevelop {
		return errors.New("explanation cannot prepare native work")
	}

	if strings.TrimSpace(request.Instruction) == "" || len(request.Instruction) > 8192 {
		return errors.New("work preparation requires the bounded original instruction")
	}

	if snapshot.Session.ProjectID != projectID || snapshot.Session.ID != request.SessionID || snapshot.Session.Revision != request.ExpectedRevision || snapshot.SourceHash != request.ExpectedSourceHash || snapshot.Question != nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func authoringWorkDescription(snapshot core.AuthoringSnapshot) (core.AuthoringArtifactObservation, error) {
	var selected core.AuthoringArtifactObservation

	for _, artifact := range snapshot.Artifacts {
		if artifact.Ref.Role != authoringcore.ArtifactRoleOutput {
			continue
		}

		if artifact.Ref.Kind == authoringcore.ArtifactKindSpec || (selected.Ref.Path == "" && artifact.Ref.Kind == authoringcore.ArtifactKindDump) {
			selected = artifact
		}
	}

	if selected.Ref.Path == "" || strings.TrimSpace(selected.Body) == "" || len(selected.Body) > 16384 {
		return selected, errors.New("prepare a canonical dump or spec description of at most 16 KiB before native work")
	}

	return selected, nil
}

func validatePreparedAuthoringWork(prepared PreparedAuthoringWork, projectID string) error {
	if err := core.ValidateAuthoringWorkAuthorization(prepared.Scope, prepared.Authorization); err != nil {
		return err
	}

	if err := guardAuthoringWorkOwner(prepared, projectID); err != nil {
		return err
	}

	image, err := json.Marshal(prepared.Mutation)
	if err != nil {
		return err
	}

	if recordDigest(image) != prepared.Scope.RecordHash {
		return fmt.Errorf("%w: prepared description changed", workflowstate.ErrConflict)
	}

	return nil
}

func guardAuthoringWorkOwner(prepared PreparedAuthoringWork, projectID string) error {
	scope, mutation := prepared.Scope, prepared.Mutation
	if scope.ProjectID != projectID || mutation.ProjectID != projectID || mutation.Namespace != scope.Namespace || mutation.Kind != RecordTicket || mutation.Case != nil || mutation.Native == nil {
		return workflowstate.ErrConflict
	}

	return guardAuthoringWorkRecord(*mutation.Native, scope)
}

func guardAuthoringWorkRecord(request core.WorkflowMutationRequest, scope core.AuthoringWorkScope) error {
	if request.Binding != scope.Binding || request.ExpectedRevision != 0 || request.Record.Task == nil || request.Record.Task.ContractHash != scope.ContractHash || request.Record.Task.Phase != "pending" || request.Record.Task.Status != core.TicketStatusBacklog {
		return workflowstate.ErrConflict
	}

	return nil
}

func authoringWorkMutation(prepared PreparedAuthoringWork) (PreparedRecordMutation, error) {
	// Clone the retained input before adding canonical provenance. Reusing a
	// preview must not change its own immutable argument digest in memory.
	data, err := json.Marshal(prepared.Mutation)
	if err != nil {
		return PreparedRecordMutation{}, err
	}

	var result PreparedRecordMutation
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	if result.Native.Record.Metadata == nil {
		result.Native.Record.Metadata = map[string]json.RawMessage{}
	}

	for key, value := range map[string]any{"authoring_scope": prepared.Scope, "authoring_authorization": prepared.Authorization} {
		encoded, err := json.Marshal(value)
		if err != nil {
			return result, err
		}

		result.Native.Record.Metadata[key] = encoded
	}

	return result, nil
}
