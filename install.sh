#!/usr/bin/env bash
# ARTEX-God installer (one-line friendly): clones if needed, configures PostgreSQL, builds & runs.
# Usage (one-liner): curl -fsSL https://raw.githubusercontent.com/Jamesjaq/ARTEX_G/main/install.sh | bash
set -euo pipefail

REPO_URL="${ARTEX_REPO_URL:-https://github.com/Jamesjaq/ARTEX_G.git}"
REPO_DIR="${ARTEX_DIR:-ARTEX_G}"
BRANCH="${ARTEX_BRANCH:-main}"

info(){ printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok(){   printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn(){ printf '\033[33m[!]\033[0m %s\n' "$*"; }
die(){  printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }

# prompt: read from /dev/tty when piped
ask(){
  local p="$1" d="${2:-}" a
  if [ -t 0 ]; then read -rp "$p${d:+ [$d]}: " a; else read -rp "$p${d:+ [$d]}: " a < /dev/tty || a=""; fi
  echo "${a:-$d}"
}

echo "=============================="
echo "  ARTEX-God installer"
echo "=============================="

# if not inside the repo, clone it
if [ ! -f "cmd/artex" ] && [ ! -f "go.mod" ] || ! grep -q 'module' go.mod 2>/dev/null; then
  if [ -d "$REPO_DIR/.git" ]; then
    cd "$REPO_DIR"
  else
    command -v git >/dev/null 2>&1 || die "git is required. Install git and re-run."
    info "Cloning $REPO_URL ($BRANCH) -> $REPO_DIR"
    git clone --depth 1 -b "$BRANCH" "$REPO_URL" "$REPO_DIR" || git clone --depth 1 "$REPO_URL" "$REPO_DIR"
    cd "$REPO_DIR"
  fi
else
  cd "$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)" 2>/dev/null || true
fi

# database config
if [ -f config.json ]; then
  ok "config.json already exists (keeping it)"
else
  info "PostgreSQL connection (required)"
  DB_HOST="$(ask 'Database host' 127.0.0.1)"
  DB_PORT="$(ask 'Port' 5432)"
  DB_USER="$(ask 'User' artex)"
  DB_PASS="$(ask 'Password' '')"
  DB_NAME="$(ask 'Database name' artex)"
  DB_SSL="$(ask 'sslmode (disable/require)' disable)"
  cat > config.json <<JSON
{
  "database": {
    "host": "${DB_HOST}",
    "port": ${DB_PORT},
    "user": "${DB_USER}",
    "password": "${DB_PASS}",
    "dbname": "${DB_NAME}",
    "sslmode": "${DB_SSL}"
  }
}
JSON
  ok "Wrote config.json"
fi

command -v go >/dev/null 2>&1 || die "Go not found (need >= 1.26). Install: https://go.dev/dl/"
ok "Go: $(go version)"

# build frontend + embed if npm present
if command -v npm >/dev/null 2>&1 && [ -d web ]; then
  info "Building frontend (static export)..."
  ( cd web && (npm ci || npm install) && npm run build:static )
  rm -rf server/webui/dist && mkdir -p server/webui && cp -r web/out server/webui/dist
  info "Building embedded binary..."
  CGO_ENABLED=0 go build -tags embedui -trimpath -o artex ./cmd/artex
else
  warn "npm/web not available: building backend without embedded UI (run web dev separately)."
  CGO_ENABLED=0 go build -trimpath -o artex ./cmd/artex
fi
ok "Built ./artex"

echo
ok "Done. Start with: ./start.sh   (or: ./artex -addr 127.0.0.1:8787)"
info "Open http://localhost:8787 -> /setup, then configure System > LLM."
info "Autonomous run: create a Task (Goal) + configure LLM + Resume (Pause/Resume). Authorized testing only."
