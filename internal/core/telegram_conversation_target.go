package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type telegramConversationReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Conversations reads only this exact peer's latest 50 retained sessions.
func (s *TelegramReceiverStore) Conversations(ctx context.Context, peer TelegramPeerIdentity) ([]TelegramConversation, error) {
	return readTelegramConversations(ctx, s.db, peer)
}

func readTelegramConversations(ctx context.Context, reader telegramConversationReader, peer TelegramPeerIdentity) ([]TelegramConversation, error) {
	rows, err := reader.QueryContext(ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? ORDER BY phase IN ('completed','canceled','failed'),updated_at DESC,id LIMIT 50`, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID)
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck // rows.Err is checked below.
		_ = rows.Close()
	}()

	var result []TelegramConversation

	for rows.Next() {
		c, err := scanTelegramConversation(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, c)
	}

	return result, rows.Err()
}

func scanTelegramConversation(scanner interface{ Scan(...any) error }) (TelegramConversation, error) {
	var row telegramConversationRow

	err := scanner.Scan(row.scanTargets()...)

	return row.conversation, err
}

func (op telegramConversationTransaction) conversation(id string) (*TelegramConversation, error) {
	p := op.input.Update.Peer

	c, err := scanTelegramConversation(op.tx.QueryRowContext(op.ctx, `SELECT `+telegramConversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, id, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, telegramConversationDiagnostic("That conversation is unavailable here. Send /sessions to choose one of your current conversations.")
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (op telegramConversationTransaction) explicitTarget() (*TelegramConversation, error) {
	if op.input.Update.ReplyToMessageID > 0 {
		return op.replyTarget()
	}

	if op.input.Conversation != "" {
		return op.conversation(op.input.Conversation)
	}

	return nil, nil
}

func (op telegramConversationTransaction) replyTarget() (*TelegramConversation, error) {
	p := op.input.Update.Peer

	var (
		id, binding string
		revision    int64
	)

	err := op.tx.QueryRowContext(op.ctx, `SELECT conversation_id,conversation_revision,question_binding FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=? AND purpose='capture' AND invalidated_at IS NULL`, p.BotID, p.ChatID, p.TopicID, p.UserID, op.input.Update.ReplyToMessageID).Scan(&id, &revision, &binding)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, telegramConversationDiagnostic("That reply is unavailable or expired here. Saved details are unchanged. Use /sessions and an explicit current target.")
	}

	if err != nil {
		return nil, err
	}

	c, err := op.conversation(id)
	if err != nil {
		return nil, err
	}

	if c.Revision != revision || c.QuestionBinding != binding || !telegramCaptureEligible(*c) {
		return nil, telegramConversationDiagnostic("That prompt was replaced or finished. Saved details are unchanged. Use /sessions to choose an explicit current session.")
	}

	if op.input.Conversation != "" && op.input.Conversation != c.ID {
		return nil, telegramConversationDiagnostic("The explicit conversation conflicts with the reply. Saved details are unchanged. Use /sessions to choose a matching target.")
	}

	return c, nil
}

func telegramCaptureEligible(c TelegramConversation) bool {
	return c.GuidedRunID != "" || c.Phase == "collecting" || c.Phase == "clarifying"
}

func (op telegramConversationTransaction) target() (*TelegramConversation, error) {
	c, err := op.explicitTarget()
	if err != nil || c != nil {
		return c, err
	}

	_, focused, err := op.peerSelection()
	if err != nil {
		return nil, err
	}

	if focused != "" {
		c, err = op.conversation(focused)
		if err != nil {
			return nil, err
		}

		if op.targetEligible(*c) {
			return c, nil
		}
	}

	return op.soleEligible()
}

func (op telegramConversationTransaction) soleEligible() (*TelegramConversation, error) {
	sessions, err := readTelegramConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return nil, err
	}

	var eligible []TelegramConversation

	for _, c := range sessions {
		if op.targetEligible(c) {
			eligible = append(eligible, c)
		}
	}

	if len(eligible) != 1 {
		return nil, telegramConversationDiagnostic("Select a session with /focus <conversation> or /reply <conversation> <text>. Use /sessions to list choices.")
	}

	return &eligible[0], nil
}

func (op telegramConversationTransaction) validateTarget(c *TelegramConversation) error {
	project, err := op.resolveProject(c)
	if err != nil {
		return err
	}

	binding, err := telegramProjectBinding(project)
	if err != nil {
		return err
	}

	if c.ProjectBindingDigest != binding {
		return telegramConversationDiagnostic("The project root changed. Resolve the registry binding explicitly by starting a new session; this session remains unchanged.")
	}

	return nil
}

func (op telegramConversationTransaction) focus() (string, *TelegramConversation, error) {
	c, err := op.explicitTarget()
	if err != nil {
		return "", nil, err
	}

	if c == nil {
		return "", nil, telegramConversationDiagnostic("Use /focus <conversation> with an exact session ID.")
	}

	if err := op.validateTarget(c); err != nil {
		return "", nil, err
	}

	if !telegramCaptureEligible(*c) {
		return "", nil, telegramConversationDiagnostic("This conversation is " + telegramConversationState(c.Phase) + ". Use /sessions to choose a conversation that can collect details.")
	}

	p := op.input.Update.Peer
	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_peers SET focused_conversation_id=?,revision=revision+1,updated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, c.ID, formatProjectTime(op.now), p.BotID, p.ChatID, p.TopicID, p.UserID)

	return "Your next details will go to " + telegramConversationLabel(c.Topic) + " in " + op.projectLabel(c.ProjectID) + ".", c, err
}

func (op telegramConversationTransaction) list() (string, *TelegramConversation, error) {
	sessions, err := readTelegramConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return "", nil, err
	}

	if len(sessions) == 0 {
		return "No retained sessions. Use /prepare <topic> or /develop <topic>.", nil, nil
	}

	_, focus, err := op.peerSelection()
	if err != nil {
		return "", nil, err
	}

	var text strings.Builder
	text.WriteString("Your conversations\n")

	for index, c := range sessions {
		row := op.conversationListRow(c, focus)
		if telegramVisibleUnits(text.String()+row) > 1500 {
			fmt.Fprintf(&text, "\nShowing %d of %d retained conversations. References in earlier messages still select their original conversation.", index, len(sessions))

			break
		}

		text.WriteString(row)
	}

	return text.String(), nil, nil
}

func (op telegramConversationTransaction) targetEligible(c TelegramConversation) bool {
	if op.input.Action == TelegramConversationAuthor {
		return telegramAuthoringEligible(c)
	}

	if op.input.Action == TelegramConversationCancel {
		return telegramCancelable(c)
	}

	return telegramCaptureEligible(c)
}

func telegramAuthoringEligible(c TelegramConversation) bool {
	return c.GuidedRunID == "" && (telegramCaptureEligible(c) || c.Phase == "preparing" || c.Phase == "prepared" || c.Phase == "awaiting_confirmation")
}
