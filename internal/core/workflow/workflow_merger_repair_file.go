package workflow

import (
	"context"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) mergerFileKind(ctx context.Context, tree, path string) (string, error) {
	data, err := reader.RawOutput(ctx, "ls-tree", "-z", "--full-tree", tree, "--", path)
	if err != nil || data == "" {
		return "text", err
	}

	header, name, ok := strings.Cut(strings.TrimSuffix(data, "\x00"), "\t")

	fields := strings.Fields(header)
	if !ok || name != path || len(fields) != 3 || !workflowCommitID(fields[2]) {
		return "", workflowstate.ErrConflict
	}

	if fields[0] == "160000" {
		return "submodule", nil
	}

	if fields[0] == "120000" {
		return "binary", nil
	}

	content, err := reader.RawOutput(ctx, "cat-file", "blob", fields[2])
	if err != nil {
		return "", err
	}

	return workflowMergerBlobKind(content), nil
}

func workflowMergerBlobKind(content string) string {
	if strings.ContainsRune(content, '\x00') {
		return "binary"
	}

	if strings.HasPrefix(content, "version https://git-lfs.github.com/spec/v1\n") {
		return "lfs"
	}

	return "text"
}
