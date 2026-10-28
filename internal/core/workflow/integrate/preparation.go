package integrate

import (
	"fmt"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowPreparationCollisionLimit bounds retries that distinguish a prepared commit from its source identity.
const WorkflowPreparationCollisionLimit = 3

// WorkflowDistinctPreparation retries only within the existing identity-collision budget using the caller-owned builder.
func WorkflowDistinctPreparation(source string, preparation *workflowstate.IntegrationPreparation, build func() (string, error)) (string, error) {
	if preparation == nil || !ValidWorkflowCommitDate(preparation.InitialCommitterDate) {
		return "", workflowstate.ErrConflict
	}

	seconds, err := strconv.ParseInt(strings.Fields(preparation.InitialCommitterDate)[0], 10, 64)
	if err != nil || seconds > 1<<63-1-WorkflowPreparationCollisionLimit {
		return "", workflowstate.ErrConflict
	}

	for retry := 0; retry <= WorkflowPreparationCollisionLimit; retry++ {
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

	return "", fmt.Errorf("%w: prepared source identity collision budget exhausted", workflowstate.ErrConflict)
}

// WorkflowPreparationIdentity derives the original retained preparation identity from immutable proof fields.
func WorkflowPreparationIdentity(proof workflowstate.IntegrationProof, p workflowstate.IntegrationPreparation) string {
	profile := ""
	if p.Signing != nil {
		profile = p.Signing.Hash
	}

	return contenthash.Sum([]byte(fmt.Sprintf("chr-preparation/v1:%s:%s:%s:%s:%s:%s:%s:%d:%s", proof.Repository, proof.Method, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit, p.InitialCommitterDate, p.CommitterDate, p.NoOpCollisions, profile)))
}

// ValidWorkflowPreparationMetadata checks the original collision, identity and signature metadata.
func ValidWorkflowPreparationMetadata(proof workflowstate.IntegrationProof, p workflowstate.IntegrationPreparation) bool {
	if !ValidWorkflowCommitDate(p.InitialCommitterDate) || !ValidWorkflowCommitDate(p.CommitterDate) || p.NoOpCollisions < 0 || p.NoOpCollisions > WorkflowPreparationCollisionLimit {
		return false
	}

	initial, err := strconv.ParseInt(strings.Fields(p.InitialCommitterDate)[0], 10, 64)
	if err != nil || initial > 1<<63-1-WorkflowPreparationCollisionLimit || p.CommitterDate != fmt.Sprintf("%d +0000", initial+int64(p.NoOpCollisions)) {
		return false
	}

	return p.Identity == WorkflowPreparationIdentity(proof, p) && validWorkflowPreparationSignatures(proof, p)
}

// WorkflowPreparationCommitterMatches checks the host committer identity and the existing replay timestamp rule.
func WorkflowPreparationCommitterMatches(entry workflowstate.IntegrationAttribution, p workflowstate.IntegrationPreparation, replay bool) bool {
	return entry.Committer.Name == "Workflow Host" && entry.Committer.Email == "workflow@example.invalid" && (!replay || entry.Committer.Date == p.CommitterDate)
}

// CheckWorkflowPreparationSigning binds retained signature requirements and the configured signer fingerprint to a proof.
func CheckWorkflowPreparationSigning(proof workflowstate.IntegrationProof, required bool, profile string) error {
	var p *workflowstate.IntegrationPreparation
	if proof.Provenance != nil {
		p = proof.Provenance.Preparation
	}

	if !required && profile == "" && (p == nil || !p.RequiredSignatures) {
		return nil
	}

	if p == nil || !p.RequiredSignatures || p.Signing == nil || p.Signing.Hash != profile {
		return workflowstate.ErrConflict
	}

	return nil
}

// ValidWorkflowPreparationObject checks the original commit-object identity, parents, author and encoding bounds.
func ValidWorkflowPreparationObject(tree string, parents []string, author workflowstate.IntegrationIdentity, encoding string) bool {
	if !workflowstate.ValidCommitID(tree) || !validWorkflowCommitIdentity(author) || strings.ContainsAny(encoding, "\x00\r\n") || len(parents) < 1 || len(parents) > 2 {
		return false
	}

	for _, parent := range parents {
		if !workflowstate.ValidCommitID(parent) || len(parent) != len(tree) {
			return false
		}
	}

	return true
}
