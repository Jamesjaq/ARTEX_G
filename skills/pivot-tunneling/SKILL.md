---
name: pivot-tunneling
description: "Invoke this skill when a segment is reachable only through a foothold you already hold: choosing a channel, proving the route, keeping every destination inside the pivot's authorisation, and closing the route with evidence."
---

# Pivot and Tunneling

Reachability engineering. Done means: a named route recorded in the graph, one in-scope destination proven through it, an audit list of everything touched behind it, and evidence the route was closed. The foothold is a route, not a second engagement.

## Boundaries

- The pivot is not a target: do not scan, exploit, or "tidy up" the foothold beyond what establishing the route requires — deeper host work is `endpoint-discovery` / `post-exploitation` and needs its own tasking.
- Every destination behind the pivot is checked against the in-scope asset list before it is touched. Out-of-scope reachability is recorded as a dead end and left alone; needing a wider range is a scope question for the lead, not a licence.
- Intermediary infrastructure is engagement-owned (see `c2-ops`): no third-party relay, no public tunnel service. One tunnel per need, torn down at wrap-up.
- Stop condition: never disable a host firewall, sensor, or routing rule on the foothold to make a route work. The blocked route is the answer, and it is a finding about the control's shape, not an obstacle to out-technique.
- Not this skill: mapping the range once you are through (`internal-mapping`), deciding which route matters (`attack-path-mapping`), reading data behind the route (`collection-exfil`), leaving the perimeter with it (`exfil-egress`).

## Procedure

1. Pick the channel from the foothold's constraints, cheapest-first:
   - Single-service forward (smallest, easiest to close): `ssh -N -L 127.0.0.1:1433:10.0.9.20:1433 operator@foothold` — proves one edge, forwards one port.
   - Dynamic SOCKS for breadth: `ssh -o ServerAliveInterval=30 -M -S /tmp/pivot-<hostid>.sock -fND 127.0.0.1:1080 operator@foothold`. The ControlMaster socket is what makes it closable.
   - Reverse tunnel when the foothold dials out but cannot be dialled into: `chisel server --reverse --port 8443` on your own listener, then `chisel client https://<listener>:8443 R:socks` from the foothold.
   - Two-hop chain: `ssh -J footholdA,footholdB -N -D 127.0.0.1:1081` — keep the chain in one command so teardown is one kill.
   - One-shot relay when a tunnel is overkill: `socat TCP-LISTEN:8445,fork,reuseaddr OPENSSL:10.0.9.7:8443`.
   - `proxychains4` only for tools that speak no SOCKS, and know its lies: no ICMP, no raw UDP, so host discovery becomes `-Pn` and some timeouts read as "filtered" when they are only unencapsulated.
2. Prove the foothold can carry the channel before building it:
   - What it already does out and in: `ss -tn state established`, `env | grep -i proxy`, `netsh winhttp show proxy`, its own route table.
   - Whether SSH lands at all (auth method available? `list_task_credentials` for the account or key reference — cite by id, never paste), and whether egress to your listener is permitted.
   - First environmental failure = one `ask_operator` with the exact command that needs running; not three substitute tools chasing a bypass.
3. Write the route into the graph before using it:
   - `record_fact` {summary: "route R1: 10.0.9.0/24 reachable through foothold <asset id> via SOCKS 127.0.0.1:1080 (ssh -D, ControlMaster /tmp/pivot-7.sock), opened <UTC window>", detail: channel, local endpoint, remote endpoint, the CIDRs it can reach, and the exact teardown command, evidence: the command plus its first successful reply, confidence: observed, asset_ids: [foothold id, reached host ids]}.
   - A route nobody wrote down is a hole nobody can close.
4. Validate with exactly one cheap request to one in-scope address:
   - `list_assets type=ip dsl="ip==10.0.9.7"` (exact match) first — if the address is absent, stop: scope question, dead end, recorded.
   - Then `curl -m 8 --socks5-hostname 127.0.0.1:1080 http://10.0.9.7/` or `proxychains4 nc -z -w3 10.0.9.7 445`.
   - Gate: no scaling past step 4 until this reply exists. A route that answers nothing is usually the wrong interface or the wrong vantage, and sweeping it produces a false negative the whole bench will then trust.
5. Size probes to the channel: a SOCKS path is slow and lossy. One host at a time, the same port set `internal-mapping` uses, `-Pn -T2 --max-retries 1`, no `--script`, no segment-wide parallelism. Prefer the foothold's own hints (mounts, `known_hosts`, DFS roots, LMHOSTS, DNS suffix, `ip route`) over discovery through the tunnel — those names are free and they point at the hosts worth asking.
6. Multiplex instead of multiplying: one ControlMaster serving many forwards beats ten tunnels, and every extra listener is another thing to justify and to close. Never leave a forward bound to 0.0.0.0 on the operator side when 127.0.0.1 does the job.
7. Attribute every result to the route, not to the network: the same address answering differently direct vs via pivot is a segmentation fact for `internal-mapping` — record both readings with the route name. Before claiming the far host replied, confirm the forward is yours: `ss -tlnp | grep 1080`, because a local service on the same port is a classic false positive.
8. Keep the audit list as you go: destination, port, method, timestamp, result, in-scope yes/no. That list is what makes the pivot defensible in the report and answerable to a SOC call.
9. Close the route and prove closure:
   - `ssh -O exit -S /tmp/pivot-<hostid>.sock operator@foothold`; terminate chisel/socat processes.
   - `ss -tnp | grep -E '1080|1433'` showing nothing listening.
   - `record_fact` the teardown with that output, referencing R1's fact id — open and closed are one pair, or the lead cannot tell whether the hole still exists.
10. At wrap-up, list every route you opened with its status and owner. An orphaned forward left on a foothold is a finding against us, not against the client.

## Evidence

- Positive: the channel command, the local↔remote endpoint pair, the first successful reply line, the open window, the fact id the route was recorded under, the teardown command and its output, and (for segmentation claims) the same destination's result from the direct vantage.
- Negative, honest: "reverse tunnel from foothold to engagement listener :8443 attempted 3 times, 14:10-14:16 UTC: TLS handshake completed then reset before data; direct egress over 53/UDP succeeded → channel BLOCKED at the inspecting proxy; 10.0.9.0/24 NOT ASSESSED by this route". Record the block *shape* (reset vs timeout vs certificate error vs proxy 403) — it is what tells the next seat whether retrying is pointless.
- A timeout through a SOCKS path is not a closed port: state which you saw and how you distinguished them, or mark the port unassessed.
- A teardown claim without post-close `ss` output is not evidence.

## Report shape

Routes are facts. Findings describe the control's shape or the exposure it permits: "<foothold> forwards arbitrary TCP to <segment> because <observed reason: no host egress firewall / segment ACL permits any from the compromised VLAN>: proven on <M> of <N> in-scope destinations". `vulnclass` = network-segmentation / pivot-exposure, `asset_ids` = foothold plus reached hosts, `evidence` = route + audit list + teardown, and the summary states the teardown status.
