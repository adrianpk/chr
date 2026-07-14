---
schema_version: 1
name: "Dev To Main Gate"
description: "Collect objective promotion evidence before aligning dev into main."
---

This use case covers the promotion gate, report output, policy configuration, and Woodpecker integration.

The gate should compare a base and head ref, collect Git evidence, run validation, classify Chryso-managed work, list manual or external changes, and write JSON and Markdown reports under `.chr/default/reports/gates/`.

For existing projects, the gate starts as an adoption-friendly evidence mechanism. Once new work is consistently executed through Chryso, policy can become stricter about missing agent evidence.
