#!/bin/sh
# ARTEX-God supervisor (Linux / macOS)
#
# Usage:
#   ./start.sh                        run in the foreground (Ctrl-C to stop)
#   nohup ./start.sh >artex.log 2>&1 &   run persistently in the background
#   ./start.sh -addr :8787            extra arguments are passed through to artex
#
# Simple single-process setup (like ARTEX): the backend serves the app on :8787.
# Requires PostgreSQL config via config.json or ARTEX_PG_DSN. On first run open /setup,
# then configure System > LLM. Autonomy = Task Goal + LLM configured + Resume.
#
# Exit handling: 0 = normal stop, 75 = restart (self-update/rollback), other = crash backoff.
set -u

cd "$(dirname "$0")" || exit 1

# load optional .env (exports vars like ARTEX_PG_DSN)
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

BIN=./artex
[ -x "$BIN" ] || { echo "[artex] executable $BIN not found. Run ./install.sh (or: go build -o artex ./cmd/artex) first." >&2; exit 1; }

# fail fast: backend needs database config (config.json or ARTEX_PG_DSN)
if [ ! -f config.json ] && [ -z "${ARTEX_PG_DSN:-}" ]; then
  echo "[artex] no database config found." >&2
  echo "[artex]   create config.json (see config.example.json), or set ARTEX_PG_DSN, e.g.:" >&2
  echo "[artex]     export ARTEX_PG_DSN='postgres://user:pass@127.0.0.1:5432/artex?sslmode=disable'" >&2
  exit 1
fi

RESTART_CODE=75
MAX_DELAY=60

child=0
stopping=0

# Forward stop signals to artex (needed for graceful shutdown; under Docker/SIGTERM to PID 1).
forward() {
  stopping=1
  if [ "$child" -ne 0 ]; then
    kill -TERM "$child" 2>/dev/null || true
  fi
}
trap forward INT TERM

delay=1
while :; do
  "$BIN" "$@" &
  child=$!

  # A signal interrupts wait (>128). Wait again to get the real exit code.
  wait "$child"
  code=$?
  if [ "$code" -gt 128 ]; then
    wait "$child"
    code=$?
  fi
  child=0

  if [ "$stopping" -eq 1 ]; then
    echo "[artex] stopped"
    exit 0
  fi

  case "$code" in
    0)
      echo "[artex] exited normally"
      exit 0
      ;;
    "$RESTART_CODE")
      echo "[artex] restart requested (update/rollback)…"
      delay=1
      ;;
    *)
      echo "[artex] abnormal exit (code=$code), restarting in ${delay}s" >&2
      sleep "$delay"
      delay=$((delay * 2))
      [ "$delay" -gt "$MAX_DELAY" ] && delay=$MAX_DELAY
      ;;
  esac
done
