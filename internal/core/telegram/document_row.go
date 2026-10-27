package telegram

import (
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Each admission retains its table's ownership shape and quota write reservation.
type documentSnapshotRow struct {
	id, owner    string
	peer         PeerIdentity
	configDigest string
	document     tgmessage.Document
	createdAt    time.Time
}

const documentSnapshotColumns = `id,owner_kind,owner_id,bot_id,chat_id,topic_id,user_id,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at`

const insertDocumentSnapshotSQL = `INSERT INTO telegram_document_snapshots(` + documentSnapshotColumns + `) SELECT ?,'receiver',?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?` + database.TelegramDocumentQuotaPredicate

// values omits owner_kind: the statement retains its fixed receiver literal.
func (row documentSnapshotRow) values() []any {
	peer, document := row.peer, row.document

	return []any{
		row.id,
		row.owner,
		peer.BotID,
		peer.ChatID,
		peer.TopicID,
		peer.UserID,
		row.configDigest,
		document.ProjectID,
		document.Namespace,
		document.ProjectBindingDigest,
		document.SourceRef,
		document.SourceHash,
		contenthash.Sum(document.Payload),
		document.Filename,
		document.MediaType,
		len(document.Payload),
		document.Payload,
		database.FormatTime(row.createdAt),
		len(document.Payload),
		tgmessage.DocumentQuota,
	}
}
