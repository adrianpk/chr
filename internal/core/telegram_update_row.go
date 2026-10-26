package core

import "time"

type telegramUpdateRow struct {
	botID, epoch int64
	update       TelegramReceivedUpdate
	state        string
	lane         TelegramProcessingLane
	control      bool
	receivedAt   time.Time
}

const telegramUpdateColumns = `bot_id,update_id,cursor_epoch,payload_digest,update_kind,
chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at,reply_to_message_id,processing_lane,control_input`

const insertTelegramUpdateSQL = `INSERT INTO telegram_updates(` + telegramUpdateColumns + `) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

func (row telegramUpdateRow) values() []any {
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
		formatProjectTime(row.receivedAt),
		update.Text,
		row.state,
		update.Reason,
		formatProjectTime(row.receivedAt),
		update.ReplyToMessageID,
		row.lane,
		row.control,
	}
}
