---
title: "Tictactoe Project Context Raw Capture"
status: "raw"
type: "dump"
curated_path: ".chr/user/dumps/project-context.md"
---

# Tictactoe Project Context Raw Capture

I want a tiny example app that is not another HTTP service. A terminal
tic-tac-toe game is enough: it is familiar, has real rules, and should force the
agent to handle state, input validation, and end conditions without requiring a
big domain model.

Keep it boring. It should be Go, no third-party libraries, no persistence, no
networking, no fancy terminal package, and no AI in the app. The app is only
there to test whether Chryso can turn normal project context into tickets
and then have Codex workers finish the implementation.

I do not want to answer a long quiz every time I reset the scenario. Seed the
normal project and use-case context, including raw-ish notes and curated notes,
but do not provide ready-made tickets for the main implementation. The tickets
should come from Chryso's generated plan.
