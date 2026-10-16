-- +migrate Up
CREATE TABLE telegram_peers (
 bot_id INTEGER NOT NULL REFERENCES telegram_receivers(bot_id),
 chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
 selected_project_id TEXT, focused_conversation_id TEXT,
 revision INTEGER NOT NULL DEFAULT 1 CHECK(revision>0), updated_at TEXT NOT NULL,
 PRIMARY KEY(bot_id,chat_id,topic_id,user_id)
);
CREATE TABLE telegram_conversations (
 id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL DEFAULT 1 CHECK(schema_version=1),
 bot_id INTEGER NOT NULL, chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
 project_id TEXT NOT NULL, project_binding_digest TEXT NOT NULL, namespace TEXT NOT NULL,
 intent TEXT NOT NULL CHECK(intent IN ('explain','prepare','develop')),
 phase TEXT NOT NULL CHECK(phase IN ('collecting','clarifying','preparing','prepared','awaiting_confirmation','executing','waiting','completed','canceled','failed')),
 topic TEXT NOT NULL CHECK(length(CAST(topic AS BLOB))<=512),
 authoring_session_ref TEXT, invocation_id TEXT, question_binding TEXT NOT NULL DEFAULT '',
 capture TEXT NOT NULL DEFAULT '' CHECK(length(CAST(capture AS BLOB))<=65536),
 authorization_ref TEXT, last_observed_revision INTEGER,
 revision INTEGER NOT NULL DEFAULT 1 CHECK(revision>0),
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL, finished_at TEXT,
 FOREIGN KEY(bot_id,chat_id,topic_id,user_id) REFERENCES telegram_peers(bot_id,chat_id,topic_id,user_id)
);
CREATE INDEX telegram_conversation_peer ON telegram_conversations(bot_id,chat_id,topic_id,user_id,phase,updated_at);
CREATE INDEX telegram_conversation_project ON telegram_conversations(project_id,phase);
CREATE INDEX telegram_conversation_invocation ON telegram_conversations(invocation_id);
CREATE TABLE telegram_capture_parts (
 conversation_id TEXT NOT NULL REFERENCES telegram_conversations(id),
 bot_id INTEGER NOT NULL, update_id INTEGER NOT NULL, text_digest TEXT NOT NULL,
 byte_start INTEGER NOT NULL, byte_end INTEGER NOT NULL,
 PRIMARY KEY(conversation_id,bot_id,update_id),
 FOREIGN KEY(bot_id,update_id) REFERENCES telegram_updates(bot_id,update_id)
);
CREATE TABLE telegram_replies (
 bot_id INTEGER NOT NULL, chat_id INTEGER NOT NULL, topic_id INTEGER NOT NULL, message_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
 conversation_id TEXT NOT NULL REFERENCES telegram_conversations(id), request_id TEXT,
 purpose TEXT NOT NULL, conversation_revision INTEGER NOT NULL, question_binding TEXT NOT NULL DEFAULT '',
 created_at TEXT NOT NULL, invalidated_at TEXT,
 PRIMARY KEY(bot_id,chat_id,topic_id,message_id,user_id)
);
CREATE INDEX telegram_reply_conversation ON telegram_replies(conversation_id,invalidated_at,created_at);
ALTER TABLE telegram_updates ADD COLUMN reply_to_message_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE telegram_outbox ADD COLUMN conversation_id TEXT REFERENCES telegram_conversations(id);
ALTER TABLE telegram_outbox ADD COLUMN conversation_revision INTEGER;
ALTER TABLE telegram_outbox ADD COLUMN question_binding TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE telegram_outbox DROP COLUMN question_binding;
ALTER TABLE telegram_outbox DROP COLUMN conversation_revision;
ALTER TABLE telegram_outbox DROP COLUMN conversation_id;
ALTER TABLE telegram_updates DROP COLUMN reply_to_message_id;
DROP TABLE telegram_replies;
DROP TABLE telegram_capture_parts;
DROP TABLE telegram_conversations;
DROP TABLE telegram_peers;
