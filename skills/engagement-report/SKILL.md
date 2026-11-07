---
name: engagement-report
description: "Invoke this skill when findings exist and the engagement needs its deliverable: assembling confirmed findings plus the coverage statement into a report a non-technical reader can act on and a defender can reproduce from, with severity by a stated method, evidence pointers, and no live secrets or bulk data."
---

# Engagement Report (defensible deliverable)

The report is where the engagement's value becomes auditable: one finding, one proof, one stated
severity, one pointer back to recorded state. Done means a reader can answer "how do you know" for
every line without calling anyone, and a defender can reproduce each finding from the pointers alone.
A report that cannot survive that question is not late — it is wrong.

## Boundaries

- Authorized use: this seat writes about assets inside the registered scope, and only about work the
  platform recorded. A finding on an unregistered asset is a scoping incident, not a report line —
  hand it to `engagement-scoping` and leave it out of the body.
- Redaction and release are `evidence-custody`'s call, not yours; you prepare the text so it can pass
  that check, and you do not decide what may leave the platform.
- Delivery to the client is performed by the human operator over the channel they name
  (`secure-operator-comms`). This seat drafts; it never sends, publishes, or uploads a deliverable.
- You do not manufacture findings to fill the report, and you do not soften a severity to make a
  number land. Zero findings plus a coverage statement is a complete deliverable.
- Not this skill: chronology (`engagement-timeline`), detection rules (`detection-engineering`),
  fix text (`remediation-advisory`), impact derivation (`impact-analysis`).

## Procedure

Do not go to step 4 until 2 and 3 have produced a finding set with proof basis and a denominator.

1. **Pull the state you will report on**: `list_findings` for the finding set, `node_detail` per
   finding for full evidence, `get_finding_traffic` for bound request/response bodies (read them in
   segments with `binding_id`/`side`/`offset`; keep the version you read for `update_finding_report`),
   `list_facts` for the negatives, `list_untested_assets` for the coverage denominator. Write from
   these rows, not from conversation memory.
2. **Classify every finding by proof basis** and put it in the text: executed-live, executed-lab, or
   static-analysis-unexecuted. Anything never actually triggered reads "potential/suspect" in the
   title and summary, never confirmed, whatever its CVSS score says.
3. **Reconcile the finding set.** Same root cause, same endpoint, same parameter: extend the existing
   finding's evidence rather than adding a second row (note the dedup you did as a fact). Different
   entry point, parameter, or privilege level: separate findings even when one narrative covers them.
   Drop anything whose evidence pointer does not resolve — a finding with a dead pointer is a claim.
4. **Executive summary from impact, not technique.** Two paragraphs maximum. Sentence shape: *who can
   now do what, to which business system or data class, with what measured consequence*. No CVE
   identifiers, no tool names, no payload syntax before paragraph three. Answer the reader's actual
   questions: what is exposed, is it being exploited now, what must change first, what did not get
   tested. If the objective was reached, that is the first sentence; if it was not, say so plainly
   and name what stopped it.
5. **Severity with one stated method, applied consistently.** Pick the method at the top of the report
   and name it — CVSS v3.1 vector plus the environmental reasoning, or the engagement's business
   impact ladder — then show it per finding: the vector string, or the impact sentence with the
   number and denominator that justifies the band. Severity may not outrun the measurement. Where no
   number was obtainable, say what was attempted and why the impact is not measurable, and band it
   as derived rather than inflating the adjective. An unearned critical is the first line the client
   disputes.
6. **One finding, one proof, in a fixed order.** Per finding: title (asset/entry point plus attacker
   outcome), affected assets and their IDs, precondition, the mechanism in two or three sentences a
   developer can act on, the proof chain (the request that triggers it and the response that shows
   the goal achieved, cited by traffic ID and role, or the command plus its key output lines), the
   measured impact with its denominator, what the defence did block, then remediation and detection
   cross-references. Reproduction steps reference bound traffic IDs; they do not restate them.
7. **Evidence by pointer.** Node IDs, finding IDs, traffic IDs with role, credential IDs from the
   credential store, file paths inside the engagement work dir. One evidence index at the back maps
   every pointer to what it proves. Never inline a value a pointer already carries.
8. **The exclusion list, applied before drafting.** No live secret, hash, token, API key, cookie, or
   private key in prose — reference the credential-store ID. No raw dump, table export, or bulk query
   result — class, count, denominator, and at most two or three representative rows with personal
   fields reduced to type-and-ID. No victim PII: a real customer record is described by class and
   count, not quoted. No full request or response bodies with session cookies or bearer tokens.
   No unredacted hostname, IP, username, or employee name in a line the client will forward
   (placeholders are fine in anything illustrative). Screenshot lines get checked for the URL bar,
   token panes and incidental names.
9. **Coverage statement, own section.** Three tables: tested and clean (class, surfaces, method,
   window), blocked or abandoned mid-path (what stopped it), and NOT ASSESSED with the RoE clause or
   missing capability behind it. An unassessed category must never render as a pass, and a check that
   cannot report that it did not run is not a check.
10. **Appendices**: scope summary as agreed (from `engagement-scoping`), environment notes (lab image,
    sensor version where a lab verdict is claimed), tooling, and the chronology pointer
    (`engagement-timeline`) rather than a second narrative.
11. **Self-defence pass over the draft**, line by line: does this sentence have a pointer, does this
    severity have a number, does this claim need a retest to be written as fixed. Anything that fails
    gets downgraded to an observation or removed and the removal recorded as a fact. Then hand the
    draft to `evidence-custody` for the release check and to the advisory lead seat for the
    defensibility pass; revise with `update_finding_report` and version the file with a date.

## Evidence

Positive: each reported finding carries a proof basis label, a measured impact with denominator, and
pointers that resolve to recorded rows; the report's severity claims all name the same method; the
excluded-content pass at step 8 is stated as done, per section.

Negative: per class — the class, the surfaces covered, the method used, the window, and the result.
"Class X tested across surfaces Y with method Z over window T; clean" is evidence. Where the class
could not be tested, name the blocker and mark NOT ASSESSED. Never write "no issues found in" a
category you have no rows for.

## Report shape

Finding title pattern: `<asset/entry point>: <attacker outcome> (<measured scope>)` — e.g.
"Report export endpoint: cross-tenant record disclosure (3 of 20 payload variants)".
Per-finding fields the reporter needs: vulnclass, severity plus method, proof basis, affected asset
IDs, precondition, mechanism, proof pointers (traffic IDs/roles, command and output), measured impact
with denominator, blocked-what, remediation reference, detection reference, retest status.
