package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) replayCommit(ctx context.Context, tree, parent, source string) (string, error) {
	data, err := reader.rawOutput(ctx, "cat-file", "commit", source)
	if err != nil {
		return "", err
	}

	attribution, err := workflowCommitAttribution(source, data)
	if err != nil {
		return "", err
	}

	_, message, _ := strings.Cut(data, "\n\n")

	return reader.attributedPreparationCommit(ctx, tree, parent, attribution.Author, attribution.Encoding, message)
}

func (reader workflowIntegrationReader) attributedPreparationCommit(ctx context.Context, tree, parent string, author WorkflowIntegrationIdentity, encoding, message string) (string, error) {
	return reader.preparationObject(ctx, tree, []string{parent}, author, encoding, message)
}

func (reader workflowIntegrationReader) preparationCommitterDate() string {
	if reader.preparation != nil {
		return reader.preparation.CommitterDate
	}

	return fmt.Sprintf("%d +0000", time.Now().Unix())
}

func (reader workflowIntegrationReader) preparationObject(ctx context.Context, tree string, parents []string, author WorkflowIntegrationIdentity, encoding, message string) (string, error) {
	if !validWorkflowPreparationObject(tree, parents, author, encoding) {
		return "", workflowstate.ErrConflict
	}

	// Write exact message bytes: commit-tree normalizes terminal newlines. Old
	// signature headers are deliberately excluded from the rewritten object.
	header := "tree " + tree + "\n"

	for _, parent := range parents {
		header += "parent " + parent + "\n"
	}

	header += fmt.Sprintf("author %s <%s> %s\ncommitter Workflow Host <workflow@example.invalid> %s\n", author.Name, author.Email, author.Date, reader.preparationCommitterDate())
	if encoding != "" {
		header += "encoding " + encoding + "\n"
	}

	data, err := reader.signPreparationObject(ctx, header+"\n"+message)
	if err != nil {
		return "", err
	}

	if len(data) >= workflowstate.FileLimit {
		return "", workflowstate.ErrConflict
	}

	id, err := reader.rawInput(ctx, data, "hash-object", "-t", "commit", "-w", "--stdin")

	return strings.TrimSpace(id), err
}

func validWorkflowPreparationObject(tree string, parents []string, author WorkflowIntegrationIdentity, encoding string) bool {
	if !workflowCommitID(tree) || !validWorkflowCommitIdentity(author) || strings.ContainsAny(encoding, "\x00\r\n") || len(parents) < 1 || len(parents) > 2 {
		return false
	}

	for _, parent := range parents {
		if !workflowCommitID(parent) || len(parent) != len(tree) {
			return false
		}
	}

	return true
}

func (reader workflowIntegrationReader) prepareRebaseFix(ctx context.Context, proof WorkflowIntegrationProof, tree string, finding workflowstate.AgenticReviewFinding) (WorkflowIntegrationProof, error) {
	if !validWorkflowRebaseFixInput(proof, finding) {
		return proof, workflowstate.ErrConflict
	}

	if err := reader.requireRebaseFixSigner(proof); err != nil {
		return proof, err
	}

	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return proof, err
	}

	if tree == proof.PreparedTree || len(proof.Provenance.Prepared) >= workflowstate.RecordLimit {
		return proof, workflowstate.ErrConflict
	}

	author := WorkflowIntegrationIdentity{Name: "Workflow Host", Email: "workflow@example.invalid", Date: fmt.Sprintf("%d +0000", time.Now().Unix())}

	commit, err := reader.attributedPreparationCommit(ctx, tree, proof.ResultCommit, author, "", "fix(integration): "+finding.Title+"\n\n"+finding.Detail+"\n")
	if err != nil {
		return proof, err
	}

	p := *proof.Provenance
	p.IntegrationFixes = append(append([]WorkflowIntegrationFix{}, p.IntegrationFixes...), WorkflowIntegrationFix{Commit: commit, Finding: finding})
	proof.Provenance = &p
	proof.ResultCommit, proof.PreparedTree, proof.VerifiedTree = commit, tree, tree

	proof.Provenance, err = reader.integrationProvenance(ctx, proof)
	if err != nil {
		return proof, err
	}

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		return proof, err
	}

	return proof, reader.verifyPublicationHistory(ctx, proof)
}

func (reader workflowIntegrationReader) requireRebaseFixSigner(proof WorkflowIntegrationProof) error {
	p := proof.Provenance.Preparation
	if p == nil || !p.RequiredSignatures {
		return nil
	}

	profile, err := workflowSSHSigningProfile(reader.signing)
	if err != nil || p.Signing == nil || profile.Hash != p.Signing.Hash {
		return fmt.Errorf("certified host commit signing path required for integration fixes")
	}

	return nil
}

func validWorkflowRebaseFixInput(proof WorkflowIntegrationProof, finding workflowstate.AgenticReviewFinding) bool {
	return proof.Method == "rebase" && proof.Provenance != nil && validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) == nil && validateAgenticReviewFinding(finding) == nil && finding.Category != AgenticReviewFindingCosmetic
}
