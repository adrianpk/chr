package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func readWorkMessageConversation(ctx context.Context, reader conversationReader, peer PeerIdentity, messageID int64) (string, error) {
	var id string

	err := reader.QueryRowContext(ctx, `SELECT o.conversation_id FROM telegram_outbox o JOIN telegram_conversations c ON c.id=o.conversation_id AND c.bot_id=o.bot_id AND c.chat_id=o.chat_id AND c.topic_id=o.topic_id AND c.user_id=o.user_id WHERE o.bot_id=? AND o.chat_id=? AND o.topic_id=? AND o.user_id=? AND o.state='delivered' AND o.action_token IS NULL AND COALESCE(c.invocation_id,'')!='' AND (?=0 OR o.telegram_message_id=?) ORDER BY o.updated_at DESC,o.rowid DESC LIMIT 1`, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID, messageID, messageID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return id, err
}

// IsWorkReply classifies an acknowledged work reference for dispatch.
// It grants no project access, capture eligibility or native control authority.
func (s *ReceiverStore) IsWorkReply(ctx context.Context, update ReceivedUpdate) (bool, error) {
	if update.Kind != "message" || update.ReplyToMessageID <= 0 {
		return false, nil
	}

	id, err := readWorkMessageConversation(ctx, s.db, update.Peer, update.ReplyToMessageID)

	return id != "", err
}

// ResolveWorkStatusTarget permits inspection through an acknowledged
// work message after capture ends. Capture and mutation reply rules stay separate.
func (s *ReceiverStore) ResolveWorkStatusTarget(ctx context.Context, input ConversationInput, now time.Time) (OperationTarget, error) {
	if input.Update.ReplyToMessageID == 0 {
		return s.ResolveOperationTarget(ctx, input, now)
	}

	var target OperationTarget

	if _, allowed := input.Config.AuthorizedUser(input.Update.Peer.ChatID, input.Update.Peer.TopicID, input.Update.Peer.UserID, input.Update.ChatType); !allowed {
		return target, workflowstate.ErrConflict
	}

	err := s.transaction(ctx, func(tx *sql.Tx) error {
		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}
		if err := op.ensurePeer(); err != nil {
			return err
		}

		id, err := readWorkMessageConversation(ctx, tx, input.Update.Peer, input.Update.ReplyToMessageID)
		if err != nil {
			return err
		}

		if id == "" || input.Conversation != "" && input.Conversation != id {
			return workflowstate.ErrConflict
		}

		conversation, err := op.conversation(id)
		if err != nil {
			return err
		}

		target, err = op.operationTarget(conversation)

		return err
	})

	return target, err
}

// QueueWorkStatus keeps a conversation-bound read reply inspectable
// without creating a control request or changing the execution or conversation.
func (s *ReceiverStore) QueueWorkStatus(ctx context.Context, input ConversationInput, target OperationTarget, invocation string, message tgmessage.Blocks, now time.Time) error {
	if target.ConversationID == "" {
		return s.QueuePresentationResponse(ctx, input.Update, message, input.Config.Digest(), now)
	}

	return s.transaction(ctx, func(tx *sql.Tx) error {
		pending, err := readResponsePending(ctx, tx, input.Update)
		if err != nil || !pending {
			return err
		}

		op := conversationTransaction{ctx: ctx, tx: tx, input: input, now: now}

		conversation, err := op.conversation(target.ConversationID)
		if err != nil {
			return err
		}

		current, err := op.operationTarget(conversation)
		if err != nil {
			return err
		}

		if current != target || invocation == "" || conversation.InvocationID != invocation {
			return workflowstate.ErrConflict
		}

		if err := queuePresentation(ctx, tx, input.Update, message, input.Config.Digest(), now); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE telegram_outbox SET conversation_id=?,conversation_revision=?,request_id=? WHERE bot_id=? AND event_key=?`, target.ConversationID, target.ConversationRevision, conversation.AuthorizationRef, input.Update.Peer.BotID, fmt.Sprintf("update:%d", input.Update.UpdateID))

		return err
	})
}
