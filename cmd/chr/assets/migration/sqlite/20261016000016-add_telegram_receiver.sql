-- +migrate Up
CREATE TABLE telegram_receivers (
    bot_id INTEGER PRIMARY KEY CHECK(bot_id > 0),
    schema_version INTEGER NOT NULL DEFAULT 1 CHECK(schema_version > 0),
    config_digest TEXT NOT NULL CHECK(length(config_digest) = 64),
    cursor_epoch INTEGER NOT NULL DEFAULT 1 CHECK(cursor_epoch > 0),
    next_update_id INTEGER CHECK(next_update_id >= 0),
    retired_update_ranges TEXT NOT NULL DEFAULT '[]',
    health TEXT NOT NULL DEFAULT 'starting' CHECK(health IN ('starting','ready','degraded','stopped','blocked')),
    last_poll_at TEXT,
    retry_attempt INTEGER NOT NULL DEFAULT 0 CHECK(retry_attempt BETWEEN 0 AND 5),
    retry_at TEXT,
    retry_cycle INTEGER NOT NULL DEFAULT 1 CHECK(retry_cycle > 0),
    total_retry_attempts INTEGER NOT NULL DEFAULT 0 CHECK(total_retry_attempts >= 0),
    last_error_code TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL
);
CREATE TABLE telegram_updates (
    bot_id INTEGER NOT NULL REFERENCES telegram_receivers(bot_id),
    update_id INTEGER NOT NULL CHECK(update_id >= 0),
    cursor_epoch INTEGER NOT NULL CHECK(cursor_epoch > 0),
    payload_digest TEXT NOT NULL CHECK(length(payload_digest) = 64),
    update_kind TEXT NOT NULL CHECK(update_kind IN ('message','callback_query','ignored','rejected')),
    chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL CHECK(topic_id >= 0),
    user_id INTEGER NOT NULL, message_id INTEGER NOT NULL,
    chat_type TEXT NOT NULL,
    received_at TEXT NOT NULL,
    payload TEXT NOT NULL CHECK(length(CAST(payload AS BLOB)) <= 16384),
    state TEXT NOT NULL CHECK(state IN ('pending','processing','processed','ignored','rejected')),
    reason_code TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY(bot_id, update_id)
);
CREATE INDEX telegram_updates_pending ON telegram_updates(bot_id, state, update_id);
CREATE TABLE telegram_outbox (
    id TEXT PRIMARY KEY CHECK(length(id) <= 128),
    bot_id INTEGER NOT NULL REFERENCES telegram_receivers(bot_id),
    chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL CHECK(topic_id >= 0), user_id INTEGER NOT NULL,
    event_key TEXT NOT NULL,
    config_digest TEXT NOT NULL CHECK(length(config_digest) = 64),
    kind TEXT NOT NULL CHECK(kind IN ('acknowledgement','question','progress','result','diagnostic')),
    part_index INTEGER NOT NULL DEFAULT 1 CHECK(part_index BETWEEN 1 AND 8),
    part_count INTEGER NOT NULL DEFAULT 1 CHECK(part_count BETWEEN 1 AND 8),
    text TEXT NOT NULL CHECK(length(text) BETWEEN 1 AND 3900),
    state TEXT NOT NULL CHECK(state IN ('pending','sending','delivered','uncertain','blocked')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5),
    retry_at TEXT,
    retry_cycle INTEGER NOT NULL DEFAULT 1 CHECK(retry_cycle > 0),
    total_attempts INTEGER NOT NULL DEFAULT 0 CHECK(total_attempts >= 0),
    telegram_message_id INTEGER,
    last_error_code TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
    UNIQUE(bot_id, chat_id, topic_id, user_id, event_key, part_index)
);
CREATE INDEX telegram_outbox_due ON telegram_outbox(bot_id, state, retry_at);

-- +migrate Down
DROP TABLE telegram_outbox;
DROP TABLE telegram_updates;
DROP TABLE telegram_receivers;
