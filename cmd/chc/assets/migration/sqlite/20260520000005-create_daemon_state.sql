-- +migrate Up
CREATE TABLE daemon_state (
    id text PRIMARY KEY,
    status text NOT NULL,
    pid integer NOT NULL,
    started_at text,
    heartbeat_at text,
    stop_requested_at text,
    failure_reason text NOT NULL,
    updated_at text NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS daemon_state;
