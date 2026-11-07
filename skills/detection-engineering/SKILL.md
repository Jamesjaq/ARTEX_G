---
name: detection-engineering
description: "Invoke this skill when a demonstrated technique needs a rule the defender can actually run: pairing a missed or partial detection-coverage verdict with a Sigma/SPL/KQL/EQL analytic drafted against named event fields, tested in the lab first, and shipped with its false-positive cost."
---

# Detection Engineering (demonstrated action to runnable rule)

A coverage gap only closes when someone can deploy a check for it. Done means: one analytic per
actionable gap, written against event fields that exist in the client's own data, fired once in a lab
stream before it is offered, and accompanied by its measured noise, its blind side, and the source it
depends on. A rule that pages an analyst every ten minutes is a rule the SOC will silence — the
false-positive estimate is part of the deliverable, not a footnote.

## Boundaries

- Authorized use: analytics are drafted from what the engagement demonstrated, and tested against
  lab data or historical data the defender supplies. Nothing here runs a detection test against a
  production sensor to see what it does, and nothing edits the client's SIEM, EDR policy, or rule
  state — deployment is performed by the defender, in their change process.
- Behaviour over indicators: a rule keyed to the exact binary name, hash, or domain this engagement
  happened to leave behind dies at the next rebuild. State which variant the behavioural rule will not
  catch instead of writing the brittle one.
- Never recommend disabling, sampling away, de-scoping, or muting a sensor to improve a number, and
  never recommend a rule whose real effect is to break a workflow the client depends on.
- Payload construction for a lab detonation belongs to `payload-lab-development`; the endpoint
  telemetry RE behind it to `edr-bypass-re`; the verdicts that feed this queue to
  `detection-coverage`; the source-enablement asks to `telemetry-sources`.
- Stop condition: if the event fields the rule needs are not confirmed to exist in the client's data,
  the analytic is a draft with a named dependency, not a recommendation.

## Procedure

Do not go to step 4 until 2 and 3 have confirmed the log source and its real field names.

1. **Take the queue from coverage verdicts.** Work the *missed* and *partial* rows that are the
   client's to fix, ordered by the risk the demonstrated action carried, not by how easy the rule is.
   A *not instrumented* row is not your job yet — it needs a source first; route it and record the
   dependency.
2. **Name the event source, then confirm the fields as the client actually sees them.** State the
   sensor and channel: Sysmon `EventID` 1/3/7/11/13/25, Windows Security `4688` (command-line auditing
   must be on for the token to exist), `4672`, `4768/4769/4776`, `5140/5145`, `7045`, `4698/4702`,
   `1102`; EDR process-create/network/file events with their `Image`, `ParentImage`, `CommandLine`,
   `TargetFilename`, `InitiatingProcess` names as that vendor spells them; Linux `auditd` `execve`
   or eBPF process events; cloud control plane (CloudTrail `eventName`, `sourceIPAddress`,
   `userAgent`, `recipientAccountId`; Azure Activity `operationName.value`; GCP
   Cloud Audit Logs `method`/`principalEmail`); object-store access logs; identity sign-in logs
   (`conditionalAccessStatus`, `authenticationDetails`, `isInteractive`); network (Zeek `conn`/`dns`,
   NetFlow 5-tuple plus bytes and sampling rate, proxy `url`/`useragent`/`action`). If you have not
   seen an event of that type in their data, say so and request one sample event via `ask_operator` —
   a rule with guessed field names is a broken rule the client discovers on deployment day.
3. **Write the detection logic as behaviour.** Choose the shape that survives a rebuild: process
   lineage (an Office or a browser parent spawning a shell), argument structure (encoded command,
   remote tool download to a user-writable path, service creation with a binary outside the expected
   directories), sequence and impossibility (a credential used from two geographies inside the
   tolerance window, a role assumption with no console session), permission-ACL change pattern,
   persistence written outside the sanctioned paths, egress to a never-seen destination at a volume
   above the host's own baseline. Prefer one precise condition over a list of ten.
4. **Draft in the portable form, then translate.** Sigma as the reviewable draft — `logsource`
   (`product`/`category`/`service`), `detection` with named selections and an explicit `condition`,
   `falsepositives` filled with real benign sources, `level`, `tags` carrying the technique ID
   (`attack.t1053.005` style), `references`, and `status: experimental` until it has fired in lab.
   Then render the client's language: SPL, KQL, EQL, or the vendor's own rule syntax. Match their rule
   naming convention and their severity vocabulary; an analytic that cannot be found by name in their
   console is not deployable.
5. **Test in the lab before asking anyone to deploy.** In the isolated image with the same sensor and
   version (see `payload-lab-development`, `edr-bypass-re` for the lab method): run the demonstrated
   action, confirm the event lands, run the drafted rule over the lab stream, and record the fire —
   event ID, rule name, timestamp, and the field values that matched. Also run the benign baseline on
   the same image (typical admin work: patching, RDP session, installer, backup job) and confirm the
   rule does **not** fire. Lab fire first, production review second; state the lab conditions with the
   result.
6. **Measure the false-positive cost, or name it unmeasured.** Ask the defender (via `ask_operator`) to
   run the draft over at least seven days of historical events, or a sample they supply, and report:
   total hits, the top benign sources by user/host/process, the tuning applied to cut them, and the
   resulting expected pages per shift. Where historical data is unavailable, write
   `untested against historical volume` on the rule and name what would test it — do not publish an
   estimated FP rate you have not counted.
7. **State what the rule cannot see**, in the same block the rule ships in: the evasion that walks past
   it (unsigned-parent abuse, live-only tooling, a different logon type, a source that is off), the
   platforms it does not cover, and the dependency it inherits (a field to collect, a source to enable,
   a retention longer than the detection window). Pair each with the compensating row from
   `telemetry-sources` rather than leaving the reader to assume the rule is complete.
8. **Give the deployment cost and risk**: ingest volume added, licence tier implied, performance
   overhead, whether the rule needs a change window, and which existing rule it overlaps or should
   replace. Rules that duplicate an existing analytic are tuning requests, not new deployments —
   check before authoring.
9. **Ship it as an artefact plus a fact**: rule text at a path in the engagement work dir, a fact
   naming the technique ID, the log source, the fields used, the lab-fire record, the FP measurement,
   the blind side, and the client-side owner who deploys. Findings that need a detection reference get
   the pointer via `report_finding`/`update_finding_report`, and the prose goes in the report through
   `engagement-report`.
10. **Version and review**: each analytic ships with a version, the data it was tested against, and a
    review date. Ask for the first two weeks of true/false positive counts back from the defender and
    tune once, recording the tuning as a new version rather than a quiet edit.

## Evidence

Positive: the rule fired on a lab event (event ID + rule name + timestamp + matched field values) and
did not fire on the stated benign baseline; FP count over a named data period with a named denominator;
field names verified against a real sample event from the client's own data.

Negative and honest forms: "no analytic drafted — required field `<x>` is not present in the client's
`<source>`; dependency logged to `telemetry-sources`"; "rule drafted and lab-fired but untested against
historical volume: defender data unavailable in window — deployment risk stated as unknown-noise";
"no detection possible from this source alone: the action leaves no event because `<reason>` —
compensating control proposed instead". Never present an untested rule as validated, and never report
a clean lab run as a coverage claim about production.

## Report shape

Row pattern: `Detection analytic <rule name> for <technique ID> — <logsource>, lab-fired <yes/no>,
FP <n>/<period>, blind to <stated>`.
Fields: technique ID, intended verdict target (the coverage row it closes), event source and field
names, rule text path, lab conditions and fire record, FP measurement, expected pages per shift,
dependencies, deployment owner, version and review date.
