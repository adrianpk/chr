package notify

import (
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

type telegramNotificationDocumentRow struct {
	id, owner string
	target    TelegramNotificationTarget
	document  tgmessage.Document
	createdAt time.Time
}

const telegramNotificationDocumentColumns = `id,owner_id,destination,config_digest,project_id,namespace,project_binding_digest,source_ref,source_hash,content_hash,filename,media_type,byte_count,payload,created_at`

const insertTelegramNotificationDocumentSQL = `INSERT INTO telegram_notification_documents(` + telegramNotificationDocumentColumns + `) SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?` + database.TelegramDocumentQuotaPredicate

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
		contenthash.Sum(d.Payload),
		d.Filename,
		d.MediaType,
		len(d.Payload),
		d.Payload,
		database.FormatTime(row.createdAt),
		len(d.Payload),
		tgmessage.DocumentQuota,
	}
}
