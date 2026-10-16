-- +migrate Up
ALTER TABLE telegram_conversations ADD COLUMN first_authoring_work_request TEXT REFERENCES telegram_requests(id);

-- +migrate Down
ALTER TABLE telegram_conversations DROP COLUMN first_authoring_work_request;
