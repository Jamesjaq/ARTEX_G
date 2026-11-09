# ARTEX-God

An LLM-driven autonomous security assessment platform (Go backend + Next.js frontend) with an integrated **legion/campaign engine**: planner + workers (“soldiers”) coordinate through a shared fact–intent graph, recall, and an offline KB to run end-to-end missions.

> **Authorized use only.** ARTEX-God is for security testing on systems you own or have explicit permission to assess.

## Quick start (simple, tested)

**Requirements:** a running **PostgreSQL**, **Go** (>= 1.26), and **Node/npm** (to build/embed the web UI). One process serves everything on a single port.

```bash
git clone https://github.com/Jamesjaq/ARTEX_G.git
cd ARTEX_G
./install.sh
```
`install.sh` prompts for your PostgreSQL connection (writes `config.json` if missing; keeps an existing `config.json`), builds the frontend, and builds the `./artex` binary (with the embedded UI when npm is available).

**Configure the database** (must match *your* PostgreSQL — host/port/user/password/dbname), then start:
```bash
# edit config.json to your real Postgres (install.sh prompts if creating it),
#   e.g. host 127.0.0.1, port 5432 (or 5433 etc), user/password/dbname/sslmode.
#   OR export a DSN instead (takes priority over config.json):
#   export ARTEX_PG_DSN='postgres://user:pass@127.0.0.1:5432/artex?sslmode=disable'

./start.sh                      # → http://localhost:8787
```
Open **http://localhost:8787** → complete **`/setup`** (first run) → **System → LLM** (configure a model). That’s it.

> **One-liner (installs + builds; you still need PostgreSQL and to set correct DB config):**
> ```bash
> curl -fsSL https://raw.githubusercontent.com/Jamesjaq/ARTEX_G/main/install.sh | bash
> ```

## Running an autonomous mission
1. **Configure an LLM** (System → LLM).
2. **Create a Task** (`/function/tasks` → New): set a clear **Goal** (and scope/hints). “Direct dispatch of first intent” helps it start fast.
3. **Resume** the task (Pause/Resume). Autonomy runs when **goal + LLM configured + task running/resumed**. Status: `exploring / paused / stalled / idle`.
4. Steer with **hints/scope**; review **graph, assets, findings, evidence**.

## Highlights
- **Campaign / legion mode (integrated, vendored, no-CTF):** 14 native `campaign_*` tools (`init/status/frontier/graph/validate/fact/intent/done/hint/recall/kb_search/doc/claim/complete`) wired to `goals, planner, mainagent, worker, auto, pentest, reporter, retester`. Per-task graph store with best-effort bridging into ARTEX **assets, findings, and exploration**. Bundled doctrine/templates and an offline KB (`campaign/`, 82 entries).
- **Enforced mission workflow:** recon → assets → intents → evidence → findings; `claim → work → fact/intent → done`; **no duplicates** (`campaign_recall`/`campaign_kb_search` + graph/assets before rework), **feedback** (record tried + blocker + reason; reuse intel), **graph hygiene** (`campaign_validate`), **multi-target scale** (bucket/prioritize, WIP limits, delta rescans, timebox/stop-loss, rate-limit fragile targets).
- **94 skills** (web/API/auth, AD/cloud, exploit-dev/fuzzing, post-ex, OSINT, reporting, etc.) + the `campaign-graph` skill.
- **Structured Kali tooling (parseable output):** custom wrappers for `nmap, masscan, nuclei, httpx, ffuf, feroxbuster, netexec, sqlmap` (scoped/low-risk), `amass` (passive), `searchsploit`, plus the full Kali CLI via shell. Vetted **MCP** support (e.g. Playwright browser MCP, enable at `/system/mcp`).
- **ICS/SCADA/OT coverage (no RF/SDR/expensive hardware):** DNP3, Siemens S7comm, Schneider Modicon/UMAS, Modbus, EtherNet/IP-CIP (Rockwell/AB), IEC-60870-5-104, OPC UA, BACnet, GE SRTP, OMRON FINS, Mitsubishi MELSEC, Triconex, Codesys, PROFINET, HMI/EWS, segmentation/advisory chains. **Active, in-scope assessment is allowed — STRICTLY NO write/control/actuation** (no register/object writes, no setpoints/mode/force, no start/stop, no operate/control commands).
- **English dashboard & API tool layer**, single Go binary (embedded UI) + PostgreSQL.

## Configuration & usage notes
- **Database:** `config.json` (see `config.example.json`) or `ARTEX_PG_DSN` (env, takes priority). `./start.sh` fails fast with a hint if neither is set.
- **start.sh:** single-process supervisor (passes args to `./artex`). Defaults: `:8787`, data next to the binary. Examples: `./start.sh -addr 127.0.0.1:8792 -data ./data`, or `ARTEX_PG_DSN=... ./start.sh`. Handles normal stop, restart (exit 75), and crash backoff.
- **Manual build (embedded UI):**
  ```bash
  cd web && npm ci && npm run build:static && cd ..
  rm -rf server/webui/dist && cp -r web/out server/webui/dist
  CGO_ENABLED=0 go build -tags embedui -trimpath -o artex ./cmd/artex
  ```
- **MCP:** `/system/mcp` — add only trusted, local/vetted MCPs. Browser/Playwright MCP supported; prefer **bash + structured custom tools** over auto-fetching arbitrary MCPs.
- **Secrets:** `config.json`, `jwt.key`, and local data are gitignored — never commit them.

## Troubleshooting
- **`password authentication failed` / connection refused / DB errors:** your `config.json` (or `ARTEX_PG_DSN`) doesn’t match your Postgres. Check `host/port/user/password/dbname/sslmode` (note: Postgres often runs on `5432`, but your instance may use another port, e.g. `5433`). Verify with `pg_isready -h HOST -p PORT`.
- **`address already in use` (8787/8788/3000):** another instance is running. Stop it (`kill <pid>`) or choose different ports, e.g. `./start.sh -addr 127.0.0.1:8792 -proxy ''`.
- **No UI / dev dashboard:** the recommended path is the **embedded UI** on the backend port (build with `-tags embedui` as above; `install.sh` does this when npm is present). A separate Next dev dashboard is optional: `cd web && npm ci && AUTOPENTEST_API=http://127.0.0.1:8787 npx next dev -p 3000 -H 127.0.0.1`.

## Project layout
- `cmd/artex`, `server/`, `agent/`, `db/` — backend and agent engine
- `web/` — Next.js dashboard (English)
- `campaign/` — vendored fact–intent graph engine, doctrine, templates, KB
- `skills/` — 94 skill packs (incl. `campaign-graph`, OT/ICS, AD/cloud/web)

## Safety & scope
Authorized testing only. Respect scope and law. **ICS/OT:** active, in-scope, read/identify/non-destructive assessment — **no write/control/actuation**. Rate-limit and avoid stressing fragile targets.

## Credits
Built on the ARTEX/ARTEX-God lineage, extended here with the legion/campaign engine, expanded skills/KB/tooling, ICS coverage, and an English dashboard. Upstream reference: [Xiao1an/ARTEX-God](https://github.com/Xiao1an/ARTEX-God).
