package integrate

import (
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CheckProof checks the exact query, ordered method history and complete retained provenance.
func CheckProof(query workflowstate.IntegrationQuery, proof workflowstate.IntegrationProof, repairs RepairProof) error {
	query.SourceDirectory = ""
	if proof.IntegrationQuery != query || !slices.Contains([]string{"merge", "rebase", "squash"}, proof.Method) || proof.PreparedTree != proof.VerifiedTree || len(proof.History) == 0 || len(proof.History) > workflowstate.RecordLimit {
		return workflowstate.ErrConflict
	}

	for _, id := range []string{proof.BaseCommit, proof.PreparedTree, proof.VerifiedTree, proof.ObservedTarget} {
		if !workflowstate.ValidCommitID(id) {
			return workflowstate.ErrConflict
		}
	}

	hash, err := HistoryHash(proof)
	if err != nil || hash != proof.HistoryHash {
		return workflowstate.ErrConflict
	}

	return checkHistory(proof, repairs)
}

func checkHistory(proof workflowstate.IntegrationProof, repairs RepairProof) error {
	if proof.Method != "rebase" && len(proof.History) != 1 {
		return workflowstate.ErrConflict
	}

	previous := proof.BaseCommit
	for _, entry := range proof.History {
		if entry.Parent != previous || !validHistoryCommit(entry) {
			return workflowstate.ErrConflict
		}

		previous = entry.Commit
	}

	last := proof.History[len(proof.History)-1]
	if last.Source != proof.SourceCommit || !historyTipMatches(proof, last) {
		return workflowstate.ErrConflict
	}

	return CheckProvenance(proof, repairs)
}

func validHistoryCommit(entry workflowstate.IntegrationCommit) bool {
	for _, id := range []string{entry.Source, entry.Commit, entry.Tree, entry.Parent} {
		if !workflowstate.ValidCommitID(id) {
			return false
		}
	}

	return true
}

// HistoryHash hashes the original ordered history and optional provenance encoding.
func HistoryHash(proof workflowstate.IntegrationProof) (string, error) {
	data, err := json.Marshal(struct {
		Method, Source, Base, Tree string
		History                    []workflowstate.IntegrationCommit
		Provenance                 *workflowstate.IntegrationProvenance `json:",omitempty"`
	}{proof.Method, proof.SourceCommit, proof.BaseCommit, proof.PreparedTree, proof.History, proof.Provenance})

	return contenthash.Sum(data), err
}

// ParentsMatch checks the configured merge or squash parent shape.
func ParentsMatch(proof workflowstate.IntegrationProof, parents []string) bool {
	switch proof.Method {
	case "merge":
		return len(parents) == 2 && parents[1] == proof.SourceCommit
	case "squash":
		return len(parents) == 1
	default:
		return false
	}
}

func historyTipMatches(proof workflowstate.IntegrationProof, last workflowstate.IntegrationCommit) bool {
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

// CheckProvenance checks bounded attribution, preparation metadata and the bound repair evidence.
func CheckProvenance(proof workflowstate.IntegrationProof, repairs RepairProof) error {
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
