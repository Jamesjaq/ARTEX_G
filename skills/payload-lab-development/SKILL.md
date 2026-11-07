---
name: payload-lab-development
description: "Invoke this skill when a payload, dropper or implant artifact is needed: it must be built and detonated only in a lab image the engagement owns, measured against the team's own sensor stack, and reported as a control blind spot rather than shipped as an evasion recipe."
---

# Payload Lab Development (build and lab detonation, never a live host)

An artifact is worth having when you can say what it is, at what version, with what hash, and which sensor caught it at which stage — measured in a lab you control.

Done means a detonation matrix with a named sensor and verdict per cell, mapped to ATT&CK technique IDs, plus a clean split between what becomes a finding (the control's blind spot) and what stays in the lab (the recipe).

"This class is detectable and should stay detectable" is a successful outcome, not a failed build.

## Boundaries

- Authorised use only: build for a purpose the engagement can name, detonate on lab images this engagement owns, or on a client-named lab twin under written approval.
- Nothing built here executes on a production host.
- Lab-only binds hard here: the output is a detonation result and a finding about the control's blind spot.
- Do not write, store or hand on a step-by-step for defeating a production sensor — not in a fact, not in a finding, not in the artifact's README, not in a chat reply. "Injection class produced no event from any sensor in the matrix" is the deliverable; the offsets, generator profiles and bypass ordering behind it are not.
- This skill is NOT endpoint telemetry RE (`edr-bypass-re`), NOT the channel (`c2-ops`), NOT staying on a host (`persistence-cleanup`), NOT whether a platform control neutralises a primitive (`mitigation-status`), NOT turning a crash into a chain (`exploit-development`).
- Stop condition: no compiler toolchain, no lab image, or no isolation from production means you report the gap and stop. A detonation result you did not observe is worse than none.
- Boot and firmware-level artifacts are lab-only at any approval level, because the client cannot evict them.
- Third-party services are not detonation targets: never submit a client-related artifact to a public scanning service to see what it says. That ships the client's material into someone else's telemetry.

## Procedure

1. Write the artifact spec before code: purpose, target platform and version, the one technique class it demonstrates, the intended detonation image, and the kill/removal method.
2. One artifact, one class. Composite artifacts make every later verdict ambiguous.
3. Record the spec as a `record_fact` against your intent id.
4. Build reproducibly in the task work directory: keep source, toolchain name and version, and build flags.
5. Embed a version string in the artifact and run `sha256sum <artifact>` the moment it exists. Hash-before-detonation is a gate, not a formality: an artifact nobody can identify at wrap-up cannot be burned down.
6. Choose the detonation matrix — the grid every variant passes through:
   - sensorless baseline image;
   - Windows Defender at current definitions with default policy;
   - Sysmon with the config actually deployed;
   - the client's EDR product and version from a trial install, isolated from their production tenant.
7. Keep every image airgapped from the client's network. An unavailable row reads "not run", and the matrix is then reported as incomplete.
8. Check the environment first: this task image ships no compiler for PE targets, no Windows guest tooling and no hypervisor.
9. Lab VM provisioning, sensor trial installs and console access are operator actions — request them through `ask_operator` and wait. Never fabricate a console alert to fill a cell.
10. Detonate one cell at a time and capture per stage: write-to-disk, first execution, network attempt, child process, persistence write.
11. For each stage, record which source alerted and its text verbatim (AV real-time, AMSI, ETW-TI, Sysmon event ids 1/7/8/10/11/25, EDR console) and what it blocked.
12. One `record_fact` row per cell. Merged observations go in that row's detail, not into ten rows.
13. Baseline before claiming a delta: run the sensorless image and the instrumented image on the same artifact and hash, so "no telemetry" cannot be an artefact of the run rather than of the control.
14. Convert results into a control statement, then map it: coverage language per class ("<class> produced no event from <source> across <n> of <N> detonations"), with the ATT&CK technique and sub-technique IDs attached.
15. Label proof basis executed-lab with sensor version and date.
16. Where a class was caught, say so and stop treating it as a gap.
17. Decide what ships and what stays. Ships: technique IDs, sensor versions, lab conditions, per-stage verdicts, the coverage gap, and the recommended detection surface.
18. Stays in the lab: the binary, the source, the recipe that turned a caught class into an uncaught one, the evasion ordering, and generator profiles.
19. If blue needs a re-run for a purple session, hand the operator the hash and the artifact by reference — their decision, not yours.
20. Account for noise and side effects: what the detonations generated on shared lab infrastructure (alert storms, quota burn, an EDR trial hitting a licence limit). Honest reporting includes the mess.
21. Burn-down: mark each artifact active / superseded / retired with its hash. Retired artifacts are destroyed with an attestation and land in the teardown inventory alongside `campaign-infrastructure`'s hosts.
22. Reuse gate: an artifact built for another engagement does not run here — its hash is already somebody's telemetry. Check the ledger in `opsec-tradecraft` before any reuse decision.

Gate: do not go to 10 until 4-7 produced a hashed artifact and a matrix with named sensor versions; do not go to 14 until every cell reads observed, caught, or not run.

## Evidence

- Positive: artifact hash and version; each lab image with sensor product, version, date and isolation state; the full matrix with per-stage source names and alert text; technique IDs; and the coverage conclusion stated as a gap or as an effective control.
- A console alert from a live production endpoint is not this skill's evidence — report that as an availability incident instead.
- Negative, honestly written: "process-injection class: 6 detonations across 4 lab images over <window>; no event from AMSI, ETW-TI, Sysmon 8/10 or <EDR> console — reported as an uninstrumented surface, not as channel safety."
- Where a sensor could not be installed or reached, write NOT ASSESSED with the blocking reason.
- Weak-signal honesty: one un-alerted run is `inferred`; a class needs repeats before it is called a gap. Never write "undetectable" — write which named sources did not emit, and which were never present.
- Signing keys, certificates and captured secrets go to `save_task_credential` and are cited by ID; no key material or token in prose.

## Report shape

Title: `Lab detonation: <sensor product version> emits no telemetry for <technique id> at <stage>`.

Fields: artifact hash and version; technique IDs; lab image inventory with isolation status; matrix completeness (n of N cells run) plus the baseline row; per-stage alert text; sensor version and date; executed-lab marker; what was withheld as lab-only; and the burn-down state.
