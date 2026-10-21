-- +migrate Up
CREATE TABLE telegram_notification_documents (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL UNIQUE REFERENCES notification_events(id),
    destination TEXT NOT NULL,
    config_digest TEXT NOT NULL CHECK(length(config_digest)=64),
    project_id TEXT NOT NULL,
    namespace TEXT NOT NULL,
    project_binding_digest TEXT NOT NULL CHECK(length(project_binding_digest)=64),
    source_ref TEXT NOT NULL CHECK(length(source_ref) BETWEEN 1 AND 4096),
    source_hash TEXT NOT NULL CHECK(length(source_hash)=64),
    content_hash TEXT NOT NULL CHECK(length(content_hash)=64),
    filename TEXT NOT NULL CHECK(length(filename) BETWEEN 1 AND 128),
    media_type TEXT NOT NULL,
    byte_count INTEGER NOT NULL CHECK(byte_count BETWEEN 1 AND 2097152),
    payload BLOB CHECK(payload IS NULL OR length(payload)=byte_count),
    created_at TEXT NOT NULL,
    terminal_at TEXT,
    purged_at TEXT,
    CHECK((payload IS NULL)=(purged_at IS NOT NULL))
);
CREATE INDEX telegram_notification_document_retention ON telegram_notification_documents(terminal_at,id) WHERE payload IS NOT NULL;
ALTER TABLE notification_events ADD COLUMN telegram_presentation TEXT;
ALTER TABLE notification_events ADD COLUMN telegram_presentation_hash TEXT CHECK(telegram_presentation_hash IS NULL OR length(telegram_presentation_hash)=64);
ALTER TABLE notification_events ADD COLUMN telegram_document_id TEXT REFERENCES telegram_notification_documents(id);
ALTER TABLE notification_events ADD COLUMN telegram_destination TEXT;
ALTER TABLE notification_events ADD COLUMN telegram_namespace TEXT;
ALTER TABLE notification_events ADD COLUMN telegram_project_binding_digest TEXT CHECK(telegram_project_binding_digest IS NULL OR length(telegram_project_binding_digest)=64);
ALTER TABLE notification_events ADD COLUMN telegram_config_digest TEXT CHECK(telegram_config_digest IS NULL OR length(telegram_config_digest)=64);
ALTER TABLE notification_events ADD COLUMN telegram_attempts INTEGER NOT NULL DEFAULT 0 CHECK(telegram_attempts BETWEEN 0 AND 2);
ALTER TABLE notification_events ADD COLUMN telegram_plain_fallback INTEGER NOT NULL DEFAULT 0 CHECK(telegram_plain_fallback IN (0,1));
ALTER TABLE notification_events ADD COLUMN telegram_uncertain INTEGER NOT NULL DEFAULT 0 CHECK(telegram_uncertain IN (0,1));
ALTER TABLE notification_events ADD COLUMN telegram_message_id INTEGER CHECK(telegram_message_id IS NULL OR telegram_message_id>0);
ALTER TABLE notification_events ADD COLUMN telegram_file_id TEXT CHECK(telegram_file_id IS NULL OR (length(telegram_file_id) BETWEEN 1 AND 256 AND telegram_document_id IS NOT NULL)) CHECK(telegram_message_id IS NULL OR telegram_document_id IS NULL OR telegram_file_id IS NOT NULL);

-- +migrate Down
ALTER TABLE notification_events DROP COLUMN telegram_file_id;
ALTER TABLE notification_events DROP COLUMN telegram_message_id;
ALTER TABLE notification_events DROP COLUMN telegram_uncertain;
ALTER TABLE notification_events DROP COLUMN telegram_plain_fallback;
ALTER TABLE notification_events DROP COLUMN telegram_attempts;
ALTER TABLE notification_events DROP COLUMN telegram_config_digest;
ALTER TABLE notification_events DROP COLUMN telegram_project_binding_digest;
ALTER TABLE notification_events DROP COLUMN telegram_namespace;
ALTER TABLE notification_events DROP COLUMN telegram_destination;
ALTER TABLE notification_events DROP COLUMN telegram_document_id;
ALTER TABLE notification_events DROP COLUMN telegram_presentation_hash;
ALTER TABLE notification_events DROP COLUMN telegram_presentation;
DROP TABLE telegram_notification_documents;
