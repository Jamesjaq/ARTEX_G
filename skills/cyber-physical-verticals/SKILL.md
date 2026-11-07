---
name: cyber-physical-verticals
description: "Invoke this skill when the in-scope asset is a regulated or constrained-device vertical — medical/IoMT, automotive, satellite/ground segment, telecom signalling or similar — where findings carry safety and legal weight: vertical-specific rules of engagement, patient/vehicle/uplink stop rules, and routing every consequence to the lead and the operator before anything else."
---

# Cyber-Physical Verticals (regulated & constrained devices)

The method does not change in these verticals; the stakes and the hard limits do. A clinical, automotive, aerospace or telecom asset can injure people, ground a vehicle, or lose a spacecraft, and it is bound by regulation the server room is not. Done means exposure and configuration assessed from outside, in-service devices untouched, every finding named against the governing standard and routed through the humans who own the risk before it goes anywhere else.

## Boundaries

- Authorized use only: a device/system in written scope with the operating organisation's and, where required, the regulator's/owner's consent. Ownership or service-status not established -> stop and record the question.
- Never touch anything in service: a device attached to a live patient, a vehicle in operation, an uplink in use, or a live network's signalling plane. Where the only way to confirm a consequence is to interact with an in-service device, produce the plan and risk statement for the operator instead.
- The stop rules below and `industrial-safety-bounds` override any finding, any client request, and any authorization an agent holds.
- Bench/hardware contact is `hardware-interface-lab`; firmware images are `firmware-iot`; process/control networks are `ot-ics-assessment`; RF capture is `radio-sdr`. This skill adds the vertical's rules-of-engagement, stop lines, and standards framing.

## Procedure

1. SAFETY STOP CONDITIONS — evaluate first, before any other step; re-check before any action that could reach an in-service asset. STOP and hand up (operator + verticals_lead) if the action would:
   - interact with, reboot, reconfigure, or denial-of-service a device attached to a PATIENT or in a clinical setting (no in-service IoMT contact, no infusion/pacing/monitor interaction);
   - act on a VEHICLE in service or motion, its brakes/steering/ADAS/ECU, or anything while the drivetrain is live — bench/OBDII work only on an immobilised, owner-consented vehicle;
   - disturb an UPLINK, downlink, TT&C channel, or an in-use satellite/ground-segment path (no commanding, no RF that could desense or interfere);
   - touch a LIVE network's SIGNALLING PLANE (SS7/SIP/Diameter) in a way that could disrupt service to real subscribers;
   - transmit/jam/deauthenticate on radio in a way that denies service — receive-only by default, active RF only with the operator's approval naming frequency and window;
   - cross a jurisdictional/regulatory line (band, data-residency, medical-device privacy) the engagement does not cover.
   Any "yes" -> do not send; record the blocked direction and route it (step 6).
2. Fix the vertical's Rules of Engagement in writing before testing. Answer these as facts:
   - what is in service vs a spare/lab unit?
   - is patient / vehicle / airworthiness data involved, and under what privacy regime?
   - who grants the maintenance / ground-test window, and is any active step permitted — by whose signature?
   - which regulator/standard governs?
   If the RoE cannot be answered, the vertical is NOT ASSESSED — do not infer permission.
3. Classify the asset and its constraint honestly into three piles:
   - assessable-from-here: network/API/management-interface posture, firmware supplied for offline analysis;
   - lab-only: physical contact, bench, in-vehicle bus, RF;
   - out-of-reach for this platform.
   Name which pile each objective is in and route lab-only work to the operator rather than pretending.
4. Assess exposure and configuration from the outside, in-service devices untouched:
   - which management/diagnostic/telemetry interface is reachable, and from where;
   - which update channel (OTA / T&C / firmware) is trusted and how it verifies;
   - which protocol speaks unauthenticated, and which hard defaults or exposed APIs exist;
   - advisory-to-banner matches are inventory; a finding only when the path is demonstrated to exist — and the demonstration stops at the boundary, never inside the device's safety function.
5. Name the standard each finding is written against so the client's compliance owner can act:
   - medical: IEC 62368 + FDA premarket cybersecurity guidance + HIPAA for patient data;
   - automotive: ISO/SAE 21434 + UNECE R155/R156 + WP.29;
   - space: NIST SP 800-183 + the operator's ground-segment policy;
   - telecom: applicable 3GPP/GSMA and carrier rules; IEC 62443 for the industrial/constrained overlap.
   Cite the obligation, not a generic CVE.
6. Note the vertical's own disclosure/regulatory obligation before writing up: some findings trigger a duty to notify the manufacturer, a CERT/regulator, or to respect an embargo (coordinated disclosure for vehicle or medical-device flaws; incident-reporting timelines for clinical/telecom systems). Record the obligation as a fact so the report owner can meet it — do not self-publish a vulnerability that a disclosure regime covers.
7. ROUTING RULE (do not skip): every consequence-bearing finding goes to the verticals_lead AND the operator FIRST — via `ask_operator`/`report_finding` with the lead notified — before it is written up broadly or shared with the client contact. These verticals carry safety and legal weight the report owner must see before distribution; a patient-safety or vehicle-safety impact is never self-published from a seat.
8. Do not go to step 4 until 1–3 have produced a written RoE and an in-service vs lab classification; exposure testing without that is how an assessment becomes an incident.

## Evidence

- Positive: the asset class + vertical, the RoE/standard it is assessed against, the exposure path as observed (source, window, in-service vs lab unit), the specific regulatory obligation met/violated, and a consequence statement routed through `industrial-safety-bounds` — modelled, not produced. State proof basis and any deliberate non-contact.
- Negative: "In-service IoMT monitor NOT assessed for remote configuration — RoE grants lab-unit only and no patient-adjacent contact permitted; management-interface exposure checked on the lab spare over window Z (reachable/unreachable recorded), live-device behaviour NOT ASSESSED by design." Name what you refused to touch and the stop rule behind it — that is a correct, reportable outcome.
- No PHI/PII, no live keys, no proprietary configs in prose; secrets by reference via `save_task_credential`; sample counts with denominators, never bulk records.

## Report shape

Title: `<vertical: asset> — <exposure/config/update-path finding> vs <standard>` [severity gated by physical consequence]. Fields: vertical + RoE reference + in-service/lab status, governing standard + obligation, exposure path (with window), consequence statement (routed through `industrial-safety-bounds`, modelled not executed), anything deliberately NOT tested and the stop rule behind it, remediation, and an explicit "routed to lead + operator before distribution" marker.
