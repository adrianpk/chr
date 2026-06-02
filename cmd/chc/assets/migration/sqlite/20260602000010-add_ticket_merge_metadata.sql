-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN merge_state text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_result text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_actor text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_commit text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merged_at text DEFAULT NULL;

ALTER TABLE indexed_tickets
ADD COLUMN merge_blocker text NOT NULL DEFAULT '';

CREATE INDEX indexed_tickets_project_merge_state_idx
ON indexed_tickets (project_id, merge_state);

-- +migrate Down
DROP INDEX IF EXISTS indexed_tickets_project_merge_state_idx;
