-- +migrate Up
ALTER TABLE execution_records
ADD COLUMN workspace_mode text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_project_checkout_dir text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_execution_dir text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_base_ref text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_base_sha text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_ticket_branch text NOT NULL DEFAULT '';

-- +migrate Down
SELECT 1;
