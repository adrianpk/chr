-- +migrate Up
CREATE TABLE indexed_project_metadata (
    project_id text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    context text NOT NULL,
    source_path text NOT NULL,
    source_hash text NOT NULL,
    schema_version integer NOT NULL,
    indexed_at text NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX indexed_project_metadata_name_idx
ON indexed_project_metadata (name);

-- +migrate Down
DROP TABLE IF EXISTS indexed_project_metadata;
