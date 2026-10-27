package core

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// Preserve the complete configuration revision while parsing strict role fields.
func TestHarnessSelectionConfiguration(t *testing.T) {
	for _, test := range []struct {
		name, body, worker, generation string
		invalid                        bool
	}{
		{"legacy", "custom:\n  retained: true\n", "", "", false},
		{"both roles", "harnesses:\n  worker: coding\n  generation: writing\n", "coding", "writing", false},
		{"clear", "harnesses:\n  worker: null\n  generation: writing\n", "", "writing", false},
		{"duplicate", "harnesses:\n  worker: a\n  worker: b\n", "", "", true},
		{"foreign field", "harnesses:\n  executable: private-path\n", "", "", true},
		{"bad reference", "harnesses:\n  generation: ../private-path\n", "", "", true},
		{"bad indentation", "harnesses:\n generation: writing\n", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, err := parseWorkflowConfig([]byte(test.body))
			if (err != nil) != test.invalid {
				t.Fatalf("configuration error = %v", err)
			}

			if !test.invalid && (config.Harnesses.Worker != test.worker || config.Harnesses.Generation != test.generation || config.Revision() != contenthash.Sum([]byte(test.body))) {
				t.Fatalf("configuration = %#v", config.Harnesses)
			}
		})
	}
}
