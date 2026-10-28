package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowTrackerImageResult struct {
	Status TrackerStatus
	Output []byte
}
type workflowTrackerMembership struct {
	Set   workflowstate.Record
	Tasks []workflowstate.Record
}

// workflowTrackerObserver assembles read-only tracker images from exact Store-owned snapshot and receipt evidence.
type workflowTrackerObserver struct{ store *Store }

func (host workflowTrackerObserver) image(snapshot workflowstate.Snapshot, setID string) (workflowTrackerImageResult, error) {
	trackerMembers, err := workflowTrackerMembers(snapshot, setID)
	if err != nil {
		return workflowTrackerImageResult{Status: TrackerStatus{}, Output: nil}, err
	}

	result := TrackerStatus{Schema: "chr-tracker/v1", ProjectID: host.store.projectID, Namespace: host.store.layout.Namespace(), SetID: setID, Path: "tracker/" + setID + ".md"}
	for _, record := range append([]workflowstate.Record{trackerMembers.Set}, trackerMembers.Tasks...) {
		result.Sources = append(result.Sources, TrackerSource{ID: record.ID, Path: record.Path, Revision: record.Revision, Hash: record.SourceHash})
	}

	if err := (workflowTrackerObserver{store: host.store}).receiptSources(&result, trackerMembers.Set, trackerMembers.Tasks); err != nil {
		return workflowTrackerImageResult{Status: result, Output: nil}, err
	}

	sort.Slice(result.Sources, func(i, j int) bool { return result.Sources[i].Path < result.Sources[j].Path })

	data, err := json.Marshal(result.Sources)
	if err != nil {
		return workflowTrackerImageResult{Status: result, Output: nil}, err
	}

	result.SourceHash = contenthash.Sum(data)

	output := renderWorkflowTracker(result, trackerMembers.Set, trackerMembers.Tasks)
	if len(output) > workflowstate.FileLimit {
		return workflowTrackerImageResult{Status: result, Output: nil}, fmt.Errorf("tracker output exceeds 1 MiB")
	}

	return workflowTrackerImageResult{Status: result, Output: output}, nil
}

func workflowTrackerMembers(snapshot workflowstate.Snapshot, setID string) (workflowTrackerMembership, error) {
	var (
		set   workflowstate.Record
		tasks []workflowstate.Record
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
		return workflowTrackerMembership{Set: set, Tasks: nil}, fmt.Errorf("adopted workflow set not found: %s", setID)
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

	return workflowTrackerMembership{Set: set, Tasks: tasks}, nil
}

func (host workflowTrackerObserver) receiptSources(result *TrackerStatus, set workflowstate.Record, tasks []workflowstate.Record) error {
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
		data, err := host.store.readWorkflowTrackerFile(ref)
		if err != nil {
			return err
		}

		result.Sources = append(result.Sources, TrackerSource{Path: ref, Hash: contenthash.Sum(data)})
	}

	return nil
}

//nolint:wsl_v5 // Adjacent writes form one generated Markdown document.
func renderWorkflowTracker(result TrackerStatus, set workflowstate.Record, tasks []workflowstate.Record) []byte {
	var output bytes.Buffer
	output.WriteString("---\nworkflow_projection: chr-tracker/v1\nstate_authority: tickets\n")
	fmt.Fprintf(&output, "slice_set_id: %s\nsource_hash: %s\n---\n\n", opsmetadata.Quote(set.ID), result.SourceHash)
	output.WriteString("# Workflow tracker\n\nGenerated view. Ticket records and receipts own progress. Do not execute or import this file.\n\n")
	fmt.Fprintf(&output, "Set: %s (%s). Closure receipt: %s.\n\n", opsmetadata.TableCell(set.ID), opsmetadata.TableCell(set.Set.Status), opsmetadata.TableCell(set.Set.ClosureRef))
	output.WriteString("| Slice | Phase | Pull request | Delivery receipt |\n| --- | --- | --- | --- |\n")
	ordered := append([]workflowstate.WorkflowSlice(nil), set.Set.Slices...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })
	for _, slice := range ordered {
		fmt.Fprintf(&output, "| %s | %s | %s | %s |\n", opsmetadata.TableCell(slice.ID), opsmetadata.TableCell(slice.Phase), opsmetadata.TableCell(slice.PullRequest), opsmetadata.TableCell(slice.DeliveryRef))
	}
	output.WriteString("\n| Ticket | Slice | Status | Phase | Delivery receipt |\n| --- | --- | --- | --- | --- |\n")
	for _, task := range tasks {
		fmt.Fprintf(&output, "| %s | %s | %s | %s | %s |\n", opsmetadata.TableCell(task.Task.Key), opsmetadata.TableCell(task.Task.SliceID), opsmetadata.TableCell(task.Task.Status), opsmetadata.TableCell(task.Task.Phase), opsmetadata.TableCell(task.Task.DeliveryRef))
	}

	return output.Bytes()
}
