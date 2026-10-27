package app

import (
	"strings"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestTelegramAuthoringBridgeParity reaches the same generation/refinement
// handlers through explicit controls and real HTTP-backed ordinary language.
func TestTelegramAuthoringBridgeParity(t *testing.T) {
	for _, natural := range []bool{false, true} {
		name := "explicit"
		if natural {
			name = "natural"
		}

		t.Run(name, func(t *testing.T) {
			f := newTelegramConversationAppFixture(t)

			var conversation atomic.Pointer[string]

			empty := ""
			conversation.Store(&empty)

			var interpretations atomic.Int64

			calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
				body := "## Requirements\n\nPreserve retries."
				if input.Action == "refine" {
					body += "\n\nClarify the existing validation."
				}

				return telegramAuthoringServiceOutput(t, body, "")
			}, func(input ConversationIntentContext) ConversationIntent {
				interpretations.Add(1)

				return telegramAuthoringParityIntent(input, *conversation.Load())
			})

			start, spec, refine, plan := "/prepare retry scope", "/spec", "/refine spec", "/plan"
			if natural {
				start, spec, refine, plan = "prepare retry scope", "generate the specification", "refine the specification validation", "generate the plan"
			}

			f.send(t, start, 0)

			selected := f.sessions(t)[0].ID
			conversation.Store(&selected)
			telegramAuthoringParityCommands(t, f, natural, selected, spec, refine, plan)

			snapshot := telegramAuthoringBridgeSnapshot(t, f, selected)
			if len(snapshot.Session.Outputs) != 2 || calls.Load() != 3 || snapshot.Question != nil {
				t.Fatalf("preparation parity: %+v calls=%d", snapshot, calls.Load())
			}

			telegramAuthoringParityAssertions(t, f, snapshot, natural, interpretations.Load())
		})
	}
}

func telegramAuthoringParityIntent(input ConversationIntentContext, conversation string) ConversationIntent {
	switch input.Message {
	case "prepare retry scope":
		return ConversationIntent{OperationID: "prepare", Arguments: ConversationIntentArguments{Text: "retry scope"}}
	case "generate the specification":
		return ConversationIntent{OperationID: "spec", Arguments: ConversationIntentArguments{Conversation: conversation}}
	case "generate the plan":
		return ConversationIntent{OperationID: "plan", Arguments: ConversationIntentArguments{Conversation: conversation}}
	default:
		return ConversationIntent{OperationID: "refine", Arguments: ConversationIntentArguments{Conversation: conversation, Action: "spec", Text: input.Message}}
	}
}

func telegramAuthoringParityCommands(t *testing.T, f *telegramConversationAppFixture, natural bool, conversation, spec, refine, plan string) {
	t.Helper()

	if !natural {
		spec += " " + conversation
		refine += " " + conversation + " clarify the existing validation"
		plan += " " + conversation
	}

	f.send(t, spec, 0)
	f.send(t, refine, 0)
	f.send(t, plan, 0)
}

func telegramAuthoringParityAssertions(t *testing.T, f *telegramConversationAppFixture, snapshot authoringcore.Snapshot, natural bool, interpretations int64) {
	t.Helper()

	for _, artifact := range snapshot.Artifacts {
		if artifact.Ref.Kind == authoringcore.ArtifactKindSpec && !strings.Contains(artifact.Body, "Clarify the existing validation") {
			t.Fatal("refinement did not reach the shared writer")
		}
	}

	expected := int64(0)
	if natural {
		expected = 4
	}

	if interpretations != expected {
		t.Fatal("explicit control interpretation parity", interpretations, expected)
	}

	telegramAuthoringBridgeNoWork(t, f)
}
