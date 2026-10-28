package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (reader workflowIntegrationReader) commitAttribution(ctx context.Context, commit string) (workflowstate.IntegrationAttribution, error) {
	data, err := reader.RawOutput(ctx, "cat-file", "commit", commit)
	if err != nil {
		return workflowstate.IntegrationAttribution{}, err
	}

	return workflowintegrate.WorkflowCommitAttribution(commit, data)
}

func (reader workflowIntegrationReader) sourceAttributions(ctx context.Context, proof workflowstate.IntegrationProof) ([]workflowstate.IntegrationAttribution, error) {
	line, err := reader.Output(ctx, "rev-list", "--reverse", "--topo-order", "--max-count="+strconv.Itoa(workflowstate.RecordLimit+1), proof.OriginCommit+".."+proof.SourceCommit)

	commits := strings.Fields(line)
	if err != nil || len(commits) == 0 || len(commits) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("integration source manifest is empty or exceeds commit bounds")
	}

	entries := make([]workflowstate.IntegrationAttribution, 0, len(commits))
	size := 0

	for _, commit := range commits {
		entry, err := reader.commitAttribution(ctx, commit)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(entry)
		if err != nil || len(data) >= workflowstate.FileLimit-size {
			return nil, fmt.Errorf("integration source attribution exceeds file bounds")
		}

		size += len(data)

		entries = append(entries, entry)
	}

	return entries, nil
}
