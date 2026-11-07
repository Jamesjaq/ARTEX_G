---
name: asset-fingerprinting
description: "Invoke this skill when a task must establish what a service, stack or version actually is: product and framework identification, version narrowing, front-component (CDN / WAF / reverse proxy) detection, and the observed-versus-inferred marker that tells the next seat whether it is holding a fact or a guess."
---

# Asset Fingerprinting

Fingerprinting turns "a web server on 443" into "nginx 1.24 behind Cloudflare, origin is Spring Boot 2.7, backend pool spans two builds". Identification is an asset attribute, never a finding. Done = every asset the intent names carries a product/version (or an explicit `unidentified` marker), the probe that produced it, and a confidence label.

## Boundaries

- Fingerprint only the assets the intent names. If the intent names none, that is a question for your lead, not a licence to choose a target.
- A host that appears in a banner, a redirect target, or a sibling SAN is intel, not a target: record the ownership question as a fact and do not touch it.
- No exploitation here. Banner-plus-CVE claims belong to vulnerability-intel; attacking the stack belongs to web-exploitation and infra-exploitation; mapping the frontend's endpoint list belongs to api-recon; internals of a binary or package belong to rea.
- Read-mostly only: GETs, banner grabs, TLS handshakes, one or two cheap probes per question. Malformed-protocol fuzzing, login attempts, and any method that can lock, restart or fill a service are not fingerprinting.
- Port work stays on the one host and the ports the asset names — never a range sweep.

## Procedure

1. Take stock before sending anything: what is already registered (`list_assets`) so you do not re-register it, and which responses for this host are already recorded. New requests only where the answer is genuinely missing.
2. Cheap tells first: `Server`, `X-Powered-By`, `X-AspNet-Version`, `Via`, cookie names (JSESSIONID / PHPSESSID / ASP.NET_SessionId / connect.sid), the shape of the 404 and the 500, the default page, `robots.txt`, `security.txt`, `.well-known/*`, the favicon hash.
3. TLS layer: `openssl s_client -connect host:443 -servername host </dev/null` — issuer, SAN list, protocol versions, cipher order. A managed edge shows up here as a mismatch against the claimed `Server`.
4. Known-artifact probes — a file served from disk outranks a header, because headers are settable: framework manifests and shipped docs (`/composer.json`, `/package.json`, Spring `/actuator/info`, `/wp-json`, `/wp-content/themes/*/style.css`, Django `/static/admin/`, Next `__NEXT_DATA__` buildId, `/server-status`, `/web-console`), plus `CHANGELOG.md`, `readme.html`, `VERSION`, sourcemap references, `asset-manifest.json`. Existence plus one decisive line is the evidence; do not mine contents beyond the version tell.
5. Non-web services: banner-grab the exact port (`nmap -sV -Pn -p <ports> <host>`, `nc host port`, `openssl s_client -alpn h2`). Record the protocol identity and the banner; do not guess an OS from it.
6. Identify the front component before concluding anything about origin behaviour: CDN/WAF edge header sets, edge-specific error pages, TLS fingerprint. If an edge is answering, a clean response is the edge's answer, not the origin's.
7. Narrow the version and label how you got there: **exact** (version string from a file on disk), **range** (hash or manifest matched against release artifacts), **inferred** (behaviour or a spoofable header). Record which evidence is spoofable.
8. Falsify before registering: repeat the same request a few times — disagreeing answers mean a backend pool, so the honest asset row is "pool spanning A and B", not an average.
9. Register one asset row per identified service: product, version or range, components, front component, confidence, and a one-line evidence note (the probe plus the decisive output). Judgments and unresolved questions go to `record_fact`, not the asset row.
10. Hand off as a fact: name the flaw classes this stack makes worth tasking ("actuator `/env` exposed, config exposure worth an intent"). You do not run the exploitation step yourself.

## Evidence

A positive states product + version or range, the probe, the layer it came from (header / file / TLS / banner / behaviour), and observed-versus-inferred. Header-only identification says so and is marked inferred.

A negative names method and window: "probed host:port with the header, TLS and known-artifact sets for candidates A, B, C (11 requests, window 09:12-09:15); no version-bearing artifact returned; closest match is A on <tell>, so the row reads `A-class, version unassessed`". A test your own environment blocked is a separate line — name the block and mark that layer NOT ASSESSED.

## Report shape

Asset rows, not findings: `<product> <version|range> on <host:port> [behind <edge>]` with confidence, evidence probe, spoofable flags, component list. The one reportable case is disclosure itself: "Version/build disclosure via `<path>` on `<host>`" with the path, the class of what it reveals, and the asset it describes.
