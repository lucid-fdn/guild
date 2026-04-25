#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT}/scripts/lib/server.sh"

ADDR="${GUILD_SIM_ADDR:-$(pick_guild_addr)}"
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

for pair in \
  "governance-policies spec/examples/governance-policy.example.json" \
  "promotion-gates spec/examples/promotion-gate.example.json" \
  "approval-requests spec/examples/approval-request.example.json" \
  "commons-entries spec/examples/commons-entry.example.json"; do
  set -- ${pair}
  curl -fsS -X POST "${BASE_URL}/api/v1/$1" \
    -H "Content-Type: application/json" \
    --data-binary "@$2" >/dev/null
done

go run ./cli/cmd/guild eval-submit \
  --base-url "${BASE_URL}" \
  --suite examples/replay-suite.example.json \
  --wait >/dev/null

go run ./cli/cmd/guild replay-export \
  --base-url "${BASE_URL}" \
  --taskpack-id "4e1fe00c-6303-453c-8cb6-2c34f84896e4" \
  --file "${DATA_DIR}/replay.json"

PYTHONPATH=sdk/python/src BASE_URL="${BASE_URL}" python3 - <<'PY'
import os
import urllib.request
import json
from guild_client import GuildClient

base = os.environ["BASE_URL"]
client = GuildClient(base)
status = client.get_status()
assert status["name"] == "guild"
bundle = client.export_replay_bundle("4e1fe00c-6303-453c-8cb6-2c34f84896e4")
assert bundle["root_taskpack_id"] == "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
assert bundle["dri_bindings"], "expected DRI binding"
assert any(a["kind"] == "benchmark_result" for a in bundle["artifacts"])
assert any(a["kind"] == "skill_candidate" for a in bundle["artifacts"])
for path in [
    "/api/v1/governance-policies",
    "/api/v1/promotion-gates",
    "/api/v1/approval-requests",
    "/api/v1/commons-entries",
    "/api/v1/evaluation-jobs",
]:
    with urllib.request.urlopen(base + path) as response:
        payload = json.loads(response.read().decode("utf-8"))
        assert payload["items"], f"expected items for {path}"
PY

GITHUB_STEP_SUMMARY="${DATA_DIR}/github-step-summary.md" GUILD_AGENTDESK_TS_SKIP_BUILD=1 "${ROOT}/scripts/agentdesk-ts.sh" >/dev/null
grep -q "Agent Work Contract: passed" "${DATA_DIR}/github-step-summary.md"

echo "simulation-ok ${BASE_URL}"
