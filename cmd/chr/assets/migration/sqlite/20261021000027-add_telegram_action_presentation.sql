-- +migrate Up
ALTER TABLE telegram_actions ADD COLUMN guard_version INTEGER NOT NULL DEFAULT 0 CHECK(guard_version>=0);
ALTER TABLE telegram_actions ADD COLUMN presentation_hash TEXT CHECK(presentation_hash IS NULL OR length(presentation_hash)=64);

-- +migrate Down
ALTER TABLE telegram_actions DROP COLUMN presentation_hash;
ALTER TABLE telegram_actions DROP COLUMN guard_version;
