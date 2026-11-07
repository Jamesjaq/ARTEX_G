---
name: opsec-tradecraft
description: "Invoke this skill when an action will leave an indicator the defender can collect, or when an identity, domain, host, certificate, account or artifact is being reused within or across engagements: maintain the indicator ledger, decide reuse, pace activity, and state what the SOC will see and whether that is intended."
---

# OPSEC Tradecraft (indicator ledger, reuse decisions, exposure budget)

Tradecraft here is measurement of our own footprint, not invisibility: every action emits indicators, the defender correlates them over a window, and the engagement's remaining access is what survives that spending.

Done means a current ledger, a recorded decision for every reuse call, an exposure statement per tasked action ("the SOC will see X, via Y, within Z — intended / not intended"), and a post-run audit of what we actually revealed versus what we meant to.

## Boundaries

- Authorised use only: this governs indicators this engagement generates against its own scope, on its own infrastructure. It never applies to a host, identity or service outside the registered scope.
- The line to hold: observing and modelling what a defender can see is assessment. Acting to disable, blind, poison, tamper with or spoof a real production sensor, log pipeline or appliance is an availability and safety action — not tradecraft. It needs the operator's explicit approval for that host, and the usual answer is a lab equivalent. Prefer the lab and say which one you used.
- Nothing here proposes modifying the platform's guard or intercept configuration, deny rules or enabled-tool set to reduce exposure. Those surfaces belong to the human operator, and a block arriving from one is the platform doing its job.
- This skill is NOT channel and beacon mechanics (`c2-ops`), NOT provisioning, aging and burn-down of hosts and domains (`campaign-infrastructure`), NOT operator identity separation and client-data containment (`secure-operator-comms`), NOT lab detonation results (`payload-lab-development`), NOT recognising a synthetic surface (`deception-detection`), NOT what to do once a tripwire has fired (`engagement-deconfliction`).
- Stop condition: when an indicator is burned, activity using it stops before it is re-planned. "One more check-in" on a burnt address is how a client's IR team ends up with a coherent timeline of an engagement they can no longer authorise.

## Procedure

1. Open the ledger at engagement start, one row per indicator, carrying: value or class, type, owning cell, first-used date, engagements touched, status, and burn evidence.
2. Cover the indicator types that actually get collected: source IP, egress range, domain, URL path, TLS certificate and connection fingerprint, file and binary hash, named pipe, mutex, service name, registry path, account name, mail template, user-agent string, jitter rhythm, tooling path, timezone and locale.
3. Client-visible values go by reference (asset id or credential id), never pasted live.
4. Register before use, not after: each new piece of infrastructure or artifact is in the ledger and in `insert_assets` / `record_fact` before it emits anything.
5. An indicator that is not in the ledger cannot be rotated, and will be rediscovered as a target by another seat.
6. Run the pre-action exposure check on every task you accept or hand down, in four questions:
   - what does this emit, in ledger classes;
   - who is positioned to collect it — endpoint sensor, proxy/NDR, mail gateway, WAF/CDN, SIEM correlation, third-party reputation feed, human analyst;
   - how fast does it arrive — seconds, check-in cycle, daily telemetry batch, feed refresh;
   - does it connect to what came before, inside that defender's correlation window.
7. Record the answer as a fact before execution. It is the artefact the auditor later checks against what actually happened.
8. Score loudness against the sequence, not against silence: a quiet step inside a loud sequence is still loud.
9. If a front is already generating alerts, either hold the next step or record that it will be correlated. Do not pretend the sequence resets.
10. Treat pacing as a decision with a stated reason: choose check-in density, sweep breadth and burst size from the engagement's noise budget, and write down why.
11. The audit asks later whether the tempo was justified or careless. Timing a step to a defender's shift change is still an exposure choice with a date, and low-and-slow is not a substitute for an objective.
12. Reuse decisions, with the gate. Before reusing anything, read its row:
    - touched another engagement → do not reuse, no convenience exceptions;
    - burned anywhere → mark it burned and rotate;
    - firm-level shared identity (an abuse contact, the firm's domain) → may cross only if the register says that class may, and record where the permission came from.
13. Reuse to save lead time costs remaining access. Write the denial as a fact so the lead budgets re-provisioning through `campaign-infrastructure`.
14. Compartment audit as you go: verify no single indicator spans two functions (payload hosting plus listener, mail plus redirect).
15. One burnt cell must not cascade; where it already does, that is a finding about our own preparation, recorded for the operator.
16. Post-run audit at each phase boundary and at wrap-up. Sweep what we actually revealed:
    - `search_all_worker_traces` and `get_worker_trace` for commands, paths and accounts used;
    - `traffic_search` / `traffic_get` for the bodies and headers we sent;
    - `Read` and `Grep` over the work dir for absolute paths, hostnames, personal usernames, real client names and pasted secrets.
17. Diff observed indicators against the ledger: every unlisted indicator is a gap in the ledger, every listed indicator nobody checked is stale.
18. Write each hit up as a dated finding for your lead with the exact location (trace, traffic id, file path) and one concrete remediation — rotate, restrip, destroy, or re-scope.
19. Do not silently edit another seat's evidence, and do not delete a file that is evidence of what leaked.
20. Hand the exposure statement to the lead at close: indicators spent, indicators burned, what the defender demonstrably correlated (per `detection-coverage`), what remains usable, and which future engagements are affected by what we revealed here.
21. That last line is the one clients and partners care about most.

Gate: do not go to 12 until the ledger from 1-3 records first-use dates; do not go to 20 until 16 produced a diff, even if the diff is empty.

## Evidence

- Positive: the indicator with type, cell, first-use and status, plus the observed location proving it was emitted (traffic id, trace line, log excerpt), and the defender surface expected to hold it.
- An exposure claim carries the same proof bar as a finding: cite the artefact, not the intent.
- The coverage-style statement is tradecraft's real output: "the SOC's NDR should correlate <source IP> to <N> check-ins over <window>, from <source list>; alert fired / did not fire per `detection-coverage`." Say plainly whether that visibility was intended by the RoE.
- Negative: "no cross-engagement reuse found — 23 ledger rows checked against engagement history for <window> via <method>; zero shared indicators."
- And the honest gap: "TLS certificate and connection-fingerprint classes never recorded; reuse status unknown → NOT ASSESSED, ledger gap." A negative with no method is worthless; a method with no result is a result.
- Never write a plan to blind or alter a defender's sensor into any artefact. Where the temptation exists, record the constraint as a fact for the lead and let them raise it with the operator.
- Captured credentials and keys stay in `save_task_credential`, cited by ID; no live value enters a ledger row, fact, finding or report.

## Report shape

Title: `Engagement exposure: <indicator class> reusable / burned — <correlation consequence for the client's stack>`.

Fields: indicator reference; type and cell; first and last use dates; engagements touched; emission evidence location; defender surfaces that plausibly hold it; correlation window; intended-versus-actual visibility; reuse decision with reason; rotation or destruction action; residual exposure budget handed to the lead.
