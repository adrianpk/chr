package app

import (
	"encoding/json"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

// Worker model context preserves its frozen identity while excluding private
// launcher references. Conflicting runtime identity and privileged roles fail.
func TestPiWorkerInput(t *testing.T) {
	for _, scenario := range []string{"valid", "legacy", "missing binding", "changed model", "changed runtime", "merger", "finite budget"} {
		t.Run(scenario, func(t *testing.T) {
			input, options := piWorkerInputFixture(t)

			input, options = alterPiWorkerFixture(scenario, input, options)

			data, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}

			prompt, err := piWorkerPrompt(data, options)
			if scenario != "valid" {
				if err == nil {
					t.Fatal("conflicting worker input accepted")
				}

				return
			}

			if err != nil || strings.Contains(prompt, "private-credential-reference") || !strings.Contains(prompt, input.Body) || !strings.Contains(prompt, input.HarnessBinding.BindingHash) {
				t.Fatalf("invalid private-context boundary: %v", err)
			}
		})
	}
}

func piWorkerInputFixture(t *testing.T) (core.WorkflowTaskInput, pi.WorkerRuntimeOptions) {
	t.Helper()

	digest := strings.Repeat("a", 64)

	binding, err := harnesscore.SealHarnessBinding(harnesscore.Binding{
		SchemaVersion: 1, Role: harnesscore.Worker, Harness: pi.Identity, Source: "project", SelectionRevision: digest,
		Profile: "worker", ProfileRevision: digest, ModelProvider: "fixture", Model: "model", RuntimeRef: "sha256:" + digest, ProtocolRevision: pi.ProtocolRevision,
		ExecutableRef: harnesscore.ExecutableRef{Hash: digest, CommandHash: digest}, Capabilities: harness.Capabilities{Worker: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	input := core.WorkflowTaskInput{Schema: "chr-task-input/v2", HarnessBinding: &binding, Body: "Frozen task body", Worker: []string{"private-credential-reference"}}
	options := pi.WorkerRuntimeOptions{Image: binding.RuntimeRef, Provider: binding.ModelProvider, Model: binding.Model}

	return input, options
}

func alterPiWorkerFixture(scenario string, input core.WorkflowTaskInput, options pi.WorkerRuntimeOptions) (core.WorkflowTaskInput, pi.WorkerRuntimeOptions) {
	switch scenario {
	case "legacy":
		input.Schema = "chr-task-input/v1"
	case "missing binding":
		input.HarnessBinding = nil
	case "changed model":
		options.Model = "replacement"
	case "changed runtime":
		options.Image = "sha256:" + strings.Repeat("b", 64)
	case "merger":
		input.Merger = &core.WorkflowMergerInput{}
	case "finite budget":
		input.AgentCall = &core.WorkflowAgentCall{}
	}

	return input, options
}
