-- +migrate Up
ALTER TABLE telegram_receivers ADD COLUMN retry_actor TEXT NOT NULL DEFAULT '' CHECK(length(retry_actor)<=128);
ALTER TABLE telegram_receivers ADD COLUMN retry_started_at TEXT;
ALTER TABLE telegram_receivers ADD COLUMN retention_gap_at TEXT;
ALTER TABLE telegram_receivers ADD COLUMN retention_gap_from_poll_at TEXT;
ALTER TABLE telegram_outbox ADD COLUMN retry_actor TEXT NOT NULL DEFAULT '' CHECK(length(retry_actor)<=128);
ALTER TABLE telegram_outbox ADD COLUMN retry_started_at TEXT;
ALTER TABLE telegram_conversations ADD COLUMN compacted_at TEXT;
ALTER TABLE telegram_requests ADD COLUMN compacted_at TEXT;

-- +migrate Down
ALTER TABLE telegram_requests DROP COLUMN compacted_at;
ALTER TABLE telegram_conversations DROP COLUMN compacted_at;
ALTER TABLE telegram_outbox DROP COLUMN retry_started_at;
ALTER TABLE telegram_outbox DROP COLUMN retry_actor;
ALTER TABLE telegram_receivers DROP COLUMN retention_gap_from_poll_at;
ALTER TABLE telegram_receivers DROP COLUMN retention_gap_at;
ALTER TABLE telegram_receivers DROP COLUMN retry_started_at;
ALTER TABLE telegram_receivers DROP COLUMN retry_actor;
