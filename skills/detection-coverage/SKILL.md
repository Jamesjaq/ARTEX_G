---
name: detection-coverage
description: "Invoke this skill when the engagement must state what the defender's sensor stack actually saw: measuring a purple or authorised red-run against the client's telemetry, tagging each demonstrated action with an ATT&CK technique and returning a covered/partial/missed/not-instrumented verdict."
---

# Detection Coverage (what the sensor stack saw, mapped to ATT&CK)

Coverage is the blue-team-side measurement of the same run: for every action the engagement
demonstrated, one verdict on whether the stack produced an event, whether a detection consumed it,
and whether an alert reached an analyst. Done means a technique-by-technique table with a denominator,
the honest "not instrumented" rows separated from the misses, and a gap list someone can task from.
The number that matters is the coverage rate, not the count of techniques that evaded.

## Boundaries

- Authorized use: the run being measured happened inside the engagement's scope; purple-team pairing
  is consented and scheduled with the defender, never done to them. If the defender is not in the
  loop, the deliverable is a coverage statement from what we can see plus an explicit limitation, not
  a guess at their telemetry.
- Detection-evasion content stays lab-only. Characterising a control's blind spot happens by
  detonating in an image the engagement controls, with the product, version and lab conditions named
  in the result. Nothing here runs an evasion against a production sensor to see what happens; where
  only production exists, the technique goes in **not tested** with the reason.
- No recommendation to disable, blind, sample away, or de-scope a defender's sensor. That is an
  availability action against the client's own protection.
- Sensor presence and configuration questions are answered from client documentation, policy state and
  permitted read access — never by tampering with or uninstalling an agent.
- Not this skill: writing the analytic (`detection-engineering`), the sensor estate and what is off
  (`telemetry-sources`), endpoint telemetry RE method (`edr-bypass-re`), the final prose
  (`engagement-report`).

## Procedure

Do not go to step 4 until 2 and 3 have produced the action list and the paired data request.

1. **Fix the technique universe from what was demonstrated.** Pull the actions from recorded state —
   `list_findings`, `list_facts`, and the worker traces that carry the actual command or request
   (`get_worker_trace`, `search_all_worker_traces`). Tag from what ran, never from what a tool is
   nominally capable of. An attempted-and-blocked action still earns a row: it is coverage evidence.
   An imagined action earns nothing.
2. **Write one row per demonstrated action** with: the action in the tool's own terms (the command,
   the request, the API call), the platform and host record it ran on, the time window (start and end,
   timezone), the intended telemetry source that should have produced an event, and the technique tag.
   Technique selection: ATT&CK Enterprise for host and cloud work, ICS for a control network in scope,
   ATLAS where the target was an AI/agent system; use the sub-technique where one applies, and cite
   the technique ID exactly. On the mitigation side, pair with the D3FEND counterpart when the client
   reads in those terms.
3. **Request the defender's answer per action** (`ask_operator`, one query at a time, with the window
   and the source named). The SOC-side pull is performed by the human: they run the SIEM query or open
   the EDR console and return the event count, the event sample with field names, the detection rule
   or alert name if one exists, the alert timestamp, and the analyst role who confirmed. The platform
   has no read path into the client's SIEM — record that as the method limit rather than inventing
   telemetry you did not see.
4. **Assign a verdict from one fixed scale and never blur it:**
   - **covered** — an event existed, a detection consumed it, and it surfaced as an alert or a
     named rule hit inside the response window.
   - **partial** — an event existed but detection was absent, or it matched a rule that was disabled,
     untriaged, suppressed, or delivered to a dashboard nobody reads. State which half failed.
   - **missed** — the source collects the event, the event was present for the window, and nothing
     detected it. This is the row detection engineering works from.
   - **not instrumented** — the event class this action would produce is not produced, collected, or
     ingested in this stack at all. Different cause, different fix, different owner from *missed*;
     cite the `telemetry-sources` row that proves it (a source that is off, unlicensed, or not
     deployed to that segment).
   - **not tested** — the action did not run in this engagement, or ran but its telemetry could not
     be read. Name the reason: no authorised window, lab-only technique, defender unavailable, sensor
     data not retrievable.
5. **Compute the coverage with the denominator stated**: N demonstrated actions over the run, of which
   covered c, partial p, missed m, not instrumented i, not tested t — with the platform/segment
   breakdown, because a rate from one host does not describe an estate. Say explicitly that *i* is a
   data-availability problem and *m* is a detection-quality problem; reporting the second as the first
   sends the client to buy something they already own.
6. **Add the response-side measures where the defender can supply them**: time from action to first
   event visible (ingest lag), time from event to alert, time from alert to analyst action, and
   whether the alert text named the right host and process or only a generic family. A detection that
   fires four hours late and lands in a low-priority queue is a partial, and say so.
7. **Matrix view** at the end: techniques by tactic, verdict-coloured, with untested cells rendered as
   untested. An empty cell that reads as a pass is the most expensive error in this report.
8. **Route every gap**: *missed* and *partial* rows go to `detection-engineering` as analytic
   candidates; *not instrumented* rows go to `telemetry-sources` as enablement or gap statements;
   rows where the fix is a platform capability or configuration question go to `mitigation-status`.
   Cross-reference each row to the finding it explains, so the client can read a gap next to the risk
   it enabled.
9. **Label lab versus live per row.** A verdict earned in the lab image says lab, names the product and
   version, and states that no production sensor was exercised to obtain it. A live verdict names the
   consented window and the defender who supplied the data.

## Evidence

Positive: a row with a technique ID, the action verbatim, the window and clock source, the named
telemetry source, the defender-returned event sample with its field names and alert name, and the
verdict. Coverage claims carry counts and denominators, and the source of each number (which query,
over which period, supplied by whom).

Negative and honest forms: "action X run at T on host H; queried source D over T to T+30m via the
SOC; zero events returned — verdict missed" with the query text quoted. For an unreadable source:
"verdict not testable: no read path to <source> in this engagement; no event data returned; technique
recorded as not tested, not as covered". Never convert an unanswered query into a clean result, and
never report an absence of telemetry as evidence of stealth.

## Report shape

Row pattern: `<technique ID> <technique name> — <verdict> on <platform/host> (<source>, <window>)`.
Fields per row: action verbatim, intended source, event present Y/N, detection present Y/N, alert
surfaced Y/N, lag to first visibility, lab or live, evidence pointer (event ID / alert ID / node ID),
gap owner. Summary block: the N/c/p/m/i/t counts with denominators per segment.
