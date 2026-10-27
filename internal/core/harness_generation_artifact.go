package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness"
)

func harnessGenerationOperationPath(project, id string) (string, error) {
	return workflowUpgradePath(project, filepath.Join("tmp", "agent-runs", "operations", contenthash.Sum([]byte(id))+".json"))
}

func readHarnessGenerationOperation(project, id, purpose string) (HarnessGenerationOperation, error) {
	var record harnessGenerationOperationFile

	path, err := harnessGenerationOperationPath(project, id)
	if err != nil {
		return record.Operation, err
	}

	data, err := readHarnessGenerationFile(path)
	if err != nil {
		return record.Operation, err
	}

	if err := decodeHarnessGenerationFile(data, &record); err != nil {
		return record.Operation, err
	}

	checksum := record.Checksum
	record.Checksum = ""

	encoded, err := json.Marshal(record)
	if err != nil || contenthash.Sum(encoded) != checksum || record.Schema != "chr-harness-operation/v1" || record.Operation.ID != id || record.Operation.Purpose != purpose {
		return record.Operation, errors.New("invalid retained generation operation")
	}

	record.Operation.program, record.Operation.project = record.Program, project

	record.Operation.configDir, record.Operation.executionDir = record.ConfigDir, record.ExecutionDir
	if record.Resources != nil {
		record.Operation.resources = *record.Resources
	}

	return record.Operation, validateGenerationOperation(record.Operation)
}

func writeHarnessGenerationOperation(operation HarnessGenerationOperation) error {
	path, err := harnessGenerationOperationPath(operation.project, operation.ID)
	if err != nil {
		return err
	}

	if err := harnessGenerationOperationCapacity(filepath.Dir(path)); err != nil {
		return err
	}

	record := harnessGenerationOperationFile{Schema: "chr-harness-operation/v1", Operation: operation, Program: operation.program, ConfigDir: operation.configDir, ExecutionDir: operation.executionDir}
	if operation.resources != (harness.GenerationResources{}) {
		record.Resources = &operation.resources
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}

	record.Checksum = contenthash.Sum(encoded)

	return writeHarnessGenerationFile(path, record)
}

func harnessGenerationOperationCapacity(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }() //nolint:errcheck // Read-only directory close.

	entries, err := file.ReadDir(workflowstate.WorkflowRecordLimit + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if len(entries) >= workflowstate.WorkflowRecordLimit {
		return errors.New("generation operation capacity exhausted")
	}

	return nil
}

func readHarnessGenerationFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // The bounded read checks file IO.

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > harnessGenerationRecordLimit {
		return nil, errors.New("generation artifact must be a private regular file within 16 KiB")
	}

	data, err := io.ReadAll(io.LimitReader(file, harnessGenerationRecordLimit+1))
	if err != nil || len(data) > harnessGenerationRecordLimit {
		return nil, errors.New("generation artifact exceeds read boundary")
	}

	return data, nil
}

func decodeHarnessGenerationFile(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("trailing generation artifact data")
	}

	return nil
}

func writeHarnessGenerationFile(path string, record any) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	return atomicBoundedWorkflowFile(path, data, harnessGenerationRecordLimit)
}
