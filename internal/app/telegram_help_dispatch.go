package app

import (
	"context"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) queueReadReply(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig, reply telegramReadReply) error {
	if reply.PlainText != "" {
		return r.store.QueueFormattedReadResponse(ctx, update, reply.Text, reply.PlainText, config.Digest(), time.Now())
	}

	return r.store.QueuePresentationResponse(ctx, update, core.TelegramMessageBlocks{InitialText: reply.Text}, config.Digest(), time.Now())
}

func telegramHelpRequest(update core.TelegramReceivedUpdate) (telegramReadReply, bool) {
	if update.Kind != "message" {
		return telegramReadReply{}, false
	}
	name, group := telegramCommandWord(strings.TrimSpace(update.Text))
	if name != "/help" {
		return telegramReadReply{}, false
	}
	if group == "" {
		return telegramBasicHelp(), true
	}
	return telegramGroupedHelp(group), true
}
