package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// DumpDraftEditorSession owns a private temporary draft and its cleanup. Finish
// must be called after the editor exits, including when execution fails.
type DumpDraftEditorSession struct {
	Command DumpEditorCommand
	Finish  func(error) (string, error)
}

// OpenDumpDraftEditor edits unsaved text without touching its canonical source.
func (s service) OpenDumpDraftEditor(content string) (DumpDraftEditorSession, error) {
	if len(content) > workflowstate.FileLimit {
		return DumpDraftEditorSession{}, fmt.Errorf("dump draft exceeds 1 MiB")
	}

	directory, err := os.MkdirTemp("", "chr-dump-editor-")
	if err != nil {
		return DumpDraftEditorSession{}, err
	}

	path := filepath.Join(directory, "draft.md")
	if err = os.WriteFile(path, []byte(content), 0o600); err != nil {
		return DumpDraftEditorSession{}, errors.Join(err, os.RemoveAll(directory))
	}

	command, err := s.dumpEditor.EditorCommand(s.ctx, path)
	if err != nil {
		return DumpDraftEditorSession{}, errors.Join(err, os.RemoveAll(directory))
	}

	return DumpDraftEditorSession{Command: command, Finish: func(runErr error) (string, error) {
		var (
			result  string
			readErr error
		)
		if runErr == nil {
			result, readErr = readDumpEditorFile(path)
		}

		return result, errors.Join(runErr, readErr, os.RemoveAll(directory))
	}}, nil
}

func readDumpEditorFile(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("editor draft must remain a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	content, readErr := io.ReadAll(io.LimitReader(file, workflowstate.FileLimit+1))

	closeErr := file.Close()
	if len(content) > workflowstate.FileLimit {
		return "", errors.Join(fmt.Errorf("editor draft exceeds 1 MiB"), readErr, closeErr)
	}

	return string(content), errors.Join(readErr, closeErr)
}
