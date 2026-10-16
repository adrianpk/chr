-- +migrate Up
ALTER TABLE telegram_updates ADD COLUMN processing_lane TEXT NOT NULL DEFAULT 'ordinary' CHECK(processing_lane IN ('ordinary','decision','urgent'));
ALTER TABLE telegram_updates ADD COLUMN control_input INTEGER NOT NULL DEFAULT 0 CHECK(control_input IN (0,1));
CREATE INDEX telegram_updates_lane ON telegram_updates(bot_id,state,processing_lane,update_id);
CREATE INDEX telegram_updates_peer_lane ON telegram_updates(bot_id,chat_id,topic_id,user_id,state,processing_lane,update_id);

-- +migrate Down
DROP INDEX telegram_updates_peer_lane;
DROP INDEX telegram_updates_lane;
ALTER TABLE telegram_updates DROP COLUMN control_input;
ALTER TABLE telegram_updates DROP COLUMN processing_lane;
