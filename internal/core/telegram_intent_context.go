package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// TelegramIntentSnapshot binds interpretation to one peer's observed context.
type TelegramIntentSnapshot struct {
	Selected string
	Focused  string
	Sessions []TelegramConversation
	Digest   string
}

// TelegramIntentContext reads selection and sessions in one transaction. Its
// digest can guard target resolution after a provider call without holding a TX.
func (s *TelegramReceiverStore) TelegramIntentContext(ctx context.Context, input TelegramConversationInput) (snapshot TelegramIntentSnapshot, err error) {
	if _, ok := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !ok {
		return snapshot, ErrWorkflowConflict
	}

	err = s.transaction(ctx, func(tx *sql.Tx) error {
		op := telegramConversationTransaction{ctx: ctx, tx: tx, input: input, now: time.Now()}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		var err error

		snapshot, err = op.intentSnapshot()

		return err
	})

	return snapshot, err
}

func (op telegramConversationTransaction) intentSnapshot() (TelegramIntentSnapshot, error) {
	var (
		snapshot TelegramIntentSnapshot
		err      error
	)

	snapshot.Selected, snapshot.Focused, err = op.peerSelection()
	if err != nil {
		return snapshot, err
	}

	snapshot.Sessions, err = readTelegramConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return snapshot, err
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}

	snapshot.Digest = workflowHash(encoded)

	return snapshot, nil
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
