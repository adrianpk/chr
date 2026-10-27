package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func guidedGitReader(root string) workflowIntegrationReader {
	environment := []string{}

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") {
			environment = append(environment, entry)
		}
	}

	environment = append(environment, "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1")

	return workflowIntegrationReader{dir: root, environment: environment}
}

func guidedRepositoryIdentity(ctx context.Context, reader workflowIntegrationReader) (string, error) {
	common, err := reader.output(ctx, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}

	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return "", err
	}

	return "git-common:" + contenthash.Sum([]byte(common)), nil
}

func (r *fileAuthoringSessionRepository) observeGuidedCheckpoint(ctx context.Context, id string, operation authoringcore.GuidedUseCaseOperation, stage string, paths []string) (authoringcore.GuidedStageCheckpoint, error) {
	if len(paths) == 0 || len(paths) > 1024 {
		return authoringcore.GuidedStageCheckpoint{}, errors.New("guided checkpoint requires 1..1024 owned files")
	}

	reader := guidedGitReader(r.projectDir)

	repository, err := guidedRepositoryIdentity(ctx, reader)
	if err != nil {
		return authoringcore.GuidedStageCheckpoint{}, err
	}

	base, err := reader.output(ctx, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return authoringcore.GuidedStageCheckpoint{}, err
	}

	checkpoint := authoringcore.GuidedStageCheckpoint{OperationID: operation.ID, Stage: stage, Repository: repository, BaseCommit: base, Ref: guidedCheckpointRef(id, operation.Digest)}

	size := 0

	for _, path := range paths {
		content, err := r.authoringRead(filepath.FromSlash(path))
		if err != nil {
			return checkpoint, err
		}

		size += len(content)
		if size > 8*workflowstate.FileLimit {
			return checkpoint, errors.New("guided checkpoint exceeds 8 MiB")
		}

		checkpoint.Artifacts = append(checkpoint.Artifacts, authoringcore.GuidedCommittedArtifact{Path: path, Hash: contenthash.Sum(content)})
	}

	return checkpoint, nil
}

func (r *fileAuthoringSessionRepository) commitGuidedCheckpoint(ctx context.Context, checkpoint authoringcore.GuidedStageCheckpoint, digest string) (result authoringcore.GuidedStageCheckpoint, err error) {
	if !validGuidedCheckpointRef(checkpoint.Ref, digest) {
		return checkpoint, workflowstate.ErrConflict
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	reader := guidedGitReader(r.projectDir)
	result = checkpoint
	// show-ref distinguishes a missing ref from an unavailable/corrupt object.
	if _, err := reader.output(ctx, "show-ref", "--verify", "--quiet", checkpoint.Ref); err == nil {
		result.Commit, err = reader.output(ctx, "rev-parse", "--verify", checkpoint.Ref+"^{commit}")
		if err != nil {
			return result, err
		}

		return result, validateGuidedCheckpoint(ctx, r.projectDir, result)
	}

	index, err := os.CreateTemp("", "chr-guided-index-")
	if err != nil {
		return result, err
	}

	indexPath := index.Name()
	if err := index.Close(); err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, os.Remove(indexPath)) }()

	if err := os.Remove(indexPath); err != nil {
		return result, err
	}

	reader.environment = append(reader.environment, "GIT_INDEX_FILE="+indexPath)

	tree, err := r.guidedCheckpointTree(ctx, reader, checkpoint)
	if err != nil {
		return result, err
	}

	result.Commit, err = reader.rawInput(ctx, "Guided "+checkpoint.Stage+" checkpoint\n\nOperation: "+digest+"\n", "commit-tree", tree, "-p", checkpoint.BaseCommit)

	result.Commit = strings.TrimSpace(result.Commit)
	if err != nil {
		return result, err
	}

	if _, err := reader.output(ctx, "update-ref", checkpoint.Ref, result.Commit, ""); err != nil {
		return result, err
	}

	return result, validateGuidedCheckpoint(ctx, r.projectDir, result)
}

func (r *fileAuthoringSessionRepository) guidedCheckpointTree(ctx context.Context, reader workflowIntegrationReader, checkpoint authoringcore.GuidedStageCheckpoint) (string, error) {
	if _, err := reader.output(ctx, "read-tree", checkpoint.BaseCommit); err != nil {
		return "", err
	}

	for _, artifact := range checkpoint.Artifacts {
		content, err := r.authoringRead(filepath.FromSlash(artifact.Path))
		if err != nil {
			return "", err
		}

		if contenthash.Sum(content) != artifact.Hash {
			return "", workflowstate.ErrConflict
		}

		blob, err := reader.rawInput(ctx, string(content), "hash-object", "-w", "--stdin")
		if err != nil {
			return "", err
		}

		blob = strings.TrimSpace(blob)
		if _, err := reader.output(ctx, "update-index", "--add", "--cacheinfo", "100644", blob, artifact.Path); err != nil {
			return "", err
		}
	}

	return reader.output(ctx, "write-tree")
}

func validateGuidedCheckpoint(ctx context.Context, root string, checkpoint authoringcore.GuidedStageCheckpoint) error {
	for _, artifact := range checkpoint.Artifacts {
		if _, err := ReadGuidedCommittedArtifact(ctx, root, checkpoint, artifact.Path); err != nil {
			return err
		}
	}

	return nil
}

// ReadGuidedCommittedArtifact verifies retained Git bytes without reading a working copy.
func ReadGuidedCommittedArtifact(ctx context.Context, root string, checkpoint authoringcore.GuidedStageCheckpoint, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if !regexp.MustCompile(`^[a-f0-9]{40}([a-f0-9]{24})?$`).MatchString(checkpoint.Commit) || !filepath.IsLocal(path) || filepath.ToSlash(filepath.Clean(path)) != path {
		return nil, errors.New("invalid guided checkpoint commit or path")
	}

	reader := guidedGitReader(root)

	identity, err := guidedRepositoryIdentity(ctx, reader)
	if err != nil {
		return nil, err
	}

	if identity != checkpoint.Repository {
		return nil, workflowstate.ErrConflict
	}

	for _, artifact := range checkpoint.Artifacts {
		if artifact.Path != path {
			continue
		}

		content, err := reader.rawOutput(ctx, "show", checkpoint.Commit+":"+path)
		if err != nil {
			return nil, fmt.Errorf("guided committed input unavailable: %w", err)
		}

		if contenthash.Sum([]byte(content)) != artifact.Hash {
			return nil, errors.New("guided committed input hash mismatch")
		}

		return []byte(content), nil
	}

	return nil, errors.New("artifact is outside the guided checkpoint")
}

func validGuidedCheckpointRef(ref, digest string) bool {
	parts := strings.Split(ref, "/")

	return len(parts) == 5 && parts[0] == "refs" && parts[1] == "chr" && parts[2] == "guided" && len(parts[3]) == 64 && parts[4] == digest
}

func guidedCheckpointRef(id, digest string) string {
	return "refs/chr/guided/" + contenthash.Sum([]byte(id)) + "/" + digest
}

// CheckGuidedCheckpointReadiness observes local Git prerequisites without mutations.
func CheckGuidedCheckpointReadiness(ctx context.Context, root string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reader := guidedGitReader(root)
	for _, args := range [][]string{{"rev-parse", "--verify", "HEAD^{commit}"}, {"var", "GIT_AUTHOR_IDENT"}, {"var", "GIT_COMMITTER_IDENT"}} {
		if _, err := reader.output(ctx, args...); err != nil {
			return errors.New("git checkpoints require a repository with an initial commit and a configured author identity")
		}
	}

	return nil
}
