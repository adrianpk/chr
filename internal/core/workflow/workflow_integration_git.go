package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OSWorkflowIntegrationGit fetches into an owned disposable bare repository.
// Repository-to-transport mapping is trusted host configuration, not task input.
type OSWorkflowIntegrationGit struct {
	repository string
	transport  string
}

// ErrWorkflowHistoryDivergence means fetched target history excludes the known
// result object. Transport errors and unavailable objects do not prove removal.
var ErrWorkflowHistoryDivergence = errors.New("canonical target history excludes the integration result")

// NewWorkflowIntegrationGit stores an explicit canonical repository/transport pair.
// Production callers use the canonical HTTPS repository itself as the transport.
func NewWorkflowIntegrationGit(repository, transport string) OSWorkflowIntegrationGit {
	return OSWorkflowIntegrationGit{repository: repository, transport: transport}
}

// VerifyWorkflowIntegration fetches the canonical target and proves its result.
func (git OSWorkflowIntegrationGit) VerifyWorkflowIntegration(ctx context.Context, query workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error) {
	return git.verifyIntegration(ctx, query, nil)
}

// VerifyPreparedWorkflowIntegration preserves the retained proof schema while
// checking original attribution and every prepared object on the remote target.
func (git OSWorkflowIntegrationGit) VerifyPreparedWorkflowIntegration(ctx context.Context, query workflowstate.IntegrationQuery, expected workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	if err := workflowintegrate.CheckProof(query, expected, workflowIntegrationRepairs{}); err != nil {
		return workflowstate.IntegrationProof{}, err
	}

	return git.verifyIntegration(ctx, query, &expected)
}

func (git OSWorkflowIntegrationGit) verifyIntegration(ctx context.Context, query workflowstate.IntegrationQuery, expected *workflowstate.IntegrationProof) (proof workflowstate.IntegrationProof, err error) {
	if err := git.validateQuery(query); err != nil {
		return proof, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	dir, err := os.MkdirTemp("", "chr-integration-proof-")
	if err != nil {
		return proof, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: dir}}

	reader, err = reader.withSourceCredentials(ctx, query.SourceDirectory)
	if err != nil {
		return proof, err
	}

	if _, err := reader.Output(ctx, "init", "--bare", "--quiet"); err != nil {
		return proof, err
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", query.SourceDirectory, query.SourceCommit); err != nil {
		return proof, fmt.Errorf("read retained integration source: %w", err)
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", git.transport, "refs/heads/"+query.TargetBranch+":refs/heads/target"); err != nil {
		return proof, fmt.Errorf("fetch canonical integration target: %w", err)
	}

	if expected == nil {
		return reader.verify(ctx, query)
	}

	if err := reader.readRetainedIntegrationResult(ctx, query); err != nil {
		return proof, err
	}

	return reader.verifyRetained(ctx, query, expected)
}

func (reader workflowIntegrationReader) readRetainedIntegrationResult(ctx context.Context, query workflowstate.IntegrationQuery) error {
	if _, err := reader.Output(ctx, "cat-file", "-e", query.ResultCommit+"^{commit}"); err == nil {
		return nil
	}

	_, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", query.SourceDirectory, query.ResultCommit)

	return err
}

func (git OSWorkflowIntegrationGit) validateQuery(query workflowstate.IntegrationQuery) error {
	repository, err := workflowstate.CanonicalRepository(git.repository)
	if err != nil || repository != query.Repository || git.transport == "" || strings.HasPrefix(git.transport, "-") || strings.ContainsAny(git.transport, "\x00\r\n") {
		return fmt.Errorf("canonical integration transport identity mismatch")
	}

	if !workflowstate.ValidBranch(query.TargetBranch) || query.SourceDirectory == "" {
		return fmt.Errorf("integration requires an exact target branch and retained source")
	}

	for _, id := range []string{query.SourceCommit, query.SourceTree, query.OriginCommit, query.ResultCommit} {
		if !workflowstate.ValidCommitID(id) {
			return fmt.Errorf("integration requires full immutable Git identities")
		}
	}

	return nil
}

func (reader workflowIntegrationReader) verify(ctx context.Context, query workflowstate.IntegrationQuery) (workflowstate.IntegrationProof, error) {
	return reader.verifyRetained(ctx, query, nil)
}

func (reader workflowIntegrationReader) verifyRetained(ctx context.Context, query workflowstate.IntegrationQuery, expected *workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	proof := workflowstate.IntegrationProof{IntegrationQuery: query}
	if expected != nil {
		proof = *expected
		proof.History = nil
	}

	proof.SourceDirectory = ""

	if err := reader.observeIntegrationTarget(ctx, &proof); err != nil {
		return proof, err
	}

	if err := reader.verifyMethod(ctx, &proof); err != nil {
		return proof, err
	}

	if expected == nil && proof.Method != "rebase" {
		provenance, err := reader.combinedProvenance(ctx, proof)
		if err != nil {
			return proof, err
		}

		proof.Provenance = provenance
	}

	var err error

	proof.HistoryHash, err = workflowintegrate.HistoryHash(proof)
	if err != nil {
		return proof, err
	}

	if !workflowRetainedProofMatches(proof, expected) {
		return proof, workflowstate.ErrConflict
	}

	return proof, workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{})
}

func (reader workflowIntegrationReader) tree(ctx context.Context, commit string) (string, error) {
	return reader.Output(ctx, "rev-parse", commit+"^{tree}")
}

func (reader workflowIntegrationReader) parents(ctx context.Context, commit string) ([]string, error) {
	line, err := reader.Output(ctx, "rev-list", "--parents", "-n", "1", commit)

	fields := strings.Fields(line)
	if err != nil || len(fields) == 0 || fields[0] != commit {
		return nil, workflowstate.ErrConflict
	}

	return fields[1:], nil
}

func (reader workflowIntegrationReader) preparedTree(ctx context.Context, origin, base, source string) (string, error) {
	line, err := reader.Output(ctx, "merge-tree", "--write-tree", "--merge-base="+origin, base, source)
	if err != nil || !workflowstate.ValidCommitID(line) {
		return "", fmt.Errorf("integration combined tree cannot be prepared cleanly")
	}

	return line, nil
}

func (reader workflowIntegrationReader) observeIntegrationTarget(ctx context.Context, proof *workflowstate.IntegrationProof) error {
	target, err := reader.Output(ctx, "rev-parse", "refs/heads/target^{commit}")
	if err != nil {
		return err
	}

	proof.ObservedTarget = target
	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", proof.ResultCommit, target); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return fmt.Errorf("%w: %w", ErrWorkflowHistoryDivergence, err)
		}

		return fmt.Errorf("provider result is absent from the canonical target: %w", err)
	}

	tree, err := reader.tree(ctx, proof.SourceCommit)
	if err != nil || tree != proof.SourceTree {
		return workflowstate.ErrConflict
	}

	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.SourceCommit); err != nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowRetainedProofMatches(proof workflowstate.IntegrationProof, expected *workflowstate.IntegrationProof) bool {
	return expected == nil || (proof.HistoryHash == expected.HistoryHash && proof.BaseCommit == expected.BaseCommit && proof.PreparedTree == expected.PreparedTree)
}

type workflowIntegrationReader struct {
	projectcore.GitReader
	preparation *workflowstate.IntegrationPreparation
	signing     *SSHCommitSigning
}
