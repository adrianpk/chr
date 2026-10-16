-- +migrate Up
CREATE TABLE telegram_requests (
 id TEXT PRIMARY KEY CHECK(length(id) BETWEEN 1 AND 128),
 bot_id INTEGER NOT NULL, update_id INTEGER NOT NULL,
 operation_ordinal INTEGER NOT NULL CHECK(operation_ordinal BETWEEN 0 AND 7),
 conversation_id TEXT REFERENCES telegram_conversations(id),
 operator_id TEXT NOT NULL CHECK(length(operator_id) BETWEEN 1 AND 128),
 project_id TEXT NOT NULL, project_binding_digest TEXT NOT NULL CHECK(length(project_binding_digest)=64), namespace TEXT NOT NULL,
 operation_id TEXT NOT NULL CHECK(length(operation_id) BETWEEN 1 AND 128),
 arguments TEXT NOT NULL CHECK(length(CAST(arguments AS BLOB))<=65536),
 argument_digest TEXT NOT NULL CHECK(length(argument_digest)=64),
 expected_conversation_revision INTEGER,
 domain_guard TEXT NOT NULL DEFAULT '',
 prepared_input TEXT CHECK(length(CAST(prepared_input AS BLOB))<=1048576),
 state TEXT NOT NULL CHECK(state IN ('prepared','awaiting_confirmation','ready','applying','confirmed','rejected','blocked','reconcile_required')),
 domain_receipt_ref TEXT,
 result TEXT CHECK(length(CAST(result AS BLOB))<=65536),
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
 UNIQUE(bot_id,update_id,operation_ordinal),
 FOREIGN KEY(bot_id,update_id) REFERENCES telegram_updates(bot_id,update_id)
);
CREATE INDEX telegram_request_pending ON telegram_requests(bot_id,state,updated_at);
CREATE INDEX telegram_request_conversation ON telegram_requests(conversation_id,created_at);
CREATE INDEX telegram_request_operation ON telegram_requests(project_id,operation_id);
CREATE TABLE telegram_actions (
 token TEXT PRIMARY KEY CHECK(length(token)=32),
 request_id TEXT NOT NULL REFERENCES telegram_requests(id),
 bot_id INTEGER NOT NULL, chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
 conversation_revision INTEGER NOT NULL CHECK(conversation_revision>0),
 guard_digest TEXT NOT NULL CHECK(length(guard_digest)=64),
 allowed_choices TEXT NOT NULL CHECK(length(CAST(allowed_choices AS BLOB))<=1024),
 state TEXT NOT NULL CHECK(state IN ('pending','applying','consumed','expired','superseded')),
 expires_at TEXT NOT NULL,
 consumed_by_request TEXT UNIQUE REFERENCES telegram_requests(id),
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE INDEX telegram_action_pending ON telegram_actions(bot_id,state,expires_at);
ALTER TABLE telegram_outbox ADD COLUMN request_id TEXT REFERENCES telegram_requests(id);
ALTER TABLE telegram_outbox ADD COLUMN action_token TEXT REFERENCES telegram_actions(token);
ALTER TABLE telegram_replies ADD COLUMN action_token TEXT REFERENCES telegram_actions(token);

-- +migrate Down
ALTER TABLE telegram_replies DROP COLUMN action_token;
ALTER TABLE telegram_outbox DROP COLUMN action_token;
ALTER TABLE telegram_outbox DROP COLUMN request_id;
DROP TABLE telegram_actions;
DROP TABLE telegram_requests;
