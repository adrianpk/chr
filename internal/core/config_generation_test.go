package core

import (
	"context"
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Preparation cannot replace the native configuration fence with its own
// approval or select a profile/revision that the actual locked image did not admit.
func TestHarnessGenerationConfigFence(t *testing.T) {
	for _, field := range []string{"revision", "profile"} {
		t.Run(field, func(t *testing.T) {
			project := t.TempDir()
			program := generationBindingProgram(t, project)
			writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: alpha\n")

			_, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: "configuration-fence", Purpose: "plan_generation"}, func(ctx context.Context, config WorkspaceConfig) (harnesshost.GenerationPreparation, error) {
				prepared, err := generationBindingResolver(program, "model-alpha")(ctx, config)
				if err != nil {
					return prepared, err
				}

				if field == "revision" {
					prepared.Binding.Selection.SelectionRevision = contenthash.Sum([]byte("other image"))
				} else {
					prepared.Binding.Selection.Profile = "other"
				}

				prepared.Binding, err = harnesscore.SealGenerationBinding(prepared.Binding)
				prepared.Admit = func(harnesscore.Selection) error { return nil }

				return prepared, err
			})
			if !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("preparation bypassed the actual configuration image", err)
			}
		})
	}
}
