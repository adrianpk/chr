package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrAssistedCurationMissing permits an initial curation before dump synthesis.
// Other conflicts must preserve the prepared source and existing destination.
var ErrAssistedCurationMissing = errors.New("assisted curated answer source is missing")

type assistedDumpImages struct {
	Result  AssistedArtifactSynthesisResult
	Changes []workflowstate.FileChange
}

type assistedDumpTarget struct {
	Binding dumpMutationBinding
	Target  dumpMutationTarget
}

type assistedDumpSessionImage struct {
	Result  AssistedArtifactSynthesisResult
	Session workflowstate.FileChange
}

func synthesizeAssistedDump(ctx context.Context, opts AssistedArtifactSynthesisOptions) (AssistedArtifactSynthesisResult, error) {
	operation, err := normalizeAssistedDumpOperation(opts)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, err
	}

	r := operation.repository

	retained, err := r.assistedDumpReceipt(operation.operationID, operation.digest)
	if err != nil || retained.Exists {
		return retained.Result, err
	}

	if err := r.projectDumpOperationAvailable(operation.operationID); err != nil {
		return retained.Result, err
	}

	prepared, err := r.prepareAssistedDump(ctx, operation)
	if err != nil {
		return prepared.Result, err
	}

	result := prepared.Result

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, prepared.Changes); err != nil {
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

	return operation, authoringstate.ValidateSessionID(operationID)
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
	if err != nil || len(encoded) > workflowstate.FileLimit || len(opts.Session.Topic) > 512 {
		return "", errors.New("assisted dump input exceeds its bound")
	}

	return contenthash.Sum(encoded), nil
}

func (r *fileAuthoringSessionRepository) prepareAssistedDump(ctx context.Context, operation assistedDumpOperation) (assistedDumpImages, error) {
	var result AssistedArtifactSynthesisResult

	opts := operation.options

	session, err := r.assistedDumpSessionImage(opts.Session)
	if err != nil {
		return assistedDumpImages{Result: result}, err
	}

	source, err := r.assistedDumpSourceImage(opts)
	if err != nil {
		return assistedDumpImages{Result: result}, err
	}

	prepared, err := r.assistedDumpPreparation(ctx, opts, operation.operationID)
	if err != nil {
		return assistedDumpImages{Result: result}, err
	}

	images, err := r.dumpMutationChanges(prepared.Binding, prepared.Target)
	if err != nil {
		return assistedDumpImages{Result: result}, err
	}

	changes := images.Changes

	sessionResult, err := assistedDumpResult(opts, images.Result, session, changes)
	if err != nil {
		return assistedDumpImages{Result: sessionResult.Result}, err
	}

	result, session = sessionResult.Result, sessionResult.Session

	history, err := r.assistedDumpCommittedImages(result, changes)
	if err != nil {
		return assistedDumpImages{Result: result}, err
	}

	receipt, err := assistedDumpReceiptChange(operation.digest, result)

	changes = append(changes, history...)

	return assistedDumpImages{Result: result, Changes: append(changes, session, source, receipt)}, err
}

func (r *fileAuthoringSessionRepository) assistedDumpPreparation(ctx context.Context, opts AssistedArtifactSynthesisOptions, operationID string) (assistedDumpTarget, error) {
	request := DumpMutationRequest{OperationID: operationID, Action: "capture", Title: opts.Session.Topic}

	target, err := r.dumpMutationTarget(ctx, request)
	if err != nil {
		return assistedDumpTarget{Target: target}, err
	}

	content := RenderAssistedSynthesizedArtifact(opts.Session, opts.CuratedSummary, opts.SourceRef, authoringDumpRef(target.Destination))

	body, err := dumpEditableBody([]byte(content))
	if err != nil {
		return assistedDumpTarget{Target: target}, err
	}

	request.Content = string(body)
	if err := validateDumpMutation(request); err != nil {
		return assistedDumpTarget{Target: target}, err
	}

	prepared, err := defaultDumpCapture(request)(ctx, target.Destination, target.RawReference)

	binding := dumpMutationBinding{Request: request, Destination: target.Destination, RawReference: target.RawReference, Prepared: &prepared}
	if err != nil {
		return assistedDumpTarget{Binding: binding, Target: target}, err
	}

	return assistedDumpTarget{Binding: binding, Target: target}, validateDumpPreparation(binding, prepared)
}

func assistedDumpResult(opts AssistedArtifactSynthesisOptions, mutation ProjectDumpMutationResult, sessionImage workflowstate.FileChange, changes []workflowstate.FileChange) (assistedDumpSessionImage, error) {
	ref := authoringDumpRef(mutation.Destination)

	session := opts.Session
	if len(session.Outputs) >= 16 && !assistedSessionHasOutput(session, ref.Path) || len(session.Provenance) >= 64 {
		return assistedDumpSessionImage{Session: sessionImage}, errors.New("assisted dump session capacity exhausted")
	}

	if !assistedSessionHasOutput(session, ref.Path) {
		session.Outputs = append(session.Outputs, ref)
	}

	now := assistedArtifactSynthesisNow(opts.Now)
	session.Provenance = append(session.Provenance, authoringcore.AuthoringProvenanceLink{Relation: authoringcore.AuthoringProvenanceRelationGenerated, Source: opts.SourceRef, Target: ref, Decision: "Synthesized assisted dump " + mutation.OperationID, RecordedAt: now})

	session.UpdatedAt = now
	if session.Revision > 0 {
		session.Revision++
	}

	paths := AuthoringSessionLayout(opts.ProjectDir, session.ID)

	content, err := RenderAuthoringSession(session, paths)
	if err != nil {
		return assistedDumpSessionImage{Session: sessionImage}, err
	}

	sessionImage.After = []byte(content)
	fullPath := filepath.Join(opts.ProjectDir, filepath.FromSlash(ref.Path))
	result := AssistedArtifactSynthesisResult{Session: session, Layout: AuthoringArtifactLayout{Ref: ref, FullPath: fullPath, Dir: filepath.Dir(fullPath)}, Mutation: &mutation}
	result.Content = string(authoringImageMap(changes)[ref.Path])

	return assistedDumpSessionImage{Result: result, Session: sessionImage}, nil
}

func assistedSessionHasOutput(session authoringcore.AuthoringSession, reference string) bool {
	for _, ref := range session.Outputs {
		if ref.Path == reference {
			return true
		}
	}

	return false
}

func (r *fileAuthoringSessionRepository) assistedDumpCommittedImages(result AssistedArtifactSynthesisResult, changes []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
	artifact := ProjectDumpArtifact{Reference: result.Mutation.Destination, SourceHash: result.Mutation.SourceHash, MetadataHash: result.Mutation.MetadataHash}
	images := authoringImageMap(changes)

	var history []workflowstate.FileChange

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
