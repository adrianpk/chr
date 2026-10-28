package workflow

import (
	"context"
	"fmt"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (git OSWorkflowPreparationGit) preparationReader(dir string, query PreparationQuery) (workflowIntegrationReader, error) {
	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: dir}, signing: git.signing}

	date := git.committerTime
	if date.IsZero() {
		date = time.Now().UTC()
	}

	initial := fmt.Sprintf("%d +0000", date.Unix())

	reader.preparation = &workflowstate.IntegrationPreparation{InitialCommitterDate: initial, CommitterDate: initial, RequiredSignatures: query.RequiredSignatures}
	if !query.RequiredSignatures {
		reader.signing = nil
		if query.SigningProfile != "" {
			return reader, workflowstate.ErrConflict
		}

		return reader, nil
	}

	profile, err := workflowSSHSigningProfile(git.signing)
	if err != nil || profile.Hash != query.SigningProfile {
		return reader, fmt.Errorf("certified host commit signing path required")
	}

	reader.preparation.Signing = &profile

	return reader, nil
}

func (reader workflowIntegrationReader) prepareDistinctResult(ctx context.Context, proof workflowstate.IntegrationProof) (string, error) {
	build := func() (string, error) {
		if proof.Method == "rebase" {
			return reader.prepareRebase(ctx, proof)
		}

		return reader.prepareCombinedCommit(ctx, proof)
	}

	return workflowintegrate.DistinctResult(proof.SourceCommit, reader.preparation, build)
}

func (reader workflowIntegrationReader) addPreparationMetadata(ctx context.Context, proof workflowstate.IntegrationProof, provenance *workflowstate.IntegrationProvenance) error {
	metadata := reader.preparation
	if metadata == nil && proof.Provenance != nil {
		metadata = proof.Provenance.Preparation
	}

	if metadata == nil {
		return nil
	}

	p := *metadata
	p.Identity = workflowintegrate.PreparationIdentity(proof, p)
	p.Signatures = nil

	for i, entry := range provenance.Prepared {
		mode := workflowintegrate.HostFixCommitter
		if i < len(provenance.Prepared)-len(provenance.IntegrationFixes) {
			mode = workflowintegrate.ReplayCommitter
		}

		if !workflowintegrate.CommitterMatches(entry, p, mode) {
			return workflowstate.ErrConflict
		}

		if p.RequiredSignatures {
			result, err := reader.verifyPreparationSignature(ctx, entry.Commit, p.Signing)
			if err != nil {
				return err
			}

			p.Signatures = append(p.Signatures, result)
		}
	}

	provenance.Preparation = &p

	return nil
}
