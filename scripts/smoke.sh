#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT}/scripts/lib/server.sh"

ADDR="${GUILD_SMOKE_ADDR:-$(pick_guild_addr)}"
BASE_URL="http://${ADDR}"
DATA_DIR="$(mktemp -d)"
LOG_FILE="${DATA_DIR}/guildd.log"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  rm -rf "${DATA_DIR}"
}
trap cleanup EXIT

cd "${ROOT}"

start_guildd "${ADDR}" "${DATA_DIR}/data" "${LOG_FILE}"
wait_for_guildd "${BASE_URL}" "${LOG_FILE}"

go run ./cli/cmd/guild conformance --base-url "${BASE_URL}" >/dev/null
go run ./cli/cmd/guild replay-export \
  --base-url "${BASE_URL}" \
  --taskpack-id "4e1fe00c-6303-453c-8cb6-2c34f84896e4" \
  --file "${DATA_DIR}/replay.json"

grep -q '"schema_version": "v1alpha1"' "${DATA_DIR}/replay.json"
grep -q '"promotion_records"' "${DATA_DIR}/replay.json"

go build -o "${DATA_DIR}/guild" ./cli/cmd/guild
mkdir -p "${DATA_DIR}/agentdesk"
(
  cd "${DATA_DIR}/agentdesk"
  "${DATA_DIR}/guild" agentdesk init --workspace smoke >/dev/null
  MANDATE_ID="$("${DATA_DIR}/guild" agentdesk mandate create "Fix failing auth tests" --writable "src/auth/**,tests/auth/**" | awk '{print $2}')"
  "${DATA_DIR}/guild" agentdesk claim --id "${MANDATE_ID}" --agent smoke-agent --ttl-minutes 30 > claim.json
  grep -q '"agent": "smoke-agent"' claim.json
  "${DATA_DIR}/guild" agentdesk preflight --id "${MANDATE_ID}" --action write --path src/auth/login.ts > preflight.json
  grep -q '"decision": "allow"' preflight.json
  "${DATA_DIR}/guild" agentdesk context compile --id "${MANDATE_ID}" --role coder > context.json
  grep -q '"mandate_id": "'"${MANDATE_ID}"'"' context.json
  printf '<testsuite failures="0"></testsuite>\n' > test-results.xml
  printf '["src/auth/login.ts"]\n' > changed-files.json
  "${DATA_DIR}/guild" agentdesk proof add --id "${MANDATE_ID}" --kind test_report --path test-results.xml >/dev/null
  "${DATA_DIR}/guild" agentdesk proof add --id "${MANDATE_ID}" --kind changed_files --path changed-files.json >/dev/null
  "${DATA_DIR}/guild" agentdesk handoff create --id "${MANDATE_ID}" --to reviewer --summary "Smoke run is ready for review." >/dev/null
  "${DATA_DIR}/guild" agentdesk verify --id "${MANDATE_ID}" > verify.json
  grep -q '"ready": true' verify.json
  "${DATA_DIR}/guild" agentdesk close --id "${MANDATE_ID}" >/dev/null
  "${DATA_DIR}/guild" agentdesk replay export --id "${MANDATE_ID}" --file agentdesk-replay.json
  grep -q '"root_taskpack_id": "'"${MANDATE_ID}"'"' agentdesk-replay.json
)

echo "smoke-ok ${BASE_URL}"
