---
key: "CHC-2"
title: "Implement CLI game loop"
synopsis: "Wire the board model into an interactive standard-input CLI game."
status: "ready_for_dev"
priority: "high"
severity: "normal"
type: "feature"
tags: ["qa", "serialized-loop", "cli"]
use_case: "cli-gameplay"
project_area: "cli"
source_kind: "qa_fixture"
source_ref: ".chc/user/fixtures/serialized-loop-tickets"
slice: "Slice 2"
branch: ""
review_refs: []
review_state: ""
review_result: ""
review_comments_url: ""
review_round: ""
blocking_findings: []
merge_state: ""
merge_result: ""
merge_actor: ""
merge_commit: ""
merge_blocker: ""
merged_at: ""
created_at: "2026-06-20T12:05:00Z"
updated_at: "2026-06-20T12:05:00Z"
---

# Description

Implement the command-line gameplay loop using the board behavior from CHC-1.

The CLI should alternate X and O turns, read moves from standard input, reject
invalid input without losing the current turn, and exit cleanly on win or draw.

# Checklist

- [ ] Reuse the board model rather than duplicating game rules in `main`.
- [ ] Add command-level tests that drive input and assert output.
- [ ] Keep prompts and messages stable enough for automated tests.

# Acceptance

- A complete valid game can be played from standard input.
- Invalid input is handled without ending the game.
- The CLI reports wins and draws clearly.

# Comments

# Blocker

# Validation
