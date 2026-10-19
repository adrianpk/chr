package pi

import (
	"errors"
	"strings"
	"testing"

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

	fixtureJSON(t, scenario, map[string]any{"type": "tool_execution_start", "toolCallId": id, "toolName": name})

	if scenario == "inspection unfinished" {
		return
	}

	end := map[string]any{"type": "tool_execution_end", "toolCallId": id, "toolName": name, "result": map[string]any{"content": []map[string]string{{"type": "text", "text": "unapproved tool output"}}}, "isError": false}
	fixtureJSON(t, scenario, end)

	if scenario == "inspection duplicate" {
		fixtureJSON(t, scenario, end)
	}
}
