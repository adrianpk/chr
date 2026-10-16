-- +migrate Up
ALTER TABLE telegram_requests ADD COLUMN model_attempt INTEGER NOT NULL DEFAULT 0 CHECK(model_attempt BETWEEN 0 AND 5);
ALTER TABLE telegram_requests ADD COLUMN model_retry_at TEXT;
ALTER TABLE telegram_requests ADD COLUMN model_error_code TEXT NOT NULL DEFAULT '' CHECK(length(model_error_code)<=64);
ALTER TABLE telegram_requests ADD COLUMN model_result TEXT CHECK(length(CAST(model_result AS BLOB))<=65536);

-- +migrate Down
ALTER TABLE telegram_requests DROP COLUMN model_result;
ALTER TABLE telegram_requests DROP COLUMN model_error_code;
ALTER TABLE telegram_requests DROP COLUMN model_retry_at;
ALTER TABLE telegram_requests DROP COLUMN model_attempt;
