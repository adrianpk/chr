package app

import "forge.adrianpk.com/chr/internal/webui"

// The Web reader selects the same original documents as the other interfaces.
func guidedWebDocuments(observed GuidedUseCaseInspection) []webui.GuidedUseCaseArtifact {
	documents := []webui.GuidedUseCaseArtifact{}
	if observed.CommittedDump != "" {
		documents = append(documents, webui.GuidedUseCaseArtifact{Stage: "dump", Content: observed.CommittedDump})
	}

	for _, kind := range []string{"quiz", "specification", "plan", "tracker"} {
		for _, artifact := range observed.Artifacts {
			if guidedOriginalMatches(*observed.Session, artifact.Stage, artifact.Reference, kind) {
				documents = append(documents, webui.GuidedUseCaseArtifact{Stage: kind, Content: artifact.Content})

				break
			}
		}
	}

	return documents
}
