package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

type caseMutationJournal struct {
	Schema   string                     `json:"schema"`
	Changes  []workflowstate.FileChange `json:"changes"`
	Checksum string                     `json:"checksum"`
}

func (s CaseMetadataService) commitCaseMutation(request CaseMutationRequest, digest string) (CaseMutationResult, error) {
	metadata, err := s.prepareCaseMutation(request.Metadata)
	if err != nil {
		return CaseMutationResult{}, err
	}

	content := []byte(RenderCaseMetadata(metadata))
	metadata.SourceHash, metadata.Exists = contenthash.Sum(content), true
	result := CaseMutationResult{Metadata: metadata, ReceiptRef: caseMutationReceiptRef(request.Binding.RequestID), ArgumentDigest: digest}
	receipt := caseMutationReceipt{Schema: "chr-case-mutation/v1", Request: request, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return result, err
	}

	receipt.Checksum = contenthash.Sum(encoded)

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return result, err
	}

	journal := caseMutationJournal{Schema: "chr-case-journal/v1", Changes: []workflowstate.FileChange{{Path: metadata.SourcePath, Before: request.Metadata.SourceHash, After: content}, {Path: result.ReceiptRef, After: encoded}}}
	if err := s.writeCaseMutationJournal(journal); err != nil {
		return result, err
	}

	return result, s.recoverCaseMutation()
}

func (s CaseMetadataService) prepareCaseMutation(metadata CaseMetadata) (CaseMetadata, error) {
	if err := ValidateCaseMetadata(metadata); err != nil {
		return metadata, err
	}

	metadata, err := s.prepareCaseMutationPath(metadata)
	if err != nil {
		return metadata, err
	}

	if !strings.HasPrefix(metadata.SourcePath, "cases/") || filepath.Ext(metadata.SourcePath) != ".md" || strings.Contains(strings.TrimPrefix(metadata.SourcePath, "cases/"), "/") {
		return metadata, errors.New("invalid canonical case path")
	}

	if err := s.guardNewCaseMutation(metadata); err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (s CaseMetadataService) writeCaseMutationJournal(journal caseMutationJournal) error {
	encoded, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	journal.Checksum = contenthash.Sum(encoded)

	encoded, err = json.Marshal(journal)
	if err != nil {
		return err
	}

	if err := s.checkCaseJournal(journal); err != nil {
		return err
	}

	path, err := s.caseMutationPath("case-mutations/pending.json")
	if err != nil {
		return err
	}

	return opsfile.WriteAtomic(path, encoded)
}

func (s CaseMetadataService) recoverCaseMutation() error {
	path, err := s.caseMutationPath("case-mutations/pending.json")
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	var journal caseMutationJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		return err
	}

	if err := s.checkCaseJournal(journal); err != nil {
		return err
	}

	for index, change := range journal.Changes {
		if err := s.applyCaseJournalImage(change, index); err != nil {
			return err
		}
	}

	if err := s.confirmCaseMutation(journal); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return opsfile.SyncDirectory(filepath.Dir(path))
}

func (s CaseMetadataService) checkCaseJournal(journal caseMutationJournal) error {
	checksum := journal.Checksum
	journal.Checksum = ""

	data, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	if journal.Schema != "chr-case-journal/v1" || len(journal.Changes) != 2 || checksum != contenthash.Sum(data) {
		return errors.New("invalid case mutation journal")
	}

	if !strings.HasPrefix(journal.Changes[0].Path, "cases/") || !strings.HasPrefix(journal.Changes[1].Path, "case-mutations/receipts/") {
		return errors.New("invalid case journal targets")
	}

	for _, change := range journal.Changes {
		if _, err := s.caseJournalImage(change); err != nil {
			return err
		}
	}

	return nil
}

func (s CaseMetadataService) caseJournalImage(change workflowstate.FileChange) (bool, error) {
	path, err := s.caseMutationPath(change.Path)
	if err != nil {
		return false, err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) && change.Before == "" {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if contenthash.Sum(data) == contenthash.Sum(change.After) {
		return true, nil
	}

	if contenthash.Sum(data) != change.Before {
		return false, fmt.Errorf("%w: case image diverged", workflowstate.ErrConflict)
	}

	return false, nil
}

type caseMutationReceiptLookup struct {
	Result CaseMutationResult
	Found  bool
}

func (s CaseMetadataService) readCaseMutationReceipt(request CaseMutationRequest, digest string) (caseMutationReceiptLookup, error) {
	path, err := s.caseMutationPath(caseMutationReceiptRef(request.Binding.RequestID))
	if err != nil {
		return caseMutationReceiptLookup{}, err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return caseMutationReceiptLookup{}, nil
	}

	if err != nil {
		return caseMutationReceiptLookup{}, err
	}

	var receipt caseMutationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return caseMutationReceiptLookup{}, err
	}

	if err := validateCaseMutationReceipt(receipt, request, digest); err != nil {
		return caseMutationReceiptLookup{}, err
	}

	return caseMutationReceiptLookup{Result: receipt.Result, Found: true}, nil
}

func validateCaseMutationReceipt(receipt caseMutationReceipt, request CaseMutationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	data, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-case-mutation/v1" || checksum != contenthash.Sum(data) || receipt.Result.ArgumentDigest != digest {
		return fmt.Errorf("%w: case receipt changed", workflowstate.ErrConflict)
	}

	if receipt.Request.ProjectID != request.ProjectID || receipt.Request.Namespace != request.Namespace || receipt.Request.Binding != request.Binding || receipt.Result.ReceiptRef != caseMutationReceiptRef(request.Binding.RequestID) {
		return fmt.Errorf("%w: case receipt scope changed", workflowstate.ErrConflict)
	}

	return nil
}

func (s CaseMetadataService) confirmCaseMutation(journal caseMutationJournal) error {
	for _, change := range journal.Changes {
		applied, err := s.caseJournalImage(change)
		if err != nil {
			return err
		}

		if !applied {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func (s CaseMetadataService) guardNewCaseMutation(metadata CaseMetadata) error {
	path, err := s.caseMutationPath(metadata.SourcePath)
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
	if !metadata.Exists && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return errors.Join(workflowstate.ErrConflict, err)
	}

	if !metadata.Exists || contenthash.Sum(data) != metadata.SourceHash || metadata.Slug != CaseSlugFromPath(path) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s CaseMetadataService) prepareCaseMutationPath(metadata CaseMetadata) (CaseMetadata, error) {
	if metadata.Exists {
		return metadata, validateCaseMutationObservation(metadata)
	}

	if metadata.SourcePath != "" || metadata.SourceHash != "" {
		return metadata, errors.New("case creation cannot replace an existing path")
	}

	path, err := s.boundedCaseMetadataPath(metadata.Name)
	if err != nil {
		return metadata, err
	}

	relative, err := filepath.Rel(workspacecore.UserPath(s.projectDir), path)
	if err != nil {
		return metadata, err
	}

	metadata.SourcePath = filepath.ToSlash(relative)
	metadata.Slug = CaseSlugFromPath(path)

	return metadata, nil
}

func validateCaseMutationObservation(metadata CaseMetadata) error {
	if metadata.SourcePath == "" || metadata.SourceHash == "" {
		return errors.New("case update requires its observed source image")
	}

	return nil
}

func (s CaseMetadataService) applyCaseJournalImage(change workflowstate.FileChange, index int) error {
	applied, err := s.caseJournalImage(change)
	if err != nil {
		return err
	}

	if applied {
		return nil
	}

	target, err := s.caseMutationPath(change.Path)
	if err != nil {
		return err
	}

	if err := opsfile.WriteAtomic(target, change.After); err != nil {
		return err
	}

	if s.afterMutationReplace != nil {
		if err := s.afterMutationReplace(index); err != nil {
			return err
		}
	}

	return nil
}

func (s CaseMetadataService) boundedCaseMetadataPath(name string) (string, error) {
	base := slugifyCaseName(name)
	for suffix := 1; suffix <= 4096; suffix++ {
		slug := base
		if suffix > 1 {
			slug = fmt.Sprintf("%s-%d", base, suffix)
		}

		path, err := s.caseMutationPath("cases/" + slug + ".md")
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return path, nil
		} else if err != nil {
			return "", err
		}
	}

	return "", errors.New("case identity allocation capacity exhausted")
}
