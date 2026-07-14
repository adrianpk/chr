---
key: "CHR-1"
title: "Implement board model"
synopsis: "Create the tic-tac-toe board state and win detection behavior."
status: "ready_for_dev"
priority: "high"
severity: "critical"
type: "feature"
tags: ["qa", "serialized-loop", "gameplay"]
use_case: "cli-gameplay"
project_area: "gameplay"
source_kind: "qa_fixture"
source_ref: ".chr/default/fixtures/serialized-loop-tickets"
slice: "Slice 1"
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
created_at: "2026-06-20T12:00:00Z"
updated_at: "2026-06-20T12:00:00Z"
---

# Description

Implement the game board representation for the CLI tic-tac-toe project.

The implementation should support legal move placement, occupied-cell
rejection, draw detection, and win detection for rows, columns, and diagonals.

# Checklist

- [ ] Decide the smallest board API needed by the CLI.
- [ ] Add table-driven unit tests for wins, draws, invalid positions, and
      occupied cells.
- [ ] Keep the implementation in Go's standard library.

# Acceptance

- Board logic is covered by deterministic tests.
- Invalid moves do not mutate board state.
- Win and draw detection are explicit and readable.

# Comments

# Blocker

# Validation
