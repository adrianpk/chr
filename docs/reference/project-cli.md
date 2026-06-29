# Project CLI

## Purpose

Project commands manage the user-level registry of local projects.

The registry is used by future daemon, indexing, and messaging workflows. It does not replace project-local `.chr/` files.

## Add

Register a project:

```sh
chr project add /path/to/project
```

The project must already contain a `.chr/` directory.

By default, the registered project name is the basename of the project path. Override it with:

```sh
chr project add --name my-project /path/to/project
```

## List

List registered projects:

```sh
chr project list
```

Rows are tab-separated:

```text
name	path	enabled	id
```

## Development Root

Project commands support a development root:

```sh
chr project add --dev-root ._workspace /path/to/project
chr project list --dev-root ._workspace
```

The development root redirects user-level config, data, cache, and state paths away from real user directories.
