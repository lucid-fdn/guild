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

echo "smoke-ok ${BASE_URL}"
