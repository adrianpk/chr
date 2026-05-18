-- +migrate Up
CREATE TABLE schema_migrations (
    version text PRIMARY KEY,
    name text NOT NULL,
    applied_at text NOT NULL
);

CREATE TABLE projects (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE,
    path text NOT NULL UNIQUE,
    enabled integer NOT NULL,
    registered_at text NOT NULL,
    last_indexed_at text
);

-- +migrate Down
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS schema_migrations;
