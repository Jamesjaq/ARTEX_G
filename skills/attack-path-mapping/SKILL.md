---
name: attack-path-mapping
description: "Invoke this skill when internal inventory exists and the question is how to get from the position you hold to the objective the task names: enumerating hops, ranking paths by reachability, and expressing each path as evidence pointers in the graph."
---

# Attack Path Mapping

Inventory is raw material; this is the reasoning layer. Done means a ranked set of concrete routes from where you are to the named objective, every hop labelled observed or inferred and backed by node ids a lead can re-read — plus the paths you considered and closed.

## Boundaries

- The objective is not yours to choose: read it from `graph_overview` goals or your tasking and map to that. A path to something more interesting is a hint for your lead, not a change of plan.
- Only authorised starting positions: every start node is a foothold or identity this engagement already holds, registered in the asset graph or referenced by id in the credential store.
- You map routes; you do not walk them. Executing hops is `lateral-movement` / `post-exploitation`; relay and delegation mechanics `privilege-relay`; certificate templates `ad-pki-certificate`; device and conditional-access trust `device-trust-mdm`; consequence numbers `impact-analysis`; the data itself `collection-exfil`.
- Stop condition: if a path only closes by harming service (removing a control, blinding telemetry, forcing a failover, coercing a person), write it as a derived hypothesis with the missing link named, and stop there.

## Procedure

1. Fix both endpoints before reasoning:
   - Start = {foothold asset id, the identity you hold (cite credential rows by id via `list_task_credentials`, never by value), what that identity is *proven* to do}.
   - Goal = the goal node's text, decomposed into the object that actually satisfies it (a named store, host, account class, or data set) and its asset ids. A goal phrased as "domain admin" still needs the object: which domain, which DC, which asset row.
2. Work the substrate you already have rather than launching a fresh collection storm:
   - `graph_overview` for goals, frontier and existing findings; `list_facts q=<keyword>`; `node_detail` on each candidate to get its detail and asset links.
   - `list_assets` for surface: `type=service dsl="service_name=smb"`, or pull `type=ip` rows and read their `open_ports` arrays.
   - Where directory collection exists (BloodHound-style graph files, LDAP dumps, `nxc -M` output from `ad_enumerator`), consume those artefacts by node id; re-ingesting the domain is a different seat's task and needs tasking, not your turn budget.
3. Build edges from a fixed vocabulary so paths stay comparable and queryable:
   - identity: principal → group → right, and the ACE/role that grants it
   - access: host → host over a named service (admin SMB, WinRM, RDP, SSH, DCOM)
   - route: host → segment only reachable via a pivot (`pivot-tunneling`)
   - trust: domain → domain, forest, external/federated trust, cloud tenant link
   - delegation: unconstrained / constrained / RBCD on a service account
   - data: share, path, or store → the objective object, plus the service owning it
   - egress: host → the objective seen from outside (backup plane, SaaS tenant, CI system)
4. Record each candidate path as one fact, hops in order, with pointers — `record_fact` with the batch `facts` array:
   - summary: "P2: foothold A (asset 118, user U) -> file01 (asset 126, admin SMB, observed) -> DC01 (asset 120, WinRM as U, inferred from group membership in fact 44)"
   - detail: per-hop from/to asset ids, supporting node ids, the privilege each hop requires, what must additionally be true for the last hop, and the one measurement that would settle it
   - evidence: the single command line behind the strongest hop; confidence: observed|inferred; asset_ids: the chain
   - One fact per path, not per hop: a graph full of hop fragments hides the route nobody can read.
5. Rank on reachability, not drama. Sort keys in this order:
   - proven hops vs unproven hops (count both, in the summary)
   - prerequisites held vs needed — list each missing credential, right, or host, and mark it a decision or a dead end
   - controls observed in the way (tiered admin model, LAPS-managed local admin, just-in-time elevation, EDR policy on the transit host)
   - noise cost of the route, then distance to the goal object
   - Ties break toward fewer unproven edges. Name the axis that decided each ranking: an unexplained order is a guess.
6. Cheapest-edge proof only, and only inside the authorisation: take the weakest edge on the top-ranked path and convert inferred → observed with one request (a group-membership read, one authenticated session list, `whoami /groups` / `id -a`). Do not walk the chain end to end — that is movement, and a path you proved by moving is a path the client's telemetry now owns. Gate: steps 4 and 5 must already be in the graph before step 6, so a partial proof still leaves a usable map.
7. Close what does not work; dead ends are half the deliverable. One fact per rejected path quoting the rejecting evidence (the ACCESS_DENIED line, the LAPS policy, the absent ACE, the segment with no route) in the form "closed by <control>, as <identity>, <window>". Leads plan from the closed list as much as the open one.
8. Label what you cannot test: "the directory data shows this edge; testing requires an identity we do not hold" → `confidence: inferred`, then `add_intent` / `add_hint` the single measurement that would settle it ("read ACEs on the OU named in fact 39").
9. Where a path genuinely closes on the objective, link the evidence node to the goal (`prove_goal` where your tool set exposes it) instead of describing the route in chat prose.
10. Re-rank after each new fact from another seat: a path that ranked third moves first when `internal-mapping` registers the management VLAN. Path analysis is a state you refresh, not a one-shot answer.
11. Keep the counterfactual short list with the paths: for each top route, the one defence that makes it dead in practice (tiering, just-in-time, LAPS, PAW-only console access, EDR on the transit host). Leads price routes by that, and it keeps a theoretical path from being briefed as an imminent one.

## Evidence

- A positive path claim: every hop has an asset id + supporting node id + observed|inferred label, and the terminal statement is measured — "reaches 1 of 3 candidate DC shares; 2 denied for this identity", never "gives domain admin".
- A ranking claim names its deciding axis and the evidence behind the score.
- A negative states the test: "edge A→B as user U over SMB, one authenticated session at 13:41 UTC: STATUS_ACCESS_DENIED; no alternate identity held → edge NOT ESTABLISHED (closed)". Where the substrate was missing: "trust edges NOT ASSESSED — no directory collection in the graph; one LDAP trust query would settle it". Absence of data is never "no path exists".
- Mark the proof basis per path: executed-live for edges you actually tested, derived-from-graph for edges that exist only in collected directory/asset data.

## Report shape

Paths are facts; report only when the route itself is the gap. Title: "<start>: <identity> reaches <goal object> in <N> hops, <M> of <K> proven". `vulnclass` = attack-path, `severity` capped by unproven hops and by the named blocking control, `asset_ids` = the chain, `evidence` = the hop table with node pointers and the proof basis.
