package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SpecMutationRequest creates an independent document or saves the exact
// observed identity. Empty Status on save preserves its native lifecycle.
type SpecMutationRequest struct {
	GuidedApproval                                              *GuidedSpecApproval `json:",omitempty"`
	OperationID, Action, Namespace, Reference, ExpectedRevision string
	ProjectID, CaseSlug, Title, Body, Status                    string
}

// SpecMutationResult acknowledges a committed operation without copying bodies.
type SpecMutationResult struct{ OperationID, Reference, Revision string }

type specMutationReceipt struct {
	Schema, Digest, Checksum string
	Result                   SpecMutationResult
}

// MutateSpec serializes body, metadata and idempotent receipt in the existing
// recoverable authoring journal. No model or discovery artifact is required.
func MutateSpec(ctx context.Context, projectDir string, request SpecMutationRequest) (result SpecMutationResult, err error) {
	if err := validateSpecMutation(request); err != nil {
		return result, err
	}

	root, err := CleanContextProjectDir(projectDir)
	if err != nil {
		return result, err
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return result, err
	}

	digest := contenthash.Sum(append([]byte(root+"\x00spec-write\x00"), encoded...))
	r := &fileSessionRepository{projectDir: root}
	err = WithMutation(ctx, root, func(ctx context.Context) error {
		lookup, readErr := r.specReceipt(request.OperationID, digest)
		result, err = lookup.Result, readErr

		if err != nil || lookup.Found {
			return err
		}

		result, err = r.commitSpec(ctx, request, digest)

		return err
	})

	return result, err
}

func validateSpecMutation(request SpecMutationRequest) error {
	if err := authoringstate.ValidateSessionID(request.OperationID); err != nil {
		return err
	}

	if err := validateSpecMutationContent(request); err != nil {
		return err
	}

	if err := quizcore.ValidateScope(quizcore.Scope{ProjectID: request.ProjectID, CaseSlug: request.CaseSlug}); err != nil {
		return err
	}

	switch request.Action {
	case "create":
		return validateSpecCreate(request)
	case "save":
		return validateSpecSave(request)
	default:
		return errors.New("unsupported spec mutation")
	}
}

func (r *fileSessionRepository) commitSpec(ctx context.Context, request SpecMutationRequest, digest string) (SpecMutationResult, error) {
	result := SpecMutationResult{OperationID: request.OperationID}
	if err := r.authoringStateCapacity("receipts"); err != nil {
		return result, err
	}

	document, err := r.specMutationTarget(ctx, request, digest)
	if err != nil {
		return result, err
	}

	changes, err := specDocumentChanges(document, request)
	if err != nil {
		return result, err
	}

	result.Reference = document.Reference
	result.Revision = contenthash.Sum([]byte(strings.Join([]string{r.projectDir, document.Reference, contenthash.Sum(changes[0].After), contenthash.Sum(changes[1].After)}, "\x00")))

	receipt, err := specReceiptChange(digest, result)
	if err != nil {
		return result, err
	}

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournalVersion(projectDumpJournalSchema, append(changes, receipt)); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

func (r *fileSessionRepository) specMutationTarget(ctx context.Context, request SpecMutationRequest, digest string) (SpecDocument, error) {
	if request.CaseSlug != "" {
		if _, err := NewCaseMetadataService(r.projectDir).ReadBounded(ctx, request.CaseSlug); err != nil {
			return SpecDocument{}, err
		}
	}

	if request.Action == "save" {
		return r.specSaveTarget(ctx, request)
	}

	return r.newSpecDocument(request, digest)
}

func (r *fileSessionRepository) specSaveTarget(ctx context.Context, request SpecMutationRequest) (SpecDocument, error) {
	document, err := ReadSpecDocument(ctx, r.projectDir, request.Reference)
	if err != nil {
		return document, err
	}

	if document.Revision != request.ExpectedRevision || document.CaseSlug != request.CaseSlug || document.ProjectID != "" && document.ProjectID != request.ProjectID {
		return document, workflowstate.ErrConflict
	}

	return document, r.guardGuidedSpecApproval(ctx, request, document)
}

func (r *fileSessionRepository) newSpecDocument(request SpecMutationRequest, digest string) (SpecDocument, error) {
	id := "00000-spec-" + digest[:32]
	document := SpecDocument{Reference: strings.Join([]string{opscore.DirectoryName, request.Namespace, "spec", id, opscore.ArtifactBodyFile()}, "/")}

	path, err := r.authoringPath(filepath.Dir(document.Reference))
	if err != nil {
		return document, err
	}

	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return document, errors.Join(workflowstate.ErrConflict, err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: r.projectDir, Namespace: request.Namespace})
	if err != nil {
		return document, err
	}

	document.metadata = opscore.NewArtifactMetadata(opscore.ArtifactOptions{Layout: layout, Type: "spec", Title: request.Title, Producer: "chr"}, id, time.Now().UTC())
	document.images = []workflowstate.FileChange{{Path: document.Reference}, {Path: filepath.Join(filepath.Dir(document.Reference), "metadata.json")}}

	return document, nil
}

func validateSpecMutationContent(request SpecMutationRequest) error {
	if strings.TrimSpace(request.Title) == "" || len(request.Title) > 512 || strings.ContainsAny(request.Title, "\r\n") || strings.TrimSpace(request.Body) == "" || len(request.Body) > workflowstate.FileLimit {
		return errors.New("spec title and body are required; title is at most 512 bytes and body at most 1 MiB")
	}

	if request.Status != "" && request.Status != "draft" && request.Status != "approved" {
		return errors.New("spec writes require draft or approved status")
	}

	return nil
}

func validateSpecCreate(request SpecMutationRequest) error {
	if request.GuidedApproval != nil {
		return errors.New("guided approval requires its generated draft revision")
	}

	if !opsnamespace.Valid(request.Namespace) || request.Reference != "" || request.ExpectedRevision != "" {
		return errors.New("spec creation requires a namespace and new document identity")
	}

	return nil
}

func validateSpecSave(request SpecMutationRequest) error {
	if len(request.ExpectedRevision) != 64 || request.Namespace != "" {
		return errors.New("spec save requires its observed revision")
	}

	return validateSpecReference(request.Reference)
}
