# Slice 2: Agent Identity and Protocol Context

Status: delivered
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-2-context-agent-protocols`
PR: #79

## Purpose

Materialize stable agent identity and project protocol context before later
slices add ticket documents, code context, Ralph attempt evidence, and Docker
artifact writing.

## Delivered Behavior

The repository now has an agent/protocol context loader that can:

- define the default Go/software/infrastructure agent identity;
- load the root `AGENTS.md` and nested `AGENTS.md` files that apply to a target
  path;
- load selected stable `.chr/` protocol and runtime files;
- render `agent.md` content from the default identity;
- render `protocols.md` content from applicable project protocol files;
- record source provenance through `ContextSourceInventoryEntry` values.

## Implementation Notes

The loader keeps protocol selection bounded. It includes:

- `AGENTS.md` files along the project-to-target path;
- `.chr/manifest.md`;
- `.chr/config.yml`;
- `.chr/agents/**/*.md`;
- `.chr/runtime/agent/Dockerfile`.

It intentionally does not include tickets, runtime run logs, result artifacts,
or broader project documents. Those are handled by later slices.

## Contracts Added or Changed

New internal contracts:

- `AgentIdentity`
- `AgentProtocolContextOptions`
- `AgentProtocolContext`
- `ContextDocument`
- `DefaultAgentIdentity`
- `LoadAgentProtocolContext`
- `RenderAgentContext`
- `RenderProtocolsContext`

Source provenance uses the existing context source inventory model from Slice 1.

## Files of Interest

- `internal/chr/context_agent_protocols.go`
- `internal/chr/context_agent_protocols_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chr -run 'TestLoadAgentProtocolContext|TestRenderAgentContext'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice renders protocol partials as strings but does not yet write them to
the run artifact directory or link them from the main worker `context.md`.
Docker-compatible materialization is planned for Slice 7.
