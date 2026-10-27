package app

import (
	"slices"
	"strings"
	"unicode/utf16"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramReadReply struct {
	Text, PlainText string
}

const telegramHelpOverflow = "The complete command list exceeds the Telegram help limit. Use the local command reference for the full list."

func renderTelegramHelp(entries []telegramHelpEntry, prefix string) telegramReadReply {
	if len(entries) == 0 {
		return telegramReadReply{}
	}

	entries = slices.Clone(entries)
	slices.SortFunc(entries, compareTelegramHelp)

	var formatted, plain strings.Builder
	formatted.WriteString(escapeTelegramHelp(prefix))
	plain.WriteString(prefix)

	category := ""

	for index, entry := range entries {
		name := telegramHelpCategory(entry.Category)
		if index == 0 || name != category {
			if index > 0 {
				formatted.WriteString("\n\n")
				plain.WriteString("\n\n")
			}

			formatted.WriteString(telegramHelpEntity(name, "*"))
			plain.WriteString(name)

			category = name
		}

		formatted.WriteString("\n" + telegramHelpEntity(entry.Syntax, "`"))
		plain.WriteString("\n" + entry.Syntax)
	}

	reply := telegramReadReply{Text: formatted.String(), PlainText: plain.String()}
	if !telegramHelpFits(reply.Text) || !telegramHelpFits(reply.PlainText) {
		text := prefix + telegramHelpOverflow
		if !telegramHelpFits(text) {
			text = telegramHelpOverflow
		}

		return telegramReadReply{Text: text}
	}

	return reply
}

func telegramHelpFits(text string) bool {
	return text != "" && core.BoundTelegramReceiverText(text) == text && len(utf16.Encode([]rune(text))) <= 3200
}

func compareTelegramHelp(left, right telegramHelpEntry) int {
	lc, rc := telegramHelpCategory(left.Category), telegramHelpCategory(right.Category)
	if order := telegramHelpOrder(lc) - telegramHelpOrder(rc); order != 0 {
		return order
	}

	if category := strings.Compare(lc, rc); category != 0 {
		return category
	}

	return strings.Compare(left.Syntax, right.Syntax)
}

func telegramHelpCategory(category string) string {
	if strings.TrimSpace(category) == "" {
		return "Other"
	}

	return category
}

func telegramHelpOrder(category string) int {
	categories := []string{"Execution", "Tickets", "Slice sets", "Use cases", "Projects", "Authoring", "Conversations", "System controls", "Other"}
	if index := slices.Index(categories, category); index >= 0 {
		return index
	}

	return len(categories)
}

func telegramHelpEntity(text, delimiter string) string {
	unsafe := "`\\\r\n"
	if delimiter == "*" {
		unsafe += "_*["
	}

	if strings.ContainsAny(text, unsafe) {
		return escapeTelegramHelp(text)
	}

	return delimiter + text + delimiter
}

func escapeTelegramHelp(text string) string {
	return strings.NewReplacer("\\", "\\\\", "_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[").Replace(text)
}

// Keep the first help view focused; the complete native catalogue remains
// explicitly available without inspecting or changing a pending work binding.
func telegramBasicHelp() telegramReadReply {
	var entries []telegramHelpEntry

	for _, entry := range telegramHelpMetadata() {
		switch entry.Category {
		case "Tickets", "Slice sets", "Use cases":
			continue
		}

		if entry.Category == "Execution" && entry.ID != "work" {
			continue
		}

		if entry.ID == "harness" || entry.ID == "prepare-work" || entry.ID == "explain" || entry.ID == core.TelegramActionAnswerCommand {
			continue
		}

		entries = append(entries, entry)
	}

	reply := renderTelegramHelp(entries, "")

	reply.Text += "\n\nSend /help all to choose an advanced command group."
	if reply.PlainText != "" {
		reply.PlainText += "\n\nSend /help all to choose an advanced command group."
	}

	return reply
}

func telegramGroupedHelp(group string) telegramReadReply {
	categories := map[string]string{"execution": "Execution", "tickets": "Tickets", "slices": "Slice sets", "cases": "Use cases", "projects": "Projects", "authoring": "Authoring", "conversations": "Conversations", "system": "System controls"}

	if group == "all" {
		return telegramReadReply{Text: "Choose a command group:\n\n/help execution\n/help tickets\n/help slices\n/help cases\n/help projects\n/help authoring\n/help conversations\n/help system"}
	}

	category, ok := categories[group]
	if !ok {
		return telegramReadReply{Text: "Send /help all to choose a command group."}
	}

	var entries []telegramHelpEntry

	for _, entry := range telegramHelpMetadata() {
		if entry.Category == category {
			entries = append(entries, entry)
		}
	}

	return renderTelegramHelp(entries, "")
}
