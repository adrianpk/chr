-- +migrate Up
CREATE TABLE execution_records (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    ticket_key text NOT NULL,
    ticket_path text NOT NULL,
    status text NOT NULL,
    runtime_image text NOT NULL,
    runtime_dockerfile_path text NOT NULL,
    runtime_container_workdir text NOT NULL,
    runtime_validation_command text NOT NULL,
    artifact_run_dir text NOT NULL,
    artifact_result_path text NOT NULL,
    artifact_log_path text NOT NULL,
    artifact_validation_path text NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    started_at text,
    finished_at text,
    updated_at text NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX execution_records_project_ticket_idx
ON execution_records (project_id, ticket_key);

CREATE INDEX execution_records_project_status_idx
ON execution_records (project_id, status);

-- +migrate Down
DROP TABLE IF EXISTS execution_records;
