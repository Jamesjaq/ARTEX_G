---
name: impact-analysis
description: "Invoke this skill when a demonstrated reach needs its consequence stated: deriving blast radius, data class, availability and safety exposure from reach, replication and trust relationships, without executing the destructive act."
---

# Impact Analysis

Severity is a conclusion, and this is where it gets earned without spending it. Done means: for each demonstrated reach, a consequence chain expressed in numbers and denominators, every link pointing at registered evidence, the harmful act explicitly marked as not performed, and the claim labelled derived rather than proven where that is the truth.

## Boundaries

- Derive, do not demonstrate: no deletion, encryption, overwrite, job disable, service stop, reboot, firmware flash, network cut, or credential-lockout test. Where an impact claim needs an act that would harm the client's service, patients, drivers or operators, the act does not happen and the claim stays labelled derived.
- Read-only inputs: this seat reasons over findings, facts, assets, and captured traffic that already exist. A missing measurement becomes an intent for another seat, not a probe by you.
- Process-safety consequences (setpoints, interlocks, protective functions, physical actuators) are not computed here — that is `industrial-safety-bounds` with the OT lead, and a derived safety claim never becomes a demonstrated one.
- Not this skill: the counts themselves (`data-objective-evidence`), the fix (`remediation-advisory`), detection mapping (`detection-coverage`), report assembly (`engagement-report`).
- Stop condition: never re-test a destructive primitive to make an impact statement more persuasive.

## Procedure

1. Take the reach as registered, not as remembered: `list_findings`, `node_detail` on the finding and the facts beneath it (parents and yields are in the graph; `expand_digest` for cold chains), `graph_overview` for goals and coverage. Build a tuple set: {identity held, host asset id, action permitted, object reached, count already measured}. Anything with no count goes back to `data-objective-evidence` before you reason from it.
2. Enumerate consequence axes per tuple — axes of consequence, not new probes:
   - confidentiality: which data classes, how many rows/objects/files, whose
   - integrity: which objects the held identity can write, from readable ACLs/grants and writable-path listings (a permission read, not a write)
   - availability: which services that identity could stop, from the inventory — replication targets, DCs, DNS, hypervisor management, telephony, payment paths, the backup plane
   - trust radius: what that identity's group memberships and delegated rights imply elsewhere (`attack-path-mapping` edges)
   - downstream population: users, tenants, customers, sites, or plants served by the affected system
3. Bound the radius with registered denominators instead of adjectives. Pull the numbers that already exist in the graph: host counts by `c_segment` from `list_assets type=ip`, DC and OU counts from directory facts, replica sets from database facts, node counts from cluster management inventories, protection groups covering M hosts from `backup-recovery-targets`, tenant counts from the app seat's facts. Every figure carries the node id that produced it.
4. Express each consequence as a conditional chain with its links labelled: "given <held identity> has <permission> over <object> (fact f41, observed), an actor could <act> against <N of M objects>; consequence = <business effect>. The act was NOT performed." Then name the one control that breaks each link — that is what turns the analysis into something `remediation-advisory` can act on.
5. Reason about replication and trust, since they set the real radius:
   - replication: readable replica/snapshot lists, Always On / binlog / async replica counts, object-storage versioning and cross-region copies, AD DC count and naming-master roles. A single readable DB host may imply the same rows in K copies, and copies inherit the least-controlled access path.
   - trust: which accounts this host's identity can act as, and where that identity is honoured (service accounts on N hosts, SQL linked servers, keytab-held delegates, federated app registrations). A host is a doorway; the radius is the trust edge count times the assets they reach.
   - management planes: hypervisor, backup console, MDM, CI, cloud control plane — one credential here typically re-radiates across the whole estate.
6. Translate to business terms in two layers: the client's own operations (transactions blocked, sites unable to authenticate, build pipeline trust), and the client's customers (records exposed, service degraded). Use their units: per-customer, per-site, per-transaction, per-shift. Keep jurisdiction or regulatory statements to describing the data class; the compliance conclusion is the report owner's.
7. Cap the severity at the measurement. Where the destructive act is unproven, write the cap reason into the finding so the reporter cannot lift it: "availability impact derived, not demonstrated: power-state action never attempted". An unearned critical is the first finding a client disputes.
8. Write the unbounded items as bounded statements: "radius lower-bounded at 14 of 14 hosts inventoried through this plane; upper bound unknown because user-count data was not readable" beats "potentially all". List what one read would settle, as an intent or a hint.
9. Emit per finding: a measurement line from `data-objective-evidence` plus the conditional chain plus the not-performed statement plus the control that would break the chain, via `update_finding_report` or a new fact. Then one cross-finding blast-radius paragraph for the lead: which single credential, if removed, collapses most of the derived impact.
10. Sanity-check the whole set for double counting: if five findings derive their impact from one compromised admin account, say so once, and label the rest as "same root, same consequence".

## Evidence

- A derived impact needs: the reach fact with its pointer, the permission read that supports the implication (observed), the numbers and their denominators, the as-of window, the conditional chain with per-link labels, and the explicit line "<act> was not performed; impact derived from permission state".
- A proven impact is only ever the minimal sample from `collection-exfil` — a count and 2-3 masked rows — never the completed harmful act.
- Negative: "availability impact not quantified: management plane permitted listing 14 of 14 VMs, no power action attempted by design → radius from inventory, disruption unproven (derived)". And where inputs were missing: "blast radius NOT ASSESSED for <system>: no replication topology fact in the graph; the read that would settle it is <one command>".

## Report shape

Title: "<reach statement> on <system>: derived consequence <effect> for <N of M> <objects/population>; blocking control: <X>". Fields the reporter needs: derived or proven, act not performed, numerator + denominator + unit + as-of window, the pointer list, the unbounded items with the read that would bound them, and the client-vs-customer split.
