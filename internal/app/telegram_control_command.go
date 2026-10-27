package app

import (
	"strconv"
	"strings"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramControlCommand struct {
	Input      telegramcore.ConversationInput
	Action     string
	Invocation string
	Parameters *telegramControlParameters
}

func telegramControlRegistry() []ConversationIntentOperation {
	return []ConversationIntentOperation{{ID: "control", Description: "Inspect or control an exact retained native invocation", Arguments: "Optional project/conversation/namespace; action=status|pause|cancel|resume|continue|reconcile|approve|grant|answer_decision; text=exact invocation (empty only for a bound conversation); grant record_json={kind,amount,reason}; answer_decision record_json={question_id,answer_action,option,reason}; no actor, authority or revision fields"}}
}

type telegramControlDescriptor struct {
	Name, Arguments string
}

func telegramControlCommands() []telegramControlDescriptor {
	const scope = "[--project <id>] [--conversation <id>] [--namespace <name>] "

	var commands []telegramControlDescriptor
	for _, name := range []string{"status", "pause", "cancel", "resume", "continue", "reconcile", "approve"} {
		commands = append(commands, telegramControlDescriptor{name, scope + "[execution]"})
	}

	return append(commands,
		telegramControlDescriptor{"grant", scope + "<execution> <kind> <amount> <reason>"},
		telegramControlDescriptor{"answer-decision", scope + "<execution> <question-id> <resolve|defer> <option-number> <reason>"},
	)
}

func telegramControlHelp() string {
	var lines []string
	for _, command := range telegramControlCommands() {
		lines = append(lines, telegramCommandSyntax("/"+command.Name, command.Arguments))
	}

	return strings.Join(lines, "\n")
}

func parseTelegramControl(update telegramcore.ReceivedUpdate) (telegramControlCommand, bool, error) {
	command := telegramControlCommand{Input: telegramcore.ConversationInput{Update: update}}
	name, body := telegramCommandWord(strings.TrimSpace(update.Text))

	command.Action = strings.ReplaceAll(strings.TrimPrefix(name, "/"), "-", "_")
	if update.Kind != "message" || !strings.HasPrefix(name, "/") || !validTelegramControlAction(command.Action) {
		return command, false, nil
	}

	if body == "" && update.ReplyToMessageID == 0 && (command.Action == "cancel" || command.Action == "status") {
		return command, false, nil
	}

	body, err := parseTelegramRecordScope(&command.Input, body)
	if err == nil {
		err = decodeTelegramControlCommand(&command, body)
	}

	if err != nil {
		return command, true, err
	}

	return command, true, validateTelegramControlCommand(command)
}

func validTelegramControlAction(action string) bool {
	for _, command := range telegramControlCommands() {
		if strings.ReplaceAll(command.Name, "-", "_") == action {
			return true
		}
	}

	return false
}

func validateTelegramControlCommand(command telegramControlCommand) error {
	if !validTelegramControlAction(command.Action) || command.Invocation != "" && (!validWorkflowControlIdentity(command.Invocation) || len(strings.Fields(command.Invocation)) != 1) {
		return ErrRecordUnsupported
	}

	return validateTelegramControlParameters(command.Action, command.Parameters)
}

func telegramIntentControl(update telegramcore.ReceivedUpdate, intent ConversationIntent) (telegramControlCommand, bool, error) {
	a := intent.Arguments

	command := telegramControlCommand{Input: telegramcore.ConversationInput{Update: update, Project: a.Project, Conversation: a.Conversation, Namespace: a.Namespace}, Action: a.Action, Invocation: a.Text}
	if intent.OperationID != "control" {
		return command, false, nil
	}

	if a.RecordJSON != "" {
		command.Parameters = &telegramControlParameters{}
		if err := decodeOperationJSON(a.RecordJSON, command.Parameters); err != nil {
			return command, true, err
		}
	}

	return command, true, validateTelegramControlCommand(command)
}

func decodeTelegramControlCommand(command *telegramControlCommand, body string) error {
	if command.Action != "grant" && command.Action != "answer_decision" {
		command.Invocation = body

		return nil
	}

	fields := strings.Fields(body)

	minimum := 4
	if command.Action == "answer_decision" {
		minimum = 5
	}

	if len(fields) < minimum {
		return ErrRecordUnsupported
	}

	command.Invocation = fields[0]
	p := &telegramControlParameters{}

	command.Parameters = p
	if command.Action == "grant" {
		amount, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return ErrRecordUnsupported
		}

		p.Kind, p.Amount, p.Reason = fields[1], amount, strings.Join(fields[3:], " ")

		return nil
	}

	option, err := strconv.Atoi(fields[3])
	if err != nil {
		return ErrRecordUnsupported
	}

	p.QuestionID, p.AnswerAction, p.Option, p.Reason = fields[1], fields[2], option, strings.Join(fields[4:], " ")

	return nil
}
