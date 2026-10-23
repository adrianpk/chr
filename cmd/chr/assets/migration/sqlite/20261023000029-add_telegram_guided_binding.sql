-- +migrate Up
ALTER TABLE telegram_conversations ADD COLUMN guided_run_id TEXT CHECK(guided_run_id IS NULL OR length(guided_run_id) BETWEEN 1 AND 128);
CREATE INDEX telegram_guided_bindings ON telegram_conversations(bot_id,id) WHERE guided_run_id IS NOT NULL;

-- +migrate Down
DROP INDEX telegram_guided_bindings;
ALTER TABLE telegram_conversations DROP COLUMN guided_run_id;
