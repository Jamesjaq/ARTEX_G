---
name: remediation-advisory
description: "Invoke this skill when confirmed findings need the fix the client can actually ship: harden, isolate, restore and detect recommendations tied to the proven root cause, ranked by a stated priority method, each with what it costs and what it cannot claim before retest."
---

# Remediation Advisory (close the demonstrated gap)

A finding without a fix the client can deploy is a cost, not a service. Done means: one or two
recommendations per finding, each naming the system, configuration path or code location that carries
the root cause, the effort and the dependency, the compensating control while the real fix waits, a
priority order with a stated method, and a retest bar written before the fix lands. "Segment the
network" is not a recommendation; a specific trust edge to remove, with the dependency that must move
first, is.

## Boundaries

- Authorized use: advice is written only for findings on registered in-scope assets and only from
  demonstrated state. Nothing here authorises the agent to change a client system, close a port,
  disable an account, rotate a key or reboot anything — the client's own change process performs every
  fix, and the platform's approval gates stay as they are.
- Do not recommend disabling, muting or removing a protection the client depends on to make a finding
  go away, and do not propose a fix that assumes staff, budget, licence spend or a maintenance window
  nobody has said exists.
- No fix claim survives without retest: "resolved" appears only where the re-run of the demonstrated
  action was stopped. Everything else is "mitigation applied, unverified".
- Not this skill: severity and impact wording (`engagement-report`, `impact-analysis`), the analytic
  itself (`detection-engineering`), whether a vendor advisory actually stops the technique
  (`mitigation-status`), the cleanup inventory of our own persistence (`persistence-cleanup`), and any
  change near a setpoint, interlock or safety function (`industrial-safety-bounds`).

## Procedure

Do not go to step 4 until 2 and 3 have named the root cause per finding and the affected asset IDs.

1. **Take the confirmed finding set** (`list_findings`, `node_detail` per row, `get_finding_traffic`
   for the proof chain). Work only from findings with a proof basis and a measured impact; a suspect or
   static-only item gets a verification action as its recommendation, not a fix.
2. **Name the root cause each fix must close.** Read the proof chain back one step from the symptom:
   the missing authorisation check at that handler, the default credential on that appliance, the trust
   edge that lets host A authenticate to B without constraint, the group membership that grants the
   right, the unsign-in baseline that lets an unmanaged device reach the tenant, the missing
   server-side validation the client believed the framework handled. If the recommendation does not
   reference that cause, it will not stop the demonstrated chain and you should rewrite it.
3. **Write the fixes in four categories, one or two per finding**, each carrying: the system and its
   version, the exact configuration path, policy name, code location or resource identifier, the owner
   role inside the client, effort (config change / code change / architecture / procurement), the
   dependency that must move first, and what breaks if they do it.
   - **Harden**: patch or upgrade to the version that closes the mechanism; disable the unused feature,
     method or legacy protocol; enforce the missing validation, rate limit or authorisation check;
     remove default and shared credentials; require MFA and conditional access on the exposed plane;
     turn on the audit that was off.
   - **Isolate**: remove the specific trust edge, restrict the management plane to a jump host or an
     allow-listed range, split the tenant or the account boundary that let the reach exist, move the
     data store off the path, restrict the service identity's scope, or pull the segment off the
     routable network. Say which edge, in which direction, and which legitimate dependency needs
     re-routing first.
   - **Restore**: what the demonstrated access left behind and what it could reach — rebuild the host
     rather than clean it where persistence was installed, remove accounts and keys created during the
     run, revoke sessions and refresh tokens issued through the flaw, rotate every credential the chain
     touched (by credential-store ID, never by value), restore data from a recovery point before the
     intrusion window and verify integrity of what comes after.
   - **Detect (interim)**: the analytic that shortens the same path until the fix ships — pull the text
     from `detection-engineering`, and name it explicitly as interim so nobody reads it as the fix.
4. **Compensating control where the root fix is slow.** Vendor EOL, an architecture decision, a change
   window in another quarter: propose the control that reduces exposure now (virtual patching at the
   WAF, network-level allow-list, credential-rotation cadence, an approval gate on the workflow,
   bringing the plane behind a broker), and state its cost in latency, workflow friction, licence, or
   the new failure mode it introduces — plus the residual risk it does not remove.
5. **Priority order with a stated method.** Rank on: whether the demonstrated path still works right
   now, reachability without credentials or with low-privilege credentials, data class and blast radius
   behind it, exploit complexity, and how cheap the fix is relative to the exposure. Publish the ranking
   rule at the top of the section and apply it; if everything is critical, the reader learns nothing and
   the first two lines will be the ones they fund. Group into now / this sprint / this quarter with a
   dependency chain, not a flat list.
6. **Say whose fix it is when it is not theirs**: a vendor product limitation (with the advisory
   reference and the version that resolves it), a fourth-party or SaaS provider setting, a shared cloud
   responsibility boundary, an EOL component requiring replacement. Naming the owner is the difference
   between an advisory and a complaint.
7. **Write the retest acceptance bar per fix, before the fix exists**: the exact action re-run, the
   exact expected outcome (the request refused with 403, no cross-tenant row returned, the command
   blocked by policy, the credential rejected), the asset and identity context it must be re-run from,
   the window, and the artefact that proves it (traffic IDs, node IDs, the detection event). Where the
   retest is not authorised, or the window has closed, record `retest not performed` with the reason —
   never an assumed pass.
8. **Deliver**: attach the recommendation text and retest bar to the finding via
   `update_finding_report`, and record a fact per fix naming the root cause it closes, its effort, its
   dependency and its owner, so the advisory lead can assemble the report section
   (`engagement-report`). Anything you cannot state concretely because the client's environment is
   unknown goes to `ask_operator` as one question — an unasked question is not a recommendation.

## Evidence

Positive: each recommendation cites the finding ID and the root-cause line it closes, names the system
plus the exact config path / code location / resource, states effort, dependency and owner, and carries
a retest bar with the action, the expected blocking outcome, and the pointer that will prove it. Restore
items list the credential IDs and account identifiers rotated or revoked, by reference.

Negative and honest forms: "no client-side fix available — vendor advisory <ref> pending for version
<v>; compensating control <x> proposed; residual exposure <y> stated"; "recommendation not verifiable
in this engagement: retest window closed on <date>, therefore mitigation status = unverified";
"isolation advice withheld: the dependency that would break is unknown and the client has not supplied
the routing inventory — question raised <date>". Never write that a fix works from a changelog alone,
and never let an untested mitigation appear as closed.

## Report shape

Row pattern: `<finding ID>: <fix verb> <specific system/edge/code location> — priority <now|sprint|quarter>,
owner <role>, effort <e>, dependency <d>, retest <bar pointer or "not performed">`.
Fields: finding ID, root cause, category (harden/isolate/restore/detect), exact change location,
compensating control and its cost, whose fix it is, retest acceptance bar, retest status, residual risk
statement.
