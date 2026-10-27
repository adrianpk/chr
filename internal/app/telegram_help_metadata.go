package app

import telegramcore "forge.adrianpk.com/chr/internal/core/telegram"

type telegramHelpEntry struct {
	ID, Category, Syntax string
}

// The command owners supply syntax; the renderer never reconstructs a catalogue.
func telegramHelpMetadata() []telegramHelpEntry {
	var entries []telegramHelpEntry

	for _, operation := range telegramReadRegistry() {
		category := "System controls"
		if operation.ID == telegramReadProjects {
			category = "Projects"
		}

		entries = append(entries, telegramHelpEntry{operation.Command, category, operation.Command})
	}

	for _, command := range telegramConversationRegistry() {
		category := "Conversations"
		if command.Action == telegramcore.ConversationSelect {
			category = "Projects"
		}

		entries = append(entries, telegramHelpEntry{command.Name, category, telegramCommandSyntax(command.Name, command.Arguments)})
	}

	for _, command := range telegramAuthoringRegistry() {
		entries = append(entries, telegramHelpEntry{command.Name, "Authoring", telegramCommandSyntax(command.Name, command.Arguments)})
	}

	entries = append(entries,
		telegramHelpEntry{"explain", "Authoring", telegramExplanationUsage()},
		telegramHelpEntry{"prepare-work", "Authoring", telegramPreparedWorkUsage()},
		telegramHelpEntry{"work", "Execution", telegramWorkUsage()},
		telegramHelpEntry{"harness", "Projects", telegramHarnessUsage},
		telegramHelpEntry{telegramcore.ActionAnswerCommand, "System controls", telegramcore.ActionAnswerUsage},
	)

	return append(entries, telegramRecordControlMetadata()...)
}

func telegramRecordControlMetadata() []telegramHelpEntry {
	var entries []telegramHelpEntry
	for _, command := range telegramControlCommands() {
		entries = append(entries, telegramHelpEntry{command.Name, "Execution", telegramCommandSyntax("/"+command.Name, command.Arguments)})
	}

	for _, command := range telegramRecordRegistry() {
		category := "Tickets"

		switch command.Kind {
		case RecordSlice, RecordSliceSet:
			category = "Slice sets"
		case RecordUseCase:
			category = "Use cases"
		}

		entries = append(entries, telegramHelpEntry{command.Name, category, telegramRecordUsage(command)})
	}

	return entries
}

func telegramCommandSyntax(name, arguments string) string {
	if arguments == "" {
		return name
	}

	return name + " " + arguments
}
