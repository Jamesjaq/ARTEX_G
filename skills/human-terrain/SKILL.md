---
name: human-terrain
description: "Invoke this skill when a vector's final step is a real person — designing and gating a phishing, vishing, SMS, watering-hole or service-desk scenario, measuring the control around the person rather than the person, and handing the authorised send to a human."
---

# Human Terrain (people-facing vectors, prepared not executed)

The human element is a control surface: the mail gateway, the browser and endpoint stack, the
identity provider, the service desk's verification habit, the approval path behind a callback.
This skill designs the scenario that measures that stack, records what the stack did, and stops
at the line where a real person would be contacted. Done means an approved, in-scope, dated
test with a process result — never a story about an individual.

## Boundaries

- Authorized use only, and authorization here is **specific**: the RoE names the recipients or the
  population, the channel, the window and the consent basis. A scope that says "the domain" does
  not authorise mailing people in it.
- **A human performs the contact.** No seat sends mail, places a call, texts, messages, posts or
  approaches anyone. The deliverable of this skill is a scenario package plus an `ask_operator`
  request naming recipient set, method, window and the clause that authorises it. That is not
  politeness: a message cannot be unsent, so it is in the irreversible class the operator owns.
- No impersonative audio or video generation of any real person, and no clone of a real vendor's
  login page carrying a real customer's credential. Engagement-owned identities and lab-hosted
  lookalikes only.
- Assess the process, not the person. Report "the service desk reset a password after a static
  answer, verified nothing" — never an employee's name, performance or personal detail. A result
  that reads as a verdict on an individual is rewritten before it is filed.
- Vulnerable-person rules: no targeting selected for susceptibility (health, age, distress), no
  lure that exploits urgency or fear beyond the ordinary business pretext, no repeated
  re-contact of someone who did not take the bait.
- Not this skill: the technical delivery chain behind a lure (`initial-access`, `client-side`,
  `payload-lab-development`), mail-store reading after access (`messaging-collab`), site entry
  (`physical-access-advisory`), our own indicator hygiene (`opsec-tradecraft`).
- Stop condition: if the RoE does not authorise a human-facing test, produce the assessment plan
  and the exposure statement and mark the test NOT ASSESSED. That is a legitimate outcome.

## Procedure

Do not build a lure (step 4) before step 3 has produced a named recipient set inside the RoE.

1. **Read the authorization as a set of permissions, not a vibe.** Extract from the RoE: allowed
   channels, allowed pretext domains, recipient population, window, prohibited content, the
   callback/deconfliction contact, and the reporting limit. Anything the RoE does not name is out.
2. **Choose the control question before the scenario.** What is being measured: gateway
   detection of an unexpected attachment class, proxy blocking of a newly registered domain,
   MFA-and-verification habit at the desk, or the approval path on a high-impact request. The
   question decides the minimum viable scenario and keeps the run from becoming a contest.
3. **Define the recipient set from the population the RoE grants**, by role and address, with the
   count. Keep it the smallest set that answers the question. Record the set as a scoped asset/fact
   list (`insert_assets`, `record_fact`) so the send and the results line up row-for-row.
4. **Build the scenario package** — all of it, in one place: pretext and its business plausibility;
   the sending identity (engagement-owned, aged per `campaign-infrastructure`); the message body;
   the link or attachment target (lab-hosted, per `client-side`/`payload-lab-development`); the
   interaction the test expects (click, open, reply, approve-or-refuse); the observables (click
   telemetry, gateway verdict, endpoint detonation result, helpdesk ticket); and the stop rule.
5. **Instrument the measurement before the send.** Where will each observable land? Click/open
   events on the lab host, gateway verdicts from the mail-store or the client's reporting, endpoint
   results from `edr-bypass-re`-style lab measurement, ticket text from the agreed contact. An
   unmeasured send produces a finding about nothing.
6. **Set the non-target guardrails.** Exclude named do-not-contact people (executives, the SOC
   contact, anyone the client names), cap re-sends at one, and set a hard abort if a real incident
   is declared (`engagement-deconfliction`).
7. **File the approval request.** One `ask_operator` per scenario: recipient set and count, channel,
   pretext, payload class, window, the clause that authorises it, the stop rule, and who is
   notified if the blue team calls. Include the exact text so the operator is approving what will
   actually be sent, not a description of it.
8. **After a human sends it**, record timing, the per-recipient outcome class (delivered /
   quarantined / blocked / clicked / submitted / reported-to-desk), and the control verdicts.
   Aggregate; do not narrate individuals.
9. **Close the loop**: state which control caught what at which stage, which stage had no
   visibility, and what the client's own detection saw (hand to `detection-engineering`).

## Evidence

Positive: the authorization clause quoted; recipient count and set reference; the exact content
approved; per-stage counts with a denominator (delivered n/N, clicked n/N, reported n/N); a dated
observable per claim; and the control verdict that follows from those numbers.

Negative and honest-gap forms: "no channel authorised for voice — vishing NOT ASSESSED, plan
delivered"; "gateway verdicts unavailable: client declined mail-store access, so delivery outcome
is inferred from our own send logs only"; "0 of 12 clicked within the window; measurement window
was 72h, single send, no re-send permitted". Never claim a human behaved a way without the
observable that shows it, and never count a click that the telemetry cannot date.

## Report shape

`Human-element control gap: <stage> did not <expected behaviour>` — fields: authorization
reference, population and count, channel, content summary (not the PII), per-stage counts with
denominators, the control that did work, the process fix (verification step, approval path,
policy or training control), and the note that no individual is identified in the deliverable.
