package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// DumpMutationRequest binds a capture, append or full-body save to one operation.
// Capture may omit ExpectedRevision to extend the latest general body under the
// owner. Append and save require the revision returned by ReadDumpDocument.
type DumpMutationRequest struct {
	OperationID, Action, Namespace, Reference, ExpectedRevision string
	Title, Content, CaseSlug                                    string
}

// DumpCapturePreparation contains provider-prepared documents, retained before
// publication so a retry never prepares different content for one operation.
type DumpCapturePreparation struct{ Content, RawContent string }

// DumpCapturePreparer receives the resolved canonical and raw component refs.
type DumpCapturePreparer func(context.Context, string, string) (DumpCapturePreparation, error)

// DumpDocumentObservation binds editable content to its resolved identity and root.
// Content retains the complete stored image; Body excludes its frontmatter.
type DumpDocumentObservation struct {
	Artifact                ProjectDumpArtifact
	Revision, Content, Body string
}

// ReadDumpDocument resolves old references and returns an exact body/metadata revision.
func ReadDumpDocument(ctx context.Context, projectDir, reference string) (DumpDocumentObservation, error) {
	root, err := filepath.Abs(projectDir)
	if err != nil || strings.TrimSpace(projectDir) == "" {
		return DumpDocumentObservation{}, errors.New("project directory is required")
	}

	artifact, err := ResolveProjectDumpReference(ctx, root, reference)
	if err != nil {
		return DumpDocumentObservation{}, err
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	fileImages, err := r.projectDumpImages(artifact)
	body := fileImages.Body

	observed := DumpDocumentObservation{Artifact: artifact, Revision: dumpDocumentRevision(root, artifact), Content: string(body)}
	if err != nil {
		return observed, err
	}

	editable, err := dumpEditableBody(body)
	if err != nil {
		return observed, err
	}

	observed.Body = string(editable)

	return observed, r.checkProjectDumpInspection(ctx)
}

func dumpDocumentRevision(root string, artifact ProjectDumpArtifact) string {
	return contenthash.Sum([]byte(strings.Join([]string{root, artifact.Reference, artifact.SourceHash, artifact.MetadataHash}, "\x00")))
}

// MutateDump publishes content, history, metadata and receipt through the shared
// authoring journal. General captures always resolve the project-wide destination.
func MutateDump(ctx context.Context, projectDir string, request DumpMutationRequest, prepare DumpCapturePreparer) (ProjectDumpMutationResult, error) {
	root, err := filepath.Abs(projectDir)
	if err != nil || strings.TrimSpace(projectDir) == "" {
		return ProjectDumpMutationResult{}, errors.New("project directory is required")
	}

	return (&fileAuthoringSessionRepository{projectDir: root}).mutateDump(ctx, request, prepare)
}

func (r *fileAuthoringSessionRepository) mutateDump(ctx context.Context, request DumpMutationRequest, prepare DumpCapturePreparer) (result ProjectDumpMutationResult, err error) {
	if err := validateDumpMutation(request); err != nil {
		return result, err
	}

	digest, err := dumpMutationDigest(r.projectDir, request)
	if err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		var found bool

		receiptLookup, receiptErr := r.projectDumpReceipt(request.OperationID, digest)

		result, found, err = receiptLookup.Result, receiptLookup.Found, receiptErr
		if err != nil || found {
			return err
		}

		binding, readErr := r.prepareDumpMutation(ctx, request, digest, prepare)
		if readErr != nil {
			return readErr
		}

		result, err = r.commitDumpMutation(ctx, binding)

		return err
	})

	return result, err
}

func validateDumpMutation(request DumpMutationRequest) error {
	if err := validateAuthoringSessionID(request.OperationID); err != nil {
		return fmt.Errorf("dump operation identity: %w", err)
	}

	if err := validateDumpMutationInput(request); err != nil {
		return err
	}

	switch request.Action {
	case "capture":
		return validateDumpCapture(request)
	case "append", "save":
		return validateDumpEdit(request)
	default:
		return errors.New("unsupported dump mutation")
	}
}

func validateDumpMutationInput(request DumpMutationRequest) error {
	if len(request.Content) > WorkflowFileLimit || strings.TrimSpace(request.Content) == "" || len(request.Title) > 512 {
		return errors.New("dump content is required and bounded to 1 MiB; title is bounded to 512 bytes")
	}

	if len(request.ExpectedRevision) > 64 || len(request.Namespace) > 128 || len(request.CaseSlug) > 128 {
		return errors.New("dump revision or scope exceeds its bound")
	}

	return nil
}

func validateDumpEdit(request DumpMutationRequest) error {
	if request.Reference == "" || len(request.ExpectedRevision) != 64 || request.CaseSlug != "" || request.Namespace != "" {
		return errors.New("dump edits require an exact document reference and observed revision")
	}

	return ValidateProjectDumpReference(request.Reference)
}

func validateDumpCapture(request DumpMutationRequest) error {
	if request.Reference != "" || strings.TrimSpace(request.Title) == "" || len(request.Content) > 65536 {
		return errors.New("capture requires a title and at most 64 KiB of content")
	}

	if request.CaseSlug == "" {
		if request.Namespace != "" {
			return errors.New("general capture derives its namespace from the project destination")
		}

		return nil
	}

	if !opsnamespace.Valid(request.Namespace) || len(request.Namespace) > 128 || len(request.CaseSlug) > 128 {
		return errors.New("case capture requires an existing case and namespace")
	}

	return nil
}

func dumpMutationDigest(root string, request DumpMutationRequest) (string, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	return contenthash.Sum(append([]byte(root+"\x00dump-write\x00"), encoded...)), nil
}
