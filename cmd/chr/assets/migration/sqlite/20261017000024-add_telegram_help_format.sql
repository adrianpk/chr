-- +migrate Up
ALTER TABLE telegram_outbox ADD COLUMN plain_text TEXT NOT NULL DEFAULT '' CHECK(length(plain_text)<=3900);
ALTER TABLE telegram_outbox ADD COLUMN parse_mode TEXT NOT NULL DEFAULT '' CHECK(parse_mode='' OR (parse_mode='Markdown' AND length(plain_text)>0));
ALTER TABLE telegram_outbox ADD COLUMN plain_fallback INTEGER NOT NULL DEFAULT 0 CHECK(plain_fallback IN (0,1) AND (plain_fallback=0 OR parse_mode='Markdown'));

-- +migrate Down
ALTER TABLE telegram_outbox DROP COLUMN plain_fallback;
ALTER TABLE telegram_outbox DROP COLUMN parse_mode;
ALTER TABLE telegram_outbox DROP COLUMN plain_text;
