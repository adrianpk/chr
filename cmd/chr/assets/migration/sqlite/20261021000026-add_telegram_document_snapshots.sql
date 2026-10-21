-- +migrate Up
CREATE TABLE telegram_document_snapshots (
    id TEXT PRIMARY KEY,
    owner_kind TEXT NOT NULL CHECK(owner_kind='receiver'),
    owner_id TEXT NOT NULL UNIQUE,
    bot_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    topic_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
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
CREATE INDEX telegram_document_retention ON telegram_document_snapshots(terminal_at,id) WHERE payload IS NOT NULL;
ALTER TABLE telegram_outbox ADD COLUMN document_id TEXT REFERENCES telegram_document_snapshots(id);
ALTER TABLE telegram_outbox ADD COLUMN telegram_file_id TEXT CHECK(telegram_file_id IS NULL OR length(telegram_file_id) BETWEEN 1 AND 256);
ALTER TABLE telegram_outbox ADD COLUMN presentation_hash TEXT CHECK(presentation_hash IS NULL OR length(presentation_hash)=64);

-- +migrate Down
ALTER TABLE telegram_outbox DROP COLUMN presentation_hash;
ALTER TABLE telegram_outbox DROP COLUMN telegram_file_id;
ALTER TABLE telegram_outbox DROP COLUMN document_id;
DROP TABLE telegram_document_snapshots;
