package core

import (
	"context"
	"database/sql"
	"time"
)

// TelegramIntentContext reads one exact peer's selection and retained sessions.
// The application supplies current authorization before invoking a provider.
func (s *TelegramReceiverStore) TelegramIntentContext(ctx context.Context, input TelegramConversationInput) (selected, focused string, sessions []TelegramConversation, err error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return "", "", nil, ErrWorkflowConflict
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		var err error

		selected, focused, err = op.peerSelection()
		if err != nil {
			return err
		}

		sessions, err = readTelegramConversations(ctx, tx, input.Update.Peer)

		return err
	})

	return selected, focused, sessions, err
}

// ResolveTelegramConversationTarget freezes the same exact/focused/sole routing
// used by capture operations, before an interpreted decision is retained.
func (s *TelegramReceiverStore) ResolveTelegramConversationTarget(ctx context.Context, input TelegramConversationInput) (TelegramOperationTarget, error) {
	var target TelegramOperationTarget
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return target, ErrWorkflowConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		conversation, err := op.target()
		if err != nil {
			return err
		}

		target, err = op.operationTarget(conversation)

		return err
	})

	return target, err
}
