-- +migrate Up
CREATE TABLE indexed_tickets (
    project_id text NOT NULL,
    path text NOT NULL,
    relative_path text NOT NULL,
    key text NOT NULL,
    title text NOT NULL,
    status text NOT NULL,
    priority text NOT NULL,
    severity text NOT NULL,
    type text NOT NULL,
    use_case text NOT NULL,
    project_area text NOT NULL,
    created_at text,
    updated_at text,
    file_mtime text NOT NULL,
    file_size integer NOT NULL,
    content_hash text NOT NULL,
    indexed_at text NOT NULL,
    PRIMARY KEY (project_id, path),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX indexed_tickets_project_key_idx
ON indexed_tickets (project_id, key);

CREATE INDEX indexed_tickets_project_status_idx
ON indexed_tickets (project_id, status);

CREATE INDEX indexed_tickets_project_type_idx
ON indexed_tickets (project_id, type);

CREATE TABLE indexed_ticket_tags (
    project_id text NOT NULL,
    ticket_path text NOT NULL,
    tag text NOT NULL,
    PRIMARY KEY (project_id, ticket_path, tag),
    FOREIGN KEY (project_id, ticket_path)
        REFERENCES indexed_tickets(project_id, path)
        ON DELETE CASCADE
);

CREATE INDEX indexed_ticket_tags_project_tag_idx
ON indexed_ticket_tags (project_id, tag);

-- +migrate Down
DROP TABLE IF EXISTS indexed_ticket_tags;
DROP TABLE IF EXISTS indexed_tickets;
