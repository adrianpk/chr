-- +migrate Up
CREATE TABLE notification_events (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    type text NOT NULL,
    severity text NOT NULL,
    status text NOT NULL,
    source text NOT NULL,
    channel text NOT NULL,
    ticket_key text NOT NULL,
    execution_id text NOT NULL,
    loop_run_id text NOT NULL,
    daemon_pid integer NOT NULL,
    summary text NOT NULL,
    detail text NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    delivered_at text,
    updated_at text NOT NULL
);

CREATE INDEX notification_events_project_created_idx
ON notification_events (project_id, created_at);

CREATE INDEX notification_events_project_status_idx
ON notification_events (project_id, status);

-- +migrate Down
DROP TABLE IF EXISTS notification_events;
