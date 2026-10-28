package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s service) prepareNativeRecordCreation(ctx context.Context, store *workflowcore.Store, binding workflowstate.MutationBinding, request RecordMutationRequest) (workflowcore.MutationRequest, error) {
	if request.Title == nil || strings.TrimSpace(*request.Title) == "" {
		return workflowcore.MutationRequest{}, errors.New("record creation requires a title")
	}

	if request.Identity != "" || request.ExpectedRevision != 0 || request.ExpectedHash != "" {
		return workflowcore.MutationRequest{}, errors.New("creation reserves its identity and initial revision")
	}

	var (
		record workflowcore.Record
		err    error
	)

	switch request.Kind {
	case RecordTicket:
		record, err = s.prepareTicketRecord(request)
	case RecordSliceSet:
		record, err = s.prepareSetRecord(ctx, store, request)
	default:
		return workflowcore.MutationRequest{}, ErrRecordUnsupported
	}

	return workflowcore.MutationRequest{Binding: binding, Title: *request.Title, Record: record}, err
}

func (s service) prepareTicketRecord(request RecordMutationRequest) (workflowcore.Record, error) {
	if request.Synopsis == nil || strings.TrimSpace(*request.Synopsis) == "" {
		return workflowcore.Record{}, errors.New("ticket creation requires a synopsis")
	}

	if unsupportedTicketCreation(request) {
		return workflowcore.Record{}, ErrRecordUnsupported
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.Record{}, err
	}

	number, err := s.reserveNextTicketKeyNumber(config.TicketKeyPrefix)
	if err != nil {
		return workflowcore.Record{}, err
	}

	key := fmt.Sprintf("%s-%d", config.TicketKeyPrefix, number)
	now := s.now().UTC()

	body, err := ticketcore.NewRenderer().Render(ticketcore.Ticket{Key: key, Title: *request.Title, Synopsis: *request.Synopsis, Status: ticketcore.StatusBacklog, UseCase: request.UseCase, Body: recordString(request.Body), CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return workflowcore.Record{}, err
	}

	contract, err := json.Marshal(struct{ Title, Synopsis, Body string }{*request.Title, *request.Synopsis, recordString(request.Body)})

	return workflowcore.Record{Kind: "ticket", Body: []byte(body), Task: &workflowcore.Task{Schema: workflowcore.WorkflowSchema, Key: key, Status: string(ticketcore.StatusBacklog), Phase: "pending", UseCase: request.UseCase, ContractHash: recordDigest(contract)}}, err
}

func (s service) prepareSetRecord(ctx context.Context, store *workflowcore.Store, request RecordMutationRequest) (workflowcore.Record, error) {
	if request.Validation == nil || request.PlanRef == "" || request.PlanHash == "" {
		return workflowcore.Record{}, errors.New("slice-set creation requires an observed plan and explicit closing validation")
	}

	if request.Slice != nil || request.Synopsis != nil || request.Context != nil || request.Body != nil || request.ParentID != "" {
		return workflowcore.Record{}, ErrRecordUnsupported
	}

	body, err := store.ReadContract(ctx, request.PlanRef, request.PlanHash)
	if err != nil {
		return workflowcore.Record{}, err
	}

	return workflowcore.Record{Kind: "plan", Body: body, Set: &workflowstate.WorkflowSet{Schema: workflowcore.WorkflowSchema, Status: "proposed", UseCase: request.UseCase, PlanRef: request.PlanRef, PlanHash: request.PlanHash, ClosingValidation: *request.Validation}}, nil
}

func recordString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func recordDigest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func (s service) prepareCaseRecordMutation(ctx context.Context, binding workflowstate.MutationBinding, request RecordMutationRequest) (authoringcore.CaseMutationRequest, error) {
	result := authoringcore.CaseMutationRequest{Binding: binding, ProjectID: s.projectContext.ProjectID, Namespace: "default"}
	if unsupportedCaseRecordFields(request) {
		return result, ErrRecordUnsupported
	}

	metadata, err := s.caseMutationTarget(ctx, request)
	if err != nil {
		return result, err
	}

	if request.Title != nil {
		metadata.Name = *request.Title
	}

	if request.Synopsis != nil {
		metadata.Description = *request.Synopsis
	}

	if request.Context != nil {
		metadata.Context = *request.Context
	}

	result.Metadata = metadata

	return result, nil
}

func unsupportedTicketCreation(request RecordMutationRequest) bool {
	return request.Slice != nil || request.Validation != nil || request.PlanRef != "" || request.PlanHash != "" || request.Context != nil || request.ParentID != ""
}

func (s service) caseMutationTarget(ctx context.Context, request RecordMutationRequest) (authoringcore.CaseMetadata, error) {
	if request.Action == "create" {
		if request.Identity != "" || request.ExpectedHash != "" || request.ExpectedRevision != 0 {
			return authoringcore.CaseMetadata{}, errors.New("case creation reserves its identity")
		}

		return authoringcore.CaseMetadata{SchemaVersion: authoringcore.CaseMetadataSchemaVersion}, nil
	}

	if !validRecordIdentity(request.Identity) || request.ExpectedHash == "" {
		return authoringcore.CaseMetadata{}, workflowstate.ErrConflict
	}

	metadata, err := authoringcore.NewCaseMetadataService(s.projectContext.ProjectDir).ReadBounded(ctx, request.Identity)
	if err != nil {
		return metadata, err
	}

	if metadata.SourceHash != request.ExpectedHash {
		return metadata, workflowstate.ErrConflict
	}

	return metadata, nil
}

func unsupportedCaseRecordFields(request RecordMutationRequest) bool {
	return request.ParentID != "" || request.Slice != nil || request.Validation != nil || request.PlanRef != "" || request.PlanHash != "" || request.UseCase != "" || request.Body != nil
}
