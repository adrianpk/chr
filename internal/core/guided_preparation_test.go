package core

import (
	"fmt"
	"strings"
	"testing"
	"time"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestGuidedUseCasePreparationContracts rejects generated drift from the approved hierarchy and source.
func TestGuidedUseCasePreparationContracts(t *testing.T) {
	for _, variant := range []string{"valid", "identity", "source", "scope", "order", "addition", "gate", "foreign dependency"} {
		t.Run(variant, func(t *testing.T) {
			input, generated := workflowPlanFixture(), workflowPlanFixture()

			switch variant {
			case "identity":
				generated.MaterializationID = "another-generation"
			case "source":
				generated.PlanHash = "another-source"
			case "scope":
				generated.UseCase = "another-case"
			case "order":
				generated.Slices[0].Tasks[0].Order = 9
			case "addition":
				task := generated.Slices[0].Tasks[0]
				task.ID, task.Order = "added-task", 4
				generated.Slices[0].Tasks = append(generated.Slices[0].Tasks, task)
			case "gate":
				generated.Slices[0].Tasks[0].Validation = workflowstate.Validation{}
			case "foreign dependency":
				generated.Slices[0].Tasks[0].DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: "older-unrelated-ticket"}}
			}

			err := ValidateGuidedContracts(input, generated)
			if (err == nil) != (variant == "valid") {
				t.Fatal("unexpected generated contract authority", err)
			}
		})
	}
}

// TestGuidedUseCasePreparationBound refuses a plan that cannot fit its exact bounded admission.
func TestGuidedUseCasePreparationBound(t *testing.T) {
	var order []generationcore.PlanOrderSlice

	for i := 1; i <= 2; i++ {
		slice := generationcore.PlanOrderSlice{ID: fmt.Sprintf("slice-%d", i), Order: i, Title: "Bounded integration"}
		for j := 1; j <= 65; j++ {
			slice.Tasks = append(slice.Tasks, generationcore.PlanOrderTask{ID: fmt.Sprintf("task-%d-%d", i, j), Order: j, Title: "Task"})
		}

		order = append(order, slice)
	}

	if err := generationcore.ValidatePlannedOrder(order); err != nil {
		t.Fatal("fixture is not a valid independent plan", err)
	}

	if err := ValidateGuidedPlannedOrder(order); err == nil {
		t.Fatal("oversized guided selection was accepted")
	}
}

// TestGuidedUseCasePreparationManualArchive preserves independent namespace plans during standalone archive.
func TestGuidedUseCasePreparationManualArchive(t *testing.T) {
	root := t.TempDir()
	independent := orderedPlanCandidate()

	independent.Body = "Independent default plan."
	if _, err := WriteGeneratedPlan(t.Context(), root, independent); err != nil {
		t.Fatal(err)
	}

	scoped := orderedPlanCandidate()

	scoped.Scope.Namespace, scoped.Body = "research", "Original research plan."
	if _, err := WriteGeneratedPlan(t.Context(), root, scoped); err != nil {
		t.Fatal(err)
	}

	archive, found, err := ArchiveGeneratedPlan(t.Context(), root, scoped.Scope, time.Now().UTC())
	if err != nil || !found {
		t.Fatal("scoped manual archive failed", err)
	}

	document, err := ReadPlanDocument(t.Context(), root, archive.RelativePath)
	if err != nil || !strings.Contains(document.Body, scoped.Body) {
		t.Fatal("scoped plan archived without its canonical envelope", err)
	}

	paths, err := GeneratedPlanLayout(root, independent.Scope)
	if err != nil {
		t.Fatal(err)
	}

	document, err = ReadPlanDocument(t.Context(), root, paths.RelativePath)
	if err != nil || !strings.Contains(document.Body, independent.Body) {
		t.Fatal("independent default plan changed", err)
	}
}
