package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// OSWorkflowIntegrationGit fetches into an owned disposable bare repository.
// Repository-to-transport mapping is trusted host configuration, not task input.
type OSWorkflowIntegrationGit struct {
	repository string
	transport  string
}

// NewWorkflowIntegrationGit stores an explicit canonical repository/transport pair.
// Production callers use the canonical HTTPS repository itself as the transport.
func NewWorkflowIntegrationGit(repository, transport string) OSWorkflowIntegrationGit {
	return OSWorkflowIntegrationGit{repository: repository, transport: transport}
}

// VerifyWorkflowIntegration fetches the canonical target and proves its result.
func (git OSWorkflowIntegrationGit) VerifyWorkflowIntegration(ctx context.Context, query WorkflowIntegrationQuery) (WorkflowIntegrationProof, error) {
	return git.verifyIntegration(ctx, query, nil)
}

// VerifyPreparedWorkflowIntegration preserves the retained proof schema while
// checking original attribution and every prepared object on the remote target.
func (git OSWorkflowIntegrationGit) VerifyPreparedWorkflowIntegration(ctx context.Context, query WorkflowIntegrationQuery, expected WorkflowIntegrationProof) (WorkflowIntegrationProof, error) {
	if err := validateWorkflowIntegrationProof(query, expected); err != nil {
		return WorkflowIntegrationProof{}, err
	}

	return git.verifyIntegration(ctx, query, &expected)
}

func (git OSWorkflowIntegrationGit) verifyIntegration(ctx context.Context, query WorkflowIntegrationQuery, expected *WorkflowIntegrationProof) (proof WorkflowIntegrationProof, err error) {
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

	reader := workflowIntegrationReader{dir: dir}
	if _, err := reader.output(ctx, "init", "--bare", "--quiet"); err != nil {
		return proof, err
	}

	if _, err := reader.output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", query.SourceDirectory, query.SourceCommit); err != nil {
		return proof, fmt.Errorf("read retained integration source: %w", err)
	}

	if _, err := reader.output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", git.transport, "refs/heads/"+query.TargetBranch+":refs/heads/target"); err != nil {
		return proof, fmt.Errorf("fetch canonical integration target: %w", err)
	}

	if expected == nil {
		return reader.verify(ctx, query)
	}

	return reader.verifyRetained(ctx, query, expected)
}

func (git OSWorkflowIntegrationGit) validateQuery(query WorkflowIntegrationQuery) error {
	repository, err := workflowCanonicalRepository(git.repository)
	if err != nil || repository != query.Repository || git.transport == "" || strings.HasPrefix(git.transport, "-") || strings.ContainsAny(git.transport, "\x00\r\n") {
		return fmt.Errorf("canonical integration transport identity mismatch")
	}

	if !validWorkflowPRBranch(query.TargetBranch) || query.SourceDirectory == "" {
		return fmt.Errorf("integration requires an exact target branch and retained source")
	}

	for _, id := range []string{query.SourceCommit, query.SourceTree, query.OriginCommit, query.ResultCommit} {
		if !workflowCommitID(id) {
			return fmt.Errorf("integration requires full immutable Git identities")
		}
	}

	return nil
}

type workflowIntegrationReader struct {
	dir         string
	environment []string
	preparation *WorkflowIntegrationPreparation
	signing     *WorkflowSSHCommitSigning
}

func (reader workflowIntegrationReader) output(ctx context.Context, args ...string) (string, error) {
	data, err := reader.rawOutput(ctx, args...)

	return strings.TrimSpace(data), err
}

func (reader workflowIntegrationReader) rawOutput(ctx context.Context, args ...string) (string, error) {
	return reader.rawInput(ctx, "", args...)
}

func (reader workflowIntegrationReader) rawInput(ctx context.Context, input string, args ...string) (string, error) {
	command := reader.process(ctx, args...)
	command.Stdin = strings.NewReader(input)
	stdout, stderr := &workflowProviderOutput{}, &workflowCommandOutput{}
	command.Stdout, command.Stderr = stdout, stderr

	if err := command.Run(); err != nil {
		return "", fmt.Errorf("integration Git observation failed: %w", err)
	}

	if stdout.overflow {
		return "", fmt.Errorf("integration Git output exceeds file bounds")
	}

	return stdout.buffer.String(), nil
}

func (reader workflowIntegrationReader) process(ctx context.Context, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, "git", append([]string{"--no-replace-objects", "--literal-pathspecs", "-c", "core.hooksPath=/dev/null", "-C", reader.dir}, args...)...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}

		return err
	}

	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1")
	if reader.environment != nil {
		command.Env = reader.environment
	}

	command.WaitDelay = time.Second

	return command
}

func (reader workflowIntegrationReader) verify(ctx context.Context, query WorkflowIntegrationQuery) (WorkflowIntegrationProof, error) {
	return reader.verifyRetained(ctx, query, nil)
}

func (reader workflowIntegrationReader) verifyRetained(ctx context.Context, query WorkflowIntegrationQuery, expected *WorkflowIntegrationProof) (WorkflowIntegrationProof, error) {
	proof := WorkflowIntegrationProof{WorkflowIntegrationQuery: query}
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

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		return proof, err
	}

	if !workflowRetainedProofMatches(proof, expected) {
		return proof, ErrWorkflowConflict
	}

	return proof, validateWorkflowIntegrationProof(query, proof)
}

func (reader workflowIntegrationReader) tree(ctx context.Context, commit string) (string, error) {
	return reader.output(ctx, "rev-parse", commit+"^{tree}")
}

func (reader workflowIntegrationReader) parents(ctx context.Context, commit string) ([]string, error) {
	line, err := reader.output(ctx, "rev-list", "--parents", "-n", "1", commit)

	fields := strings.Fields(line)
	if err != nil || len(fields) == 0 || fields[0] != commit {
		return nil, ErrWorkflowConflict
	}

	return fields[1:], nil
}

func (reader workflowIntegrationReader) preparedTree(ctx context.Context, origin, base, source string) (string, error) {
	line, err := reader.output(ctx, "merge-tree", "--write-tree", "--merge-base="+origin, base, source)
	if err != nil || !workflowCommitID(line) {
		return "", fmt.Errorf("integration combined tree cannot be prepared cleanly")
	}

	return line, nil
}

func (reader workflowIntegrationReader) observeIntegrationTarget(ctx context.Context, proof *WorkflowIntegrationProof) error {
	target, err := reader.output(ctx, "rev-parse", "refs/heads/target^{commit}")
	if err != nil {
		return err
	}

	proof.ObservedTarget = target
	if _, err := reader.output(ctx, "merge-base", "--is-ancestor", proof.ResultCommit, target); err != nil {
		return fmt.Errorf("provider result is absent from the canonical target: %w", err)
	}

	tree, err := reader.tree(ctx, proof.SourceCommit)
	if err != nil || tree != proof.SourceTree {
		return ErrWorkflowConflict
	}

	if _, err := reader.output(ctx, "merge-base", "--is-ancestor", proof.OriginCommit, proof.SourceCommit); err != nil {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowRetainedProofMatches(proof WorkflowIntegrationProof, expected *WorkflowIntegrationProof) bool {
	return expected == nil || (proof.HistoryHash == expected.HistoryHash && proof.BaseCommit == expected.BaseCommit && proof.PreparedTree == expected.PreparedTree)
}
