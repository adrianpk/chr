package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrAssistedCurationMissing permits an initial curation before dump synthesis.
// Other conflicts must preserve the prepared source and existing destination.
var ErrAssistedCurationMissing = errors.New("assisted curated answer source is missing")

func synthesizeAssistedDump(ctx context.Context, opts AssistedArtifactSynthesisOptions) (AssistedArtifactSynthesisResult, error) {
	operation, err := normalizeAssistedDumpOperation(opts)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	r := operation.repository

	retained, found, err := r.assistedDumpReceipt(operation.operationID, operation.digest)
	if err != nil || found {
		return retained, err
	}

	if err := r.projectDumpOperationAvailable(operation.operationID); err != nil {
		return retained, err
	}

	result, changes, err := r.prepareAssistedDump(ctx, operation)
	if err != nil {
		return result, err
	}

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

type assistedDumpOperation struct {
	repository          *fileAuthoringSessionRepository
	options             AssistedArtifactSynthesisOptions
	operationID, digest string
}

func normalizeAssistedDumpOperation(opts AssistedArtifactSynthesisOptions) (assistedDumpOperation, error) {
	var operation assistedDumpOperation

	r, ok := opts.Repository.(*fileAuthoringSessionRepository)
	if !ok {
		return operation, workflowstate.ErrConflict
	}

	root, err := filepath.Abs(opts.ProjectDir)
	if err != nil {
		return operation, err
	}

	repositoryRoot, err := filepath.Abs(r.projectDir)
	if err != nil || root != repositoryRoot {
		return operation, workflowstate.ErrConflict
	}

	opts.ProjectDir = root

	opts.TargetName = strings.TrimSpace(opts.TargetName)
	if err := normalizeAssistedDumpSource(&opts); err != nil {
		return operation, err
	}

	digest, err := assistedDumpDigest(opts)
	if err != nil {
		return operation, err
	}

	operationID := opts.OperationID
	if operationID == "" {
		operationID = "assisted-dump-" + digest[:40]
	}

	repository := *r
	repository.projectDir = root
	operation = assistedDumpOperation{repository: &repository, options: opts, operationID: operationID, digest: digest}

	return operation, validateAuthoringSessionID(operationID)
}

func assistedDumpDigest(opts AssistedArtifactSynthesisOptions) (string, error) {
	summary := opts.CuratedSummary
	summary.GeneratedAt = time.Time{}
	input := struct {
		Root, SessionID, ProjectID, Namespace, Topic, Source, TargetName string
		AllowReplace                                                     bool
		Summary                                                          AssistedCuratedAnswerSummary
	}{opts.ProjectDir, opts.Session.ID, opts.Session.ProjectID, opts.Session.Namespace, opts.Session.Topic, opts.SourceRef.Path, opts.TargetName, opts.AllowReplace, summary}

	encoded, err := json.Marshal(input)
	if err != nil || len(encoded) > WorkflowFileLimit || len(opts.Session.Topic) > 512 {
		return "", errors.New("assisted dump input exceeds its bound")
	}

	return contenthash.Sum(encoded), nil
}

func (r *fileAuthoringSessionRepository) prepareAssistedDump(ctx context.Context, operation assistedDumpOperation) (AssistedArtifactSynthesisResult, []workflowFileChange, error) {
	var result AssistedArtifactSynthesisResult

	opts := operation.options

	session, err := r.assistedDumpSessionImage(opts.Session)
	if err != nil {
		return result, nil, err
	}

	source, err := r.assistedDumpSourceImage(opts)
	if err != nil {
		return result, nil, err
	}

	binding, target, err := r.assistedDumpPreparation(ctx, opts, operation.operationID)
	if err != nil {
		return result, nil, err
	}

	changes, mutation, err := r.dumpMutationChanges(binding, target)
	if err != nil {
		return result, nil, err
	}

	result, session, err = assistedDumpResult(opts, mutation, session, changes)
	if err != nil {
		return result, nil, err
	}

	history, err := r.assistedDumpCommittedImages(result, changes)
	if err != nil {
		return result, nil, err
	}

	receipt, err := assistedDumpReceiptChange(operation.digest, result)

	changes = append(changes, history...)

	return result, append(changes, session, source, receipt), err
}

func (r *fileAuthoringSessionRepository) assistedDumpPreparation(ctx context.Context, opts AssistedArtifactSynthesisOptions, operationID string) (dumpMutationBinding, dumpMutationTarget, error) {
	request := DumpMutationRequest{OperationID: operationID, Action: "capture", Title: opts.Session.Topic}

	target, err := r.dumpMutationTarget(ctx, request)
	if err != nil {
		return dumpMutationBinding{}, target, err
	}

	content := RenderAssistedSynthesizedArtifact(opts.Session, opts.CuratedSummary, opts.SourceRef, authoringDumpRef(target.Destination))

	body, err := dumpEditableBody([]byte(content))
	if err != nil {
		return dumpMutationBinding{}, target, err
	}

	request.Content = string(body)
	if err := validateDumpMutation(request); err != nil {
		return dumpMutationBinding{}, target, err
	}

	prepared, err := defaultDumpCapture(request)(ctx, target.Destination, target.RawReference)

	binding := dumpMutationBinding{Request: request, Destination: target.Destination, RawReference: target.RawReference, Prepared: &prepared}
	if err != nil {
		return binding, target, err
	}

	return binding, target, validateDumpPreparation(binding, prepared)
}

func assistedDumpResult(opts AssistedArtifactSynthesisOptions, mutation ProjectDumpMutationResult, sessionImage workflowFileChange, changes []workflowFileChange) (AssistedArtifactSynthesisResult, workflowFileChange, error) {
	ref := authoringDumpRef(mutation.Destination)

	session := opts.Session
	if len(session.Outputs) >= 16 && !assistedSessionHasOutput(session, ref.Path) || len(session.Provenance) >= 64 {
		return AssistedArtifactSynthesisResult{}, sessionImage, errors.New("assisted dump session capacity exhausted")
	}

	if !assistedSessionHasOutput(session, ref.Path) {
		session.Outputs = append(session.Outputs, ref)
	}

	now := assistedArtifactSynthesisNow(opts.Now)
	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{Relation: AuthoringProvenanceRelationGenerated, Source: opts.SourceRef, Target: ref, Decision: "Synthesized assisted dump " + mutation.OperationID, RecordedAt: now})

	session.UpdatedAt = now
	if session.Revision > 0 {
		session.Revision++
	}

	paths := AuthoringSessionLayout(opts.ProjectDir, session.ID)

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, sessionImage, err
	}

	sessionImage.After = []byte(content)
	fullPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(ref.Path))
	result := AssistedArtifactSynthesisResult{Session: session, Layout: AuthoringArtifactLayout{Ref: ref, FullPath: fullPath, Dir: filepath.Dir(fullPath)}, Mutation: &mutation}
	result.Content = string(authoringImageMap(changes)[ref.Path])

	return result, sessionImage, nil
}

func assistedSessionHasOutput(session AuthoringSession, reference string) bool {
	for _, ref := range session.Outputs {
		if ref.Path == reference {
			return true
		}
	}

	return false
}

func (r *fileAuthoringSessionRepository) assistedDumpCommittedImages(result AssistedArtifactSynthesisResult, changes []workflowFileChange) ([]workflowFileChange, error) {
	artifact := ProjectDumpArtifact{Reference: result.Mutation.Destination, SourceHash: result.Mutation.SourceHash, MetadataHash: result.Mutation.MetadataHash}
	images := authoringImageMap(changes)

	var history []workflowFileChange

	for _, name := range []string{"artifact.md", "metadata.json"} {
		path := filepath.Join(filepath.Dir(artifact.Reference), name)

		change, err := r.projectDumpHistoryChange(projectDumpHistoryRef(artifact)+"/"+name, images[path])
		if err != nil {
			return nil, err
		}

		history = append(history, change)
	}

	return history, nil
}
