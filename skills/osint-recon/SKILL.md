---
name: osint-recon
description: "Invoke this skill when a task needs public-source collection about an authorized target: infrastructure records, published code and packages, indexed documents, organisation and role structure, and leak surface — each governed by its own policy gate, ownership check and dating requirement."
---

# OSINT / Public-Source Collection

What the target already told the world, and what the world says about it: domains and infrastructure, technologies, org structure, exposed documents, published code. Done = a sourced, dated, corroborated intel set plus an explicit list of the source classes you could not use and why.

## Boundaries

- Authorized use first: collect only about the target this engagement is authorised against, from source classes its rules permit. Third-party platforms are queried **about** the target — never tested, registered on, or worked around to get the answer.
- Ownership before anything else. An item becomes an asset only from a registered in-scope asset, an authoritative registry record, or the target's own publication. A similarly-named domain is an unresolved-ownership fact, not something to probe.
- People are context, not dossiers: role, function, and exposure relevant to the objective. No bulk profile scraping, no contact-detail aggregation, no purchased broker data.
- Program and engagement rules pick the source list: some programs bar cached-engine results, some require notice before touching a suspected leak, some log your queries in a way that identifies the engagement. Read the rules before choosing a method, and honour the strictest applicable reading.
- Anything that authenticates as a real person, circumvents a login wall or a paywall, pays a broker, or reaches out to a human is one `ask_operator` request, not a workaround.
- Not: active scanning against the perimeter (bbot-recon, network-pentest), identifying what a service is (asset-fingerprinting), the application's own API surface (api-recon), the sending step of a human vector (initial-access, human-terrain), or the reputation of the infrastructure you query from (opsec-tradecraft).

## Procedure

1. Resolve the target's own vocabulary first: legal entity names, brands, product names, domains, ASNs and IP blocks, registry records. Bounty scope pages, `security.txt`, job ads, engineering blogs and conference talks name internal stack directly and cost nothing.
2. Classify the source before using it, and gate it: (a) certificates, DNS, public registries — passive, always in bounds; (b) search engines, web archives, code hosting — public, but respect rate limits and terms, no login-wall or paywall circumvention; (c) scanning platforms (Shodan, Censys, ZoomEye, InternetDB, VirusTotal) — query their data about the target, and only their data; (d) professional and social platforms — role-level detail only, no bulk collection; (e) breach and leak corpora — only when the authorisation names them, and stop at class plus count.
3. Passive infrastructure set: certificate transparency (`crt.sh`, dedupe wildcard parents), RDAP/WHOIS, SPF/DMARC/MX, PTR on the named ranges, archived URL harvest. For breadth across these, run bbot-recon in passive mode rather than re-implementing the chain.
4. Published-service posture from scanning-platform records: which services are exposed, what banners they showed, last-seen date. State "as of <date>" on every item and flag anything the engagement will act on for live re-verification.
5. Code and package surface: search hosting and package registries for entity/product/org names plus high-signal shapes (hostnames, bucket names, internal URLs, key prefixes). Published org packages name frameworks and endpoints without touching the target. A discovered credential is recorded by reference through `save_task_credential` and cited by ID — never used, never pasted.
6. Document surface: site-scoped queries for indexed office docs, PDFs, config dumps, `.env`, backup artifacts, plus archived versions of removed pages. Real customer or personal data stops at class, record count, and the URL that served it; do not keep copies of personal data in the work dir and do not quote rows.
7. Corroborate before registering: two independent sources, or one authoritative source (registry record, CT log entry, the target's own statement). A single-source claim stays a fact marked unverified.
8. Date everything. OSINT rots: a subdomain from a 2021 archive may be dangling or decommissioned. Retrieval date plus "needs live verification" is the minimum honest form for anything a seat downstream will spend a probe on.
9. Emit: assets (only after ownership resolution), facts carrying source URL + query + retrieval date, and the gap list of source classes not used, with the reason (no key, egress allowlist, program rule, budget) and what each gap leaves unassessed.

## Evidence

Positive: the item, the source and query, the retrieval date, the corroboration or the single-source flag, and — for anything registered — the ownership basis. Exposure claims carry the class disclosed, a count with its denominator, and a couple of non-personal representative markers (a hostname, a key prefix), never a dump of real records.

Negative: name the source classes actually queried, the query used, and the window — and keep the distinction sharp between "queried, nothing returned" (closed) and "not queried: no key for that platform / barred by the program rule" (NOT ASSESSED, with what would unblock it).

## Report shape

Most output is facts and assets. The reportable shape is "Publicly exposed `<artifact class>` at `<source location>` discloses `<consequence>`" with URL, retrieval date, class, count, corroboration. Unverified leads go to your lead as facts.
