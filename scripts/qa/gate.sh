#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
project_dir="$(mktemp -d)"

cleanup() {
	rm -rf "$project_dir"
}
trap cleanup EXIT HUP INT TERM

cat >"$project_dir/go.mod" <<'EOF'
module gateqa.test

go 1.25
EOF

cat >"$project_dir/gateqa_test.go" <<'EOF'
package gateqa

import "testing"

func TestGateQA(t *testing.T) {}
EOF

mkdir -p "$project_dir/.chc/user/config"

cat >"$project_dir/.chc/user/config/gates.yml" <<'EOF'
gates:
  dev_main:
    validation_command: go test ./...
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: false
EOF

git -C "$project_dir" init >/dev/null
git -C "$project_dir" -c user.name="ChrysoCoord QA" -c user.email="qa@example.test" add .
git -C "$project_dir" -c user.name="ChrysoCoord QA" -c user.email="qa@example.test" commit -m "base" >/dev/null

base_ref="$(git -C "$project_dir" rev-parse HEAD)"

cat >"$project_dir/README.md" <<'EOF'
# Gate QA

This file gives the gate a concrete Git change to report.
EOF

git -C "$project_dir" -c user.name="ChrysoCoord QA" -c user.email="qa@example.test" add README.md
git -C "$project_dir" -c user.name="ChrysoCoord QA" -c user.email="qa@example.test" commit -m "change" >/dev/null

output_file="$project_dir/gate-output.txt"
go_cache="$project_dir/go-build-cache"

GOCACHE="$go_cache" go run "$repo_root/cmd/chc" gate dev-main --dir "$project_dir" --base "$base_ref" --head HEAD >"$output_file"

if ! grep -q "status: passed" "$output_file"; then
	echo "qa gate failed: expected passed status" >&2
	cat "$output_file" >&2
	exit 1
fi

json_count="$(find "$project_dir/.chc/user/reports/gates" -maxdepth 1 -type f -name '*.json' | wc -l | tr -d ' ')"
markdown_count="$(find "$project_dir/.chc/user/reports/gates" -maxdepth 1 -type f -name '*.md' | wc -l | tr -d ' ')"

if [ "$json_count" != "1" ] || [ "$markdown_count" != "1" ]; then
	echo "qa gate failed: expected one JSON report and one Markdown report" >&2
	echo "json_count=$json_count markdown_count=$markdown_count" >&2
	exit 1
fi

echo "qa gate ok: dev-main gate passed and wrote JSON/Markdown reports"
