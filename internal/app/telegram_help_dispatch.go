package app

import (
	"context"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) queueReadReply(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig, reply telegramReadReply) error {
	if reply.PlainText != "" {
		return r.store.QueueFormattedReadResponse(ctx, update, reply.Text, reply.PlainText, config.Digest(), time.Now())
	}

	return r.store.QueueReadResponse(ctx, update, core.BoundTelegramReceiverText(reply.Text), config.Digest(), time.Now())
}
