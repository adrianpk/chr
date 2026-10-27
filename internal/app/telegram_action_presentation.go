package app

import (
	"encoding/json"
	"fmt"
	"strings"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramControlPreviewText(preview telegramControlPreview) string {
	control, status := preview.Request, preview.Status

	text := fmt.Sprintf("%s for work %s?\n\nCurrent state: %s.\n%s", telegramControlTitle(control.Action), status.ID, status.Status, telegramControlConsequence(control.Action))
	if grant := control.BudgetGrant; grant != nil {
		text += fmt.Sprintf("\nAllowance: %d %s. Reason: %s", grant.Amount, grant.Kind, grant.Reason)
	}

	if answer := control.DecisionAnswer; answer != nil {
		text += fmt.Sprintf("\nQuestion: %s\nAnswer: %s; choice: %s\nReason: %s", answer.DecisionID, answer.Action, answer.Choice, answer.Reason)
	}

	return text
}

func telegramControlTitle(action string) string {
	titles := map[string]string{"pause": "Request pause", "cancel": "Cancel execution", "resume": "Resume execution", "continue": "Continue one step", "reconcile": "Reconcile retained state", "approve": "Approve the current review", "grant": "Grant additional allowance", "answer_decision": "Confirm the selected answer"}

	return titles[action]
}

func telegramControlConsequence(action string) string {
	consequences := map[string]string{
		"pause":           "Requests a pause at the existing safe boundary. A running worker may still be stopping.",
		"cancel":          "Cancels this execution under its existing stop rules. Saved results remain available.",
		"resume":          "Makes the retained execution eligible to resume with its original scope and remaining allowance.",
		"continue":        "Runs at most one native checkpoint with the original scope, policy and remaining allowance.",
		"reconcile":       "Checks retained execution state under the existing recovery rules; it does not grant a new attempt allowance.",
		"approve":         "Approves only the currently observed review and its bound candidate.",
		"grant":           "Adds only the displayed allowance to this execution; other limits remain in force.",
		"answer_decision": "Records this answer for the currently observed integration question and candidate.",
	}

	return consequences[action]
}

func telegramRecordResultText(result RecordMutationResult) string {
	record := result.Record

	return fmt.Sprintf("Record saved: %s %s — %s.\nState: %s.\nCanonical record: %s", record.Kind, record.Identity, record.Title, record.Phase, record.Reference)
}

func telegramPreparedWorkReplay(request telegramcore.OperationRequest) string {
	var result AuthoringWorkResult
	if json.Unmarshal(request.Result, &result) != nil {
		return "This ticket preparation was already confirmed. Inspect the original result before requesting another preparation."
	}

	return "This ticket preparation was already confirmed.\n\n" + telegramPreparedWorkResultText(result)
}

func telegramActionReplayText(request telegramcore.OperationRequest) string {
	if request.OperationID == "authoring.work" {
		return telegramPreparedWorkReplay(request)
	}

	text := "This action was already confirmed. Its original result is retained."

	if strings.HasPrefix(request.OperationID, "records.") {
		var result RecordMutationResult
		if json.Unmarshal(request.Result, &result) == nil {
			text += "\n\n" + telegramRecordResultText(result)
		}
	}

	if details := telegramActionReplayDetails(request); details != "" {
		text += "\n\n" + details
	}

	return text
}

func telegramActionReplayDetails(request telegramcore.OperationRequest) string {
	switch request.OperationID {
	case "work.admit":
		var result telegramWorkAdmissionResult
		if json.Unmarshal(request.Result, &result) == nil {
			return fmt.Sprintf("Original admission: %s %s; %d tickets. Retained result state: %s. Use /status %s for current state.", result.Selector.Extent, result.Selector.Identity, result.Members, result.Status, result.Invocation.ID)
		}
	case "work.control":
		var status WorkflowWorkStatus
		if json.Unmarshal(request.Result, &status) == nil {
			return fmt.Sprintf("Original control receipt retained for work %s. Retained result state: %s. Use /status %s for current state.", status.ID, status.Status, status.ID)
		}
	case telegramcore.HarnessSelectionOperation:
		var facts struct {
			Snapshot *harnesscore.HarnessSelectionSnapshot `json:"snapshot"`
		}
		if json.Unmarshal(request.Result, &facts) == nil && facts.Snapshot != nil {
			return fmt.Sprintf("Original defaults saved for %s. Worker: %s. Generation: %s. Only new operations use these defaults.", facts.Snapshot.ProjectRef, telegramHarnessSavedRole(facts.Snapshot.Worker), telegramHarnessSavedRole(facts.Snapshot.Generation))
		}
	}

	return ""
}

func telegramOutcomeMessage(family, text string) *tgmessage.Blocks {
	return &tgmessage.Blocks{Family: family, InitialText: text}
}
