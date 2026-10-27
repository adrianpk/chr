package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// CaseMutationRequest binds a metadata write to its caller and observed image.
type CaseMutationRequest struct {
	Binding   workflowstate.MutationBinding `json:"binding"`
	ProjectID string                        `json:"project_id"`
	Namespace string                        `json:"namespace"`
	Metadata  CaseMetadata                  `json:"metadata"`
}

// CaseMutationResult retains the original metadata and a project-relative receipt.
type CaseMutationResult struct {
	Metadata       CaseMetadata `json:"metadata"`
	ReceiptRef     string       `json:"receipt_ref"`
	ArgumentDigest string       `json:"argument_digest"`
}

type caseMutationReceipt struct {
	Schema   string              `json:"schema"`
	Request  CaseMutationRequest `json:"request"`
	Result   CaseMutationResult  `json:"result"`
	Checksum string              `json:"checksum"`
}

// SaveBound reuses the case owner and its exclusion for an immutable request.
func (s CaseMetadataService) SaveBound(ctx context.Context, request CaseMutationRequest) (result CaseMutationResult, err error) {
	normalized, err := s.normalizeCaseMutation(request)
	if err != nil {
		return result, err
	}

	request, digest := normalized.Request, normalized.Digest

	err = s.withMutationLock(ctx, func() error {
		lookup, readErr := s.readCaseMutationReceipt(request, digest)
		if readErr != nil || lookup.Found {
			result = lookup.Result

			return readErr
		}

		if err := s.reserveCaseMutation(request, digest); err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		var writeErr error

		result, writeErr = s.commitCaseMutation(request, digest)

		return writeErr
	})
	if err == nil {
		result.Metadata.SourcePath = filepath.Join(workspacecore.UserPath(s.projectDir), filepath.FromSlash(result.Metadata.SourcePath))
	}

	return result, err
}

type normalizedCaseMutation struct {
	Request CaseMutationRequest
	Digest  string
}

func (s CaseMetadataService) normalizeCaseMutation(request CaseMutationRequest) (normalizedCaseMutation, error) {
	request.Binding.RequestID = strings.TrimSpace(request.Binding.RequestID)

	request.Binding.Operator = strings.TrimSpace(request.Binding.Operator)
	if err := request.Binding.Validate(); err != nil {
		return normalizedCaseMutation{Request: request}, err
	}

	if request.ProjectID == "" || !opsnamespace.Valid(request.Namespace) {
		return normalizedCaseMutation{Request: request}, errors.New("case mutation requires its target scope")
	}

	if request.Metadata.SourcePath != "" {
		relative, err := filepath.Rel(workspacecore.UserPath(s.projectDir), request.Metadata.SourcePath)
		if err != nil {
			return normalizedCaseMutation{Request: request}, err
		}

		if _, err := s.caseMutationPath(relative); err != nil {
			return normalizedCaseMutation{Request: request}, err
		}

		request.Metadata.SourcePath = filepath.ToSlash(relative)
	}

	request.Metadata.SchemaVersion = CaseMetadataSchemaVersion

	data, err := json.Marshal(request)
	if err != nil {
		return normalizedCaseMutation{Request: request}, err
	}

	if len(data) > 64*1024 {
		return normalizedCaseMutation{Request: request}, errors.New("case mutation arguments exceed 64 KiB")
	}

	return normalizedCaseMutation{Request: request, Digest: contenthash.Sum(data)}, nil
}

func (s CaseMetadataService) withMutationLock(ctx context.Context, action func() error) (err error) {
	path, err := s.caseMutationPath("case-mutations/owner.lock")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	if err := opsfile.Lock(ctx, file); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)) }()

	if err := s.recoverCaseMutation(); err != nil {
		return err
	}

	return action()
}

func (s CaseMetadataService) caseMutationPath(relative string) (string, error) {
	if !validCaseMutationRelative(relative) {
		return "", errors.New("invalid case owner path")
	}

	root, err := filepath.Abs(s.projectDir)
	if err != nil {
		return "", err
	}

	path := filepath.Join(workspacecore.UserPath(root), relative)
	for parent := path; parent != root; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("case owner cannot traverse a symlink")
		}
	}

	return path, nil
}

func caseMutationReceiptRef(id string) string {
	return "case-mutations/receipts/" + contenthash.Sum([]byte(id)) + ".json"
}

func (s CaseMetadataService) reserveCaseMutation(request CaseMutationRequest, digest string) error {
	path, err := s.caseMutationPath("case-mutations/bindings/" + contenthash.Sum([]byte(request.Binding.RequestID)) + ".json")
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(struct {
		Schema, ProjectID, Namespace, Digest string
		Binding                              workflowstate.MutationBinding
	}{"chr-case-binding/v1", request.ProjectID, request.Namespace, digest, request.Binding})
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
	if err == nil {
		if string(data) != string(encoded) {
			return fmt.Errorf("%w: case mutation payload changed", workflowstate.ErrConflict)
		}

		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := s.caseMutationCapacity(); err != nil {
		return err
	}

	return opsfile.WriteAtomic(path, encoded)
}

func (s CaseMetadataService) caseMutationCapacity() error {
	path, err := s.caseMutationPath("case-mutations/bindings")
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

	entries, readErr := dir.ReadDir(4097)

	closeErr := dir.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return errors.Join(readErr, closeErr)
	}

	if len(entries) >= 4096 {
		return errors.New("case mutation binding capacity exhausted")
	}

	return nil
}

func validCaseMutationRelative(relative string) bool {
	return filepath.IsLocal(relative) && filepath.Clean(relative) == relative && (strings.HasPrefix(filepath.ToSlash(relative), "cases/") || strings.HasPrefix(filepath.ToSlash(relative), "case-mutations/"))
}
