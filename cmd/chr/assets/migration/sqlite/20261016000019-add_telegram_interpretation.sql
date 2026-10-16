-- +migrate Up
ALTER TABLE telegram_updates ADD COLUMN interpretation_attempt INTEGER NOT NULL DEFAULT 0 CHECK(interpretation_attempt BETWEEN 0 AND 5);
ALTER TABLE telegram_updates ADD COLUMN interpretation_retry_at TEXT;
ALTER TABLE telegram_updates ADD COLUMN interpretation_error_code TEXT NOT NULL DEFAULT '' CHECK(length(interpretation_error_code)<=64);
ALTER TABLE telegram_updates ADD COLUMN interpretation TEXT CHECK(length(CAST(interpretation AS BLOB))<=65536);

-- +migrate Down
ALTER TABLE telegram_updates DROP COLUMN interpretation;
ALTER TABLE telegram_updates DROP COLUMN interpretation_error_code;
ALTER TABLE telegram_updates DROP COLUMN interpretation_retry_at;
ALTER TABLE telegram_updates DROP COLUMN interpretation_attempt;
