package authoring

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type normalizedAuthoringRequest struct {
	Request ContinuationRequest
	Digest  string
}

func normalizeAuthoringRequest(request ContinuationRequest) (normalizedAuthoringRequest, error) {
	if err := request.Binding.Validate(); err != nil {
		return normalizedAuthoringRequest{Request: request}, err
	}

	if request.Action == "create" && request.SessionID == "" {
		request.SessionID = "author-" + contenthash.Sum([]byte(request.Binding.RequestID))[:24]
	}

	if err := authoringstate.ValidateSessionID(request.SessionID); err != nil {
		return normalizedAuthoringRequest{Request: request}, err
	}

	if request.ProjectID == "" || !opsnamespace.Valid(request.Namespace) {
		return normalizedAuthoringRequest{Request: request}, errors.New("authoring requires project and namespace")
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return normalizedAuthoringRequest{Request: request}, err
	}

	if len(encoded) > 512*1024 {
		return normalizedAuthoringRequest{Request: request}, errors.New("authoring request exceeds 512 KiB")
	}

	return normalizedAuthoringRequest{Request: request, Digest: contenthash.Sum(encoded)}, nil
}

func validateAuthoringRequest(request ContinuationRequest) error {
	if len(request.Topic) > 512 || len(request.Capture) > 65536 || len(request.Instructions) > 16384 || len(request.OriginalInstruction) > 16384 {
		return errors.New("authoring input exceeds its bound")
	}

	if !validAuthoringContinuationKind(request.Target) {
		return errors.New("unsupported authoring target")
	}

	if request.Action == "create" {
		return validateAuthoringCreation(request)
	}

	if !validAuthoringObservation(request) {
		return workflowstate.ErrConflict
	}

	return validateAuthoringActionArguments(request)
}

func validateAuthoringActionArguments(request ContinuationRequest) error {
	switch request.Action {
	case "ask", "generate", "refine":
		if request.Question != nil {
			return workflowstate.ErrConflict
		}
	case "answer", "respond":
		if request.Question == nil || strings.TrimSpace(request.Instructions) == "" {
			return workflowstate.ErrConflict
		}
	default:
		return errors.New("unsupported authoring action")
	}

	return nil
}

func (r *fileSessionRepository) reserveAuthoringRequest(request ContinuationRequest, digest string) (authoringRequestBinding, error) {
	binding := authoringRequestBinding{Schema: "chr-authoring-request/v1", Digest: digest, Request: request}

	data, err := r.authoringRead(authoringRequestRef(request.Binding.RequestID))
	if err == nil {
		return readExistingAuthoringBinding(data, digest)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return binding, err
	}

	if err := r.authoringRequestCapacity(); err != nil {
		return binding, err
	}

	binding.CreatedAt = time.Now().UTC()

	return binding, r.saveAuthoringRequest(binding)
}

func (r *fileSessionRepository) saveAuthoringRequest(binding authoringRequestBinding) error {
	path, err := r.authoringPath(authoringRequestRef(binding.Request.Binding.RequestID))
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(binding)
	if err != nil {
		return err
	}

	return opsfile.WriteAtomic(path, encoded)
}

func (r *fileSessionRepository) authoringRequestCapacity() error {
	return r.authoringStateCapacity("requests")
}

func (r *fileSessionRepository) authoringStateCapacity(directory string) error {
	path, err := r.authoringPath(authoringRuntimeRef(directory))
	if err != nil {
		return err
	}

	dir, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	defer func() {
		//nolint:errcheck // Read-only directory close cannot change stored state.
		_ = dir.Close()
	}()

	entries, err := dir.ReadDir(workflowstate.RecordLimit + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if len(entries) >= workflowstate.RecordLimit {
		return errors.New("authoring " + strings.TrimSuffix(directory, "s") + " capacity exhausted")
	}

	return nil
}

type authoringReceiptLookup struct {
	Result ContinuationResult
	Found  bool
}

func (r *fileSessionRepository) readAuthoringReceipt(request ContinuationRequest, digest string) (authoringReceiptLookup, error) {
	var receipt authoringMutationReceipt

	data, err := r.authoringRead(authoringReceiptRef(request.Binding.RequestID))
	if errors.Is(err, os.ErrNotExist) {
		return authoringReceiptLookup{Result: receipt.Result}, nil
	}

	if err != nil {
		return authoringReceiptLookup{Result: receipt.Result}, err
	}

	if err := authoringstate.Decode(data, &receipt); err != nil {
		return authoringReceiptLookup{Result: receipt.Result}, err
	}

	if err := validateAuthoringReceipt(receipt, request, digest); err != nil {
		return authoringReceiptLookup{Result: receipt.Result}, err
	}

	return authoringReceiptLookup{Result: receipt.Result, Found: true}, nil
}

func readExistingAuthoringBinding(data []byte, digest string) (authoringRequestBinding, error) {
	var binding authoringRequestBinding
	if err := authoringstate.Decode(data, &binding); err != nil {
		return binding, err
	}

	payload, err := json.Marshal(binding.Request)
	if err != nil {
		return binding, err
	}

	if binding.Attempts < 0 || binding.Attempts > 5 || binding.Schema != "chr-authoring-request/v1" || binding.Digest != digest || contenthash.Sum(payload) != digest {
		return binding, workflowstate.ErrConflict
	}

	return binding, nil
}

func validateAuthoringReceipt(receipt authoringMutationReceipt, request ContinuationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-authoring-receipt/v1" || receipt.Digest != digest || receipt.Result.ArgumentDigest != digest || checksum != contenthash.Sum(encoded) {
		return workflowstate.ErrConflict
	}

	if receipt.Result.RequestID != request.Binding.RequestID || receipt.Result.Session.ID != request.SessionID || receipt.Result.ReceiptRef != filepath.ToSlash(authoringReceiptRef(request.Binding.RequestID)) {
		return workflowstate.ErrConflict
	}

	return nil
}

func validAuthoringContinuationKind(kind ArtifactKind) bool {
	switch kind {
	case ArtifactKindDump, ArtifactKindQuiz, ArtifactKindSpec, ArtifactKindPlan:
		return true
	default:
		return false
	}
}

func validateAuthoringCreation(request ContinuationRequest) error {
	if request.Topic == "" || request.Capture == "" || request.ExpectedRevision != 0 || request.ExpectedSourceHash != "" || request.Question != nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func validAuthoringObservation(request ContinuationRequest) bool {
	return request.ExpectedRevision > 0 && len(request.ExpectedSourceHash) == 64 && request.Topic == "" && request.Capture == "" && request.OriginalInstruction == ""
}
