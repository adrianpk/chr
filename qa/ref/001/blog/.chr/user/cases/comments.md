---
schema_version: 1
name: "Comments"
description: "Add simple comments to published blog posts."
---

Readers should be able to add short comments to published posts and retrieve the
comments for a post through the HTTP API.

The first comments implementation should be realistic but small. It should keep
storage in memory, avoid moderation workflows, and preserve the existing
standard-library HTTP design.
