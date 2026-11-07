---
name: engagement-deconfliction
description: "Invoke this skill when the real world reacts to the engagement: a SOC or blue-team contact, a vendor or ISP abuse call, a WAF block page, a ticket from an owner, a phone call, or a client who says stop. Contain the engagement, classify the signal into one of three cases, and hand the operator one question."
---

# Engagement Deconfliction (a tripwire fired: contain, classify, hand up)

The job at this moment is containment of the engagement, not of the target. Someone outside the bench now knows something.

The only safe response is to stop, capture what fired, work out which of our actions caused it, classify the signal honestly, and put one recommendation in front of the operator.

Done means activity is paused, evidence is preserved, the classification is written, and the decision is a human's.

## Boundaries

- Authorised use only, and authorisation is exactly what makes the reaction survivable. Nothing here is a guide to talking a defender out of their job or to keeping activity running against a client who said stop.
- The owner's "stop" is terminal until the operator says otherwise, and so is any credible request from a named blue contact.
- Never answer a suspected-real incident with more activity to "see what happens".
- Never contact the client's SOC, IR team, an abuse desk, a vendor or an ISP ourselves, and never impersonate a real party — not the client, not the firm, not a domain registrant. Contact with the client's defenders is the operator's call, and only theirs.
- Do not weaken the platform's safety surfaces while containing: no proposal to edit guard or intercept rules, deny lists or the enabled-tool set to make the reaction go away. A block or refusal produced by this platform is the platform working, not the target's defence — it is not a tripwire, and it is not something to route around.
- This skill is NOT recognising a synthetic surface before it catches anyone (`deception-detection`), NOT what the defender saw and how their coverage held (`detection-coverage`), NOT the ledger that says which indicator burned (`opsec-tradecraft`), NOT anything whose final step is a person (`human-terrain`), NOT the RoE that names the blue contact in the first place (`engagement-scoping`).
- Stop condition on data: preserve before you clean. Do not delete, rotate or restripe anything until the evidence of what fired and what caused it is captured — cleanup destroys the incident record otherwise.

## Procedure

1. Freeze first. `pause_task` on the front that plausibly caused it.
2. `kill_work` on running intents against the named asset, its same-segment neighbours, and anything sharing that session, credential or indicator.
3. Stop new tasking yourself and `add_task_hint` so every seat sees the freeze.
4. Do not close channels yet: closing destroys evidence and, for a listener, can read as escalation.
5. Capture the trigger artefact verbatim with exact time and timezone: the mail, ticket or call notes (who, which address, number, quoted text); the block or error page body and its full response via `traffic_get` / `traffic_blob`; any rule or request id; the URL and asset; the channel it arrived on.
6. Write it as a `record_fact` with confidence=observed, and keep the raw file in the task work directory.
7. Build the candidate-cause list. Walk backwards from the trigger time and pull `get_worker_trace` / `search_all_worker_traces`, `graph_overview`, `list_assets` and the `opsec-tradecraft` ledger rows.
8. For each candidate record: which action touched that asset, from which source indicator, at what minute, and what else was running concurrently.
9. Include benign possibilities: their own change window, a scheduled scan, another vendor, an unrelated customer of the same ISP, or an alert from something we never touched.
10. Label each correlation honestly: strong (same asset, same indicator, minutes apart); plausible (same asset, different source, wider window); weak (adjacent, timing circumstantial); or no credible link.
11. Over-attributing to us wastes the engagement's remaining goodwill; under-attributing surprises the operator mid-decision. Say which way the evidence leans and how far.
12. Map blast radius: which fronts are now unsafe, which indicators the defender now holds, whether a listener or implant is still resident on a client-side host, and whether anything still running could re-trigger. This decides what stays paused after the event.
13. Classify into exactly one of three cases, and state the tests you applied, because the wrong label is how an engagement becomes an incident:
   - Case A — agreed blue-contact deconfliction: the contact arrived on the channel named in the RoE, from a person or address listed as the engagement's blue/IR contact, consistent with the agreed process. Verify against the RoE contact list and the channel set recorded in `secure-operator-comms`; if the RoE names no blue contact, it is not Case A.
   - Case B — real unsolicited complaint: a party not in the RoE contacted us — an ISP or hosting abuse desk, a software vendor, a different business unit or tenant, an unrelated customer, an internal team never told — or a named contact reports an effect outside the agreed scope, or outside the agreed window. Treat as real until the operator rules otherwise, and assume a third party may already hold data about us.
   - Case C — sensor artefact: we can see a trip or block from our own position with no human reaction attached — our request hit a WAF page or rate limit, a decoy signalled (see `deception-detection`), a listener log shows a rejected check-in. No external party has contacted anyone; beyond preserving the observation and adjusting tradecraft, no human action is required.
   - A block issued by this platform's own guard or intercept layer is none of the three: it is an internal control and never counts as defender contact.
14. Compose the handoff as one `ask_operator` containing: the trigger artefact with when and how it arrived; the classification with the tests applied; the likely cause with correlation strength and timestamps; the blast radius; and one recommendation drawn from — continue as scoped / stand down this front / stand down the engagement / notify the blue contact / destroy a named resident artefact now.
15. State plainly what you have already paused, then wait. Do not run more activity while waiting.
16. Record the decision as a fact, then execute exactly it: resume fronts, keep them paused, or begin teardown. Teardown follows `c2-ops` and `campaign-infrastructure` discipline, and any captured secret material goes through `save_task_credential` by reference rather than into prose.
17. Record the defender's view for the report: what they saw, how fast, through which source, and what their people did about it — written up through `detection-coverage`. It never includes a lesson in how to avoid being caught next time.
18. Write the post-incident list for the next engagement: what about our preparation caused the exposure — a missing blue-contact row in the RoE, unaged infrastructure, a reused indicator, an unplanned action. Send it to your lead as a fact, and never quietly tidy another seat's evidence to make the timeline read better.

Gate: do not go to 7 until 1 has frozen activity and 5 has preserved the artefact; do not go to 16 until the operator has answered 14.

## Evidence

- Positive (an event record): trigger artefact verbatim with source, exact timestamp and timezone; sender named or explicitly named-unknown; candidate causes ranked with correlation strength and an evidence line for each; the classification with its tests; blast radius list; the operator's decision text with timestamp; and the actions taken after it, per front.
- Negative / non-event, honestly written: "trigger observed at <time>: WAF block page on 2 of 41 requests to <asset>; no external contact received within <window>; classified Case C on the basis that no party outside the engagement has communicated. Fronts resumed after the operator's confirmation."
- "No contact received" is bounded by the window you actually watched — state it. A block page nobody outside saw is a control observation, not a deconfliction event.
- Where the classification cannot be settled, record the ambiguity and leave the front paused: "Case A vs B unresolved — sender address absent from the RoE contact list and unverified; recommendation was stand down pending operator check."
- Never paste a third party's personal contact detail, a token or a client credential into a fact; cite by reference and let the operator's channel hold the rest.

## Report shape

Title: `Deconfliction event <date>: <case A|B|C> triggered by <action> against <asset> — <disposition>`.

Fields: trigger artefact and arrival channel; sender and verification status; timestamps in the engagement timezone; likely cause with correlation strength; affected fronts and indicators; classification with the tests applied; operator decision and who gave it; actions executed; defender-visibility summary for `detection-coverage`; resident artefacts needing client-side removal; and the preparation-lessons list.
