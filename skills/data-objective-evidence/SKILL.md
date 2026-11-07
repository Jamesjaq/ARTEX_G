---
name: data-objective-evidence
description: "Invoke this skill when a finding or objective claims a consequence and needs it quantified: turning 'we could read customer records' into '1.2M of 3.4M rows reachable from this role, 3 sampled, evidence pointers f41/a2b1'."
---

# Data Objective Evidence

This is the seat that answers "how bad, actually" with a number, its denominator, and pointers a reader can re-open. Done means every objective claim in the task has one measurement line attached — or an explicit statement of why the number is not obtainable — and no finding ships with a severity that outruns its own evidence.

## Boundaries

- Read and reason only: this skill works the graph, the finding store and captured traffic. It does not send new packets. Where a number needs a fresh measurement, write the one query that would produce it and hand it up as an intent/hint; do not go measure it yourself.
- Sensitive values are not evidence. Reference objects by id and describe the class; the number and the pointer carry the argument.
- Not this skill: obtaining the data (`collection-exfil`), crossing the perimeter (`exfil-egress`), deriving what the number means for the business (`impact-analysis`), assembling the deliverable (`engagement-report`).
- Stop condition: never inflate an adjective to cover a missing number, and never move a claim from derived to proven for comfort.

## Procedure

1. Restate the objective as a measurement. Take the goal text from your tasking, or `graph_overview` where your tool set exposes it, and rewrite it in the form "<N> of <M> <unit> <verb>able by <identity> on <asset>" — units being rows, records, objects, accounts, tenants, hosts, commands executed, or minutes of availability. If the objective cannot be phrased that way, that ambiguity is the first thing to hand back to the lead.
2. Inventory the existing evidence rather than re-running work: `list_findings`, `list_facts q=<store|keyword>`, `node_detail` on each candidate (facts carry the counts in `detail`; findings carry `evidence` and bound traffic ids), plus `get_worker_trace` / `search_all_worker_traces` when a count exists only in a worker's output. Build a candidate list of {node id, asset id, identity used, action, result, count present|missing}.
3. For each claim, extract or mark missing exactly four things: numerator (what succeeded), population/denominator (what was attempted or what exists), method (the query/request that produced it), and window (as-of timestamp). Anything absent is written as absent — "count not recorded in f41; the single measurement that yields it is `SELECT count(*) FROM customers WHERE tenant_id<>1`".
4. Source denominators from the platform rather than inventing them:
   - asset/scope coverage: `list_untested_assets` and `list_assets` counts by type (hosts attempted vs hosts in scope)
   - request-level counts: `traffic_search` by asset/time and read the returned totals (attempts vs 200s vs 403s), citing traffic ids with their roles
   - data volumes: the aggregate queries another seat already recorded (counts, `reltuples`, `dbStats`, bucket listings)
   - estate coverage: protection groups and recovery points from `backup-recovery-targets` facts
   - Where a store is not in your tool set, cite the fact that already carries the number and name the missing read as a follow-up instead of guessing it.
5. Normalise units and vocabulary before comparing anything: rows vs tables vs databases vs hosts vs accounts vs tenants vs customers. One sentence per claim naming the unit, the data class, the identity, and the as-of window. Mixed units are how "3.4M records" and "12 tables" end up in the same paragraph as the same size.
6. Assemble the pointer chain for each number: the fact ids, finding ids, asset ids and traffic ids (with roles) that a reader opens to re-verify. Target shape: "1.2M of 3.4M customer rows readable as app_user (count from fact f41; 3 sampled rows masked, fact f43; queries bound as traffic refs a2b1 proof, c7d2 baseline; assets 118, 126)".
7. Re-grade severity against the measurement, not the class name: an access-control flaw reaching 3 of 3.4M rows in one seeded tenant is not the same sentence as one reaching 1.2M production rows; a command executing on 1 of 14 hosts is not "code execution on the estate". Where severity outruns the number, say which number caps it and why.
8. Split the two impacted populations explicitly: the client's own systems vs the client's customers. The second usually carries the regulatory and reputational weight; name the data class that makes it so (PII, health, payment, authentication material) and leave the legal conclusion to the report's owner.
9. Record the objective statement as one roll-up fact per goal — numerator, denominator, unit, identity, window, pointers, and the unproven remainder — then link it to the goal node (`prove_goal` where available) only if the assembled evidence actually satisfies the objective as written.
10. Review pass over `list_findings`: every finding either carries a measurement line or carries an explicit "not measurable because <reason>, attempted <X, Y>". A claim with neither is returned to its author, not reworded by you.

## Evidence

- A quantitative claim is complete when it has: numerator, population, method (the exact capped query/request), window, the identity/privilege it was done as, the proof basis (executed-live / executed-lab / derived-from-graph), and at least one re-openable pointer per number.
- A negative or partial is written with the same parts: "CRM objective not proven — 2 of 5 candidate read paths reached data, and both samples came from the seeded test tenant; production rows 0 of 3 attempts, denials quoted in facts f41/f44/f47; payment-card column encrypted at rest, key not reached → NOT ASSESSED".
- Never present an estimate as a count: `reltuples` and size-based guesses get labelled estimates with their source.
- Zero is a result: "0 of 20 identifier variants returned another tenant's record (all 403)" is evidence of a working control and belongs in the deliverable, not the floor.

## Report shape

One line per objective: "<N> of <M> <unit> <reachable|readable|modifiable> as <role>; <K> sampled; blocked: <what>; pointers: <fact/finding/asset/traffic ids>". In the finding itself, the measured impact replaces any adjective in the summary, and `evidence` ends with the pointer list so the reporter can rebuild the number without re-running the query.
