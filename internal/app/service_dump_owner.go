package app

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

func (s service) CreateDump(request DumpCreateRequest) (DumpOperationResult, error) {
	mutation := authoringcore.DumpMutationRequest{OperationID: dumpRequestID(request.OperationID), ExpectedRevision: request.ExpectedRevision, Action: "capture", Namespace: request.Namespace, CaseSlug: request.Scope.CaseSlug, Title: strings.TrimSpace(request.Title), Content: request.Content}
	if request.Scope.CaseSlug != "" {
		layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: request.Namespace})
		if err != nil {
			return DumpOperationResult{}, err
		}

		mutation.Namespace = layout.Namespace()
	}

	result, err := authoringcore.MutateDump(s.ctx, s.projectContext.ProjectDir, mutation, func(ctx context.Context, destination, raw string) (authoringcore.DumpCapturePreparation, error) {
		preparer, _, err := s.bindGeneration(ctx, "dump-"+mutation.OperationID, "dump_curation")
		if err != nil {
			if !errors.Is(err, errLegacyGenerationUnavailable) {
				return authoringcore.DumpCapturePreparation{}, err
			}

			preparer = s
			preparer.ctx = ctx
			preparer.dumpCurator = fallbackDumpCurationProvider{Reason: err.Error()}
		}

		curation, err := preparer.curateDump(request)
		if err != nil {
			return authoringcore.DumpCapturePreparation{}, err
		}

		return authoringcore.DumpCapturePreparation{Content: renderCuratedDumpMarkdown(mutation.Title, request.Scope, raw, curation), RawContent: renderRawDumpMarkdown(mutation.Title, request.Scope, destination, request.Content, curation)}, nil
	})
	if err != nil {
		return DumpOperationResult{}, err
	}

	return s.finishDumpMutation(result, "Captured"), nil
}

func dumpRequestID(id string) string {
	if id == "" {
		return rand.Text()
	}

	return id
}

func (s service) AppendDump(request DumpAppendRequest) (DumpOperationResult, error) {
	return s.editDump(authoringcore.DumpMutationRequest{OperationID: request.OperationID, ExpectedRevision: request.ExpectedRevision, Action: "append", Reference: request.RelativePath, Content: request.Content}, "Appended")
}

func (s service) SaveDump(request DumpSaveRequest) (DumpOperationResult, error) {
	return s.editDump(authoringcore.DumpMutationRequest{OperationID: request.OperationID, ExpectedRevision: request.ExpectedRevision, Action: "save", Reference: request.RelativePath, Content: request.Content}, "Saved")
}

func (s service) editDump(request authoringcore.DumpMutationRequest, action string) (DumpOperationResult, error) {
	result, err := authoringcore.MutateDump(s.ctx, s.projectContext.ProjectDir, request, nil)
	if err != nil {
		return DumpOperationResult{}, err
	}

	return s.finishDumpMutation(result, action), nil
}

func (s service) finishDumpMutation(mutation authoringcore.ProjectDumpMutationResult, action string) DumpOperationResult {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	s.ctx = ctx

	result := DumpOperationResult{Committed: true, Mutation: mutation, RelativePath: mutation.Destination, Message: action + " dump " + mutation.Destination + "."}

	result.SourceHash, result.MetadataHash = mutation.SourceHash, mutation.MetadataHash
	if read, err := s.ReadDump(DumpReadRequest{RelativePath: mutation.Destination}); err == nil && read.SourceHash == mutation.SourceHash && read.MetadataHash == mutation.MetadataHash {
		result.Content, result.Revision = read.Content, read.Revision
	}

	_, indexErr := s.workspaceIndex()

	catalog, catalogErr := s.ListDumps()
	if indexErr != nil || catalogErr != nil {
		result.ProjectionStale = true
		result.Message += " Catalog refresh is pending."

		return result
	}

	result.Dumps = catalog.Dumps

	return result
}
