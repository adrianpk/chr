package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

type conversationReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Conversations reads only this exact peer's latest 50 retained sessions.
func (s *ReceiverStore) Conversations(ctx context.Context, peer PeerIdentity) ([]Conversation, error) {
	return readConversations(ctx, s.db, peer)
}

func readConversations(ctx context.Context, reader conversationReader, peer PeerIdentity) (result []Conversation, err error) {
	rows, err := reader.QueryContext(ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? ORDER BY phase IN ('completed','canceled','failed'),updated_at DESC,id LIMIT 50`, peer.BotID, peer.ChatID, peer.TopicID, peer.UserID)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, c)
	}

	return result, rows.Err()
}

func scanConversation(scanner interface{ Scan(...any) error }) (Conversation, error) {
	var row conversationRow

	err := scanner.Scan(row.scanTargets()...)

	return row.conversation, err
}

func (op conversationTransaction) conversation(id string) (*Conversation, error) {
	p := op.input.Update.Peer

	c, err := scanConversation(op.tx.QueryRowContext(op.ctx, `SELECT `+conversationColumns+` FROM telegram_conversations WHERE id=? AND bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, id, p.BotID, p.ChatID, p.TopicID, p.UserID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, conversationDiagnosticError("That conversation is unavailable here. Send /sessions to choose one of your current conversations.")
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (op conversationTransaction) explicitTarget() (*Conversation, error) {
	if op.input.Update.ReplyToMessageID > 0 {
		return op.replyTarget()
	}

	if op.input.Conversation != "" {
		return op.conversation(op.input.Conversation)
	}

	return nil, nil
}

func (op conversationTransaction) replyTarget() (*Conversation, error) {
	p := op.input.Update.Peer

	var (
		id, binding string
		revision    int64
	)

	err := op.tx.QueryRowContext(op.ctx, `SELECT conversation_id,conversation_revision,question_binding FROM telegram_replies WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=? AND message_id=? AND purpose='capture' AND invalidated_at IS NULL`, p.BotID, p.ChatID, p.TopicID, p.UserID, op.input.Update.ReplyToMessageID).Scan(&id, &revision, &binding)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, conversationDiagnosticError("That reply is unavailable or expired here. Saved details are unchanged. Use /sessions and an explicit current target.")
	}

	if err != nil {
		return nil, err
	}

	c, err := op.conversation(id)
	if err != nil {
		return nil, err
	}

	if c.Revision != revision || c.QuestionBinding != binding || !captureEligible(*c) {
		return nil, conversationDiagnosticError("That prompt was replaced or finished. Saved details are unchanged. Use /sessions to choose an explicit current session.")
	}

	if op.input.Conversation != "" && op.input.Conversation != c.ID {
		return nil, conversationDiagnosticError("The explicit conversation conflicts with the reply. Saved details are unchanged. Use /sessions to choose a matching target.")
	}

	return c, nil
}

func captureEligible(c Conversation) bool {
	return c.GuidedRunID != "" || c.Phase == ConversationCollecting || c.Phase == ConversationClarifying
}

func (op conversationTransaction) target() (*Conversation, error) {
	c, err := op.explicitTarget()
	if err != nil || c != nil {
		return c, err
	}

	selection, err := op.peerSelection()
	focused := selection.RootDigest

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

func (op conversationTransaction) soleEligible() (*Conversation, error) {
	sessions, err := readConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return nil, err
	}

	var eligible []Conversation

	for _, c := range sessions {
		if op.targetEligible(c) {
			eligible = append(eligible, c)
		}
	}

	if len(eligible) != 1 {
		return nil, conversationDiagnosticError("Select a session with /focus <conversation> or /reply <conversation> <text>. Use /sessions to list choices.")
	}

	return &eligible[0], nil
}

func (op conversationTransaction) validateTarget(c *Conversation) error {
	project, err := op.resolveProject(c)
	if err != nil {
		return err
	}

	binding, err := projectBinding(project)
	if err != nil {
		return err
	}

	if c.ProjectBindingDigest != binding {
		return conversationDiagnosticError("The project root changed. Resolve the registry binding explicitly by starting a new session; this session remains unchanged.")
	}

	return nil
}

func (op conversationTransaction) focus() (conversationReply, error) {
	c, err := op.explicitTarget()
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if c == nil {
		return conversationReply{Text: "", Conversation: nil}, conversationDiagnosticError("Use /focus <conversation> with an exact session ID.")
	}

	if err := op.validateTarget(c); err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if !captureEligible(*c) {
		return conversationReply{Text: "", Conversation: nil}, conversationDiagnosticError("This conversation is " + conversationState(c.Phase) + ". Use /sessions to choose a conversation that can collect details.")
	}

	p := op.input.Update.Peer
	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_peers SET focused_conversation_id=?,revision=revision+1,updated_at=? WHERE bot_id=? AND chat_id=? AND topic_id=? AND user_id=?`, c.ID, database.FormatTime(op.now), p.BotID, p.ChatID, p.TopicID, p.UserID)

	return conversationReply{Text: "Your next details will go to " + conversationLabel(c.Topic) + " in " + op.projectLabel(c.ProjectID) + ".", Conversation: c}, err
}

func (op conversationTransaction) list() (conversationReply, error) {
	sessions, err := readConversations(op.ctx, op.tx, op.input.Update.Peer)
	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	if len(sessions) == 0 {
		return conversationReply{Text: "No retained sessions. Use /prepare <topic> or /develop <topic>.", Conversation: nil}, nil
	}

	selection2, err := op.peerSelection()
	focus := selection2.RootDigest

	if err != nil {
		return conversationReply{Text: "", Conversation: nil}, err
	}

	var text strings.Builder
	text.WriteString("Your conversations\n")

	for index, c := range sessions {
		row := op.conversationListRow(c, focus)
		if tgmessage.VisibleUnits(text.String()+row) > 1500 {
			fmt.Fprintf(&text, "\nShowing %d of %d retained conversations. References in earlier messages still select their original conversation.", index, len(sessions))

			break
		}

		text.WriteString(row)
	}

	return conversationReply{Text: text.String(), Conversation: nil}, nil
}

func (op conversationTransaction) targetEligible(c Conversation) bool {
	if op.input.Action == ConversationAuthor {
		return authoringEligible(c)
	}

	if op.input.Action == ConversationCancel {
		return cancelable(c)
	}

	return captureEligible(c)
}

func authoringEligible(c Conversation) bool {
	return c.GuidedRunID == "" && (captureEligible(c) || c.Phase == ConversationPreparing || c.Phase == ConversationPrepared || c.Phase == ConversationAwaitingConfirmation)
}
