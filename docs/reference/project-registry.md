# Project Registry

## Purpose

The project registry stores local projects known by the user daemon.

It is user-level state. It does not replace the project-local `.chc/` directory and does not make project ticket files non-portable.

## Database

The registry uses the global user SQLite database:

```text
<user-data-dir>/chc.db
```

The user data directory comes from `chc.UserEnvironment`.

## Schema

The registry stores projects in the `projects` table.

```sql
CREATE TABLE projects (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE,
    path text NOT NULL UNIQUE,
    enabled integer NOT NULL,
    registered_at text NOT NULL,
    last_indexed_at text
);
```

The schema is created through embedded SQLite migrations before registry operations run.

## Core API

The core registry API is exposed through `chc.ProjectRegistry`.

Initial operations:

- add a project;
- list projects;
- get a project by name.

CLI commands are intentionally separate from the core repository and will be added in a later slice.
