#!/usr/bin/env bash

pick_guild_addr() {
  python3 - <<'PY'
import socket

with socket.socket() as sock:
    sock.bind(("127.0.0.1", 0))
    print(f"127.0.0.1:{sock.getsockname()[1]}")
PY
}

start_guildd() {
  local addr="$1"
  local data_dir="$2"
  local log_file="$3"

  GUILD_ADDR="${addr}" GUILD_DATA_DIR="${data_dir}" go run ./server/cmd/guildd >"${log_file}" 2>&1 &
  SERVER_PID=$!
}

wait_for_guildd() {
  local base_url="$1"
  local log_file="$2"

  for _ in {1..60}; do
    if curl -fsS "${base_url}/healthz" >/dev/null 2>&1; then
      return 0
    fi
    if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
      cat "${log_file}" >&2
      return 1
    fi
    sleep 0.25
  done

  cat "${log_file}" >&2
  echo "guildd did not become ready at ${base_url}" >&2
  return 1
}
