---
name: evidence-custody
description: "Invoke this skill when engagement artefacts need to be identified, classified, redacted and accounted for: chain of custody over evidence blobs, traffic captures, traces and exports, the PII scrub before anything is a deliverable, and the destruction list and attestation at close."
---

# Evidence Custody (chain of custody, retention class, PII scrub)

Custody is what makes an artefact usable: it can be identified, located, hashed, traced to who
produced it and who touched it since, classified against a retention obligation, and redacted before
a client or a third party reads it. Done means a manifest whose rows verify, a redaction pass that
states what it removed, a handover log, and a destruction list with an attestation at close. Custody
is also the seat that refuses a delivery, in writing, when it cannot be made safely.

## Boundaries

- Authorized use: this covers artefacts the engagement itself produced, inside the engagement's own
  storage and the client's agreed channels. Nothing here authorises collecting new material from a
  target, and nothing here is a procedure for moving client data somewhere the agreement did not name.
- **Nothing leaves to an external paste/gist/screenshot-sharing/storage tool.** No WebFetch upload,
  no third-party pastebin, no "quick share" of a capture, report draft, screenshot or trace excerpt.
  Delivery happens by the human operator, over the channel they designate (`secure-operator-comms`).
- Traffic capture is full and plaintext — it is the highest-sensitivity material in the engagement,
  not a neutral log. Treat every capture as containing credentials and personal data until the scrub
  proves otherwise.
- You do not silently edit another seat's evidence. A redaction is recorded as a redaction, on the
  row that shows what was removed and what replaced it.
- Not this skill: credential material storage mechanics (`credential-handling`), report composition
  (`engagement-report`), the sequence of events (`engagement-timeline`), our own burned indicators
  (`opsec-tradecraft`).
- Stop condition: a deliverable that cannot be produced safely is reported as the finding — which
  artefact, which class, what must change. It is not shipped stripped of its proof, and it is not
  shipped anyway.

## Procedure

Do not go to step 4 until 2 and 3 have produced a class per artefact and a verified manifest.

1. **Inventory what exists.** Evidence store blobs (content-addressed bodies for finding traffic),
   finding nodes and their bound traffic (`list_findings`, `get_finding_traffic`), exploration facts
   and assets, worker traces, the engagement work directory (scratch files, exports, downloaded
   binaries, screenshots), staged report exports (`StageFindingsExport` output directories), host
   traffic exports (`ExportHosts` archives), credential-store entries (`list_task_credentials` — IDs
   and metadata only), and any artefact the operator handed in. Record paths and counts; do not rely
   on anyone's memory of what was written where.
2. **Assign a retention class to every row**, from the scheme agreed with the client, and quote the
   obligation that sets it. Workable default:
   - **C1 client-confidential deliverable** — the report and its evidence index. Retention per
     contract; readable by the named client contacts.
   - **C2 raw capture** — full plaintext traffic, memory/disk images, unscrubbed command output.
     Never delivered; retained only if the contract says so, destroyed otherwise.
   - **C3 working/reproducible** — scratch scripts, wordlists, re-downloadable binaries, intermediate
     exports. Destroy at close.
   - **C4 engagement-ops material** — our own infrastructure detail, indicators, implant and payload
     builds. Retained by the ops rule, destroyed with `campaign-infrastructure` burn-down confirmation.
   - **C5 personal data** — anything naming or identifying customers, employees, or victims. Minimal
     set, shortest retention, scrubbed before delivery, and legal-review named where the contract
     requires it.
   Where the client has not stated a retention period or a delivery limit for a class, ask with
   `ask_operator` and record the answer as a dated fact; an unasked question is not an assumption.
3. **Build the manifest and verify the chain.** Per artefact: path or ID, sha256, byte length, the
   finding/node it supports, the source seat or task, the capture time, the class, and the location.
   Verify the bodies before handover — the evidence store already checks length and hash on open, and
   a body that fails verification is a broken chain: report which artefact failed, and stop treating
   it as proof. Re-downloadable items get a note saying so rather than a preservation claim.
4. **Run the redaction pass before delivery, and state it.** For each artefact that will become C1:
   - live secrets, hashes, tokens, cookies, private keys → replaced by the credential-store ID from
     `save_task_credential`; the value itself never appears in prose, filenames, or screenshot pixels.
   - bulk output → class, count, denominator, and at most two or three representative rows with
     personal fields reduced to type-and-ID; anything that looks like a table dump goes back to the
     author for sampling (`data-objective-evidence` owns how the sample is taken).
   - victim and employee PII → record class and count, no names, no contact details, no identifiers
     that survive into the deliverable; filenames and paths that name a person are normalised.
   - screenshots → check the URL bar, token and secret panes, email client headers, incidental
     bystander names in chat or ticket views; re-capture or crop, and record what was cropped.
   - internal hostnames/IPs where the contract calls for anonymisation → placeholder form, with the
     real value kept in the manifest row, not in the report.
   Log the pass as a fact: artefact, what was removed, what replaced it, and who performed the check.
5. **Record every handover as a dated custody event**: artefact IDs, class, format, recipient role and
   organisation, the approval reference that permitted it, the channel, and the operator who performed
   the transfer. Agents prepare handovers; a human performs them.
6. **Access control at rest**: state which location holds which class and who may read it; call out
   any artefact sitting in the work directory at a class that does not match its location (a C2 capture
   in an export folder, a C5 list in a scratch file). Route each to its owner with a one-line fix:
   move, re-strip, or destroy.
7. **Prepare the destruction list at close.** Enumerate by class what must be deleted when: C3 scratch,
   staged exports and copies outside the evidence store, downloaded binaries, temporary captures,
   screenshots not in the deliverable, and C4 material after burn-down. Each row: path, size, count,
   class, and the clause that requires its destruction. **A human performs the deletion** — hand the
   operator the list and the command per line, and never run a bulk destructive delete against a
   client system or the shared store to make a number look tidy.
8. **Write the attestation**, once the operator has confirmed the deletions: what existed, its class
   and location, manifest rows verified by hash, what was destroyed on what date by whom, what is
   retained and under which obligation, and what could not be accounted for. Get the acknowledgement
   through `ask_operator` and record it; an unacknowledged attestation is a draft.
9. **Report the unresolvable**: an artefact whose class makes delivery impossible, a body that fails
   verification, a chain break with no record, a retention conflict between contract and platform.
   Write it as a custody finding: the artefact, the class, the break, and what has to change.

## Evidence

Positive: a manifest row that verifies (path/ID, sha256, length, source, capture time, class), a dated
redaction record naming what was removed and what replaced it, and a handover entry with an approval
reference. Every delivered artefact traces to a class and to at least one dated custody event.

Negative and honest-gap forms, all reportable: "no custody record exists for artefact X — it appears
in <report section> but not in the manifest" (chain break); "C2 capture retained because the contract
gives no destruction instruction, question asked on <date>, unanswered"; "verification of body Y
failed at <date>; the finding it supported is marked unsupported pending re-proof". Never write "all
evidence accounted for" from recall, and never record a class as cleared without the pass that
cleared it.

## Report shape

Deliverable "Custody statement and evidence index": fields per artefact = ID or path, sha256, length,
source seat/task, capture time, retention class, delivery status, redaction summary, handover
date/recipient/approval ref, destruction date or retention clause. Closing block = attestation text,
operator acknowledgement date, and the unaccounted-for list.
