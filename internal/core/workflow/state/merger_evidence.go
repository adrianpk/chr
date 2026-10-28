package state

// EvidenceRefs returns the original ordered evidence references of this retained value.
func (input MergerInput) EvidenceRefs() []EvidenceRef {
	refs := []EvidenceRef{input.SourceManifestRef, input.ValidationContractRef}
	if input.OperatorDecisionRef != nil {
		refs = append(refs, *input.OperatorDecisionRef)
	}

	if input.RepairCheckpointRef != nil {
		refs = append(refs, *input.RepairCheckpointRef)
	}

	refs = append(refs, input.IntentRefs...)
	refs = append(refs, input.TargetChangeRefs...)
	refs = append(refs, input.FindingRefs...)

	for _, requirement := range input.Requirements {
		refs = append(refs, requirement.Evidence)
	}

	for _, conflict := range input.Conflicts {
		refs = append(refs, conflict.Evidence)
	}

	if input.HistoryCheckpoint != nil {
		refs = append(refs, input.HistoryCheckpoint.Evidence)
	}

	return refs
}
