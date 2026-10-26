package core

import "time"

// Each admission retains its table's ownership shape and quota write reservation.
type telegramDocumentSnapshotRow struct {
	id, owner    string
	peer         TelegramPeerIdentity
	configDigest string
	document     TelegramDocument
	createdAt    time.Time
}

type telegramNotificationDocumentRow struct {
	id, owner string
	target    TelegramNotificationTarget
	document  TelegramDocument
	createdAt time.Time
}

const telegramDocumentQuotaPredicate = ` WHERE (SELECT COALESCE(SUM(length(payload)),0) FROM telegram_document_snapshots)+(SELECT COALESCE(SUM(length(payload)),0) FROM telegram_notification_documents)+?<=?`

const telegramDocumentSnapshotColumns = `id,owner_kind,owner_id,bot_id,chat_id,topic_id,user_id,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at`

const insertTelegramDocumentSnapshotSQL = `INSERT INTO telegram_document_snapshots(` + telegramDocumentSnapshotColumns + `) SELECT ?,'receiver',?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?` + telegramDocumentQuotaPredicate

// values omits owner_kind: the statement retains its fixed receiver literal.
func (row telegramDocumentSnapshotRow) values() []any {
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
		workflowHash(document.Payload),
		document.Filename,
		document.MediaType,
		len(document.Payload),
		document.Payload,
		formatProjectTime(row.createdAt),
		len(document.Payload),
		telegramDocumentQuota,
	}
}

const telegramNotificationDocumentColumns = `id,owner_id,destination,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at`

const insertTelegramNotificationDocumentSQL = `INSERT INTO telegram_notification_documents(` + telegramNotificationDocumentColumns + `) SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?` + telegramDocumentQuotaPredicate

func (row telegramNotificationDocumentRow) values() []any {
	d, target := row.document, row.target

	return []any{
		row.id,
		row.owner,
		target.Destination,
		target.ConfigDigest,
		d.ProjectID,
		d.Namespace,
		d.ProjectBindingDigest,
		d.SourceRef,
		d.SourceHash,
		workflowHash(d.Payload),
		d.Filename,
		d.MediaType,
		len(d.Payload),
		d.Payload,
		formatProjectTime(row.createdAt),
		len(d.Payload),
		telegramDocumentQuota,
	}
}
