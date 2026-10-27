package database

// TelegramDocumentQuotaPredicate shares one byte quota across both retained-document tables.
const TelegramDocumentQuotaPredicate = ` WHERE (SELECT COALESCE(SUM(length(payload)),0) FROM telegram_document_snapshots)+(SELECT COALESCE(SUM(length(payload)),0) FROM telegram_notification_documents)+?<=?`
