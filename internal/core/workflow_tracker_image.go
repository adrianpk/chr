package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

func (s *WorkflowStore) workflowTrackerImage(snapshot WorkflowSnapshot, setID string) (WorkflowTrackerStatus, []byte, error) {
	set, tasks, err := workflowTrackerMembers(snapshot, setID)
	if err != nil {
		return WorkflowTrackerStatus{}, nil, err
	}

	result := WorkflowTrackerStatus{Schema: "chr-tracker/v1", ProjectID: s.projectID, Namespace: s.layout.Namespace(), SetID: setID, Path: "tracker/" + setID + ".md"}
	for _, record := range append([]WorkflowRecord{set}, tasks...) {
		result.Sources = append(result.Sources, WorkflowTrackerSource{ID: record.ID, Path: record.Path, Revision: record.Revision, Hash: record.SourceHash})
	}

	if err := s.workflowTrackerReceiptSources(&result, set, tasks); err != nil {
		return result, nil, err
	}

	sort.Slice(result.Sources, func(i, j int) bool { return result.Sources[i].Path < result.Sources[j].Path })

	data, err := json.Marshal(result.Sources)
	if err != nil {
		return result, nil, err
	}

	result.SourceHash = contenthash.Sum(data)

	output := renderWorkflowTracker(result, set, tasks)
	if len(output) > WorkflowFileLimit {
		return result, nil, fmt.Errorf("tracker output exceeds 1 MiB")
	}

	return result, output, nil
}

func workflowTrackerMembers(snapshot WorkflowSnapshot, setID string) (WorkflowRecord, []WorkflowRecord, error) {
	var (
		set   WorkflowRecord
		tasks []WorkflowRecord
	)

	for _, record := range snapshot.Records {
		if record.Set != nil && record.ID == setID {
			set = record
		}

		if record.Task != nil && record.Task.SetID == setID {
			tasks = append(tasks, record)
		}
	}

	if set.ID == "" {
		return set, nil, fmt.Errorf("adopted workflow set not found: %s", setID)
	}

	orders := map[string]int{}
	for _, slice := range set.Set.Slices {
		orders[slice.ID] = slice.Order
	}

	sort.Slice(tasks, func(i, j int) bool {
		if orders[tasks[i].Task.SliceID] != orders[tasks[j].Task.SliceID] {
			return orders[tasks[i].Task.SliceID] < orders[tasks[j].Task.SliceID]
		}

		return tasks[i].Task.Order < tasks[j].Task.Order
	})

	return set, tasks, nil
}

func (s *WorkflowStore) workflowTrackerReceiptSources(result *WorkflowTrackerStatus, set WorkflowRecord, tasks []WorkflowRecord) error {
	refs := map[string]bool{}
	for _, task := range tasks {
		refs[task.Task.DeliveryRef] = true
	}

	for _, slice := range set.Set.Slices {
		refs[slice.DeliveryRef] = true
	}

	refs[set.Set.ClosureRef] = true
	for _, ref := range set.Set.ClosureHistory {
		refs[ref] = true
	}

	delete(refs, "")

	for ref := range refs {
		data, err := s.readWorkflowTrackerFile(ref)
		if err != nil {
			return err
		}

		result.Sources = append(result.Sources, WorkflowTrackerSource{Path: ref, Hash: contenthash.Sum(data)})
	}

	return nil
}

//nolint:wsl_v5 // Adjacent writes form one generated Markdown document.
func renderWorkflowTracker(result WorkflowTrackerStatus, set WorkflowRecord, tasks []WorkflowRecord) []byte {
	var output bytes.Buffer
	output.WriteString("---\nworkflow_projection: chr-tracker/v1\nstate_authority: tickets\n")
	fmt.Fprintf(&output, "slice_set_id: %s\nsource_hash: %s\n---\n\n", yamlString(set.ID), result.SourceHash)
	output.WriteString("# Workflow tracker\n\nGenerated view. Ticket records and receipts own progress. Do not execute or import this file.\n\n")
	fmt.Fprintf(&output, "Set: %s (%s). Closure receipt: %s.\n\n", trackerCell(set.ID), trackerCell(set.Set.Status), trackerCell(set.Set.ClosureRef))
	output.WriteString("| Slice | Phase | Pull request | Delivery receipt |\n| --- | --- | --- | --- |\n")
	ordered := append([]WorkflowSlice(nil), set.Set.Slices...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })
	for _, slice := range ordered {
		fmt.Fprintf(&output, "| %s | %s | %s | %s |\n", trackerCell(slice.ID), trackerCell(slice.Phase), trackerCell(slice.PullRequest), trackerCell(slice.DeliveryRef))
	}
	output.WriteString("\n| Ticket | Slice | Status | Phase | Delivery receipt |\n| --- | --- | --- | --- | --- |\n")
	for _, task := range tasks {
		fmt.Fprintf(&output, "| %s | %s | %s | %s | %s |\n", trackerCell(task.Task.Key), trackerCell(task.Task.SliceID), trackerCell(task.Task.Status), trackerCell(task.Task.Phase), trackerCell(task.Task.DeliveryRef))
	}

	return output.Bytes()
}

func trackerCell(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "|", "&#124;", "\n", " ", "\r", " ").Replace(value)
}
