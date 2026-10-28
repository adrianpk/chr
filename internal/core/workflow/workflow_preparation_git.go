package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OSWorkflowPreparationGit prepares commits in an owned disposable object store
// and attaches a separate validation worktree to the source's coordinator.
type OSWorkflowPreparationGit struct {
	transport     OSWorkflowIntegrationGit
	signing       *SSHCommitSigning
	committerTime time.Time
}

// NewWorkflowPreparationGit stores a canonical identity and trusted Git transport.
func NewWorkflowPreparationGit(repository, transport string) OSWorkflowPreparationGit {
	return OSWorkflowPreparationGit{transport: NewWorkflowIntegrationGit(repository, transport)}
}

// PrepareWorkflowIntegration leaves the source and remote target unchanged.
func (git OSWorkflowPreparationGit) PrepareWorkflowIntegration(ctx context.Context, query workflowstate.PreparationQuery) (proof workflowstate.IntegrationProof, err error) {
	if err := git.validateQuery(query); err != nil {
		return proof, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	dir, err := os.MkdirTemp("", "chr-integration-prepare-")
	if err != nil {
		return proof, err
	}

	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()

	reader, err := git.preparationReader(dir, query)
	if err != nil {
		return proof, err
	}

	reader, err = reader.withSourceCredentials(ctx, query.SourceDirectory)
	if err != nil {
		return proof, err
	}

	if err := git.fetchPreparation(ctx, reader, query); err != nil {
		return proof, err
	}

	proof, err = reader.prepareIntegration(ctx, query)
	if err != nil {
		return proof, err
	}

	if err := attachWorkflowPreparation(ctx, reader, query, proof); err != nil {
		return proof, err
	}

	return proof, nil
}

func (git OSWorkflowPreparationGit) validateQuery(query workflowstate.PreparationQuery) error {
	base := query.IntegrationQuery

	base.ResultCommit = query.BaseCommit
	if err := git.transport.validateQuery(base); err != nil {
		return err
	}

	if !validWorkflowLane(query.Lane) || query.Lane.BaseCommit != query.BaseCommit || !slices.Contains([]string{"merge", "rebase", "squash"}, query.Method) {
		return workflowstate.ErrConflict
	}

	if _, err := os.Lstat(query.Lane.Worktree); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("prepared lane path must be absent")
	}

	return nil
}

func (git OSWorkflowPreparationGit) fetchPreparation(ctx context.Context, reader workflowIntegrationReader, query workflowstate.PreparationQuery) error {
	if _, err := reader.Output(ctx, "init", "--bare", "--quiet"); err != nil {
		return err
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", query.SourceDirectory, query.SourceCommit); err != nil {
		return err
	}

	if _, err := reader.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--", git.transport.transport, "refs/heads/"+query.TargetBranch+":refs/heads/target"); err != nil {
		return err
	}

	base, err := reader.Output(ctx, "rev-parse", "refs/heads/target^{commit}")
	if err != nil || base != query.BaseCommit {
		return fmt.Errorf("%w: canonical base moved during preparation", workflowstate.ErrConflict)
	}

	tree, err := reader.tree(ctx, query.SourceCommit)
	if err != nil || tree != query.SourceTree {
		return workflowstate.ErrConflict
	}

	if _, err := reader.Output(ctx, "merge-base", "--is-ancestor", query.OriginCommit, query.SourceCommit); err != nil {
		return workflowstate.ErrConflict
	}

	return nil
}

func (reader workflowIntegrationReader) prepareIntegration(ctx context.Context, query workflowstate.PreparationQuery) (workflowstate.IntegrationProof, error) {
	proof := workflowstate.IntegrationProof{IntegrationQuery: query.IntegrationQuery, BaseCommit: query.BaseCommit, ObservedTarget: query.BaseCommit}

	proof.SourceDirectory = ""
	if err := reader.verifyBase(ctx, proof); err != nil {
		return proof, err
	}

	var err error

	proof.ResultCommit, err = reader.prepareDistinctResult(ctx, proof)
	if err != nil {
		return proof, err
	}

	if err := reader.verifyMethod(ctx, &proof); err != nil {
		return proof, err
	}

	proof.Provenance, err = reader.integrationProvenance(ctx, proof)
	if err != nil {
		return proof, err
	}

	proof.HistoryHash, err = workflowintegrate.HistoryHash(proof)
	if err != nil || proof.BaseCommit != query.BaseCommit {
		return proof, workflowstate.ErrConflict
	}

	return proof, workflowintegrate.CheckProof(proof.IntegrationQuery, proof, workflowIntegrationRepairs{})
}

func (reader workflowIntegrationReader) prepareCombinedCommit(ctx context.Context, proof workflowstate.IntegrationProof) (string, error) {
	tree, err := reader.preparedTree(ctx, proof.OriginCommit, proof.BaseCommit, proof.SourceCommit)
	if err != nil {
		return "", err
	}

	parents := []string{proof.BaseCommit}
	if proof.Method == "merge" {
		parents = append(parents, proof.SourceCommit)
	}

	return reader.preparationCommit(ctx, tree, parents)
}

func (reader workflowIntegrationReader) prepareRebase(ctx context.Context, proof workflowstate.IntegrationProof) (string, error) {
	sources, err := reader.linearSource(ctx, proof)
	if err != nil {
		return "", err
	}

	previousSource, base := proof.OriginCommit, proof.BaseCommit
	for _, source := range sources {
		tree, err := reader.preparedTree(ctx, previousSource, base, source)
		if err != nil {
			return "", err
		}

		commit, err := reader.replayCommit(ctx, tree, base, source)
		if err != nil {
			return "", err
		}

		previousSource, base = source, commit
	}

	baseTree, err := reader.tree(ctx, proof.BaseCommit)
	if err != nil {
		return "", err
	}

	resultTree, err := reader.tree(ctx, base)
	if err != nil || baseTree == resultTree {
		return "", fmt.Errorf("%w: empty integration unit requires operator resolution", workflowstate.ErrConflict)
	}

	return base, nil
}

func (reader workflowIntegrationReader) preparationCommit(ctx context.Context, tree string, parents []string) (string, error) {
	author := workflowstate.IntegrationIdentity{Name: "Workflow Host", Email: "workflow@example.invalid", Date: reader.preparationCommitterDate()}

	return reader.preparationObject(ctx, workflowPreparationCommit{Tree: tree, Parents: parents, Authorship: workflowPreparationAuthorship{Author: author, Encoding: "", Message: "Prepare configured workflow integration\n"}})
}

func attachWorkflowPreparation(ctx context.Context, reader workflowIntegrationReader, query workflowstate.PreparationQuery, proof workflowstate.IntegrationProof) error {
	source := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: query.SourceDirectory}}
	branch := strings.TrimPrefix(query.Lane.Branch, "refs/heads/")
	git := executioncore.OSWorktreeGit{}

	exists, err := git.BranchExists(ctx, query.SourceDirectory, branch)
	if err != nil || exists {
		return errorsOrConflict(err)
	}

	if _, err := source.Output(ctx, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-write-fetch-head", "--", reader.Directory, proof.ResultCommit); err != nil {
		return err
	}

	if err := git.CreateBranch(ctx, query.SourceDirectory, branch, proof.ResultCommit); err != nil {
		return err
	}

	return git.AddWorktree(ctx, query.SourceDirectory, query.Lane.Worktree, branch)
}
