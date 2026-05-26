-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN synopsis text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN board_column text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN source_kind text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN source_ref text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN slice text NOT NULL DEFAULT '';

CREATE INDEX indexed_tickets_project_board_column_idx
ON indexed_tickets (project_id, board_column, priority, severity);

CREATE INDEX indexed_tickets_project_use_case_idx
ON indexed_tickets (project_id, use_case);

CREATE INDEX indexed_tickets_project_area_idx
ON indexed_tickets (project_id, project_area);

CREATE INDEX indexed_tickets_project_updated_idx
ON indexed_tickets (project_id, updated_at);

CREATE INDEX indexed_tickets_project_source_kind_idx
ON indexed_tickets (project_id, source_kind);

CREATE TABLE ticket_key_sequences (
    project_id text NOT NULL,
    prefix text NOT NULL,
    next_number integer NOT NULL,
    updated_at text NOT NULL,
    PRIMARY KEY (project_id, prefix),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS ticket_key_sequences;
DROP INDEX IF EXISTS indexed_tickets_project_source_kind_idx;
DROP INDEX IF EXISTS indexed_tickets_project_updated_idx;
DROP INDEX IF EXISTS indexed_tickets_project_area_idx;
DROP INDEX IF EXISTS indexed_tickets_project_use_case_idx;
DROP INDEX IF EXISTS indexed_tickets_project_board_column_idx;
