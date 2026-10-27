package app

import (
	"context"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func (r telegramReceiver) queueReadReply(ctx context.Context, update telegramcore.ReceivedUpdate, config telegramcore.ReceiverConfig, reply telegramReadReply) error {
	if reply.PlainText != "" {
		return r.store.QueueFormattedReadResponse(ctx, telegramcore.FormattedReply{Source: update, Markdown: reply.Text, Plain: reply.PlainText, ConfigDigest: config.Digest(), At: time.Now()})
	}

	return r.store.QueuePresentationResponse(ctx, update, tgmessage.Blocks{InitialText: reply.Text}, config.Digest(), time.Now())
}

func telegramHelpRequest(update telegramcore.ReceivedUpdate) (telegramReadReply, bool) {
	if update.Kind != "message" {
		return telegramReadReply{}, false
	}

	word := telegramCommandWord(strings.TrimSpace(update.Text))

	name, group := word.Command, word.Body
	if name != "/help" {
		return telegramReadReply{}, false
	}

	if group == "" {
		return telegramBasicHelp(), true
	}

	return telegramGroupedHelp(group), true
}
