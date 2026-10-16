package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type caseMutationJournal struct {
	Schema   string               `json:"schema"`
	Changes  []workflowFileChange `json:"changes"`
	Checksum string               `json:"checksum"`
}

func (s CaseMetadataService) commitCaseMutation(request CaseMutationRequest, digest string) (CaseMutationResult, error) {
	metadata, err := s.prepareCaseMutation(request.Metadata)
	if err != nil {
		return CaseMutationResult{}, err
	}

	content := []byte(RenderCaseMetadata(metadata))
	metadata.SourceHash, metadata.Exists = contentHash(content), true
	result := CaseMutationResult{Metadata: metadata, ReceiptRef: caseMutationReceiptRef(request.Binding.RequestID), ArgumentDigest: digest}
	receipt := caseMutationReceipt{Schema: "chr-case-mutation/v1", Request: request, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return result, err
	}

	receipt.Checksum = workflowHash(encoded)

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return result, err
	}

	journal := caseMutationJournal{Schema: "chr-case-journal/v1", Changes: []workflowFileChange{{Path: metadata.SourcePath, Before: request.Metadata.SourceHash, After: content}, {Path: result.ReceiptRef, After: encoded}}}
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

	journal.Checksum = workflowHash(encoded)

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

	return atomicWorkflowFile(path, encoded)
}

func (s CaseMetadataService) recoverCaseMutation() error {
	path, err := s.caseMutationPath("case-mutations/pending.json")
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
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

	return syncWorkflowDirectory(filepath.Dir(path))
}

func (s CaseMetadataService) checkCaseJournal(journal caseMutationJournal) error {
	checksum := journal.Checksum
	journal.Checksum = ""

	data, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	if journal.Schema != "chr-case-journal/v1" || len(journal.Changes) != 2 || checksum != workflowHash(data) {
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

func (s CaseMetadataService) caseJournalImage(change workflowFileChange) (bool, error) {
	path, err := s.caseMutationPath(change.Path)
	if err != nil {
		return false, err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) && change.Before == "" {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if workflowHash(data) == workflowHash(change.After) {
		return true, nil
	}

	if workflowHash(data) != change.Before {
		return false, fmt.Errorf("%w: case image diverged", ErrWorkflowConflict)
	}

	return false, nil
}

func (s CaseMetadataService) readCaseMutationReceipt(request CaseMutationRequest, digest string) (CaseMutationResult, bool, error) {
	path, err := s.caseMutationPath(caseMutationReceiptRef(request.Binding.RequestID))
	if err != nil {
		return CaseMutationResult{}, false, err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return CaseMutationResult{}, false, nil
	}

	if err != nil {
		return CaseMutationResult{}, false, err
	}

	var receipt caseMutationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return CaseMutationResult{}, false, err
	}

	if err := validateCaseMutationReceipt(receipt, request, digest); err != nil {
		return CaseMutationResult{}, false, err
	}

	return receipt.Result, true, nil
}

func validateCaseMutationReceipt(receipt caseMutationReceipt, request CaseMutationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	data, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-case-mutation/v1" || checksum != workflowHash(data) || receipt.Result.ArgumentDigest != digest {
		return fmt.Errorf("%w: case receipt changed", ErrWorkflowConflict)
	}

	if receipt.Request.ProjectID != request.ProjectID || receipt.Request.Namespace != request.Namespace || receipt.Request.Binding != request.Binding || receipt.Result.ReceiptRef != caseMutationReceiptRef(request.Binding.RequestID) {
		return fmt.Errorf("%w: case receipt scope changed", ErrWorkflowConflict)
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
			return ErrWorkflowConflict
		}
	}

	return nil
}

func (s CaseMetadataService) guardNewCaseMutation(metadata CaseMetadata) error {
	path, err := s.caseMutationPath(metadata.SourcePath)
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if !metadata.Exists && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return errors.Join(ErrWorkflowConflict, err)
	}

	if !metadata.Exists || contentHash(data) != metadata.SourceHash || metadata.Slug != caseSlugFromPath(path) {
		return ErrWorkflowConflict
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

	relative, err := filepath.Rel(UserPath(s.projectDir), path)
	if err != nil {
		return metadata, err
	}

	metadata.SourcePath = filepath.ToSlash(relative)
	metadata.Slug = caseSlugFromPath(path)

	return metadata, nil
}

func validateCaseMutationObservation(metadata CaseMetadata) error {
	if metadata.SourcePath == "" || metadata.SourceHash == "" {
		return errors.New("case update requires its observed source image")
	}

	return nil
}

func (s CaseMetadataService) applyCaseJournalImage(change workflowFileChange, index int) error {
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

	if err := atomicWorkflowFile(target, change.After); err != nil {
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
