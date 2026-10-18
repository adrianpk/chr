package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PlanOrderTask describes stable intended order; ticket references are assigned
// only by approved native materialization, never by generated Markdown.
type PlanOrderTask struct {
	ID        string `json:"id"`
	Order     int    `json:"task_order"`
	Title     string `json:"title"`
	TicketKey string `json:"ticket_key,omitempty"`
	TicketRef string `json:"ticket_ref,omitempty"`
}

// PlanOrderSlice groups task order within one stable integration unit.
type PlanOrderSlice struct {
	ID    string          `json:"id"`
	Order int             `json:"slice_order"`
	Title string          `json:"title"`
	Tasks []PlanOrderTask `json:"tasks"`
}

// PlanRevisionBinding identifies retained source and companion images. Hashing
// the plan body independently prevents circular plan/tracker dependencies.
type PlanRevisionBinding struct {
	Reference         string `json:"plan_ref"`
	Revision          uint64 `json:"plan_revision"`
	BodyHash          string `json:"plan_body_hash"`
	BodyRef           string `json:"body_ref"`
	OrderRef          string `json:"order_ref"`
	OrderHash         string `json:"order_hash"`
	TrackerRef        string `json:"tracker_ref"`
	TrackerHash       string `json:"tracker_hash"`
	CurrentTrackerRef string `json:"current_tracker_ref"`
}

// PlanOrderDocument is an order-only projection. Mutable processing state has
// no representation in either its structural hash or its rendered bytes.
type PlanOrderDocument struct {
	Schema                string              `json:"schema"`
	ProjectID             string              `json:"project_id"`
	Namespace             string              `json:"namespace"`
	PlanRef               string              `json:"plan_ref"`
	PlanRevision          uint64              `json:"plan_revision"`
	PlanBodyHash          string              `json:"plan_body_hash"`
	SourceObservationHash string              `json:"source_observation_hash,omitempty"`
	SourceInputs          []PlanSourceBinding `json:"source_inputs,omitempty"`
	SetID                 string              `json:"set_id,omitempty"`
	MaterializationID     string              `json:"materialization_id,omitempty"`
	Slices                []PlanOrderSlice    `json:"slices"`
}

// ValidatePlannedOrder requires bounded explicit identities and positive order,
// and prevents a model from supplying allocated ticket references.
func ValidatePlannedOrder(slices []PlanOrderSlice) error {
	if len(slices) == 0 || len(slices) > WorkflowPageLimit {
		return fmt.Errorf("plan requires 1..128 ordered slices")
	}

	ids, orders, tasks := map[string]bool{}, map[int]bool{}, map[string]bool{}
	for _, slice := range slices {
		if !workflowInputIDPattern.MatchString(slice.ID) || ids[slice.ID] || slice.Order < 1 || orders[slice.Order] || strings.TrimSpace(slice.Title) == "" {
			return fmt.Errorf("invalid plan slice identity, title or order")
		}

		ids[slice.ID], orders[slice.Order] = true, true
		if err := validatePlannedTasks(slice.Tasks, tasks); err != nil {
			return err
		}
	}

	return nil
}

func validatePlannedTasks(tasks []PlanOrderTask, ids map[string]bool) error {
	if len(tasks) == 0 || len(tasks) > WorkflowPageLimit {
		return fmt.Errorf("plan slice requires 1..128 tasks")
	}

	orders := map[int]bool{}
	for _, task := range tasks {
		if err := validatePlannedTaskIdentity(task, ids, orders); err != nil {
			return err
		}

		if task.TicketKey != "" || task.TicketRef != "" {
			return fmt.Errorf("generated order cannot allocate tickets")
		}

		ids[task.ID], orders[task.Order] = true, true
		if len(ids) > WorkflowRecordLimit {
			return fmt.Errorf("plan task capacity exceeded")
		}
	}

	return nil
}

func orderedPlanSlices(input []PlanOrderSlice) []PlanOrderSlice {
	result := append([]PlanOrderSlice(nil), input...)
	sort.Slice(result, func(i, j int) bool { return result[i].Order < result[j].Order })

	for i := range result {
		result[i].Tasks = append([]PlanOrderTask(nil), result[i].Tasks...)
		sort.Slice(result[i].Tasks, func(a, b int) bool { return result[i].Tasks[a].Order < result[i].Tasks[b].Order })
	}

	return result
}

func planOrderBytes(document PlanOrderDocument) ([]byte, []byte, error) {
	document.Slices = orderedPlanSlices(document.Slices)

	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, nil, err
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "---\nworkflow_projection: chr-tracker/v2\nstate_authority: tickets\nplan_ref: %s\nplan_revision: %d\nplan_body_hash: %s\nsource_hash: %s\n---\n\n# Workflow tracker\n\n", yamlString(document.PlanRef), document.PlanRevision, document.PlanBodyHash, workflowHash(encoded))

	if document.SetID == "" {
		output.WriteString("Planned order. Tickets are unbound; this map cannot authorize execution.\n\n")
	} else {
		fmt.Fprintf(&output, "Set: %s. Read live tickets for remaining work. This document supplies order only.\n\n", trackerCell(document.SetID))
	}

	output.WriteString("| Slice | Order | Title | Phase | Pull request | Delivery receipt |\n| --- | --- | --- | --- | --- | --- |\n")

	for _, slice := range document.Slices {
		fmt.Fprintf(&output, "| %s | %d | %s | | | |\n", trackerCell(slice.ID), slice.Order, trackerCell(slice.Title))
	}

	output.WriteString("\n| Ticket | Slice | Task | Order | Title | Reference | Status | Phase | Delivery receipt |\n| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, slice := range document.Slices {
		for _, task := range slice.Tasks {
			key := task.TicketKey
			if key == "" {
				key = "unbound"
			}

			fmt.Fprintf(&output, "| %s | %s | %s | %d | %s | %s | | | |\n", trackerCell(key), trackerCell(slice.ID), trackerCell(task.ID), task.Order, trackerCell(task.Title), trackerCell(task.TicketRef))
		}
	}

	if len(encoded) > WorkflowFileLimit || output.Len() > WorkflowFileLimit {
		return nil, nil, fmt.Errorf("order map exceeds 1 MiB")
	}

	return encoded, output.Bytes(), nil
}

func validatePlannedTaskIdentity(task PlanOrderTask, ids map[string]bool, orders map[int]bool) error {
	if !workflowInputIDPattern.MatchString(task.ID) || ids[task.ID] || task.Order < 1 || orders[task.Order] || strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf("invalid plan task identity, title or order")
	}

	return nil
}
