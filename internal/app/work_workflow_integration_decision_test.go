package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

func workflowIntegrationAnswerRequest(shown WorkflowWorkStatus, action string) WorkflowControlRequest {
	decision := shown.Integration.Decision
	question := decision.Requests[0]
	request := serviceIntegrationControl(shown, "answer_decision")
	request.DecisionRevision, request.DecisionHash = decision.Revision, decision.Hash
	request.DecisionAnswer = &core.WorkflowIntegrationAnswer{ID: "surface-answer", DecisionID: question.ID, Choice: question.Options[0], Action: action, Actor: "operator", Authority: question.Authority, Reason: "Preserve both requirements through the compatible interpretation"}

	return request
}

func workflowIntegrationCLIAnswer(t *testing.T, operations service, shown WorkflowWorkStatus, daemon bool) (WorkflowWorkStatus, error) {
	t.Helper()

	request := workflowIntegrationAnswerRequest(shown, "resolve")
	answer := request.DecisionAnswer
	args := append(workflowIntegrationCLIArguments(shown), "--decision-revision", strconv.FormatUint(request.DecisionRevision, 10), "--decision-hash", request.DecisionHash, "--record-id", answer.ID, "--decision-id", answer.DecisionID, "--choice", answer.Choice, "--answer-action", answer.Action, "--actor", answer.Actor, "--authority", answer.Authority, "--reason", answer.Reason, shown.ID)

	if !daemon {
		return workflowCLIInjectedControl(t, operations, "answer-decision", args...)
	}

	var output bytes.Buffer

	command := (commandFactory{deps: testAppDependencies(), io: IO{Stdout: &output}}).daemonWorkflowCommand()
	for _, child := range command.Commands {
		if child.Name == "answer-decision" {
			child.Action = func(ctx context.Context, cmd *cli.Command) error {
				return (workWorkflowCLI{stdout: &output}).control(ctx, cmd, operations, "answer-decision")
			}
		}
	}

	err := command.Run(t.Context(), append([]string{"test", "answer-decision", "--json"}, args...))

	var status WorkflowWorkStatus
	if output.Len() != 0 {
		err = errors.Join(err, json.Unmarshal(output.Bytes(), &status))
	}

	return status, err
}

// TestWorkflowIntegrationAnswer routes real sealed questions through work,
// daemon and manual callers, preserving budgets and rejecting stale replay.
func TestWorkflowIntegrationAnswer(t *testing.T) {
	cases := []struct {
		name   string
		submit func(*testing.T, service, WorkflowWorkStatus) (WorkflowWorkStatus, error)
	}{
		{"work CLI", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			return workflowIntegrationCLIAnswer(t, s, shown, false)
		}},
		{"daemon CLI", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			return workflowIntegrationCLIAnswer(t, s, shown, true)
		}},
		{"manual service", func(t *testing.T, s service, shown WorkflowWorkStatus) (WorkflowWorkStatus, error) {
			return s.ControlWorkExtent(t.Context(), workflowIntegrationAnswerRequest(shown, "resolve"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operations, id := serviceIntegrationDecisionReady(t)

			before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
			if err != nil || before.Integration == nil || before.Integration.Decision == nil {
				t.Fatal("missing retained service question", err)
			}

			after, err := tc.submit(t, operations, before)
			checkWorkflowIntegrationAnswer(t, before, after, "resolved", err)

			images := workflowCLIImage(t, operations.projectContext.ProjectDir)
			if _, err := tc.submit(t, operations, before); !errors.Is(err, workflowstate.ErrConflict) || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("stale displayed answer changed authority", err)
			}
		})
	}
}

func checkWorkflowIntegrationAnswer(t *testing.T, before, after WorkflowWorkStatus, state string, err error) {
	t.Helper()

	if err != nil || after.Integration == nil || after.Integration.Decision == nil {
		t.Fatal("explicit decision response failed", err)
	}

	observed := after.Integration
	if observed.Decision.State != state || len(observed.Decision.Answers) != 1 || observed.Budget.Consumed != before.Integration.Budget.Consumed || observed.Publication != "not_started" || after.IntegratedUnits != 0 || after.ReviewState == "authorized" {
		t.Fatal("decision response lost its state, consumed an agent or approved publication")
	}
}

// TestWebIntegrationAnswer submits rendered candidate/question guards through
// HTTP, rejects stale question fences, and distinguishes resolve from defer.
func TestWebIntegrationAnswer(t *testing.T) {
	for _, action := range []string{"resolve", "defer"} {
		t.Run(action, func(t *testing.T) {
			operations, id := serviceIntegrationDecisionReady(t)
			handler := webWorkflowHandler(t, operations)
			page := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?id="+id, nil)
			fields := webIntegrationForm(t, page.Body.String(), "answer_decision")

			before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
			if err != nil {
				t.Fatal(err)
			}

			webIntegrationAnswerFields(t, page.Body.String(), fields, before, action)
			checkWebIntegrationAnswerGuards(t, operations, handler, fields)

			response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
			if !strings.Contains(response.Body.String(), `data-write-status="ok"`) {
				t.Fatal("rendered operator response failed", response.Body.String())
			}

			after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
			state := "resolved"

			if action == "defer" {
				state = "waiting"
			}

			checkWorkflowIntegrationAnswer(t, before, after, state, err)
			webWorkflowRejectedControl(t, operations, handler, fields, webui.WriteStatusStale)
		})
	}
}

func webIntegrationAnswerFields(t *testing.T, body string, fields url.Values, shown WorkflowWorkStatus, action string) {
	t.Helper()

	form := webIntegrationAnswerForm(t, body)
	for _, name := range []string{"record_id", "decision_id", "decision_revision", "decision_hash", "authority"} {
		fields.Set(name, webWorkflowHidden(t, form, name))
	}

	fields.Set("choice", shown.Integration.Decision.Requests[0].Options[0])
	fields.Set("answer_action", action)
	fields.Set("actor", "operator")
	fields.Set("reason", "Preserve both requirements through the compatible interpretation")
}

func webIntegrationAnswerForm(t *testing.T, body string) string {
	t.Helper()

	for _, form := range regexp.MustCompile(`(?s)<form\b.*?</form>`).FindAllString(body, -1) {
		if strings.Contains(form, `value="workflow.answer_decision"`) {
			return form
		}
	}

	t.Fatal("Web omitted the retained question form")

	return ""
}

func checkWebIntegrationAnswerGuards(t *testing.T, operations service, handler http.Handler, fields url.Values) {
	t.Helper()

	for _, tc := range []struct{ name, key, value string }{
		{"question revision", "decision_revision", "999"},
		{"question hash", "decision_hash", strings.Repeat("0", 64)},
		{"authority", "authority", "different-authority"},
		{"choice", "choice", "undisplayed-choice"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stale := maps.Clone(fields)
			stale.Set(tc.key, tc.value)
			webWorkflowRejectedControl(t, operations, handler, stale, webui.WriteStatusStale)
		})
	}
}
