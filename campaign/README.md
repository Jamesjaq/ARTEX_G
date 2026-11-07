# campaign-graph (`cg`) — vendored, standalone (no-CTF)

A general-engagement fact–intent graph for autonomous security assessment agents.
Vendored into ARTEX so the system is self-contained (no external sappers dependency).

- **Core**: `lib/graph.py` (stdlib-only CLI/store) — typed facts (`host/service/cred/vuln/loot/artifact/note`), intents, goal, hints, frontier/validate/claim/release, recall + `doc` layers, KB, dispatch/auto/sweep. CTF/arena/flag-submission platform wiring removed.
- **KB**: `kb/entries/` (offline) + `lib/kb.py`.
- **Prompts**: `templates/` (reason/explore/verify/blindspot/command) — untrusted-data fenced, scope-disciplined, no-CTF.
- **Doctrine**: `doctrine/commander.md` (commander/scout/general/soldier, escalation ladder, intel transfer, state handover).
- **CLI**: `./cg --help` (default store `.cg/`).

Usage (standalone):
```bash
./campaign/cg init --goal "assess <authorized target>"
./campaign/cg fact "observed ..." -k service -d host=... -d port=...
./campaign/cg status && ./campaign/cg graph
./campaign/cg kb search "sqli" && ./campaign/cg recall "<gist>"
```
