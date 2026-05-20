# Semantic Code Context

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`

## Purpose

Define how `chc` should derive technical context from a repository before and
during ticket execution.

The goal is to provide the agent with relevant code abstractions without dumping
the entire repository into the prompt.

## Context Sources

### Filesystem Tree

The filesystem tree gives a broad structural view.

It can show:

- top-level directories;
- package boundaries;
- configuration files;
- tests;
- docs;
- migrations;
- generated assets.

The tree is useful for orientation but insufficient for semantic understanding.

### Text Search

Text search is useful for:

- ticket key references;
- error strings;
- command names;
- config keys;
- function or type names;
- docs and tests.

Text search is fast and should remain a first-pass tool.

### AST and Syntax Trees

AST-based extraction can identify:

- declarations;
- functions;
- methods;
- structs and interfaces;
- call sites;
- imports;
- comments attached to declarations.

For Go code, the standard `go/ast`, `go/parser`, and `go/types` packages can
support this layer. For package-level loading with types and module context,
`golang.org/x/tools/go/packages` is the likely integration point.

Tree-sitter remains useful if multi-language parsing becomes necessary.

### LSP and `gopls`

`gopls` can provide editor-grade answers such as:

- definition lookup;
- references;
- implementation lookup;
- diagnostics;
- symbol search.

This is useful when the repository shape is complex or when type-aware answers
are required quickly.

### `gograph`

`gograph` is a candidate code graph context provider.

It can help answer:

- where a symbol is defined;
- where a symbol is referenced;
- which package owns a behavior;
- what tests touch a package or function;
- what changes may be impacted by a target edit.

`gograph` may be used through a CLI or MCP integration. The CLI path is simpler
as an initial integration; MCP may be useful when agent tooling needs direct
interactive graph queries.

## When to Query Semantic Context

The context builder should query semantic code context when:

- a ticket names a function, type, package, command, endpoint, or error string;
- a ticket requires persistence or migration changes;
- a ticket asks for behavior that likely spans packages;
- a validation failure points to a symbol or package;
- a review or Ralph Loop iteration needs impact analysis;
- a change may require tests in multiple locations.

## Query Strategies

If the ticket mentions a function:

- locate the declaration;
- list callers;
- list direct tests;
- include nearby types and interfaces.

If the ticket mentions an endpoint or command:

- locate command/route registration;
- locate handler or command implementation;
- locate tests and docs for the surface.

If the ticket mentions persistence:

- locate models;
- locate repositories;
- locate migrations;
- locate tests and docs for the schema.

If the ticket mentions a package or area:

- summarize package responsibilities;
- list exported symbols;
- locate package tests;
- find recent docs or specs that mention the area.

If the ticket mentions an error string:

- locate the literal;
- inspect the surrounding control flow;
- inspect tests that assert the error.

## Context Budget

Semantic code context must be bounded.

The context builder should prefer:

- summaries over full files;
- declarations over complete implementations when enough;
- direct references before transitive references;
- tests that assert the relevant behavior;
- links or paths when full content is not needed.

## Source of Truth

Semantic tools are not the source of truth.

The repository files, tests, and validation commands remain authoritative.
Graph, AST, LSP, and search output are context aids that can be stale,
incomplete, or misleading.

## Open Questions

- Should `gograph` be integrated as a CLI first or as MCP first?
- Which AST extraction output format should become durable artifact data?
- How should code context be cached between Ralph Loop iterations?
- What maximum context budget should apply before pruning begins?
