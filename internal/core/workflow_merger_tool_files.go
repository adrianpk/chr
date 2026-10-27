package core

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"unicode/utf8"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowMergerFileHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (tools *workflowMergerTools) read(path string) (workflowMergerToolResponse, error) {
	file, err := tools.root.Open(path)
	if err != nil {
		return workflowMergerToolResponse{}, err
	}
	defer func() { _ = file.Close() }() //nolint:errcheck

	data, err := io.ReadAll(io.LimitReader(file, WorkflowFileLimit+1))
	if err != nil {
		return workflowMergerToolResponse{}, err
	}

	if len(data) > WorkflowFileLimit || !utf8.Valid(data) {
		return workflowMergerToolResponse{}, workflowstate.ErrConflict
	}

	return workflowMergerToolResponse{Content: string(data)}, nil
}

func (tools *workflowMergerTools) write(path, content string) (err error) {
	if len(content) > WorkflowFileLimit || !utf8.ValidString(content) {
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
