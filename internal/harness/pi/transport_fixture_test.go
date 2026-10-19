package pi

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

type fixtureConfig struct {
	Scenario string
}

// This child entrypoint supplies a deterministic RPC peer, never a model call.
func TestPiTransportDriver(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "--" {
		return
	}

	data, err := os.ReadFile(os.Args[len(os.Args)-1])
	if err != nil {
		t.Fatal(err)
	}

	var config fixtureConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}

	runFixture(t, config.Scenario)
	os.Exit(0)
}

func runFixture(t *testing.T, scenario string) {
	t.Helper()

	decoder := json.NewDecoder(os.Stdin)

	setup := make([]command, 3)
	for i := range setup {
		if err := decoder.Decode(&setup[i]); err != nil {
			t.Fatal(err)
		}
	}

	// Return setup responses out of order to exercise request identity ownership.
	for i := len(setup) - 1; i >= 0; i-- {
		fixtureResponse(t, scenario, setup[i])
	}

	for {
		var request command
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return
			}

			t.Fatal(err)
		}

		if request.Type == "prompt" {
			fixturePrompt(t, scenario, request)

			return
		}

		fixtureResponse(t, scenario, request)
	}
}

func fixtureResponse(t *testing.T, scenario string, request command) {
	t.Helper()

	response := map[string]any{"type": "response", "id": request.ID, "command": request.Type, "success": true}
	switch request.Type {
	case "set_model":
		response["data"] = map[string]any{"provider": "fixture", "id": "model"}
		if scenario == "wrong model" {
			response["data"] = map[string]any{"provider": "fixture", "id": "replacement"}
		}
	case "get_state":
		fixtureState(response, scenario)
	case "prompt":
		response["data"] = map[string]any{"disposition": "started"}
		if scenario == "handled" {
			response["data"] = map[string]any{"disposition": "handled"}
		}
	}

	fixtureResponseFault(response, scenario, request.Type)
	fixtureJSON(t, scenario, response)

	if scenario == "duplicate response" && request.Type == "set_model" {
		fixtureJSON(t, scenario, response)
	}

	if scenario == "backpressure" && request.Type == "get_state" {
		fixtureBackpressure(t)
	}
}

func fixtureState(response map[string]any, scenario string) {
	state := map[string]any{
		"model":         map[string]any{"provider": "fixture", "id": "model"},
		"thinkingLevel": "high", "isStreaming": false, "isCompacting": false,
		"autoCompactionEnabled": false, "messageCount": 0, "pendingMessageCount": 0,
	}
	if scenario == "thinking clamped" {
		state["thinkingLevel"] = "off"
	}

	if scenario == "occupied session" {
		state["messageCount"] = 1
	}

	response["data"] = state
}

func fixtureResponseFault(response map[string]any, scenario, kind string) {
	if kind != "set_model" {
		return
	}

	switch scenario {
	case "command rejected":
		response["success"], response["error"] = false, "Model not available"
	case "command diagnostic":
		response["success"], response["error"] = false, strings.Repeat("X", 262141)+"FIXTURE_PRIVATE_BOUNDARY_KEY"
	case "wrong correlation":
		response["id"] = "unowned"
	}
}

func fixturePrompt(t *testing.T, scenario string, request command) {
	t.Helper()

	if scenario != "late acknowledgment" {
		fixtureResponse(t, scenario, request)
	}

	if fixtureTerminalFault(t, scenario) {
		return
	}

	fixtureJSON(t, scenario, map[string]any{"type": "agent_start"})
	fixtureContinuation(t, scenario)
	fixtureInspection(t, scenario)

	message := fixtureMessage(scenario)
	fixtureJSON(t, scenario, map[string]any{"type": "message_end", "message": message})

	if scenario == "agent end only" {
		fixtureJSON(t, scenario, map[string]any{"type": "agent_end", "willRetry": false})

		return
	}

	fixtureJSON(t, scenario, map[string]any{"type": "agent_settled", "aborted": scenario == "aborted"})

	if scenario == "late acknowledgment" {
		fixtureResponse(t, scenario, request)
	}

	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		t.Fatal(err)
	}

	if scenario == "nonzero exit" {
		os.Exit(7)
	}

	if scenario == "shutdown stalled" {
		fixtureChild(t)
	}
}

func fixtureMessage(scenario string) map[string]any {
	message := map[string]any{
		"role": "assistant", "provider": "fixture", "model": "model", "stopReason": "stop",
		"content": []map[string]any{{"type": "thinking", "thinking": "not output"}, {"type": "text", "text": "{\"answer\":\"final\u2028line\u2029end\"}"}},
		"usage":   map[string]any{"input": 12, "output": 7},
	}

	switch scenario {
	case "provider failure":
		message["stopReason"] = "error"
	case "worker output":
		message["content"] = []map[string]any{{"type": "text", "text": strings.Repeat("x", 65_537)}}
	case "output overflow":
		message["content"] = []map[string]any{{"type": "text", "text": strings.Repeat("x", 257)}}
	case "unknown usage":
		delete(message, "usage")
	case "partial usage":
		message["usage"] = map[string]any{"input": 12}
	case "negative usage":
		message["usage"] = map[string]any{"input": -1, "output": 7}
	}

	return message
}

func fixtureContinuation(t *testing.T, scenario string) {
	t.Helper()

	if scenario != "continuation" && scenario != "continuation unknown usage" && scenario != "overflow usage" {
		return
	}

	message := fixtureMessage("")
	message["content"] = []map[string]any{{"type": "text", "text": `{"early":true}`}}

	message["usage"] = map[string]any{"input": 0, "output": 0}
	if scenario == "continuation unknown usage" {
		delete(message, "usage")
	}

	if scenario == "overflow usage" {
		message["usage"] = map[string]any{"input": int64(math.MaxInt64), "output": 0}
	}

	fixtureJSON(t, scenario, map[string]any{"type": "message_end", "message": message})
	fixtureJSON(t, scenario, map[string]any{"type": "agent_end", "willRetry": true})
	fixtureJSON(t, scenario, map[string]any{"type": "auto_retry_start", "attempt": 1})
	fixtureJSON(t, scenario, map[string]any{"type": "compaction_start", "reason": "overflow"})
	fixtureJSON(t, scenario, map[string]any{"type": "compaction_end", "aborted": false, "willRetry": true, "result": map[string]any{"usage": map[string]any{"input": 0, "output": 0}}})
	fixtureJSON(t, scenario, map[string]any{"type": "agent_start"})
	fixtureJSON(t, scenario, map[string]any{"type": "auto_retry_end", "success": true})
}

func fixtureJSON(t *testing.T, scenario string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	ending := "\n"

	if scenario == "fragmented" {
		data = []byte(strings.ReplaceAll(strings.ReplaceAll(string(data), `\u2028`, "\u2028"), `\u2029`, "\u2029"))
		ending = "\r\n"
	}

	data = append(data, ending...)
	for len(data) != 0 {
		size := len(data)
		if scenario == "fragmented" {
			size = min(size, 7)
		}

		if _, err := os.Stdout.Write(data[:size]); err != nil {
			t.Fatal(err)
		}

		data = data[size:]
	}
}

func fixtureBackpressure(t *testing.T) {
	t.Helper()

	for range 4096 {
		fixtureJSON(t, "", map[string]any{"type": "message_update", "assistantMessageEvent": map[string]any{"type": "text_delta", "delta": strings.Repeat("x", 256)}})
	}

	if _, err := io.WriteString(os.Stderr, strings.Repeat("diagnostic\n", 8192)); err != nil {
		t.Fatal(err)
	}
}

func fixtureTerminalFault(t *testing.T, scenario string) bool {
	t.Helper()

	switch scenario {
	case "acknowledgment only", "handled":
		return true
	case "malformed":
		fixtureRaw(t, os.Stdout, "not JSON\n")
	case "unterminated":
		fixtureRaw(t, os.Stdout, `{"type":"agent_start"}`)
	case "invalid UTF-8":
		fixtureRaw(t, os.Stdout, "{\"type\":\"message_update\",\"value\":\"\xff\"}\n")
	case "record overflow":
		fixtureRaw(t, os.Stdout, strings.Repeat("x", 1025))
	case "stdout overflow":
		fixtureBackpressure(t)
	case "stderr overflow":
		fixtureRaw(t, os.Stderr, strings.Repeat("x", 4097))
	default:
		return fixtureRuntimeFault(t, scenario)
	}

	return true
}

func fixtureRuntimeFault(t *testing.T, scenario string) bool {
	t.Helper()

	switch scenario {
	case "tool event":
		fixtureJSON(t, scenario, map[string]any{"type": "tool_execution_start", "toolName": "bash"})
	case "cancellation":
		fixtureChild(t)
	case "diagnostics":
		fixtureRaw(t, os.Stderr, "Authorization: Bearer fixture-private-token\n")

		return false
	default:
		return false
	}

	return true
}

func fixtureRaw(t *testing.T, writer io.Writer, data string) {
	t.Helper()

	if _, err := io.WriteString(writer, data); err != nil {
		t.Fatal(err)
	}
}

func fixtureChild(t *testing.T) {
	t.Helper()

	child := exec.Command("sleep", "30")

	child.Stdout, child.Stderr = os.Stdout, os.Stderr
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("child.pid", []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := child.Wait(); err != nil {
		t.Fatal(err)
	}
}
