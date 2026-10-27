package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const (
	harnessSelectionPending         = "tmp/harness-selection-change.json"
	harnessSelectionReceiptLimit    = 16_384
	harnessSelectionReceiptCapacity = 4096
)

// The intent retains only bounded choices and digests, never a copy of unrelated
// configuration. Recovery reconstructs the image from the exact prior hash.
type harnessSelectionJournal struct {
	Schema   string                        `json:"schema"`
	Request  HarnessSelectionChangeRequest `json:"request"`
	Result   HarnessSelectionSnapshot      `json:"result"`
	Before   string                        `json:"before"`
	Checksum string                        `json:"checksum"`
}

func readHarnessSelectionConfig(project string) ([]byte, error) {
	path, err := workflowUpgradePath(project, "config.yml")
	if err != nil {
		return nil, err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	return data, err
}

func requireNoHarnessSelectionChange(project string) error {
	_, err := readHarnessSelectionFile(project, harnessSelectionPending)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	return errors.New("harness selection change pending: retry the original change before upgrading")
}

func readHarnessSelectionFile(project, relative string) ([]byte, error) {
	path, err := workflowUpgradePath(project, relative)
	if err != nil {
		return nil, err
	}

	return opsfile.ReadBounded(path, harnessSelectionReceiptLimit)
}

func harnessSelectionReceiptPath(id string) string {
	return filepath.Join("harness-selection-receipts", contenthash.Sum([]byte(id))+".json")
}

func writeHarnessSelectionJournal(project string, journal harnessSelectionJournal) error {
	checksum, err := harnessSelectionJournalHash(journal)
	if err != nil {
		return err
	}

	journal.Checksum = checksum

	return writeHarnessSelectionFile(project, harnessSelectionPending, journal)
}

func writeHarnessSelectionFile(project, relative string, journal harnessSelectionJournal) error {
	path, err := workflowUpgradePath(project, relative)
	if err != nil {
		return err
	}

	data, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	return opsfile.WriteAtomicBounded(path, data, harnessSelectionReceiptLimit)
}

func harnessSelectionJournalHash(journal harnessSelectionJournal) (string, error) {
	journal.Checksum = ""
	data, err := json.Marshal(journal)

	return contenthash.Sum(data), err
}

func decodeHarnessSelectionJournal(data []byte) (harnessSelectionJournal, error) {
	var journal harnessSelectionJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		return journal, err
	}

	checksum, err := harnessSelectionJournalHash(journal)
	if err != nil || checksum != journal.Checksum || journal.Schema != "chr-harness-selection-change/v1" {
		return journal, errors.New("invalid harness selection journal")
	}

	if err := validateHarnessSelectionJournalBinding(journal); err != nil {
		return journal, err
	}

	return journal, validateHarnessSelectionJournalResult(journal)
}

func validateHarnessSelectionJournalBinding(journal harnessSelectionJournal) error {
	digest, err := harnessSelectionRequestDigest(journal.Request)

	r := journal.Result
	if err != nil || digest != r.RequestDigest || journal.Before != journal.Request.ExpectedRevision || !harnesscore.SelectionDigest(r.Revision) {
		return errors.New("invalid harness selection request evidence")
	}

	if r.ProjectRef != journal.Request.ProjectRef || r.RequestID != journal.Request.Binding.RequestID || r.ChangedBy != journal.Request.Binding.Operator {
		return errors.New("invalid harness selection result binding")
	}

	return nil
}

func validateHarnessSelectionJournalResult(journal harnessSelectionJournal) error {
	r := journal.Result
	for _, field := range []struct {
		chosen *string
		result string
	}{{journal.Request.Worker, r.Worker}, {journal.Request.Generation, r.Generation}} {
		if field.chosen != nil && *field.chosen != field.result {
			return errors.New("harness selection result changed requested choices")
		}

		if field.result != "" && !harnesscore.ValidHarnessProfileName(field.result) {
			return errors.New("invalid retained harness choice")
		}
	}

	return validateHarnessSelectionAudit(r.HarnessSelectionAudit)
}

func readHarnessSelectionReceipt(project string, request HarnessSelectionChangeRequest, digest string) (HarnessSelectionSnapshot, bool, error) {
	data, err := readHarnessSelectionFile(project, harnessSelectionReceiptPath(request.Binding.RequestID))
	if errors.Is(err, os.ErrNotExist) {
		return HarnessSelectionSnapshot{}, false, nil
	}

	if err != nil {
		return HarnessSelectionSnapshot{}, false, err
	}

	journal, err := decodeHarnessSelectionJournal(data)
	if err != nil {
		return HarnessSelectionSnapshot{}, false, err
	}

	if journal.Result.RequestDigest != digest {
		return HarnessSelectionSnapshot{}, false, workflowstate.ErrConflict
	}

	return journal.Result, true, nil
}

func recoverHarnessSelectionChange(project string, request HarnessSelectionChangeRequest, digest string) error {
	data, err := readHarnessSelectionFile(project, harnessSelectionPending)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	journal, err := decodeHarnessSelectionJournal(data)
	if err != nil {
		return err
	}

	if journal.Request.ProjectRef != request.ProjectRef {
		return workflowstate.ErrConflict
	}

	if journal.Request.Binding.RequestID == request.Binding.RequestID && journal.Result.RequestDigest != digest {
		return workflowstate.ErrConflict
	}

	return finishHarnessSelectionChange(project, journal)
}

func finishHarnessSelectionChange(project string, journal harnessSelectionJournal) error {
	current, err := readHarnessSelectionConfig(project)
	if err != nil {
		return err
	}

	observed := contenthash.Sum(current)
	if observed != journal.Before && observed != journal.Result.Revision {
		return workflowstate.ErrConflict
	}

	if observed == journal.Before {
		if err := publishRetainedHarnessSelectionImage(project, current, journal); err != nil {
			return err
		}
	}

	return finishHarnessSelectionReceipt(project, journal)
}

func publishRetainedHarnessSelectionImage(project string, current []byte, journal harnessSelectionJournal) error {
	image := harnessSelectionChangeImage(current, journal.Request, journal.Result.RequestDigest, journal.Result.ChangedAt)
	if contenthash.Sum(image) != journal.Result.Revision {
		return workflowstate.ErrConflict
	}

	path, err := workflowUpgradePath(project, "config.yml")
	if err != nil {
		return err
	}

	return opsfile.WriteAtomic(path, image)
}

func finishHarnessSelectionReceipt(project string, journal harnessSelectionJournal) error {
	if err := retainHarnessSelectionReceipt(project, journal); err != nil {
		return err
	}

	path, err := workflowUpgradePath(project, harnessSelectionPending)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return opsfile.SyncDirectory(filepath.Dir(path))
}

func retainHarnessSelectionReceipt(project string, journal harnessSelectionJournal) error {
	retained, found, err := readHarnessSelectionReceipt(project, journal.Request, journal.Result.RequestDigest)
	if err != nil {
		return err
	}

	if found {
		if retained != journal.Result {
			return workflowstate.ErrConflict
		}

		return nil
	}

	checksum, err := harnessSelectionJournalHash(journal)
	if err != nil {
		return err
	}

	journal.Checksum = checksum

	return writeHarnessSelectionFile(project, harnessSelectionReceiptPath(journal.Request.Binding.RequestID), journal)
}

func checkHarnessSelectionReceiptCapacity(project string) error {
	path, err := workflowUpgradePath(project, filepath.Join("harness-selection-receipts", ".capacity"))
	if err != nil {
		return err
	}

	dir, err := os.Open(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	entries, readErr := dir.ReadDir(harnessSelectionReceiptCapacity + 1)

	closeErr := dir.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return errors.Join(readErr, closeErr)
	}

	if len(entries) >= harnessSelectionReceiptCapacity {
		return fmt.Errorf("harness selection receipt capacity exhausted (%d)", harnessSelectionReceiptCapacity)
	}

	return nil
}
