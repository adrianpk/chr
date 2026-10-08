-- +migrate Up
ALTER TABLE indexed_tickets ADD COLUMN namespace text NOT NULL DEFAULT 'default';
ALTER TABLE indexed_tickets ADD COLUMN workflow_schema text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN workflow_revision integer NOT NULL DEFAULT 0;
ALTER TABLE indexed_tickets ADD COLUMN representation_ref text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN body_ref text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN execution_contract_hash text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN slice_set_id text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN slice_id text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN task_order integer NOT NULL DEFAULT 0;
ALTER TABLE indexed_tickets ADD COLUMN task_phase text NOT NULL DEFAULT '';
ALTER TABLE indexed_tickets ADD COLUMN slice_delivery_ref text NOT NULL DEFAULT '';
DROP INDEX indexed_tickets_project_key_idx;
CREATE UNIQUE INDEX indexed_tickets_scoped_key_idx ON indexed_tickets(project_id, namespace, key);
CREATE UNIQUE INDEX indexed_tickets_task_order_idx ON indexed_tickets(project_id, namespace, slice_id, task_order)
    WHERE slice_id <> '' AND workflow_revision > 0;
CREATE INDEX indexed_tickets_workflow_phase_idx ON indexed_tickets(project_id, namespace, task_phase, slice_set_id, slice_id, task_order);

CREATE TABLE workflow_sets (
    project_id text NOT NULL, namespace text NOT NULL, id text NOT NULL,
    path text NOT NULL, revision integer NOT NULL, content_hash text NOT NULL,
    use_case text NOT NULL, use_case_order integer NOT NULL, created_at text NOT NULL,
    status text NOT NULL, closure_ref text NOT NULL,
    PRIMARY KEY(project_id, namespace, id), UNIQUE(project_id, namespace, path)
);
CREATE UNIQUE INDEX workflow_sets_order_idx ON workflow_sets(project_id, namespace, use_case, use_case_order)
    WHERE use_case <> '' AND use_case_order > 0;
CREATE TABLE workflow_slices (
    project_id text NOT NULL, namespace text NOT NULL, id text NOT NULL,
    set_id text NOT NULL, slice_order integer NOT NULL, phase text NOT NULL,
    pull_request text NOT NULL, delivery_ref text NOT NULL,
    PRIMARY KEY(project_id, namespace, id), UNIQUE(project_id, namespace, set_id, slice_order)
);
CREATE UNIQUE INDEX workflow_slices_pr_idx ON workflow_slices(project_id, namespace, pull_request) WHERE pull_request <> '';
CREATE TABLE workflow_receipts (
    project_id text NOT NULL, namespace text NOT NULL, id text NOT NULL,
    path text NOT NULL, revision integer NOT NULL, content_hash text NOT NULL,
    candidate_id text NOT NULL, submission_id text NOT NULL, unit_kind text NOT NULL,
    unit_id text NOT NULL, pull_request text NOT NULL, confirmed_at text NOT NULL,
    PRIMARY KEY(project_id, namespace, id), UNIQUE(project_id, namespace, candidate_id, submission_id)
);
CREATE TABLE workflow_snapshots (
    project_id text NOT NULL, namespace text NOT NULL, source_hash text NOT NULL,
    PRIMARY KEY(project_id, namespace)
);

-- +migrate Down
DROP TABLE workflow_snapshots;
DROP TABLE workflow_receipts;
DROP TABLE workflow_slices;
DROP TABLE workflow_sets;
DROP INDEX indexed_tickets_workflow_phase_idx;
DROP INDEX indexed_tickets_task_order_idx;
DROP INDEX indexed_tickets_scoped_key_idx;
ALTER TABLE indexed_tickets DROP COLUMN slice_delivery_ref;
ALTER TABLE indexed_tickets DROP COLUMN task_phase;
ALTER TABLE indexed_tickets DROP COLUMN task_order;
ALTER TABLE indexed_tickets DROP COLUMN slice_id;
ALTER TABLE indexed_tickets DROP COLUMN slice_set_id;
ALTER TABLE indexed_tickets DROP COLUMN execution_contract_hash;
ALTER TABLE indexed_tickets DROP COLUMN body_ref;
ALTER TABLE indexed_tickets DROP COLUMN representation_ref;
ALTER TABLE indexed_tickets DROP COLUMN workflow_revision;
ALTER TABLE indexed_tickets DROP COLUMN workflow_schema;
ALTER TABLE indexed_tickets DROP COLUMN namespace;
CREATE UNIQUE INDEX indexed_tickets_project_key_idx ON indexed_tickets(project_id, key);
