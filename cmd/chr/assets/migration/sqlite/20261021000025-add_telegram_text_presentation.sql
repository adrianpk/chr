-- +migrate Up
ALTER TABLE telegram_outbox ADD COLUMN presentation TEXT CHECK(presentation IS NULL OR (json_valid(presentation) AND length(CAST(presentation AS BLOB))<=32768));
ALTER TABLE telegram_outbox ADD COLUMN presentation_fallback INTEGER NOT NULL DEFAULT 0 CHECK(presentation_fallback IN (0,1) AND (presentation_fallback=0 OR presentation IS NOT NULL));

-- +migrate Down
ALTER TABLE telegram_outbox DROP COLUMN presentation_fallback;
ALTER TABLE telegram_outbox DROP COLUMN presentation;
