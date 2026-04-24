#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT}/scripts/lib/server.sh"

ADDR="${GUILD_E2E_ADDR:-$(pick_guild_addr)}"
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
  --file "${DATA_DIR}/seed-replay.json"
go run ./cli/cmd/guild validate --kind replay-bundle --file "${DATA_DIR}/seed-replay.json" >/dev/null

go run ./cli/cmd/guild replay-suite \
  --base-url "${BASE_URL}" \
  --suite examples/replay-suite.example.json >/dev/null

go run ./cli/cmd/guild eval-submit \
  --base-url "${BASE_URL}" \
  --suite examples/replay-suite.example.json \
  --wait >/dev/null

PYTHONPATH=sdk/python/src BASE_URL="${BASE_URL}" python3 - <<'PY'
import os
from guild_client import GuildClient

client = GuildClient(os.environ["BASE_URL"] + "/")
assert client.get_status()["name"] == "guild"
assert len(client.list_taskpacks()) >= 1
bundle = client.export_replay_bundle("4e1fe00c-6303-453c-8cb6-2c34f84896e4")
assert bundle["schema_version"] == "v1alpha1"
assert bundle["taskpack"]["taskpack_id"] == "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
assert bundle["root_taskpack_id"] == "4e1fe00c-6303-453c-8cb6-2c34f84896e4"
assert len(bundle["taskpacks"]) >= 1
assert any(artifact["kind"] == "benchmark_result" for artifact in bundle["artifacts"])
assert any(artifact["kind"] == "skill_candidate" for artifact in bundle["artifacts"])
assert len(client._get_json("/api/v1/evaluation-jobs")["items"]) >= 1
PY

GUILD_BASE_URL="${BASE_URL}" corepack pnpm --dir examples/typescript-adapter-core exec tsx src/demo.ts >"${DATA_DIR}/adapter-replay.json"
go run ./cli/cmd/guild validate --kind replay-bundle --file "${DATA_DIR}/adapter-replay.json" >/dev/null
grep -q '"taskpack_id": "d013e9c3-3fdc-4f72-a79f-3ca30d0fe111"' "${DATA_DIR}/adapter-replay.json"

echo "e2e-ok ${BASE_URL}"
