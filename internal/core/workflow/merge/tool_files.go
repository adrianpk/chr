package merge

import (
	"io"
	"os"
	"unicode/utf8"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (tools *WorkflowMergerTools) read(path string) (workflowMergerToolResponse, error) {
	file, err := tools.root.Open(path)
	if err != nil {
		return workflowMergerToolResponse{}, err
	}
	defer func() { _ = file.Close() }() //nolint:errcheck

	data, err := io.ReadAll(io.LimitReader(file, workflowstate.FileLimit+1))
	if err != nil {
		return workflowMergerToolResponse{}, err
	}

	if len(data) > workflowstate.FileLimit || !utf8.Valid(data) {
		return workflowMergerToolResponse{}, workflowstate.ErrConflict
	}

	return workflowMergerToolResponse{Content: string(data)}, nil
}

func (tools *WorkflowMergerTools) write(path, content string) (err error) {
	if len(content) > workflowstate.FileLimit || !utf8.ValidString(content) {
		return workflowstate.ErrConflict
	}

	file, err := tools.root.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	_, err = io.WriteString(file, content)
	if err == nil {
		err = file.Sync()
	}

	return err
}
