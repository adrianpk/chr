---
schema_version: 1
name: "Blog"
description: "Small headless blog service used as a chr QA scenario."
---

The project is an early-stage Go headless blog service. It exposes HTTP
endpoints for published posts and intentionally starts without a comments
feature.

The QA goal is to let chr generate the comments specification, plan,
tracking document, and implementation tickets from seeded context. The project
also includes one manual bug ticket for the existing unknown-post behavior.
