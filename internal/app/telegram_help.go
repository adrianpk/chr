package app

import (
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramReadReply struct {
	Text, PlainText string
}

const telegramHelpOverflow = "The complete command list exceeds the Telegram help limit. Use the local command reference for the full list."

func telegramHelpReply(prefix string) telegramReadReply {
	return renderTelegramHelp(telegramHelpMetadata(), prefix)
}

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
	return text != "" && core.BoundTelegramReceiverText(text) == text && len(core.TelegramResponseParts(text, "help")) == 1
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
