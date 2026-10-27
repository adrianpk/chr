package telegram

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"unicode/utf8"

	"forge.adrianpk.com/chr/internal/core/database"
)

func (op conversationTransaction) applyCapture() (string, *Conversation, error) {
	c, err := op.target()
	if err != nil {
		return "", nil, err
	}

	if err := op.validateTarget(c); err != nil {
		return "", nil, err
	}

	if c.GuidedRunID != "" {
		return "", nil, conversationDiagnosticError("This is a guided use case. Continue through its current guided prompt; manual capture commands do not change it.")
	}

	if op.input.Action == ConversationShow {
		return c.Capture, c, nil
	}

	if op.input.Action == ConversationCancel && cancelable(*c) {
		return op.cancelCapture(c)
	}

	if !captureEligible(*c) {
		return "", nil, conversationDiagnosticError("This capture is finished. Start a new session to collect more text.")
	}

	return op.applyCaptureAction(c)
}

func (op conversationTransaction) applyCaptureAction(c *Conversation) (string, *Conversation, error) {
	switch op.input.Action {
	case ConversationCapture:
		return op.appendCapture(c)
	case ConversationDone:
		return op.finishCapture(c)
	default:
		return "", nil, conversationDiagnosticError("Unknown conversation operation. Use /help.")
	}
}

func (op conversationTransaction) appendCapture(c *Conversation) (string, *Conversation, error) {
	text := op.input.Text
	if text == "" || len(text) > 16*1024 || !utf8.ValidString(text) {
		return "", nil, conversationDiagnosticError("Each detail message must contain valid UTF-8 text of at most 16 KiB. Saved details are unchanged; send a shorter message or /done.")
	}

	var parts int
	if err := op.tx.QueryRowContext(op.ctx, `SELECT COUNT(*) FROM telegram_capture_parts WHERE conversation_id=?`, c.ID).Scan(&parts); err != nil {
		return "", nil, err
	}

	separator := ""
	if parts > 0 {
		separator = "\n"
	}

	start := len(c.Capture) + len(separator)
	if parts >= 32 || start+len(text) > 64*1024 {
		return "", nil, conversationDiagnosticError("Capture limit reached: at most 64 KiB including separators and 32 parts. Existing text is unchanged; use /done or /cancel.")
	}

	sum := sha256.Sum256([]byte(text))

	_, err := op.tx.ExecContext(op.ctx, `INSERT INTO telegram_capture_parts(conversation_id,bot_id,update_id,text_digest,byte_start,byte_end) VALUES(?,?,?,?,?,?)`, c.ID, c.Peer.BotID, op.input.Update.UpdateID, hex.EncodeToString(sum[:]), start, start+len(text))
	if err != nil {
		return "", nil, err
	}

	c.Capture += separator + text
	if err := op.saveCapture(c); err != nil {
		return "", nil, err
	}

	return captureSavedText(*c, parts+1), c, nil
}

func (op conversationTransaction) finishCapture(c *Conversation) (string, *Conversation, error) {
	c.Phase = "preparing"
	if err := op.saveCapture(c); err != nil {
		return "", nil, err
	}

	return "Your request and details are saved. Artifact generation has not run. Use /sessions to inspect the conversation.", c, nil
}

func (op conversationTransaction) cancelCapture(c *Conversation) (string, *Conversation, error) {
	c.Phase, c.Capture, c.QuestionBinding = "canceled", "", ""
	if _, err := op.tx.ExecContext(op.ctx, `DELETE FROM telegram_capture_parts WHERE conversation_id=?`, c.ID); err != nil {
		return "", nil, err
	}

	if err := op.saveCapture(c); err != nil {
		return "", nil, err
	}

	return "Conversation canceled. Its collected details were discarded.", c, nil
}

func (op conversationTransaction) saveCapture(c *Conversation) error {
	finished := sql.NullString{String: database.FormatTime(op.now), Valid: c.Phase == "canceled"}

	result, err := op.tx.ExecContext(op.ctx, `UPDATE telegram_conversations SET capture=?,phase=?,question_binding=?,revision=revision+1,updated_at=?,finished_at=? WHERE id=? AND revision=?`, c.Capture, c.Phase, c.QuestionBinding, database.FormatTime(op.now), finished, c.ID, c.Revision)
	if err != nil {
		return err
	}

	if err := requireChanged(result); err != nil {
		return err
	}

	c.Revision++
	_, err = op.tx.ExecContext(op.ctx, `UPDATE telegram_replies SET invalidated_at=? WHERE conversation_id=? AND invalidated_at IS NULL`, database.FormatTime(op.now), c.ID)

	return err
}

func cancelable(c Conversation) bool {
	return captureEligible(c) || c.Phase == "preparing" || c.Phase == "prepared" || c.Phase == "awaiting_confirmation"
}
