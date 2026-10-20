package pi

import (
	"errors"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// File-backed credentials are sanitized before truncating rejected RPC command diagnostics.
func TestPiTransportCommandDiagnostic(t *testing.T) {
	command, request := fixtureCall(t, "command diagnostic")

	response, err := (Transport{Executor: process.Executor{Secrets: []string{"FIXTURE_PRIVATE_BOUNDARY_KEY"}}}).Run(t.Context(), command, request)
	if err == nil || len(response.Output) != 0 {
		t.Fatal("rejected setup produced output")
	}

	for _, text := range []string{err.Error(), string(response.Process.Stdout), string(response.Process.Stderr)} {
		if strings.Contains(text, "FIX") {
			t.Fatal("RPC diagnostic excerpt retained a private credential fragment")
		}
	}
}

// Only correlated, host-enabled inspection may run before a later final answer.
func TestPiTransportInspection(t *testing.T) {
	for _, scenario := range []struct {
		name string
		want error
	}{
		{"inspection allowed", nil},
		{"inspection unknown usage", nil},
		{"inspection denied", ErrCapability},
		{"inspection unmatched", ErrProtocol},
		{"inspection unfinished", ErrProtocol},
		{"inspection duplicate", ErrProtocol},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			command, request := fixtureCall(t, scenario.name)
			request.ReadOnlyFiles = true

			response, err := (Transport{}).Run(t.Context(), command, request)
			if !errors.Is(err, scenario.want) || err != nil && len(response.Output) != 0 {
				t.Fatal("inspection lifecycle admitted invalid output", err)
			}

			if err == nil && !strings.Contains(string(response.Output), `"final`) {
				t.Fatal("tool output became the final result")
			}

			if scenario.name == "inspection unknown usage" && (response.Usage.Input != nil || response.Usage.Output != nil) {
				t.Fatal("unmeasured inspection continuation fabricated usage")
			}
		})
	}
}

func fixtureInspection(t *testing.T, scenario string) {
	t.Helper()

	if !strings.HasPrefix(scenario, "inspection ") {
		return
	}

	name := "chr_inspect"
	if scenario == "inspection denied" {
		name = "bash"
	} else if strings.HasPrefix(scenario, "inspection worker") {
		name = "write"
	}

	message := fixtureMessage("")
	message["stopReason"] = "toolUse"

	message["content"] = []map[string]any{{"type": "toolCall", "id": "inspect", "name": name, "arguments": map[string]string{"action": "read", "path": "decision.md"}}}
	if scenario == "inspection unknown usage" {
		delete(message, "usage")
	}

	fixtureJSON(t, scenario, map[string]any{"type": "message_end", "message": message})

	id := "inspect"
	if scenario == "inspection unmatched" {
		id = "foreign"
	}

	eventName := name
	if scenario == "inspection worker mismatch" {
		eventName = "bash"
	}

	fixtureJSON(t, scenario, map[string]any{"type": "tool_execution_start", "toolCallId": id, "toolName": eventName})

	if scenario == "inspection unfinished" {
		return
	}

	end := map[string]any{"type": "tool_execution_end", "toolCallId": id, "toolName": name, "result": map[string]any{"content": []map[string]string{{"type": "text", "text": "unapproved tool output"}}}, "isError": false}
	fixtureJSON(t, scenario, end)

	if scenario == "inspection duplicate" {
		fixtureJSON(t, scenario, end)
	}
}

// The worker profile admits only its explicit builtin tools and correlates both
// the call ID and tool name. The same subprocess is denied by generation.
func TestPiTransportWorker(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		worker   bool
		readOnly bool
		want     error
	}{
		{"inspection worker allowed", true, false, nil},
		{"inspection worker mismatch", true, false, ErrProtocol},
		{"inspection worker generation", false, false, ErrCapability},
		{"inspection worker read-only", false, true, ErrCapability},
		{"inspection worker ambiguous", true, true, ErrCapability},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			command, request := fixtureCall(t, scenario.name)
			request.WorkerTools, request.ReadOnlyFiles = scenario.worker, scenario.readOnly

			result, err := (Transport{}).Run(t.Context(), command, request)
			if !errors.Is(err, scenario.want) || (err != nil && len(result.Output) != 0) {
				t.Fatalf("tool authority: output=%q, error=%v", result.Output, err)
			}

			if err == nil && !strings.Contains(string(result.Output), `"final`) {
				t.Fatal("unapproved tool output became final")
			}
		})
	}
}

// The worker's larger normalized-output ceiling must not relax the existing
// structured-generation ceiling, even when a caller requests that larger limit.
func TestPiTransportWorkerOutput(t *testing.T) {
	for _, test := range []struct {
		name   string
		worker bool
		want   error
	}{
		{"worker", true, nil},
		{"generation", false, process.ErrLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			command, request := fixtureCall(t, "worker output")
			request.WorkerTools, request.OutputLimit = test.worker, harness.MaxRecord

			response, err := (Transport{}).Run(t.Context(), command, request)
			if !errors.Is(err, test.want) || (err == nil && len(response.Output) != 65_537) || (err != nil && len(response.Output) != 0) {
				t.Fatalf("worker output boundary: bytes=%d, error=%v", len(response.Output), err)
			}
		})
	}
}
