package runtime

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	"forge.adrianpk.com/chr/internal/harness"
)

// WorkerRuntimeMounts grants the execution checkout and its Git metadata at
// their actual paths. Linked worktrees need the common Git directory as well
// as their checkout; no other project files or home directories are imported.
func WorkerRuntimeMounts(ctx context.Context, directory string) ([]harness.RuntimeMount, error) {
	root, err := workerRuntimeRoot(directory)
	if err != nil {
		return nil, err
	}

	common, err := projectcore.ResolveRepositoryDirectory(ctx, root)
	if err != nil {
		return nil, errors.New("worker execution checkout requires available Git metadata")
	}

	if workerRuntimeHome(common) {
		return nil, errors.New("worker Git metadata cannot mount the developer home")
	}

	mounts := []harness.RuntimeMount{{Source: root, Target: root, Description: "execution checkout"}}

	rel, err := filepath.Rel(root, common)
	if err != nil {
		return nil, err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		mounts = append(mounts, harness.RuntimeMount{Source: common, Target: common, Description: "execution Git metadata"})
	}

	return mounts, nil
}

// WorkerRuntimeInstructions supplies the checkout's explicit project contract.
// It does not discover model resources or follow a linked instruction file.
func WorkerRuntimeInstructions(directory string) (data []byte, err error) {
	file, err := os.OpenFile(filepath.Join(directory, "AGENTS.md"), os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.New("worker project instructions are unavailable")
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	info, err := file.Stat()
	if err != nil || !validWorkerInstructionFile(info) {
		return nil, errors.New("worker project instructions exceed their regular-file boundary")
	}

	data, err = io.ReadAll(io.LimitReader(file, harness.MaxInput+1))
	if err != nil || len(data) > harness.MaxInput || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
		return nil, errors.New("worker project instructions exceed the context boundary")
	}

	return data, nil
}

func validWorkerInstructionFile(info os.FileInfo) bool {
	if !info.Mode().IsRegular() || info.Size() > harness.MaxInput {
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)

	return ok && stat.Nlink == 1
}

func workerRuntimeRoot(directory string) (string, error) {
	root, err := filepath.EvalSymlinks(directory)
	if err != nil || !filepath.IsAbs(root) || root == string(filepath.Separator) {
		return "", errors.New("worker requires an existing absolute execution checkout")
	}

	if workerRuntimeHome(root) {
		return "", errors.New("worker checkout cannot mount the developer home")
	}

	return root, nil
}

func workerRuntimeHome(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	if path == filepath.Clean(home) {
		return true
	}

	real, err := filepath.EvalSymlinks(home)

	return err == nil && path == real
}
