---
name: industrial-safety-bounds
description: "Invoke this skill before any action near a process setpoint, interlock, safety-instrumented function, trip logic or actuator — the governing stop rules for physical consequence: hard stops, why SIFs are never test targets, the operator-approval gate before any state change, and why 'it worked' is not a finding here."
---

# Industrial Safety Bounds (the stop layer for cyber-physical action)

This is the safety authority that overrides every other cyber-physical skill. Its purpose is not to find a flaw — it is to guarantee the assessment never becomes the cause of a physical event. When this skill and a finding disagree, this skill wins, every time, without escalation to anyone but the human who owns the process.

## Boundaries

- Authorized use only, and authorization does not lift a physical stop: even inside written scope, no setpoint change, no interlock interaction, no actuator command, no SIF test at any level an agent can hold. Process safety outranks engagement scope; a client signature cannot make a trip test safe.
- READ and REASON only. You may model consequence from documentation, architecture and advisory data. You may not produce the consequence in order to observe it.
- Applies on top of `ot-ics-assessment`, `cyber-physical-verticals`, and any embedded/hardware work — those skills hand their state-changing questions here before acting.
- If following a procedure would cross a stop below, the correct output is a blocked direction handed up, never a workaround.

## Procedure

1. HARD STOP CONDITIONS — evaluate first, before any other step. These are absolute; "it is authorized / in scope / the client asked / it is reversible" do NOT unlock them. STOP and hand to the operator (`ask_operator`) or the OT/ICS lead if the intended action would:
   - WRITE to a controller: change, download, or upload a program; set/reset a register, coil, or tag; force an I/O point.
   - MOVE a SETPOINT, alarm threshold, trip point, or control constant on a live process.
   - INTERACT with an interlock, permissive, or safety-instrumented function (SIF/SIS), or defeat/bypass a protective layer — at any authorization level.
   - COMMAND an actuator, valve, breaker, drive, or mover that affects the physical process.
   - TRIP, emergency-stop, de-energize, or restart a running process or safety function.
   - Take an action whose worst case you cannot bound, or where the fail-safe depends on network integrity you are testing.
   Any "yes" -> do not send. Record the blocked direction and why. There is no retry-with-different-flags path out of a safety stop.
2. SAFETY-INSTRUMENTED FUNCTIONS ARE NOT TEST TARGETS. A SIF/SIS — the layer that trips the process to a safe state — protects people. You do not:
   - actuate it, demand it, inhibit it, or measure its response;
   - not for a read, not for a "safe"/point-to-point test, not with sign-off;
   - reason about it from documentation only.
3. OPERATOR-APPROVAL GATE before any state change — even one that clears every stop in step 1 (a read some device might treat as a write, or a window-scoped action the OT lead already cleared). Route through `ask_operator` with:
   - the exact single command/request, the specific device and process element;
   - the expected state effect, the worst case, and the recovery step;
   - then wait for an explicit human reply. A timeout is a DENY. Never batch, never self-approve, never assume silence is consent.
4. Classify the action you were about to take — read-only/no-effect, read-with-device-dependent-risk, or state-changing. If it is state-changing it is stopped at step 1 or gated at step 3; it never proceeds on your own judgement.
5. Reason about consequence WITHOUT producing it. From docs/architecture/advisories, write the effect in engineering terms the client's safety lead can check:
   - which setpoint, which protective layer, which fail-safe;
   - which physical outcome if the exposed path were used by an attacker;
   - model it; do not run it.
6. Where evidence for a consequence claim can only come from a physical test, produce the TEST PLAN and the risk statement for a human engineer to execute under their own permit-to-work, and hand it up. Do not execute it yourself.
7. Where you are guessing about the physical effect, say so explicitly and hand it to a human engineer. An unbounded worst case is itself a stop condition.
8. Worked shape of a correct refusal: an upstream seat wants to "prove impact" by nudging a pressure trip point one step to watch the SIS respond — that is steps 1 (move a setpoint) + 2 (actuate a SIF). The correct output is NOT the nudge; it is a stop record plus a documentation-based consequence ("exposed L2 write path could command the reactor coolant pump; the low-level trip is the last protective layer; if that layer assumes network integrity it is defeated by this path"), handed to the client's safety lead with the physical verification deferred to their permit-to-work.

## Evidence

- The primary evidence here is a STOP record: "intended action <X> on <device> NOT PERFORMED — reason <which stop condition / no window / SIF>; handed to <operator/lead> on <date>". That record is a successful outcome of this skill, not a failure to report.
- A consequence analysis is a finding only when it names the setpoint, the protective layer, and the physical outcome, sourced from documentation/advisory data, and marked modelled-not-executed. "Risk was assessed" with no engineering element named is filler.
- Honest negative: "No state-changing action taken against any controller or safety function during the engagement; assessment confined to passive/read-only per `ot-ics-assessment`; SIF response NOT tested by design -> physical-consequence exposure assessed by documentation review, live verification deferred to client engineering."

## Why "it worked" is NOT a finding here

In every other skill, demonstrating that an action succeeded proves impact. Here it can mean you moved a real valve, tripped a real line, or defeated a real protective function — i.e. caused or risked physical harm to plant and people. A working write on a live process cell is an incident, not evidence. Impact is established by *reasoning* from the exposed path to the physical outcome, never by producing the outcome. If a step's success would be reported with pride and could injure someone, the correct move was to stop at step 1.

## Report shape

Title: `<physical-consequence exposure> at <process element / protective layer> (modelled, not executed)`. Fields: the exposed path (from `ot-ics-assessment`), the setpoint/interlock/SIF involved, the protective layer and its fail-safe assumption, the worst-case physical outcome, an explicit "no state-changing action taken; SIF not tested", the deferred physical verification assigned to client engineering, and any stop record generated during the assessment.
