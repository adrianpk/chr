# Slice 5 Investigation: Forgejo and Woodpecker Feedback

Status: completed
Slicing: Dev to Main Agent Evidence Gate

## Findings

Woodpecker is the first practical feedback path for this slice. Its workflow
syntax runs steps serially and a non-zero step exit fails the workflow, which is
enough for `chr gate dev-main` to block promotion through CI status.

Woodpecker can run a single `.woodpecker.yml` workflow or multiple workflows
under `.woodpecker/`. The setup helper should start with the single-file path
because it is easier to inspect and safer to generate.

Forgejo Actions exposes pull request events and repository context variables,
but this slice does not need Forgejo API writes. The first implementation should
rely on Woodpecker status visibility in Forgejo and local Markdown/JSON gate
artifacts. API comments/status enrichment can be added later if required.

## Selected Mechanism

- Generate a Woodpecker gate step.
- Let Woodpecker fail the pipeline when `chr gate dev-main` exits non-zero.
- Keep detailed evidence in `.chr/default/reports/gates/`.
- Do not require Forgejo API tokens for the first integration.

## Sources

- Woodpecker workflow syntax: https://woodpecker-ci.org/docs/usage/workflow-syntax
- Woodpecker workflows: https://woodpecker-ci.org/docs/next/usage/workflows
- Forgejo Actions reference: https://forgejo.org/docs/latest/user/actions/reference/
