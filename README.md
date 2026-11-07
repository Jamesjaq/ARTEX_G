# ARTEX-God

An LLM-driven autonomous security assessment platform (Go backend + Next.js frontend) with an integrated **legion/campaign engine**: planner + workers (“soldiers”) coordinate through a shared fact–intent graph, recall, and an offline KB to run end-to-end missions.

> **Authorized use only.** ARTEX-God is for security testing on systems you own or have explicit permission to assess.

## Highlights
- **Campaign / legion mode (integrated, vendored, no-CTF):** 14 native `campaign_*` tools (`init/status/frontier/graph/validate/fact/intent/done/hint/recall/kb_search/doc/claim/complete`) wired to `goals, planner, mainagent, worker, auto, pentest, reporter, retester`. Per-task graph store with best-effort bridging into ARTEX **assets, findings, and exploration**. Bundled doctrine/templates and an offline KB (`campaign/`, 82 entries).
- **Enforced mission workflow:** recon → assets → intents → evidence → findings; `claim → work → fact/intent → done`; **no duplicates** (`campaign_recall`/`campaign_kb_search` + graph/assets before rework), **feedback** (record tried + blocker + reason; reuse intel), **graph hygiene** (`campaign_validate`), **multi-target scale** (bucket/prioritize, WIP limits, delta rescans, timebox/stop-loss, rate-limit fragile targets).
- **94 skills** (web/API/auth, AD/cloud, exploit-dev/fuzzing, post-ex, OSINT, reporting, etc.) + the `campaign-graph` skill.
- **Structured Kali tooling (parseable output):** custom wrappers for `nmap, masscan, nuclei, httpx, ffuf, feroxbuster, netexec, sqlmap` (scoped/low-risk), `amass` (passive), `searchsploit`, plus the full Kali CLI via shell. Vetted **MCP** support (e.g. Playwright browser MCP, enable at `/system/mcp`).
- **ICS/SCADA/OT coverage (no RF/SDR/expensive hardware):** DNP3, Siemens S7comm, Schneider Modicon/UMAS, Modbus, EtherNet/IP-CIP (Rockwell/AB), IEC-60870-5-104, OPC UA, BACnet, GE SRTP, OMRON FINS, Mitsubishi MELSEC, Triconex, Codesys, PROFINET, HMI/EWS, segmentation/advisory chains. **Active, in-scope assessment is allowed — STRICTLY NO write/control/actuation** (no register/object writes, no setpoints/mode/force, no start/stop, no operate/control commands).
- **English dashboard & API tool layer**, Go backend + Next.js frontend, PostgreSQL-backed.

## Quick start (one-line installer)
```bash
curl -fsSL https://raw.githubusercontent.com/Jamesjaq/ARTEX_G/main/install.sh | bash
```
The installer clones (or uses the current repo), helps you set up PostgreSQL config, and builds the binary (embedding the frontend when Node/npm is available).

## Manual install (from source)
**Requirements:** Go (>= 1.26 recommended), Node/npm (to build/embed the frontend), and a running **PostgreSQL**. Configure an LLM in the UI (System → LLM).

```bash
git clone https://github.com/Jamesjaq/ARTEX_G.git
cd ARTEX_G
cp config.example.json config.json   # fill your database connection
./install.sh                         # or follow the steps below
```

Build & run manually:
```bash
# build frontend (optional but recommended)
cd web && npm ci && npm run build:static && cd ..
rm -rf server/webui/dist && cp -r web/out server/webui/dist
# build backend (embed UI)
CGO_ENABLED=0 go build -tags embedui -trimpath -o artex ./cmd/artex
# run
./start.sh                           # → http://localhost:8787
```
Open **http://localhost:8787**, complete `/setup` (admin password), then configure **System → LLM**.

## Running an autonomous mission
1. **Configure an LLM** (System → LLM).
2. **Create a Task** (`/function/tasks` → New task): set a clear **Goal** (and scope/hints). “Direct dispatch of first intent” can kick it off fast.
3. **Resume** the task (task detail/list Pause-Resume). Autonomy runs when you have **goal + LLM configured + task running/resumed**. Status: `exploring / paused / stalled / idle`.
4. Steer with **hints/scope**; review **graph, assets, findings, evidence**. For ICS, keep assessments read-identify only (no write/control).

## Configuration
- `config.json` (see `config.example.json`) holds the database connection. `jwt.key`, local data, and secrets are gitignored — never commit them.
- MCP servers: `/system/mcp` (add only trusted, local/vetted MCPs). Browser/Playwright MCP is supported; the system favors **bash + structured custom tools** over auto-fetching arbitrary MCPs.

## Project layout
- `cmd/artex`, `server/`, `agent/`, `db/` — backend and agent engine
- `web/` — Next.js dashboard (English)
- `campaign/` — vendored fact–intent graph engine, doctrine, templates, KB
- `skills/` — 94 skill packs (incl. `campaign-graph`, OT/ICS, AD/cloud/web)

## Safety & scope
Authorized testing only. Respect scope and law. **ICS/OT:** active, in-scope, read/identify/non-destructive assessment — **no write/control/actuation**. Rate-limit and avoid stressing fragile targets.

## Credits
Built on the ARTEX/ARTEX-God lineage, extended here with the legion/campaign engine, expanded skills/KB/tooling, ICS coverage, and an English dashboard. Upstream reference: [Xiao1an/ARTEX-God](https://github.com/Xiao1an/ARTEX-God).
