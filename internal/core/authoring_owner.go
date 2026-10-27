package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

type authoringOwnerContextKey struct{}

type authoringOwnerClaim struct {
	root   string
	active *atomic.Bool
}

// WithAuthoringMutation excludes every shared authoring writer and recovers its
// pending journal before a new change. The context permits nested owner calls;
// no lock identity is serialized or supplied by model output.
func WithAuthoringMutation(ctx context.Context, projectDir string, action func(context.Context) error) (err error) {
	if projectDir == "" {
		return errors.New("authoring project directory is required")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	root, err := filepath.Abs(projectDir)
	if err != nil {
		return err
	}

	if holdsAuthoringOwner(ctx, root) {
		return action(ctx)
	}

	r := &fileAuthoringSessionRepository{projectDir: root}

	lockPath, err := r.authoringPath(filepath.Join(workspacecore.DirectoryName, "authoring", "owner.lock"))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	if err := acquireWorkflowLock(ctx, file); err != nil {
		return err
	}

	if err := r.recoverAuthoringMutation(); err != nil {
		return err
	}

	claim := authoringOwnerClaim{root: root, active: &atomic.Bool{}}

	claim.active.Store(true)
	defer claim.active.Store(false)

	return action(context.WithValue(ctx, authoringOwnerContextKey{}, claim))
}

func (r *fileAuthoringSessionRepository) authoringPath(relative string) (string, error) {
	if !filepath.IsLocal(relative) || filepath.Clean(relative) != relative {
		return "", errors.New("invalid authoring relative path")
	}

	root, err := filepath.Abs(r.projectDir)
	if err != nil {
		return "", err
	}

	path := filepath.Join(root, relative)
	for parent := path; parent != root; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("authoring owner cannot traverse a symlink")
		}
	}

	return path, nil
}

func authoringRuntimeRef(suffix string) string {
	return filepath.Join(workspacecore.DirectoryName, "authoring", suffix)
}

func authoringRequestRef(id string) string {
	return authoringRuntimeRef("requests/" + contenthash.Sum([]byte(id)) + ".json")
}

func authoringReceiptRef(id string) string {
	return authoringRuntimeRef("receipts/" + contenthash.Sum([]byte(id)) + ".json")
}

func (r *fileAuthoringSessionRepository) authoringSessionPaths(id string) (AuthoringSessionPaths, error) {
	if err := authoringstate.ValidateSessionID(id); err != nil {
		return AuthoringSessionPaths{}, err
	}

	paths := AuthoringSessionLayout(r.projectDir, id)
	for _, path := range []string{paths.SessionPath, paths.TranscriptPath, paths.ArtifactsPath} {
		relative, err := filepath.Rel(r.projectDir, path)
		if err != nil {
			return paths, err
		}

		if _, err := r.authoringPath(relative); err != nil {
			return paths, err
		}
	}

	return paths, nil
}

func (r *fileAuthoringSessionRepository) authoringRead(relative string) ([]byte, error) {
	path, err := r.authoringPath(relative)
	if err != nil {
		return nil, err
	}

	return readWorkflowFile(path)
}

func authoringRelativePaths(projectDir string, paths AuthoringSessionPaths) (AuthoringSessionPaths, error) {
	result := paths

	pairs := []struct {
		target *string
		value  string
	}{{&result.SessionDir, paths.SessionDir}, {&result.SessionPath, paths.SessionPath}, {&result.TranscriptPath, paths.TranscriptPath}, {&result.ArtifactsPath, paths.ArtifactsPath}}
	for _, pair := range pairs {
		relative, err := filepath.Rel(projectDir, pair.value)
		if err != nil || !filepath.IsLocal(relative) {
			return result, fmt.Errorf("invalid authoring session path: %w", err)
		}

		*pair.target = filepath.ToSlash(relative)
	}

	return result, nil
}

func validAuthoringArtifactRef(namespace string, ref AuthoringArtifactRef) bool {
	parts := strings.Split(filepath.ToSlash(ref.Path), "/")

	return len(parts) == 5 && parts[0] == OpsDirectoryName && parts[1] == namespace && parts[4] == OpsArtifactBodyFile() && parts[2] == authoringArtifactType(ref.Kind)
}

func holdsAuthoringOwner(ctx context.Context, root string) bool {
	claim, ok := ctx.Value(authoringOwnerContextKey{}).(authoringOwnerClaim)

	return ok && claim.root == root && claim.active != nil && claim.active.Load()
}
