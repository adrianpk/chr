package core

import (
	"bytes"
	"encoding/json"
	"fmt"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// planOrderImages retains the exact order and tracker bytes hashed together.
type planOrderImages struct {
	Order   []byte
	Tracker []byte
}

func planOrderBytes(document generationcore.PlanOrderDocument) (planOrderImages, error) {
	document.Slices = generationcore.OrderedPlanSlices(document.Slices)

	encoded, err := json.Marshal(document)
	if err != nil {
		return planOrderImages{}, err
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "---\nworkflow_projection: chr-tracker/v2\nstate_authority: tickets\nplan_ref: %s\nplan_revision: %d\nplan_body_hash: %s\nsource_hash: %s\n---\n\n# Workflow tracker\n\n", opsmetadata.Quote(document.PlanRef), document.PlanRevision, document.PlanBodyHash, contenthash.Sum(encoded))

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
		return planOrderImages{}, fmt.Errorf("order map exceeds 1 MiB")
	}

	return planOrderImages{Order: encoded, Tracker: output.Bytes()}, nil
}
