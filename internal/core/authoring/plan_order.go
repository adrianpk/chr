package authoring

import (
	"bytes"
	"encoding/json"
	"fmt"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanOrderImages retains the exact order and tracker bytes hashed together.
type PlanOrderImages struct {
	Order   []byte
	Tracker []byte
}

// EncodePlanOrder renders the original ordered JSON and tracker images.
func EncodePlanOrder(document generationcore.PlanOrderDocument) (PlanOrderImages, error) {
	document.Slices = generationcore.OrderedPlanSlices(document.Slices)

	encoded, err := json.Marshal(document)
	if err != nil {
		return PlanOrderImages{}, err
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "---\nworkflow_projection: chr-tracker/v2\nstate_authority: tickets\nplan_ref: %s\nplan_revision: %d\nplan_body_hash: %s\nsource_hash: %s\n---\n\n# Workflow tracker\n\n", opsmetadata.Quote(document.PlanRef), document.PlanRevision, document.PlanBodyHash, contenthash.Sum(encoded))

	if document.SetID == "" {
		output.WriteString("Planned order. Tickets are unbound; this map cannot authorize execution.\n\n")
	} else {
		fmt.Fprintf(&output, "Set: %s. Read live tickets for remaining work. This document supplies order only.\n\n", opsmetadata.TableCell(document.SetID))
	}

	output.WriteString("| Slice | Order | Title | Phase | Pull request | Delivery receipt |\n| --- | --- | --- | --- | --- | --- |\n")

	for _, slice := range document.Slices {
		fmt.Fprintf(&output, "| %s | %d | %s | | | |\n", opsmetadata.TableCell(slice.ID), slice.Order, opsmetadata.TableCell(slice.Title))
	}

	output.WriteString("\n| Ticket | Slice | Task | Order | Title | Reference | Status | Phase | Delivery receipt |\n| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, slice := range document.Slices {
		for _, task := range slice.Tasks {
			key := task.TicketKey
			if key == "" {
				key = "unbound"
			}

			fmt.Fprintf(&output, "| %s | %s | %s | %d | %s | %s | | | |\n", opsmetadata.TableCell(key), opsmetadata.TableCell(slice.ID), opsmetadata.TableCell(task.ID), task.Order, opsmetadata.TableCell(task.Title), opsmetadata.TableCell(task.TicketRef))
		}
	}

	if len(encoded) > workflowstate.FileLimit || output.Len() > workflowstate.FileLimit {
		return PlanOrderImages{}, fmt.Errorf("order map exceeds 1 MiB")
	}

	return PlanOrderImages{Order: encoded, Tracker: output.Bytes()}, nil
}
