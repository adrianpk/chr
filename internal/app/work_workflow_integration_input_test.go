package app

import (
	"bytes"
	"maps"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowIntegrationPayload preserves explicit replay ownership and finding
// data at the wire boundary; only the domain owner interprets those assertions.
func TestWorkflowIntegrationPayload(t *testing.T) {
	fields := map[string]string{"entry_id": "entry", "entry_revision": "4", "budget_revision": "3", "candidate_hash": strings.Repeat("a", 64), "decision_revision": "2", "decision_hash": strings.Repeat("b", 64), "record_id": "answer", "decision_id": "replay-owner", "answer_action": "resolve", "actor": "operator", "authority": "maintainer", "reason": "Identify the explicit replay ownership"}
	for _, tc := range []struct {
		name, choice string
		payload      map[string]string
		original     string
		fix          *workflowstate.ReviewFinding
	}{
		{"source step", "Identify the original source step", map[string]string{"original_commit": strings.Repeat("c", 40)}, strings.Repeat("c", 40), nil},
		{"integration fix", "Declare an integration fix with an actionable finding", map[string]string{"fix_category": "correctness", "fix_title": "Preserve declared behavior", "fix_detail": "Retain both source and target requirements"}, "", &workflowstate.ReviewFinding{Category: "correctness", Title: "Preserve declared behavior", Detail: "Retain both source and target requirements"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := maps.Clone(fields)
			maps.Copy(input, tc.payload)

			input["choice"] = tc.choice

			request, err := workflowIntegrationControlFields("answer-decision", input)
			checkWorkflowIntegrationPayloadFence(t, request, fields, err)

			answer := request.DecisionAnswer
			if answer.OriginalCommit != tc.original || answer.Choice != tc.choice || answer.Actor != "operator" || answer.Authority != "maintainer" {
				t.Fatal("transport replaced an explicit ownership assertion")
			}

			checkWorkflowIntegrationFinding(t, answer.IntegrationFix, tc.fix)
		})
	}
}

func checkWorkflowIntegrationPayloadFence(t *testing.T, request WorkflowControlRequest, fields map[string]string, err error) {
	t.Helper()

	if err != nil || request.Integration == nil || request.Action != "answer_decision" || request.DecisionRevision != 2 || request.DecisionHash != fields["decision_hash"] || request.Integration.CandidateHash != fields["candidate_hash"] {
		t.Fatal("transport replaced a displayed decision fence", err)
	}
}

func checkWorkflowIntegrationFinding(t *testing.T, actual, expected *workflowstate.ReviewFinding) {
	t.Helper()

	if expected == nil {
		if actual != nil {
			t.Fatal("transport invented a finding")
		}

		return
	}

	if actual == nil || *actual != *expected {
		t.Fatal("transport lost the explicit integration finding")
	}
}

// TestWorkflowIntegrationInput rejects unbound, overflowing and malformed
// operator commands before resolving a project or changing retained authority.
func TestWorkflowIntegrationInput(t *testing.T) {
	guard := []string{"--revision", "1", "--entry-id", "entry", "--entry-revision", "1", "--budget-revision", "1", "--candidate-hash", strings.Repeat("a", 64)}
	body := []string{"--record-id", "extension", "--actor", "operator", "--reason", "Authorize a bounded extension", "--kind", "submission"}

	for _, tc := range []struct {
		name, expected string
		args           []string
	}{
		{"unbound", "displayed invocation revision", []string{"grant", "execution"}},
		{"negative amount", "positive bounded allowance", append(append(append([]string{"grant"}, guard...), body...), "--amount", "-1", "execution")},
		{"overflow", "invalid value", []string{"grant", "--revision", "18446744073709551616", "execution"}},
		{"missing question", "displayed revisions", append(append([]string{"answer-decision"}, guard...), "execution")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"work", tc.args[0], "--dir", t.TempDir()}, tc.args[1:]...)

			err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), tc.expected) {
				t.Fatal("malformed operator input reached project resolution", err)
			}
		})
	}
}
