#!/usr/bin/env bash
# ARTEX-God one-command starter: backend (+ optional dashboard) with supervisor.
#
# Usage:
#   ./start.sh                  # start backend + dashboard (if available)
#   START_WEB=0 ./start.sh      # backend only
#   ./start.sh -addr :8787      # extra args passed to ./artex
#
# Env (optional): ARTEX_PG_DSN, ARTEX_ADDR (default 127.0.0.1:8787),
#   ARTEX_DATA (default ./data), ARTEX_PROXY (default ''), WEB_PORT (default 3000),
#   START_WEB (1=auto start dashboard when web/ deps present, 0=off),
#   AUTOPENTEST_API (override dashboard->backend URL). A local .env is sourced if present.
#
# Backend supervisor: exit 0 = stop, 75 = restart (self-update/rollback), else crash backoff.
set -u

cd "$(dirname "${BASH_SOURCE[0]:-$0}")" || exit 1

# load optional .env (export vars)
if [ -f .env ]; then
  set -a; . ./.env; set +a
fi

ARTEX_ADDR="${ARTEX_ADDR:-127.0.0.1:8787}"
ARTEX_DATA="${ARTEX_DATA:-./data}"
ARTEX_PROXY="${ARTEX_PROXY-}"
WEB_PORT="${WEB_PORT:-3000}"
START_WEB="${START_WEB:-1}"
BIN=./artex
RESTART_CODE=75
MAX_DELAY=60

if [ ! -x "$BIN" ]; then
  echo "[artex] missing executable $BIN. Run ./install.sh (or build: go build -o artex ./cmd/artex) first." >&2
  exit 1
fi

# fail fast: backend needs database config (config.json or ARTEX_PG_DSN)
if [ ! -f config.json ] && [ -z "${ARTEX_PG_DSN:-}" ]; then
  echo "[artex] no database config found." >&2
  echo "[artex]   set ARTEX_PG_DSN, e.g.:" >&2
  echo "[artex]     export ARTEX_PG_DSN='postgres://user:pass@127.0.0.1:5432/artex?sslmode=disable'" >&2
  echo "[artex]   or create config.json (see config.example.json), then re-run ./start.sh" >&2
  exit 1
fi

# derive API URL for the dashboard
if [ -n "${AUTOPENTEST_API:-}" ]; then
  API_URL="$AUTOPENTEST_API"
elif [[ "$ARTEX_ADDR" == :* ]]; then
  API_URL="http://127.0.0.1${ARTEX_ADDR}"
else
  API_URL="http://${ARTEX_ADDR}"
fi

# display URL (localhost if bound to 0.0.0.0)
disp_addr="$ARTEX_ADDR"
[[ "$disp_addr" == 0.0.0.0:* ]] && disp_addr="127.0.0.1:${disp_addr#0.0.0.0:}"
[[ "$disp_addr" == :* ]] && disp_addr="127.0.0.1${disp_addr}"

backend_child=0
web_child=0
stopping=0

stop_all() {
  stopping=1
  [ "$web_child" -ne 0 ] && kill -TERM "$web_child" 2>/dev/null || true
  [ "$backend_child" -ne 0 ] && kill -TERM "$backend_child" 2>/dev/null || true
}
trap stop_all INT TERM

# --- backend supervisor (background) ---
(
  delay=1
  while :; do
    "$BIN" -addr "$ARTEX_ADDR" -proxy "$ARTEX_PROXY" -data "$ARTEX_DATA" "$@" &
    bpid=$!
    echo "$bpid" > .artex.backend.pid 2>/dev/null || true
    wait "$bpid"; code=$?
    if [ "$code" -gt 128 ]; then wait "$bpid" 2>/dev/null || true; code=$?; fi
    rm -f .artex.backend.pid 2>/dev/null || true
    [ "$stopping" -eq 1 ] && exit 0
    case "$code" in
      0) echo "[artex] backend exited normally"; exit 0 ;;
      "$RESTART_CODE") echo "[artex] backend requested restart (update/rollback)…"; delay=1 ;;
      *) echo "[artex] backend exited (code=$code), restarting in ${delay}s" >&2; sleep "$delay"; delay=$((delay*2)); [ "$delay" -gt "$MAX_DELAY" ] && delay=$MAX_DELAY ;;
    esac
  done
) &
backend_child=$!

echo "[artex] backend: ${API_URL}  (data: ${ARTEX_DATA})"
echo "[artex] UI:      http://${disp_addr}/  (embedded UI, if built with -tags embedui) or setup below"

# --- dashboard (optional) ---
if [ "$START_WEB" != "0" ] && [ -f web/package.json ]; then
  if command -v npx >/dev/null 2>&1; then
    if [ -d web/node_modules ]; then
      echo "[artex] dashboard: http://127.0.0.1:${WEB_PORT}  (-> ${API_URL})"
      (
        cd web && AUTOPENTEST_API="$API_URL" npx next dev -p "$WEB_PORT" -H 127.0.0.1
      ) &
      web_child=$!
    else
      echo "[artex] dashboard skipped: web/node_modules missing (run: cd web && npm ci). Set START_WEB=0 to hide." >&2
    fi
  else
    echo "[artex] dashboard skipped: npx/npm not found. Set START_WEB=0 to hide." >&2
  fi
elif [ "$START_WEB" != "0" ]; then
  echo "[artex] dashboard skipped: web/ not found. (single-binary/embedded UI can serve the app)" >&2
fi

echo "[artex] running. Ctrl+C to stop. (first run: open the UI -> /setup, then configure System > LLM)"

# wait: exit when backend supervisor exits (or on signal)
wait "$backend_child" 2>/dev/null || true
stop_all
[ "$web_child" -ne 0 ] && wait "$web_child" 2>/dev/null || true
echo "[artex] stopped"
exit 0
