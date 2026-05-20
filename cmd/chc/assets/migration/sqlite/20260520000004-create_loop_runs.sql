-- +migrate Up
CREATE TABLE loop_runs (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    status text NOT NULL,
    eligible_count integer NOT NULL,
    processed_count integer NOT NULL,
    blocked_count integer NOT NULL,
    failed_count integer NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    started_at text,
    finished_at text,
    updated_at text NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX loop_runs_project_status_idx
ON loop_runs (project_id, status);

CREATE INDEX loop_runs_project_created_idx
ON loop_runs (project_id, created_at);

-- +migrate Down
DROP TABLE IF EXISTS loop_runs;
