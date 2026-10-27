package telegram

import (
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

const actionColumns = `token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,COALESCE(consumed_by_request,''),guard_version,COALESCE(presentation_hash,'')`

type actionReadRow struct {
	action          Action
	choices, expiry string
}

func (row *actionReadRow) scanTargets() []any {
	return []any{
		&row.action.Token,
		&row.action.RequestID,
		&row.action.Peer.BotID,
		&row.action.Peer.ChatID,
		&row.action.Peer.TopicID,
		&row.action.Peer.UserID,
		&row.action.ConversationRevision,
		&row.action.GuardDigest,
		&row.choices,
		&row.action.State,
		&row.expiry,
		&row.action.AnswerRequestID,
		&row.action.GuardVersion,
		&row.action.PresentationHash,
	}
}

type actionAdmissionRow struct {
	token                string
	request              OperationRequest
	choices              string
	expiresAt, createdAt time.Time
}

const insertActionSQL = `INSERT INTO telegram_actions(token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,'pending',?,?,?)`

func (row actionAdmissionRow) values() []any {
	p := row.request.Source.Peer

	return []any{
		row.token,
		row.request.ID,
		p.BotID,
		p.ChatID,
		p.TopicID,
		p.UserID,
		row.request.Target.ConversationRevision,
		actionGuard(row.request),
		row.choices,
		database.FormatTime(row.expiresAt),
		database.FormatTime(row.createdAt),
		database.FormatTime(row.createdAt),
	}
}
