-- +migrate Up
CREATE TABLE indexed_cases (
    project_id text NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    context text NOT NULL,
    source_path text NOT NULL,
    source_hash text NOT NULL,
    schema_version integer NOT NULL,
    indexed_at text NOT NULL,
    PRIMARY KEY (project_id, slug),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX indexed_cases_project_name_idx
ON indexed_cases (project_id, name);

-- +migrate Down
DROP TABLE IF EXISTS indexed_cases;
