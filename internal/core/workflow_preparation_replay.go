package core

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	if !workflowCommitID(tree) || !workflowCommitID(parent) || !validWorkflowCommitIdentity(author) || strings.ContainsAny(encoding, "\x00\r\n") {
		return "", ErrWorkflowConflict
	}

	// Write exact message bytes: commit-tree normalizes terminal newlines. Old
	// signature headers are deliberately excluded from the rewritten object.
	header := fmt.Sprintf("tree %s\nparent %s\nauthor %s <%s> %s\ncommitter Workflow Host <workflow@example.invalid> %d +0000\n", tree, parent, author.Name, author.Email, author.Date, time.Now().Unix())
	if encoding != "" {
		header += "encoding " + encoding + "\n"
	}

	data := header + "\n" + message
	if len(data) >= WorkflowFileLimit {
		return "", ErrWorkflowConflict
	}

	id, err := reader.rawInput(ctx, data, "hash-object", "-t", "commit", "-w", "--stdin")

	return strings.TrimSpace(id), err
}

func (reader workflowIntegrationReader) prepareRebaseFix(ctx context.Context, proof WorkflowIntegrationProof, tree string, finding AgenticReviewFinding) (WorkflowIntegrationProof, error) {
	if !validWorkflowRebaseFixInput(proof, finding) {
		return proof, ErrWorkflowConflict
	}

	if err := reader.verifyPublicationHistory(ctx, proof); err != nil {
		return proof, err
	}

	if tree == proof.PreparedTree || len(proof.Provenance.Prepared) >= WorkflowRecordLimit {
		return proof, ErrWorkflowConflict
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

	proof.Provenance, err = reader.rebaseProvenance(ctx, proof)
	if err != nil {
		return proof, err
	}

	proof.HistoryHash, err = workflowIntegrationHistoryHash(proof)
	if err != nil {
		return proof, err
	}

	return proof, reader.verifyPublicationHistory(ctx, proof)
}

func validWorkflowRebaseFixInput(proof WorkflowIntegrationProof, finding AgenticReviewFinding) bool {
	return proof.Method == "rebase" && proof.Provenance != nil && validateWorkflowIntegrationProof(proof.WorkflowIntegrationQuery, proof) == nil && validateAgenticReviewFinding(finding) == nil && finding.Category != AgenticReviewFindingCosmetic
}
