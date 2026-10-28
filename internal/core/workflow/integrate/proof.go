package integrate

import (
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ValidateWorkflowIntegrationProof checks the original immutable candidate observations and retained proof.
func ValidateWorkflowIntegrationProof(query workflowstate.IntegrationQuery, proof workflowstate.IntegrationProof, repairs RepairProof) error {
	query.SourceDirectory = ""
	if proof.IntegrationQuery != query || !slices.Contains([]string{"merge", "rebase", "squash"}, proof.Method) || proof.PreparedTree != proof.VerifiedTree || len(proof.History) == 0 || len(proof.History) > workflowstate.RecordLimit {
		return workflowstate.ErrConflict
	}

	for _, id := range []string{proof.BaseCommit, proof.PreparedTree, proof.VerifiedTree, proof.ObservedTarget} {
		if !workflowstate.ValidCommitID(id) {
			return workflowstate.ErrConflict
		}
	}

	hash, err := WorkflowIntegrationHistoryHash(proof)
	if err != nil || hash != proof.HistoryHash {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationHistory(proof, repairs)
}

func validateWorkflowIntegrationHistory(proof workflowstate.IntegrationProof, repairs RepairProof) error {
	if proof.Method != "rebase" && len(proof.History) != 1 {
		return workflowstate.ErrConflict
	}

	previous := proof.BaseCommit
	for _, entry := range proof.History {
		if entry.Parent != previous || !validWorkflowIntegrationCommit(entry) {
			return workflowstate.ErrConflict
		}

		previous = entry.Commit
	}

	last := proof.History[len(proof.History)-1]
	if last.Source != proof.SourceCommit || !workflowIntegrationHistoryTipMatches(proof, last) {
		return workflowstate.ErrConflict
	}

	return ValidateWorkflowIntegrationProvenance(proof, repairs)
}

func validWorkflowIntegrationCommit(entry workflowstate.IntegrationCommit) bool {
	for _, id := range []string{entry.Source, entry.Commit, entry.Tree, entry.Parent} {
		if !workflowstate.ValidCommitID(id) {
			return false
		}
	}

	return true
}

// WorkflowIntegrationHistoryHash checks the original immutable candidate observations and retained proof.
func WorkflowIntegrationHistoryHash(proof workflowstate.IntegrationProof) (string, error) {
	data, err := json.Marshal(struct {
		Method, Source, Base, Tree string
		History                    []workflowstate.IntegrationCommit
		Provenance                 *workflowstate.IntegrationProvenance `json:",omitempty"`
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History, proof.Provenance})

	return contenthash.Sum(data), err
}

// WorkflowIntegrationParentsMatch checks the original immutable candidate observations and retained proof.
func WorkflowIntegrationParentsMatch(proof workflowstate.IntegrationProof, parents []string) bool {
	switch proof.Method {
	case "merge":
		return len(parents) == 2 && parents[1] == proof.SourceCommit
	case "squash":
		return len(parents) == 1
	default:
		return false
	}
}

func workflowIntegrationHistoryTipMatches(proof workflowstate.IntegrationProof, last workflowstate.IntegrationCommit) bool {
	if proof.Method == "rebase" && proof.Provenance != nil {
		p := proof.Provenance
		if !ValidProvenanceHeader(proof, *p) || len(p.Source) != len(proof.History) {
			return false
		}

		entry := p.Prepared[len(p.Source)-1]

		return entry.Commit == last.Commit && entry.Tree == last.Tree
	}

	return last.Commit == proof.ResultCommit && last.Tree == proof.PreparedTree
}

// ValidateWorkflowIntegrationProvenance checks the original immutable candidate observations and retained proof.
func ValidateWorkflowIntegrationProvenance(proof workflowstate.IntegrationProof, repairs RepairProof) error {
	p := proof.Provenance
	if p == nil {
		return nil
	}

	if !ValidProvenanceHeader(proof, *p) {
		return workflowstate.ErrConflict
	}

	if err := CheckProvenanceCommits(proof, *p); err != nil {
		return err
	}

	if p.Preparation != nil && !ValidPreparationMetadata(proof, *p.Preparation) {
		return workflowstate.ErrConflict
	}

	if err := repairs.Check(proof); err != nil {
		return err
	}

	data, err := json.Marshal(p)
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowstate.ErrConflict
	}

	return nil
}
