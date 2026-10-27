package core

import (
	"bytes"
	"encoding/json"
	"fmt"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func planOrderBytes(document generationcore.PlanOrderDocument) ([]byte, []byte, error) {
	document.Slices = generationcore.OrderedPlanSlices(document.Slices)

	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, nil, err
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "---\nworkflow_projection: chr-tracker/v2\nstate_authority: tickets\nplan_ref: %s\nplan_revision: %d\nplan_body_hash: %s\nsource_hash: %s\n---\n\n# Workflow tracker\n\n", yamlString(document.PlanRef), document.PlanRevision, document.PlanBodyHash, contenthash.Sum(encoded))

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

	if len(encoded) > workflowstate.FileLimit || output.Len() > workflowstate.FileLimit {
		return nil, nil, fmt.Errorf("order map exceeds 1 MiB")
	}

	return encoded, output.Bytes(), nil
}
