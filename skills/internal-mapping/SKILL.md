---
name: internal-mapping
description: "Invoke this skill when a foothold is in hand and the task is to inventory an authorised internal range: which hosts exist, on what segment, running what, and how to write that into the shared asset graph so the next seat does not re-scan it."
---

# Post-Foothold Internal Mapping

Turn "there is a network behind this host" into rows other seats can query. Done means: every address in the tasked range is either registered or explicitly unaccounted for, segment-level conclusions are facts in the shared store, and no second seat sweeps the same /24 to find out what you already knew.

## Boundaries

- The range is the one the engagement names, entered from a foothold you already hold under that authorisation. Check an address against `list_assets` and the task's scope lines BEFORE the first packet; an address that is neither registered nor named in scope is a question you record and hand up, not a host you probe.
- A discovered host is inventory, not a target. Exploiting it is `network-pentest` / `infra-exploitation` work; one host's own state is `endpoint-discovery`.
- Not this skill: routes to the objective (`attack-path-mapping`), reachability plumbing (`pivot-tunneling`), getting the foothold (`initial-access`), internet-scale breadth (`bbot-recon`).
- Discovery-only discipline: no authenticated enumeration logins, no writes to a host you are mapping, no UDP or `--script` sweeps across a segment. If a probe would read to the SOC as an incident rather than a scan, shrink it.

## Procedure

1. Read the vantage before sweeping anything. Passive, no new packets:
   - Linux foothold: `ip -4 addr; ip route; ip neigh; ss -tn state established; cat /etc/hosts /etc/resolv.conf`
   - Windows foothold: `ipconfig /all; route print; arp -a; nslookup -type=ns <domain>; net config workstation`
   - The routing table and ARP cache give the segment list; existing connections give the neighbours that already talk to this box, which is a better candidate list than a blind /24.
2. Write the vantage as one fact, because it gates everything after it:
   - `record_fact` {summary: "foothold <host> routes to 10.0.5.0/24 and 10.0.9.0/24; gw 10.0.5.1; search domain corp.example", detail: the interface/route/ARP lines, evidence: the command, confidence: observed, asset_ids: [foothold asset id]}.
   - Gate: no CIDR gets probed until it appears in an authorised scope line.
3. Ask the graph what already exists, so you insert only news and the map stays deduplicated:
   - `list_assets type=ip dsl="ip=10.0.5"` — fuzzy prefix over the text `ip` column returns the whole /24's existing rows.
   - `list_assets type=subdomain dsl="domain=corp.example"` — the internal name tree as others registered it.
   - `list_untested_assets type=ip` — scoped addresses nobody has anchored. Note the counts: they are your close-out denominator.
4. Sweep smallest-first, per segment, bounded to the tasked CIDR:
   - Liveness: the ARP cache where ARP works; otherwise `nmap -sn -Pn 10.0.5.0/24`.
   - Services on live hosts only: `nmap -Pn -sS -p 22,53,80,443,445,389,636,3269,3389,1433,3306,8080,9380 --open -T4 --max-retries 1 <ip>`.
   - A named port list beats 65535: faster through the shell, quieter to NDR, and it is the set you actually route on.
5. Enumerate the roles that decide later routes — these are the rows other seats query:
   - identity: DCs (53/88/389/636/3269), DNS forwarders, CA (80/443 certsrv)
   - shared state: file/NAS (445/2049), SQL engines (1433/3306/1521), directory-backed apps
   - management planes: hypervisor/container (443/902/10250), iDRAC/iLO on a separate management address, jump hosts, VPN concentrator, CI/CD (8080/8443)
   - the estate and the sensors: backup servers (`backup-recovery-targets`), monitoring/SIEM/NDR consoles, print servers (9100/631)
   - Sensor and management hosts are mapping facts in their own right: they tell the bench who can see it moving.
6. Register in one batch per segment with `insert_assets`, mixing types in one call:
   - `{"type":"ip","ip":"10.0.5.12","open_ports":[{"port":445,"service":"smb"},{"port":3389,"service":"rdp"}],"bound_domains":["file01.corp.example"]}`
   - `{"type":"subdomain","domain":"dc01.corp.example","record_type":"A","record_value":["10.0.5.10"]}` — name and address are two rows by design: the name joins the in-scope domain tree, the address carries ports.
   - `{"type":"service","service_name":"ldap","port":389,"ip":"10.0.5.10"}` where a service deserves its own node; `{"type":"service","url":"http://10.0.5.31:8080/","page_title":"Jenkins","technologies":["jetty"]}` for HTTP, so later seats can filter on `status_code` / `technology`.
   - The call returns `{results:[{index,id,type}]}` — keep those ids: they are the anchors every later fact points at.
7. Dedup against the external map instead of creating a second truth:
   - An internal host whose name matches a registered subdomain is that same row with an internal `record_value` — not `dc01-internal.corp.example`.
   - `insert_assets` upserts by key and append-merges `technologies`/`auth`, so re-registering a known name with the internal address is the correct merge.
   - Never invent a name to dodge a collision, and never register a host you only saw mentioned in a config file as live — that belongs in a fact with `confidence: inferred`.
8. A segment is not an asset type; the fact store owns it. One fact per segment: {summary: "10.0.5.0/24 = server segment: gw 10.0.5.1, VLAN 50 (CDP, inferred), 2 DCs, 1 file, 1 backup", detail: CIDR/gateway/VLAN + role list + member asset ids, evidence: the commands, confidence: observed|inferred}. `insert_assets` auto-computes `c_segment` (/24) on every IP row — group returned rows by `c_segment` for per-segment counts without another sweep.
9. Close coverage honestly: re-run `list_untested_assets`, diff scope addresses against registered rows, write the gap fact — "41 of 62 scoped addresses anchored in 10.0.5.0/24; 21 silent (no ARP reply, no TCP on the tested set); 10.0.9.0/24: 0 of 18 attempted, no route from the current foothold". Name which part of the network you did NOT map and why; a map with silent holes gets trusted.
10. Hand off directions, do not chase them: a host that looks like it reaches the objective becomes `add_hint` (or a fact marked `confidence: inferred`) for `attack-path-mapping` / your lead. Your deliverable is the map, not the movement.

## Evidence

- A positive host/segment claim carries: the command, the output line proving liveness or the service banner, the UTC window, and the asset row id it produced.
- `confidence: observed` only for what appeared in output. "Probably the DC" is `inferred` until 88/389 answer or a directory read confirms.
- A negative carries method + coverage + what it rules out: "nmap -Pn -p <set> over 10.0.9.0/24, 14:02-14:09 UTC, 0 responses; the foothold's route table shows a path, so this is filtered, not empty → segment NOT ASSESSED". "Nothing there" and "nothing seen" are different sentences; write the second unless evidence supports the first.
- No "no further hosts found" without a denominator (scoped count vs answered count), and no segment-level claim built on one host.

## Report shape

Mapping output is facts, not findings; a seat without `report_finding` hands candidates up via `add_hint`. Where the inventory itself is the gap: "<class>: <service> at <ip:port> reachable from <segment> as <identity>", `vulnclass` = exposure / network-architecture, `asset_ids` = the host rows, `evidence` = probe command + response lines + the coverage denominator.
