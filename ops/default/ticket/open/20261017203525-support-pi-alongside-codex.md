---
id: TKT-20261017203525
title: "Support Pi alongside Codex for execution and generation"
status: open
kind: task
severity: unclassified
priority: unclassified
scope: domain
tags: agents, pi, models
source: chat
reported_at: 2026-10-17T20:35:25Z
commits: []
---

# Support Pi alongside Codex for execution and generation

## Need

Allow users to choose Pi as an alternative to Codex in CHR. Keep Codex
supported without forcing existing users to replace their harness or migrate
their setup.

The motivation is model and provider choice: users should be able to use
models supported by Pi, including open-weight models and compatible local or
self-hosted endpoints.

## Desired Outcomes

1. Users can select Codex or Pi for headless ticket execution in Docker.
2. Users can select Codex or Pi for AI conversation and content processing,
   including Dumps, Quiz questions, Specs, Plans and related generation.
3. Users can choose the same harness for both roles or select them
   independently, with the desired model and provider supported by that harness.
4. Harness selection preserves CHR's ticket lifecycle, execution order,
   validation requirements, result semantics and workflow ownership.
5. Existing Codex setups remain usable without requiring Pi.

## Capture Boundary

This ticket records the product need. It does not prescribe protocols,
adapters, configuration schemas, container packaging or other implementation
details. Define those in a separately authorized design before implementation.
