---
name: ssrf-metadata
description: "Invoke this skill when the task is a server-side fetch on an authorized target: URI-parsing and redirect edges, internal-service reach, and cloud or workload metadata endpoints (IMDS, Azure/GCP/OpenStack metadata, kubelet, Envoy admin, ECS task metadata) reached through that fetch."
---

# SSRF & Metadata Reach

Making an in-scope application fetch a URL you choose, and reading what that proves: its network position, its URI parser, its redirect policy, and whether a workload identity endpoint sits behind it. Done = the sink proved, the filter characterised with counts, the reach class demonstrated with the smallest response, and the boundary you deliberately did not walk through.

## Boundaries

- The application performing the fetch is the target. Internal hosts it can reach are in scope only if the asset list says so: reaching an internal admin panel is the finding; continuing from it into systems nobody tasked you with is out of bounds — stop and record the boundary as a fact for your lead.
- Metadata and workload-identity endpoints are read for **reach and shape only**, using the smallest response that proves it. Never harvest another tenant's or another workload's credential to make a point; the path existing is the finding.
- Any token, key or credential you do see goes to `save_task_credential` and is cited by ID. Not into a fact string, a finding, a note, or the report draft.
- Read-only. No PUT/POST writes to internal services, no requests that can mutate or restart anything, no scanning of internal ranges. Port probing stays on the handful of ports whose answer you actually need.
- Cloud-privilege walking after you land (IAM paths, bucket listings, role assumption) belongs to cloud-attack; the internal inventory belongs to internal-mapping. Your job is the fetch and what it proves.
- Out-of-band confirmation needs an egress destination the operator named; a host that appears inside target-controlled content is never that destination.

## Procedure

1. Enumerate the URL-taking inputs on the tasked asset: webhooks and callbacks, URL imports and feed readers, avatar/attachment-by-URL, thumbnail and PDF/office renderers, URL preview and unfurl, proxy and rewrite rules, health checks, `?url=`/`?src=`/`?target=`/`?endpoint=`/`?imageUrl=` style parameters, JSON fields named `url`, `callback`, `webhook`, `fetchUri`, `href`, and header-taking inputs (`X-Imgur-Url`, `X-Forwarded-`, `Link`).
2. Separate server-side from client-side before spending anything: a preview rendered by the browser is XSS (client-side), not SSRF. Discriminate with a marker whose answer only the server could produce, plus the recorded request log for the asset — if the app's own outbound traffic never shows the fetch, it is not a fetch sink.
3. Prove the fetch with the cheapest oracle, in this order: response-difference (reflected body, differing status or length), timing against a port you know is closed versus open, then the approved out-of-band destination. One probe per oracle; the goal is "the server fetched", not a survey.
4. Characterise the filter, one axis per request: blocklist contents (loopback, RFC1918, link-local, cloud metadata hostnames, `localhost`, resolvable names versus literals), method/scheme allowlist, whether the client follows redirects and how many hops, whether it re-resolves and re-checks the IP after a redirect or only validates the string it was given. Then walk the parser edge: userinfo (`http://good.com@169.254.169.254/`), fragment/query tricks (`http://169.254.169.254#.good.com`), host-normalisation mismatch, alternate IP encodings (decimal, octal, hex, `0`, `0.0.0.0`), IPv6 forms (`::1`, `[::ffff:127.0.0.1]`, `fd00:ec2::254`), DNS rebinding against a short-TTL name you control, and `*.nip.io`/`sslip.io`-class wildcard resolvers. The full variant set to run once, in order, is in [reference.md](reference.md).
5. Redirect chain: make an approved, in-scope host that the app trusts issue a 302 to the internal destination. If it follows without re-checking, the finding is the redirect edge and the bypass is quiet — no brute force needed. Record 302-versus-follow evidence rather than a guess.
6. Internal reach: fetch the specific internal endpoint the objective needs, read-only, and take the smallest artifact that proves position (a distinctive byte, a status/`Server` pair, a title string). One response per service. Register what you proved as an ownership question in a fact — do not silently register internal hosts as assets, since the asset list is what authorises the next seat's work.
7. Metadata classes: identify the platform from what you already know about the asset, then request its endpoint with the preconditions that endpoint requires (the exact paths and required headers/verbs are in [reference.md](reference.md)). Note which requirement blocked you, because that is the control worth reporting: IMDSv2 token requirement, `Metadata: true` on Azure, `Metadata-Flavor: Google` on GCP, and the kubelet's TLS-client authentication are all real controls. If a token is required and issued, it is another tenant's or workload's material: store by reference, report the class and expiry, never the value.
8. Proofs of consequence, kept minimal and scoped to the credential you already hold by reference: which identity does that credential resolve to, and which control stopped it from going further. Nothing beyond the single demonstration.
9. Counts, always: run the variant table once, record per variant blocked/passed and which layer blocked it (application validation, egress proxy, the client library's own scheme rules). A filter that stopped 19 of 20 is a hardening note with the surviving variant named; one that stopped 0 of 20 is a finding about the parser.
10. Hand up the moment you are beyond the fetch: reach into the cloud account belongs to cloud-attack; a new internal service worth tasking goes to your lead as a fact.

## Evidence

Positive: the exact parameter and the exact URI form that produced the fetch; the class of what was reached; the artifact only that origin could have produced (a role name, an instance-id pattern, a document field, the byte count of the metadata body); variants passed versus blocked with the blocking layer named; every touched host, path and credential-ID. A response-length difference alone is not proof of reach — the proving artifact must be content.

Negative: "`<endpoint>` param `<name>`: 14 variants (loopback, link-local, decimal, IPv6, redirect, rebinding) over `<window>`, all rejected with the identical `invalid host` validation error and no hit on the approved callback — SSRF closed for this input". And when your own environment blocked the test: "`<class>` NOT ASSESSED: the rebinding resolver was unreachable from this container, so DNS-rebinding could not be attempted; the other 12 variants were blocked at application validation".

## Report shape

"Server-side request forgery via `<param>` on `<METHOD /path>` reaches `<class: internal service / workload metadata>`" with fields: sink, URI form that bypassed, fetch proof, platform and endpoint, required header/verb precondition, variant counts with the blocking layer, credential reference IDs, and the named cleanup of anything you created.
