package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const workflowPreparationCollisionLimit = 3

// WorkflowIntegrationPreparation binds host identity, bounded retries and
// signature policy to the prepared objects without adding artificial commits.
type WorkflowIntegrationPreparation struct {
	Identity             string                         `json:"identity"`
	InitialCommitterDate string                         `json:"initial_committer_date"`
	CommitterDate        string                         `json:"committer_date"`
	NoOpCollisions       int                            `json:"no_op_collisions"`
	RequiredSignatures   bool                           `json:"required_signatures"`
	Signing              *WorkflowCommitSigningProfile  `json:"signing,omitempty"`
	Signatures           []WorkflowIntegrationSignature `json:"signatures,omitempty"`
}

func (git OSWorkflowPreparationGit) preparationReader(dir string, query WorkflowPreparationQuery) (workflowIntegrationReader, error) {
	reader := workflowIntegrationReader{dir: dir, signing: git.signing}

	date := git.committerTime
	if date.IsZero() {
		date = time.Now().UTC()
	}

	initial := fmt.Sprintf("%d +0000", date.Unix())

	reader.preparation = &WorkflowIntegrationPreparation{InitialCommitterDate: initial, CommitterDate: initial, RequiredSignatures: query.RequiredSignatures}
	if !query.RequiredSignatures {
		reader.signing = nil
		if query.SigningProfile != "" {
			return reader, ErrWorkflowConflict
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

func (reader workflowIntegrationReader) prepareDistinctResult(ctx context.Context, proof WorkflowIntegrationProof) (string, error) {
	build := func() (string, error) {
		if proof.Method == "rebase" {
			return reader.prepareRebase(ctx, proof)
		}

		return reader.prepareCombinedCommit(ctx, proof)
	}

	return workflowDistinctPreparation(proof.SourceCommit, reader.preparation, build)
}

func workflowDistinctPreparation(source string, preparation *WorkflowIntegrationPreparation, build func() (string, error)) (string, error) {
	if preparation == nil || !validWorkflowCommitDate(preparation.InitialCommitterDate) {
		return "", ErrWorkflowConflict
	}

	seconds, err := strconv.ParseInt(strings.Fields(preparation.InitialCommitterDate)[0], 10, 64)
	if err != nil || seconds > 1<<63-1-workflowPreparationCollisionLimit {
		return "", ErrWorkflowConflict
	}

	for retry := 0; retry <= workflowPreparationCollisionLimit; retry++ {
		preparation.NoOpCollisions = retry
		preparation.CommitterDate = fmt.Sprintf("%d +0000", seconds+int64(retry))

		result, err := build()
		if err != nil {
			return "", err
		}

		if result != source {
			return result, nil
		}
	}

	return "", fmt.Errorf("%w: prepared source identity collision budget exhausted", ErrWorkflowConflict)
}

func workflowPreparationIdentity(proof WorkflowIntegrationProof, p WorkflowIntegrationPreparation) string {
	profile := ""
	if p.Signing != nil {
		profile = p.Signing.Hash
	}

	return workflowHash([]byte(fmt.Sprintf("chr-preparation/v1:%s:%s:%s:%s:%s:%s:%s:%d:%s", proof.Repository, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit, p.InitialCommitterDate, p.CommitterDate, p.NoOpCollisions, profile)))
}

func validWorkflowPreparationMetadata(proof WorkflowIntegrationProof, p WorkflowIntegrationPreparation) bool {
	if !validWorkflowCommitDate(p.InitialCommitterDate) || !validWorkflowCommitDate(p.CommitterDate) || p.NoOpCollisions < 0 || p.NoOpCollisions > workflowPreparationCollisionLimit {
		return false
	}

	initial, err := strconv.ParseInt(strings.Fields(p.InitialCommitterDate)[0], 10, 64)
	if err != nil || initial > 1<<63-1-workflowPreparationCollisionLimit || p.CommitterDate != fmt.Sprintf("%d +0000", initial+int64(p.NoOpCollisions)) {
		return false
	}

	return p.Identity == workflowPreparationIdentity(proof, p) && validWorkflowPreparationSignatures(proof, p)
}

func (reader workflowIntegrationReader) addPreparationMetadata(ctx context.Context, proof WorkflowIntegrationProof, provenance *WorkflowIntegrationProvenance) error {
	metadata := reader.preparation
	if metadata == nil && proof.Provenance != nil {
		metadata = proof.Provenance.Preparation
	}

	if metadata == nil {
		return nil
	}

	p := *metadata
	p.Identity = workflowPreparationIdentity(proof, p)
	p.Signatures = nil

	for i, entry := range provenance.Prepared {
		if !workflowPreparationCommitterMatches(entry, p, i < len(provenance.Prepared)-len(provenance.IntegrationFixes)) {
			return ErrWorkflowConflict
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

func workflowPreparationCommitterMatches(entry WorkflowIntegrationAttribution, p WorkflowIntegrationPreparation, replay bool) bool {
	return entry.Committer.Name == "Workflow Host" && entry.Committer.Email == "workflow@example.invalid" && (!replay || entry.Committer.Date == p.CommitterDate)
}

func checkWorkflowPreparationSigning(proof WorkflowIntegrationProof, required bool, profile string) error {
	var p *WorkflowIntegrationPreparation
	if proof.Provenance != nil {
		p = proof.Provenance.Preparation
	}

	if !required && profile == "" && (p == nil || !p.RequiredSignatures) {
		return nil
	}

	if p == nil || !p.RequiredSignatures || p.Signing == nil || p.Signing.Hash != profile {
		return ErrWorkflowConflict
	}

	return nil
}
