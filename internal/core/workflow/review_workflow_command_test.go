package workflow

import (
	"context"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowPRCommand bounds real provider streams without retaining excess
// output and consumes noisy stderr without corrupting the structured response.
func TestWorkflowPRCommand(t *testing.T) {
	cases := []struct {
		name, script string
		size         int
		fails        bool
	}{
		{"stdout limit", "head -c 1048576 /dev/zero", workflowstate.FileLimit, false},
		{"stdout overflow", "head -c 1048577 /dev/zero", 0, true},
		{"stderr overflow", "head -c 1048577 /dev/zero >&2; printf ok", 2, false},
		{"command failure", "exit 7", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := (workflowProviderCommandRunner{}).Output(t.Context(), "sh", "-c", tc.script)
			if (err != nil) != tc.fails || len(data) != tc.size {
				t.Fatalf("provider output: %d bytes, %v", len(data), err)
			}
		})
	}
}

// TestWorkflowPRCommandDeadline respects a caller's shorter deadline and kills
// only the owned direct CLI process, without leaving a shell child behind.
func TestWorkflowPRCommandDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	if _, err := (workflowProviderCommandRunner{}).Output(ctx, "sh", "-c", "exec sleep 5"); err == nil {
		t.Fatal("provider ignored its context deadline")
	}
}
