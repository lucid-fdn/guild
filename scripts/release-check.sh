#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

required_not_ignored=(
  "cli/cmd/guild/main.go"
  "server/cmd/guildd/main.go"
  "adapters/a2a/src/index.ts"
  "adapters/langgraph/src/index.ts"
  "adapters/mcp/src/index.ts"
  "adapters/mcp/src/server.ts"
  "adapters/mcp/bin/guild-agentdesk-mcp"
  "adapters/typescript/src/index.ts"
  "examples/one-task-one-dri-commons/run.sh"
  "scripts/smoke.sh"
  "scripts/e2e.sh"
  "scripts/simulation.sh"
  "scripts/agentdesk-ts.sh"
  "examples/one-task-one-dri-commons/run.sh"
  "scripts/lib/server.sh"
  "spec/taskpack.schema.json"
  "openapi/guild.v1alpha1.yaml"
)

required_executable=(
  "scripts/smoke.sh"
  "scripts/e2e.sh"
  "scripts/simulation.sh"
  "scripts/agentdesk-ts.sh"
  "scripts/lib/server.sh"
  "adapters/mcp/bin/guild-agentdesk-mcp"
)

echo "release-check: git ignore guard"
for path in "${required_not_ignored[@]}"; do
  if git check-ignore -q "${path}"; then
    echo "release-check: ${path} is ignored but must be tracked" >&2
    exit 1
  fi
done

echo "release-check: executable script guard"
for path in "${required_executable[@]}"; do
  if [[ ! -x "${path}" ]]; then
    echo "release-check: ${path} must be executable" >&2
    exit 1
  fi
done

echo "release-check: forbidden generated artifacts guard"
for path in guild guildd data ui/.next node_modules; do
  if [[ -e "${path}" ]] && ! git check-ignore -q "${path}"; then
    echo "release-check: ${path} exists and is not ignored" >&2
    exit 1
  fi
done

make verify
make e2e
make simulation
examples/one-task-one-dri-commons/run.sh

echo "release-check-ok"
