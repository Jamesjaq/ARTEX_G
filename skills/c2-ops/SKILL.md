---
name: c2-ops
description: "Invoke this skill when standing up, tasking, auditing or tearing down a command channel on infrastructure the engagement owns: listener binding and hygiene, beacon interval and jitter budgets, tasking discipline, session accounting, the kill chain for a burned listener, and the closing channel inventory."
---

# C2 Ops (command-channel hygiene on authorised infrastructure)

This is our side of the line: the channel itself — what it touches, how loud it is, who is tasking it, and that it is provably gone at the end.

Done means every listener sits on a registered engagement-owned asset, every beacon has a loudness budget agreed with your lead, every session is accounted for, and wrap-up produces an inventory a client can act on.

The deliverable is a coverage statement over authorised infrastructure, never a stealth guide.

## Boundaries

- Authorised use only: a listener binds to infrastructure this engagement owns and has registered.
- If the host serving or receiving the channel is not a registered in-scope or engagement-owned asset, there is no channel — stop and record the question.
- Never point a listener, redirector or payload download at a host outside the registered scope, and never accept a check-in from one.
- An unexpected host calling your channel is a stop-and-report event handed to `engagement-deconfliction`, not a bonus foothold.
- This skill is NOT the artifact (`payload-lab-development`), NOT host provisioning and aging (`campaign-infrastructure`), NOT the indicator ledger (`opsec-tradecraft`), NOT persistence on the client's host (`persistence-cleanup`).
- Channel shape is measured and reported as defender coverage (`telemetry-sources`, `detection-coverage`).
- Nothing here proposes blinding, disabling or spoofing a production sensor or appliance, and nothing proposes changing the platform's guard/intercept configuration — a block you see came from the platform, not from the defender.
- Volume stays inside the engagement's noise and bandwidth budget; testing a real perimeter's egress belongs to `exfil-egress` under its own approval.

## Procedure

1. Gate first. Take from your lead the approval text, the hosts this channel may serve, and the objective the channel exists for.
2. Pull `list_assets` and be able to name the owned assets in your own words before binding anything. No approval, no listener.
3. Register each piece of the channel as it exists: host, FQDN and port through `insert_assets` (type=ip, type=domain/subdomain as fits).
4. Write one `record_fact` per piece: purpose, cell, transport, burn state. An untracked listener is a liability you cannot tear down, and another seat will rediscover it as a target.
5. Check the environment before promising a step. This task image is python:3.12-slim-bookworm with curl, dig, whois, nmap, nc, jq and the ProjectDiscovery set — no C2 framework, no compiler for implant targets, no Windows lab VM.
6. Listener daemons, inbound firewall rules and TLS installs run on an owned VPS or the operator's host, so each goes up as one `ask_operator` request and you wait. Never run a listener in the container that also holds client data.
7. Bind narrow: the listener binds the owned address only, and the FQDN the implant contacts resolves only to that address.
8. Verify the resolution path before the first check-in: `dig +short <fqdn>`, then `curl -sv --resolve <fqdn>:443:<owned_ip> https://<fqdn>/`. What answers must be your listener, not recycled content, a CDN front or a previous owner's site.
9. An address you did not register answering that name is a burn indicator: record it and stop.
10. Set the loudness budget before any implant tasking, in numbers you can report: check-in interval, jitter percentage, request size, and a per-host cap on bytes per hour.
11. Get the lead's agreement, fact the numbers, and keep every later session inside them. "Slow and quiet" is not a licence to crawl for days; it is a set of numbers with a window.
12. Tasking discipline: one tasking path per channel (framework queue or operator chat, not both), one task per session, each task traceable to an intent id.
13. Scripted tasking beats an interactive shell for anything repeatable — interactive sessions are where orphans, unlogged commands and unrecorded scope events come from.
14. Session accounting as you go: record host, first-seen time, the scope entry that authorises it, and its cell.
15. A session from a host you cannot match to a scope entry is killed immediately and written up. You do not look around it first.
16. Keep it minimal: no idle sessions held for comfort, no second channel opened to work around a slow one, no repeat of a task that already produced nothing.
17. Re-count before closing and write it as live/orphaned against total seen ("4 of 11 still open; 7 closed"), never "a few".
18. Burned-listener kill chain, in this order:
    - stop tasking into it — `pause_task` the front, `kill_work` the intents using it;
    - capture what burned it before rotating (listener log excerpt, block or alert text, cert/DNS reputation hit, client detection notice), or the cause is lost;
    - write the burn into the ledger through `opsec-tradecraft` and mark the pieces burned;
    - rotate to a reserved cell with the operator re-provisioning — never reuse a burned FQDN, IP or certificate to save time;
    - confirm no live session still calls the dead address, or a stranded implant generates a second incident.
19. Teardown inventory at wrap-up: every listener, address, domain, certificate and redirection rule, plus every implant or scheduled item still resident on a client-side host.
20. Give each inventory row an owner for removal (us, or the client's own hand) and a verification method. Cross-check `persistence-cleanup`'s list: an implant the client cannot evict with their own tools is an incident you delivered.

Gate: do not go to 10 until 3-9 have produced registered, verified-owned addresses; do not go to 19 until 18 was run for every channel that died mid-engagement.

## Evidence

- Positive (channel behaviour): transport, the interval/jitter and byte figures actually used, the window, sessions seen, and what each defender source produced or did not — named by source (NDR flow record, proxy log, DNS resolver, EDR network event, or "never instrumented").
- Label the proof basis: executed-live against the authorised or consented host, or executed-lab.
- Positive (hygiene finding): a concrete exposure with blast radius — "listener accepted a check-in from 1 unregistered host over 3 beacons before closure; cause: FQDN recycled by previous owner".
- Negative: name method, window and denominator, then close honestly — "HTTPS transport checked over 40 check-ins from 2 lab hosts between <t1>-<t2>; no process-tree or DNS event from the stack in use, reported as a coverage gap, not as a safe channel."
- Where no defender source was reachable, write NOT ASSESSED with the reason instead of implying silence.
- Material captured in a session goes to the credential store via `save_task_credential` and is cited by ID. No token, cookie or key in a fact, listener-log excerpt, finding or report.

## Report shape

Title: `C2 channel <transport> on <owned FQDN>: <defender verdict> across <n> of <N> beacons in <window>`.

Fields: listener asset id and cell; interval/jitter and bytes per hour; hosts that checked in (count plus each one's scope entry); sources that saw it and what they emitted; burn events with cause and rotation state; teardown inventory with owner and verification; proof-basis label.
