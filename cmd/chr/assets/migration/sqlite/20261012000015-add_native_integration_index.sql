-- +migrate Up
CREATE TABLE workflow_integration_entries (
    project_id text NOT NULL, coordinator_id text NOT NULL, namespace text NOT NULL,
    entry_id text NOT NULL, invocation_id text NOT NULL, unit_kind text NOT NULL,
    unit_id text NOT NULL, repository_id text NOT NULL, target_ref text NOT NULL,
    state text NOT NULL, queue_sequence text NOT NULL, image_hash text NOT NULL,
    PRIMARY KEY(project_id, coordinator_id, entry_id)
);
CREATE INDEX workflow_integration_target_idx ON workflow_integration_entries
    (project_id, coordinator_id, repository_id, target_ref, state, queue_sequence, entry_id);
CREATE INDEX workflow_integration_unit_idx ON workflow_integration_entries
    (project_id, coordinator_id, namespace, invocation_id, unit_kind, unit_id);
CREATE TABLE workflow_integration_snapshots (
    project_id text NOT NULL, coordinator_id text NOT NULL, source_hash text NOT NULL,
    PRIMARY KEY(project_id, coordinator_id)
);

-- +migrate Down
DROP TABLE workflow_integration_snapshots;
DROP TABLE workflow_integration_entries;
