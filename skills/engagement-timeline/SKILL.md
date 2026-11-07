---
name: engagement-timeline
description: "Invoke this skill when the engagement needs its chronology: reconstructing what happened, in what order, from graph nodes, worker traces, findings and traffic captures, with every line dated from recorded state and anything undated labelled as inferred."
---

# Engagement Timeline (reconstruction from recorded state)

The timeline is the spine the report and the debrief are read from: which action produced which
access, when, and what proves it. Done means a dated sequence where every line points at a node,
trace, or capture ID, blocked routes sit beside the routes that worked, and no date appears that the
platform did not record. A reconstructed narrative with invented timestamps costs the engagement more
credibility than a gap does.

## Boundaries

- Authorized use: this seat reads engagement state that already exists. It sends nothing to a target,
  probes nothing to "fill in" a missing time, and re-runs nothing for a cleaner sequence.
- Dates come from recorded state only. No timestamp is written from recall, from a plausible duration,
  or from a tool's default clock without saying which clock it is.
- A platform block, a constraint refusal, or an unanswered approval is a dated event in the timeline
  like any other. Record it neutrally as "stopped by the engagement's constraint set"; this seat does
  not propose a way around it and does not characterise it as a defender action.
- Not this skill: what the artefacts may leave as (`evidence-custody`), the finding write-up
  (`engagement-report`), technique-to-ATT&CK verdicts (`detection-coverage`).

## Procedure

Do not go to step 5 until 3 has produced a source inventory with the clock statement written.

1. **Enumerate the sources for the engagement**: `get_task_graph` (per task, including related tasks)
   or `graph_overview` in-task for nodes and lineage; `list_findings` and `list_facts` for outcomes and
   negatives; `list_task_worker_traces` plus `search_task_worker_traces` (or `list_worker_traces` /
   `search_all_worker_traces` in-task) for the runs, and `get_task_worker_trace` for one run's detail;
   `traffic_search` for captured requests with their timestamps; `list_task_credentials` for what was
   registered and when (IDs and dates only — never values). Write the list of sources you actually
   read at the head of the deliverable; a source you did not query cannot support a claim.
2. **Settle the clock before ordering anything.** State which clock each source uses (task container
   time, SIEM/capture time, a timestamp quoted inside target content) and the timezone you will render
   in. Where two clocks disagree, name the offset and which one the timeline treats as authoritative.
   A timestamp that came from inside target-controlled output is noted as the target's claim, not as
   fact.
3. **Harvest events with their IDs.** For each node: its created time, kind, the intent it came from
   and the intent it yielded (`node_detail` for the full payload and lineage edges). For each worker
   run: start, end, the intent it served, and the outcome. For each finding: the action that proved
   it and the traffic IDs bound to it. For each goal: `list_goals` / `prove_goal` state changes, which
   mark the phase boundaries. Keep the ID with the event as you extract it; an event harvested without
   its pointer cannot be written into the table.
4. **Build the spine from objective downward**: goal → the intents tasked against it → the facts each
   produced → the next intent those facts triggered. That lineage, not prose memory, is what decides
   causation. Where the graph shows no edge between two events, they are sequential at best, not
   causal — write "after" and not "because".
5. **Sort and merge by recorded time**, ascending, and collapse each event to one line: date, seat or
   task that recorded it, what was done, what it produced, evidence pointer. Same-second events keep
   their recorded sub-order or are written as simultaneous.
6. **Insert the blocked paths at their own dates**, beside the successes, in the same table: route
   attempted, what stopped it (a control, a missing route, a constraint, an unanswered approval, a
   cancelled window), and the node or trace ID that shows the attempt. This is the part of the
   chronology that makes the eventual success credible, and the part the client's own team can defend
   internally.
7. **Label every inferred entry.** Where an event has no recorded time but has a position in the
   sequence, write it with `inferred (undated; position from lineage)` and no invented timestamp.
   Where a duration is inferred from two records, name both records. Inference is allowed; inference
   dressed as an observation is not.
8. **Cross-check against the finding set**: every finding's reproduction should appear as a dated
   action, and every dated action that produced a claimed impact should have a finding or a fact
   attached. A finding with no timeline entry, or a timeline event with no recorded evidence, is a
   gap — record it as a fact and route it back to the front that owns it rather than smoothing the
   sequence over.
9. **Attribution stays at seat and task level.** "The access front pivoted through host X" is a
   timeline line; anything naming a real person on the client side, or speculating what an individual
   did or knew, is not. Human-facing tests appear only as the operator-reported outcome with their
   approval reference (see `human-terrain`).
10. **Sensitive content**: reference credentials, captured tokens, personal data and bulk output by ID
    and count, never inline; run the result past `evidence-custody` before it becomes a deliverable.
11. **Render two things**: the table (the auditable artefact) and a short critical-path narrative of
    the chain that achieved the objective, each paragraph citing table rows. Date the deliverable and
    the state-read time — the timeline is a snapshot, and later events will exist.

## Evidence

Positive: a sequence where every line has a date from recorded state or an inferred label, a
recording seat/task, and an evidence pointer resolving to a node, trace, finding, or traffic ID;
clock and timezone declared; the source list at the head.

Negative and honest-gap forms: "no recorded action exists between <date A> and <date B> in the
sources read; the gap is unexplained" — written as a gap, not filled with a plausible narrative.
Where a source was unavailable (a task with no traces retained, a capture window that never started),
name the source, the period it does not cover, and what that makes undatable. Never imply a complete
chronology from a partial read.

## Report shape

Deliverable "Engagement chronology": fields per line = timestamp (or `inferred`), clock source,
recording seat/task, action, outcome, evidence pointer, status (success / blocked / not assessed).
Head block = engagement ID, sources read, clock declaration, generated-at, and the critical-path
chain with its table row references.
