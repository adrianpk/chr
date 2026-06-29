#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)"
fixture_dir="$repo_root/.prj/fixtures/mvp-agent-proving-ground/todolist"
objectives_dir="$fixture_dir/.chr/proving/objectives"

test -f "$fixture_dir/.chr/proving/objective-contract.md"

for objective in TODO-1-fix-task-id-reuse.md TODO-2-complete-task.md; do
    path="$objectives_dir/$objective"

    test -f "$path"
    grep -q "^objective_id:" "$path"
    grep -q "^ticket_key:" "$path"
    grep -q "^artifact_root:" "$path"
    grep -q "^max_attempts:" "$path"
    grep -q "^## Objective$" "$path"
    grep -q "^## Validation Commands$" "$path"
    grep -q "^## Success Outcome$" "$path"
    grep -q "^## Blocked Outcome$" "$path"
done

printf 'objectives: ok\n'
