package telegram

import (
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

type updateRow struct {
	botID, epoch int64
	update       ReceivedUpdate
	state        string
	lane         ProcessingLane
	control      bool
	receivedAt   time.Time
}

const updateColumns = `bot_id,update_id,cursor_epoch,payload_digest,update_kind,
chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at,reply_to_message_id,processing_lane,control_input`

const insertUpdateSQL = `INSERT INTO telegram_updates(` + updateColumns + `) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

func (row updateRow) values() []any {
	update := row.update

	return []any{
		row.botID,
		update.UpdateID,
		row.epoch,
		update.Digest,
		update.Kind,
		update.Peer.ChatID,
		update.Peer.TopicID,
		update.Peer.UserID,
		update.MessageID,
		update.ChatType,
		database.FormatTime(row.receivedAt),
		update.Text,
		row.state,
		update.Reason,
		database.FormatTime(row.receivedAt),
		update.ReplyToMessageID,
		row.lane,
		row.control,
	}
}
