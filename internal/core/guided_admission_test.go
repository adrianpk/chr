package core

import (
	"errors"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCaseAdmissionMembership refuses replaced, missing, foreign or additional native work.
func TestGuidedUseCaseAdmissionMembership(t *testing.T) {
	member := workflowstate.PreviewMember{Key: "PRJ-1", Path: "ticket/one/metadata.json", Revision: 1, SourceHash: strings.Repeat("a", 64), ContractHash: strings.Repeat("b", 64), SetID: "set-one"}
	reference := authoringcore.GuidedMaterializationReference{SetID: member.SetID, Members: []authoringcore.GuidedWorkMember{{Key: member.Key, Path: member.Path, Revision: member.Revision, SourceHash: member.SourceHash, ContractHash: member.ContractHash}}}

	for _, variant := range []string{"exact", "revision", "source", "contract", "path", "set", "missing", "additional"} {
		t.Run(variant, func(t *testing.T) {
			preview := workflowcore.Preview{Selector: workflowstate.Selector{Extent: "slice_set", Identity: reference.SetID}, Members: []workflowstate.PreviewMember{member}, Total: 1}
			changeGuidedPreview(&preview, variant)

			err := workflowcore.ValidateGuidedWorkPreview(reference, preview)
			if variant == "exact" && err != nil || variant != "exact" && !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("unexpected membership authority", err)
			}
		})
	}
}

func changeGuidedPreview(preview *workflowcore.Preview, variant string) {
	switch variant {
	case "revision":
		preview.Members[0].Revision++
	case "source":
		preview.Members[0].SourceHash = strings.Repeat("c", 64)
	case "contract":
		preview.Members[0].ContractHash = strings.Repeat("c", 64)
	case "path":
		preview.Members[0].Path = "ticket/foreign/metadata.json"
	case "set":
		preview.Members[0].SetID = "foreign"
	case "missing":
		preview.Members = nil
	case "additional":
		preview.Total++
		preview.Members = append(preview.Members, preview.Members[0])
	}
}
