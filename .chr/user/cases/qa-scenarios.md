---
schema_version: 1
name: "QA Scenarios"
description: "Maintain reproducible scenarios for validating Chryso workflows."
---

This use case covers versioned reference projects, resettable active projects, gallery-oriented demo data, and smoke scripts.

The QA area should make it cheap to reset a scenario, operate it through the TUI or CLI, and compare expected outcomes. The current examples include the blog scenario for workflow validation and the dialer scenario for UI gallery screenshots.

Scenarios should remain simple enough to diagnose while still resembling real early-stage Go projects.
