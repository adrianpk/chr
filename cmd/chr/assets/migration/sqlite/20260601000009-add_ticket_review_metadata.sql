-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN branch text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_refs text NOT NULL DEFAULT '[]';

ALTER TABLE indexed_tickets
ADD COLUMN review_state text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_round integer NOT NULL DEFAULT 0;

ALTER TABLE indexed_tickets
ADD COLUMN review_result text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_comments_url text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN blocking_findings text NOT NULL DEFAULT '[]';

CREATE INDEX indexed_tickets_project_review_state_idx
ON indexed_tickets (project_id, review_state);

CREATE INDEX indexed_tickets_project_branch_idx
ON indexed_tickets (project_id, branch);

-- +migrate Down
DROP INDEX IF EXISTS indexed_tickets_project_branch_idx;
DROP INDEX IF EXISTS indexed_tickets_project_review_state_idx;
