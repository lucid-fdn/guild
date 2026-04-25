#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "${DATA_DIR}"
}
trap cleanup EXIT

cd "${ROOT}"

if [[ "${GUILD_AGENTDESK_TS_SKIP_BUILD:-}" != "1" ]]; then
  corepack pnpm run build:agentdesk-ts >/dev/null
fi

CLI="${ROOT}/packages/agentdesk-cli/dist/index.js"
MCP="${ROOT}/packages/agentdesk-mcp/dist/index.js"

(
  cd "${DATA_DIR}"
  node "${CLI}" init --workspace agentdesk-ts >/dev/null
  MANDATE_ID="$(node "${CLI}" mandate create "Fix failing auth tests" --writable "src/auth/**,tests/auth/**" | awk '{print $2}')"

  node "${CLI}" next > next.json
  grep -q '"taskpack_id": "'"${MANDATE_ID}"'"' next.json

  node "${CLI}" claim --id "${MANDATE_ID}" --agent ts-agent --ttl-minutes 30 > claim.json
  grep -q '"agent": "ts-agent"' claim.json

  if node "${CLI}" next >/tmp/agentdesk-next-claimed.out 2>/tmp/agentdesk-next-claimed.err; then
    echo "agentdesk-ts: claimed mandate should be skipped by default" >&2
    exit 1
  fi
  grep -q "no open mandates found" /tmp/agentdesk-next-claimed.err

  node "${CLI}" context compile --id "${MANDATE_ID}" --role coder > context.json
  grep -q '"mandate_id": "'"${MANDATE_ID}"'"' context.json
  grep -q '"may_write"' context.json

  node "${CLI}" preflight --id "${MANDATE_ID}" --action write --path src/auth/login.ts > preflight-allow.json
  grep -q '"decision": "allow"' preflight-allow.json

  node "${CLI}" preflight --id "${MANDATE_ID}" --action write --path infra/prod/main.tf > preflight-approval.json
  grep -q '"decision": "needs_approval"' preflight-approval.json

  APPROVAL_ID="$(node "${CLI}" approval request --id "${MANDATE_ID}" --reason "Need owner consent" | awk '{print $2}')"
  node "${CLI}" approval resolve --approval-id "${APPROVAL_ID}" --decision approved --actor owner >/dev/null

  printf '<testsuite failures="0"></testsuite>\n' > test-results.xml
  printf '["src/auth/login.ts"]\n' > changed-files.json
  node "${CLI}" proof add --id "${MANDATE_ID}" --kind test_report --path test-results.xml >/dev/null
  node "${CLI}" proof add --id "${MANDATE_ID}" --kind changed_files --path changed-files.json >/dev/null
  node "${CLI}" handoff create --id "${MANDATE_ID}" --to reviewer --summary "TypeScript run is ready for review." >/dev/null

  node "${CLI}" replay export --id "${MANDATE_ID}" --file .agentdesk/replay/replay.json
  node "${CLI}" verify --id "${MANDATE_ID}" --github-report --replay-file .agentdesk/replay/replay.json > verify.json
  grep -q '"ready": true' verify.json

  grep -q "Agent Work Contract: passed" "${GITHUB_STEP_SUMMARY:-/dev/null}" 2>/dev/null || true

  node "${CLI}" close --id "${MANDATE_ID}" >/dev/null
  node "${CLI}" replay export --id "${MANDATE_ID}" > replay.json
  grep -q '"root_taskpack_id": "'"${MANDATE_ID}"'"' replay.json

  MCP_PATH="${MCP}" WORKSPACE="${DATA_DIR}" MANDATE_ID="${MANDATE_ID}" node --input-type=module <<'JS'
const { pathToFileURL } = await import("node:url");
const { handleMcpRequest } = await import(pathToFileURL(process.env.MCP_PATH).href);

const list = await handleMcpRequest({ id: 1, method: "tools/list" }, process.env.WORKSPACE);
if (!list?.result?.tools?.some((tool) => tool.name === "guild_get_next_mandate")) {
  throw new Error("MCP tools/list did not expose guild_get_next_mandate");
}

const replay = await handleMcpRequest(
  {
    id: 2,
    method: "tools/call",
    params: {
      name: "guild_export_replay_bundle",
      arguments: { taskpack_id: process.env.MANDATE_ID }
    }
  },
  process.env.WORKSPACE
);
const text = replay?.result?.content?.[0]?.text ?? "";
if (!text.includes(process.env.MANDATE_ID)) {
  throw new Error("MCP replay export did not include mandate id");
}
JS
)

echo "agentdesk-ts-ok"
