package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PrepareRecordMutation freezes validated owner input before durable dispatch.
func (s service) PrepareRecordMutation(ctx context.Context, namespace string, binding workflowstate.MutationBinding, request RecordMutationRequest) (PreparedRecordMutation, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: namespace})
	if err != nil {
		return PreparedRecordMutation{}, err
	}

	namespace = layout.Namespace()

	prepared := PreparedRecordMutation{ProjectID: s.projectContext.ProjectID, Namespace: namespace, Kind: request.Kind, Identity: request.Identity, ParentID: request.ParentID}
	if !validRecordMutationAction(request) {
		return prepared, ErrRecordUnsupported
	}

	if request.Kind == RecordUseCase {
		target, err := s.prepareCaseRecordMutation(ctx, binding, request)
		target.Namespace = namespace
		prepared.Case = &target

		return prepared, err
	}

	store, err := s.scopedRecordStore(namespace)
	if err != nil {
		return prepared, err
	}

	mutation, err := s.prepareNativeMutation(ctx, store, binding, request)
	if err != nil {
		return prepared, err
	}

	if request.Kind == RecordSlice && request.Action == "create" && request.Slice != nil {
		prepared.Identity = request.Slice.ID
	}

	prepared.Native = &mutation

	return prepared, nil
}

// ApplyPreparedRecordMutation invokes the actual owner with the retained input.
func (s service) ApplyPreparedRecordMutation(ctx context.Context, prepared PreparedRecordMutation) (RecordMutationResult, error) {
	if prepared.ProjectID != s.projectContext.ProjectID || !opsnamespace.Valid(prepared.Namespace) {
		return RecordMutationResult{}, workflowstate.ErrConflict
	}

	if err := validatePreparedRecordOwner(prepared); err != nil {
		return RecordMutationResult{}, err
	}

	if prepared.Case != nil {
		return s.applyPreparedCaseMutation(ctx, prepared)
	}

	result, err := s.ApplyWorkflowRecordMutation(ctx, prepared.Namespace, *prepared.Native)
	if err != nil {
		return RecordMutationResult{}, err
	}

	view := nativeRecordView(result.Record, prepared.Kind)
	if prepared.Kind == RecordSlice {
		views, err := filterRecordViews(sliceRecordViews(result.Record), RecordQuery{Identity: prepared.Identity, ParentID: prepared.ParentID})
		if err != nil {
			return RecordMutationResult{}, err
		}

		view = views[0]
	}

	return RecordMutationResult{Record: view, ReceiptRef: result.ReceiptRef, ArgumentDigest: result.ArgumentDigest}, nil
}

func (s service) prepareNativeRecordUpdate(ctx context.Context, store *workflowcore.Store, binding workflowstate.MutationBinding, request RecordMutationRequest) (workflowcore.MutationRequest, error) {
	record, err := s.nativeMutationTarget(ctx, store, request.Kind, request.Identity)
	if err != nil {
		return workflowcore.MutationRequest{}, err
	}

	if err := guardRecordObservation(record, request); err != nil {
		return workflowcore.MutationRequest{}, err
	}

	if unsupportedRecordDisplayPatch(request) {
		return workflowcore.MutationRequest{}, ErrRecordUnsupported
	}

	if request.Title == nil && request.Synopsis == nil {
		return workflowcore.MutationRequest{}, errors.New("record update requires a display field")
	}

	record, err = applyNativeRecordDisplay(record, request)

	return workflowcore.MutationRequest{Binding: binding, Record: record, ExpectedRevision: request.ExpectedRevision}, err
}

func (s service) nativeMutationTarget(ctx context.Context, store *workflowcore.Store, kind RecordKind, identity string) (workflowcore.Record, error) {
	if !validRecordIdentity(identity) {
		return workflowcore.Record{}, errors.New("an exact record identity is required")
	}

	snapshot, err := store.Snapshot(ctx)
	if err != nil {
		return workflowcore.Record{}, err
	}

	for _, record := range snapshot.Records {
		if kind == RecordTicket && record.Task != nil && record.Task.Key == identity {
			return record, nil
		}

		if kind == RecordSliceSet && record.Set != nil && record.ID == identity {
			return record, nil
		}
	}

	return workflowcore.Record{}, fmt.Errorf("%w: native record is absent; unadopted records require their explicit owner", ErrRecordUnsupported)
}

func guardRecordObservation(record workflowcore.Record, request RecordMutationRequest) error {
	if request.ExpectedRevision == 0 || request.ExpectedHash == "" || record.Revision != request.ExpectedRevision || record.SourceHash != request.ExpectedHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s service) applyPreparedCaseMutation(ctx context.Context, prepared PreparedRecordMutation) (RecordMutationResult, error) {
	if prepared.Kind != RecordUseCase || prepared.Case.ProjectID != prepared.ProjectID || prepared.Case.Namespace != prepared.Namespace {
		return RecordMutationResult{}, workflowstate.ErrConflict
	}

	result, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).SaveBound(ctx, *prepared.Case)
	if err != nil {
		return RecordMutationResult{}, err
	}

	view, err := caseRecordView(s.projectContext.ProjectDir, result.Metadata)
	if err != nil {
		return RecordMutationResult{}, err
	}

	return RecordMutationResult{Record: view, ReceiptRef: result.ReceiptRef, ArgumentDigest: result.ArgumentDigest}, nil
}

func (s service) prepareNativeMutation(ctx context.Context, store *workflowcore.Store, binding workflowstate.MutationBinding, request RecordMutationRequest) (workflowcore.MutationRequest, error) {
	if request.Kind == RecordSlice {
		return s.prepareSliceRecordMutation(ctx, store, binding, request)
	}

	if request.Action == "create" {
		return s.prepareNativeRecordCreation(ctx, store, binding, request)
	}

	return s.prepareNativeRecordUpdate(ctx, store, binding, request)
}

func unsupportedRecordDisplayPatch(request RecordMutationRequest) bool {
	return request.Body != nil || request.Context != nil || request.Slice != nil || request.Validation != nil || request.PlanRef != "" || request.PlanHash != "" || request.UseCase != "" || request.ParentID != ""
}

func applyNativeRecordDisplay(record workflowcore.Record, request RecordMutationRequest) (workflowcore.Record, error) {
	if record.Task != nil {
		return workflowcore.UpdateWorkflowTicketDisplay(record, request.Title, request.Synopsis)
	}

	if request.Synopsis != nil || request.Title == nil {
		return record, ErrRecordUnsupported
	}

	encoded, err := json.Marshal(*request.Title)
	record.Metadata["title"] = encoded

	return record, err
}

func validRecordMutationAction(request RecordMutationRequest) bool {
	return (request.Action == "create" || request.Action == "update") && (request.Kind == RecordUseCase || nativeRecordKind(request.Kind))
}

func validatePreparedRecordOwner(prepared PreparedRecordMutation) error {
	if (prepared.Case == nil) == (prepared.Native == nil) {
		return errors.New("prepared mutation requires one canonical owner")
	}

	if prepared.Case != nil {
		if prepared.Kind != RecordUseCase {
			return workflowstate.ErrConflict
		}

		return nil
	}

	return validatePreparedNativeKind(prepared.Kind, prepared.Native.Record)
}

func validatePreparedNativeKind(kind RecordKind, record workflowcore.Record) error {
	if kind == RecordTicket && record.Task != nil && record.Kind == "ticket" {
		return nil
	}

	if (kind == RecordSlice || kind == RecordSliceSet) && record.Set != nil && record.Kind == "plan" {
		return nil
	}

	return workflowstate.ErrConflict
}
