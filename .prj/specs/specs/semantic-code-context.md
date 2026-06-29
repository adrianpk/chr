# Semantic Code Context

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chr_ticket_execution_context_and_ralph_loop.md`
- `https://www.bytesizego.com/blog/gograph-a-go-cli-that-parses-your-codebase-for-ai-agents`

## Purpose

Define how `chr` should derive technical context from a repository before and
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

### Go AST and Syntax Trees

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

For Go repositories, this should be the primary syntax layer. Tree-sitter is
not required for the initial Go implementation.

Tree-sitter remains useful if multi-language parsing becomes necessary. It
should be treated as a later multi-language capability, not as a dependency for
the first Go context composer.

### LSP and `gopls`

`gopls` can provide editor-grade answers such as:

- definition lookup;
- references;
- implementation lookup;
- diagnostics;
- symbol search.

This is useful when the repository shape is complex or when type-aware answers
are required quickly.

`gopls` should not be part of the first context composer implementation. It
adds lifecycle, cache, workspace, and versioning concerns that are not needed
until the simpler graph-based path has been proven.

### `gograph`

`gograph` is the preferred first code graph context provider for Go projects.

It can help answer:

- where a symbol is defined;
- where a symbol is referenced;
- which package owns a behavior;
- what tests touch a package or function;
- what changes may be impacted by a target edit.

`gograph` should be integrated through the CLI first. The initial integration
should run `gograph build .` before execution, then query focused context rather
than embedding the full graph in the prompt.

Useful initial commands include:

- `gograph context <symbol>`;
- `gograph source <symbol>`;
- `gograph callers <symbol>`;
- `gograph callees <symbol>`;
- `gograph tests <symbol-or-package>`;
- `gograph deps <package>`;
- `gograph changes`;
- `gograph hotspot --top 20`.

The default build mode should be the fast AST path because it can tolerate an
incomplete or actively edited repository. `gograph build . --precise` should be
reserved for cases where the repository compiles and the ticket needs stronger
type, interface, or callgraph answers.

The generated graph and reports are context infrastructure. They help the
agent know where to look; they do not replace source files, tests, or
validation commands.

MCP integration may be useful later when agent tooling needs direct interactive
graph queries. It is not required for the first implementation.

## Initial Integration Decision

The first Go semantic context implementation should use:

- `gograph` CLI as the primary context provider;
- fast `gograph build .` by default;
- optional `gograph build . --precise` for compiling repositories and risky
  changes;
- selected query output materialized into execution artifacts;
- no Tree-sitter dependency;
- no LSP or `gopls` daemon dependency.

This keeps the context composer simple while reducing repeated repository
rediscovery across independent Codex executions and Ralph Loop attempts.

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

The builder should not pass the entire `.gograph/graph.json` to Codex by
default. It should pass targeted query output plus paths and provenance.

Across Ralph Loop attempts, the builder should reuse or refresh the graph and
materialize attempt-specific code context artifacts instead of forcing each
agent invocation to rediscover the repository from scratch.

## Source of Truth

Semantic tools are not the source of truth.

The repository files, tests, and validation commands remain authoritative.
Graph, AST, LSP, and search output are context aids that can be stale,
incomplete, or misleading.

## Open Questions

- Which AST extraction output format should become durable artifact data?
- How should code context be cached between Ralph Loop iterations?
- What maximum context budget should apply before pruning begins?
- Which tickets should trigger `gograph build . --precise` automatically?
