package pi

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

type runtimeInspector struct{ output string }

func (r runtimeInspector) RunAgentCommand(_ context.Context, command harness.AgentCommand, input string, _ time.Time) (harness.AgentCommandResult, error) {
	if command.Program != "docker" || strings.Join(command.Args[:2], " ") != "image inspect" || input != "" {
		return harness.AgentCommandResult{}, errors.New("inspection must not build, pull or launch a model")
	}

	return harness.AgentCommandResult{Stdout: []byte(r.output)}, nil
}

// Metadata failures must reject an explicit runtime rather than using Codex or
// interpreting an arbitrary Docker option as an image reference. The executor
// double isolates metadata validation; the opt-in image test covers Docker.
func TestPiRuntimeReadiness(t *testing.T) {
	valid := `{"id":"sha256:` + strings.Repeat("a", 64) + `","os":"linux","architecture":"amd64","labels":{"org.chr.harness":"pi","org.chr.pi.version":"1.3.0","org.chr.pi.protocol":"pi-rpc/1.3.0","org.chr.node.version":"22.23.3"}}`
	for _, test := range []struct {
		name      string
		reference string
		metadata  string
		wantOK    bool
	}{
		{"prepared", "chr-agent-pi:1.3.0", valid, true},
		{"arm64", "chr-agent-pi:1.3.0", strings.Replace(valid, "amd64", "arm64", 1), true},
		{"missing", "", valid, false},
		{"option", "--help", valid, false},
		{"invalid output", "local", "{}", false},
		{"mutable identity", "local", strings.Replace(valid, "sha256:", "tag:", 1), false},
		{"unsupported OS", "local", strings.Replace(valid, "linux", "windows", 1), false},
		{"unsupported platform", "local", strings.Replace(valid, "amd64", "riscv64", 1), false},
		{"Codex image", "local", strings.Replace(valid, `"pi"`, `"codex"`, 1), false},
		{"wrong version", "local", strings.ReplaceAll(valid, "1.3.0", "1.2.0"), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			image, err := InspectRuntime(context.Background(), runtimeInspector{output: test.metadata}, test.reference)
			if (err == nil) != test.wantOK || (test.wantOK && !strings.HasPrefix(image.ID, "sha256:")) {
				t.Fatalf("image = %#v, %v", image, err)
			}
		})
	}
}
