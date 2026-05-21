-- +migrate Up
CREATE TABLE remote_command_records (
    id text PRIMARY KEY,
    status text NOT NULL,
    channel text NOT NULL,
    sender_external_id text NOT NULL,
    sender_display_name text NOT NULL,
    project_selector text NOT NULL,
    intent text NOT NULL,
    ticket_key text NOT NULL,
    comment text NOT NULL,
    loop_statuses text NOT NULL,
    loop_priorities text NOT NULL,
    loop_severities text NOT NULL,
    loop_project_areas text NOT NULL,
    loop_tags text NOT NULL,
    loop_types text NOT NULL,
    loop_use_cases text NOT NULL,
    loop_max_iterations integer NOT NULL,
    raw_text text NOT NULL,
    handoff_command text NOT NULL,
    rejection_reason text NOT NULL,
    received_at text NOT NULL,
    created_at text NOT NULL
);

CREATE INDEX remote_command_records_created_idx
ON remote_command_records (created_at);

CREATE INDEX remote_command_records_status_idx
ON remote_command_records (status);

-- +migrate Down
DROP TABLE IF EXISTS remote_command_records;
