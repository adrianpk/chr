package telegram

import (
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
)

type responsePartRow struct {
	id                      string
	peer                    PeerIdentity
	key, configDigest, text string
	part, parts             int
	at                      time.Time
}

const insertResponsePartSQL = `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,part_index,part_count,state,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,'result',?,?,?,'pending',?,?)`

func (row responsePartRow) values() []any {
	return []any{
		row.id,
		row.peer.BotID,
		row.peer.ChatID,
		row.peer.TopicID,
		row.peer.UserID,
		row.key,
		row.configDigest,
		row.text,
		row.part,
		row.parts,
		database.FormatTime(row.at),
		database.FormatTime(row.at),
	}
}

type workNoticeRow struct {
	ref                           WorkReference
	kind, key, text, configDigest string
	part, parts                   int
	at                            time.Time
}

const insertWorkNoticeSQL = `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,conversation_id,conversation_revision,request_id,event_key,config_digest,kind,text,part_index,part_count,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,'pending',?,?)`

func (row workNoticeRow) values() []any {
	peer, target := row.ref.Request.Source.Peer, row.ref.Request.Target
	id := "tgd-" + contenthash.Sum(fmt.Appendf(nil, "%d:%s:%d", peer.BotID, row.key, row.part))[:32]

	return []any{
		id,
		peer.BotID,
		peer.ChatID,
		peer.TopicID,
		peer.UserID,
		target.ConversationID,
		row.ref.ConversationRevision,
		row.ref.Request.ID,
		row.key,
		row.configDigest,
		row.kind,
		row.text,
		row.part,
		row.parts,
		database.FormatTime(row.at),
		database.FormatTime(row.at),
	}
}
