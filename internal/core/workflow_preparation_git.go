package core

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
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// OSWorkflowPreparationGit prepares commits in an owned disposable object store
// and attaches a separate validation worktree to the source's coordinator.
type OSWorkflowPreparationGit struct {
	transport     OSWorkflowIntegrationGit
	signing       *WorkflowSSHCommitSigning
	committerTime time.Time
}

// NewWorkflowPreparationGit stores a canonical identity and trusted Git transport.
func NewWorkflowPreparationGit(repository, transport string) OSWorkflowPreparationGit {
	return OSWorkflowPreparationGit{transport: NewWorkflowIntegrationGit(repository, transport)}
}

// PrepareWorkflowIntegration leaves the source and remote target unchanged.
func (git OSWorkflowPreparationGit) PrepareWorkflowIntegration(ctx context.Context, query WorkflowPreparationQuery) (proof WorkflowIntegrationProof, err error) {
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

func (git OSWorkflowPreparationGit) validateQuery(query WorkflowPreparationQuery) error {
	base := query.WorkflowIntegrationQuery

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

func (git OSWorkflowPreparationGit) fetchPreparation(ctx context.Context, reader workflowIntegrationReader, query WorkflowPreparationQuery) error {
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

func (reader workflowIntegrationReader) prepareIntegration(ctx context.Context, query WorkflowPreparationQuery) (WorkflowIntegrationProof, error) {
	proof := WorkflowIntegrationProof{WorkflowIntegrationQuery: query.WorkflowIntegrationQuery, BaseCommit: query.BaseCommit, ObservedTarget: query.BaseCommit}

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

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil || proof.BaseCommit != query.BaseCommit {
		return proof, workflowstate.ErrConflict
	}

	return proof, validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof)
}

func (reader workflowIntegrationReader) prepareCombinedCommit(ctx context.Context, proof WorkflowIntegrationProof) (string, error) {
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

func (reader workflowIntegrationReader) prepareRebase(ctx context.Context, proof WorkflowIntegrationProof) (string, error) {
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
	author := WorkflowIntegrationIdentity{Name: "Workflow Host", Email: "workflow@example.invalid", Date: reader.preparationCommitterDate()}

	return reader.preparationObject(ctx, tree, parents, author, "", "Prepare configured workflow integration\n")
}

func attachWorkflowPreparation(ctx context.Context, reader workflowIntegrationReader, query WorkflowPreparationQuery, proof WorkflowIntegrationProof) error {
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
