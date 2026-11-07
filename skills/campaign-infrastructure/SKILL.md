---
name: campaign-infrastructure
description: "Invoke this skill when the engagement needs its own hosts, domains, certificates or mail channel stood up, aged, health-checked, reputation-checked or torn down: provisioning through the operator, cell design, aging lead times, reputation checks, and the burn-down inventory."
---

# Campaign Infrastructure (engagement-owned stand-up, aging and burn-down)

Our side of the line: the hosts, domains, certificates, redirect layers and mail channel the campaign runs on — none of them the target's, all of them registered, defensible and destroyable. Done means every piece is registered with a purpose and a cell, health- and reputation-checked with dates, and closes with a burn-down inventory the operator can execute. Acquisition itself is a human act: you plan, document, and hand over one request.

## Boundaries

- Authorised use only, and the authorisation points the other way here: this is infrastructure the engagement owns, paid for and named. Nothing here provisions anything against a host, service or domain that is not ours.
- Third-party systems are not free terrain. Re-pointing a legitimate platform, someone's open redirector, or an unrelated VPS as cover is an operator decision made with the client's contract in front of them — never a bench convenience.
- Registration, purchase, account creation, DNS changes and deletion are operator actions. You produce the request and a reviewable definition; the operator executes it. Do not attempt them from the task environment.
- This skill is NOT the channel or beacon discipline (`c2-ops`), NOT the indicator ledger and cross-engagement reuse call (`opsec-tradecraft`), NOT the artifact (`payload-lab-development`), NOT the RoE and scope statement (`engagement-scoping`), NOT operator comms and client-data containment (`secure-operator-comms`).
- Stop condition: any piece that cannot be traced to a paying owner, this engagement, and a destroy method is decommissioned rather than used.
- Never submit engagement- or client-derived material (a payload, an IoC list, a client domain) to a third-party reputation or scanning service as a way of checking our own hygiene.

## Procedure

1. Draw the cell topology before requesting anything. At minimum separate identities for payload hosting, redirection, listener/tasking, mail/send, and operator access; note which pieces are shared and which single burn must not cascade. Compartment is why a burnt mail domain does not cost you the listener.
2. Hand the operator one provisioning request per piece: purpose, cell, provider and jurisdiction, sizing, expected lifetime, and the destroy method. Do not pick a provider because it is fast — the operator's identity and payment method attach to it.
3. Plan lead times against the engagement calendar and write them into the plan, because they are the usual reason a campaign stalls:
   - domain: registration plus registration-data settling; expect days before it looks unremarkable, and register at least two weeks ahead of need;
   - IP/host: a fresh address from a range with bad history ages worse than the domain does;
   - certificates: ACME issuance history matters — a cert with no issuance history on a domain with no DNS history is itself a tell;
   - mail send: warm-up plus SPF/DKIM/DMARC propagation is days, not hours;
   - blocklist and cache removal after a burn runs days-to-weeks, so budget it into rotation rather than into the tasking window.
4. Register everything as it lands:
   - each host/IP/FQDN through `insert_assets` (type=ip, type=root_domain or type=subdomain as fits);
   - one `record_fact` per piece: cell, purpose, provider, provisioning date, planned teardown date, the operator who owns it;
   - the asset graph is the ledger other benches reuse — an unregistered host gets re-discovered as a target.
5. Define, do not hand-run. Bring each piece up from a repeatable definition the operator can review and destroy (compose file, Terraform snippet, install script in the work dir with its hash), not a sequence of terminal commands nobody can re-execute or roll back.
6. Health check on a stated schedule:
   - reachability and identity: `dig +short <fqdn>`, `curl -sv https://<fqdn>/`, `nc -zv <ip> <port>`;
   - certificate chain and expiry: `openssl s_client -connect <fqdn>:443 -servername <fqdn>` where the binary exists;
   - surface drift: `httpx -u <url>` against the cover backstory the piece claims;
   - record the verdict per piece with its date — a piece unchecked for two weeks is a piece you cannot trust in front of a client.
7. Reputation check before use and after anything unusual:
   - `whois <domain>` for age, registrar and previous-owner residue;
   - `dig +short -x <ip>` for reverse-DNS plausibility;
   - public DNSBL lookups on the send path if mail is in play (`dig +short <reversed-ip>.zen.spamhaus.org` and equivalents);
   - `web_search` for the domain's public history.
   Two results are red flags and both get reported: our piece is already listed or burned, or the recycled domain still shows the previous owner's live content — using it would touch a third party.
8. Age deliberately and record what aging did. A domain created this morning, a certificate issued this morning and a VPS opened this morning, presented as an established service, is an indicator set a defender reads instantly. Note age at first use in the fact, so the report can say whether exposure came from the control or from our thin preparation.
9. Burn detection and rotation: define each piece's burn signal up front (listener log pattern, block page, vendor or client notice, an alert arriving from the defender) and treat the first signal as a burn event —
   - mark the piece burned with date and evidence, stop tasking against it;
   - rotate to a reserved cell; let the operator decommission rather than rehabilitate;
   - never push a burned IP or certificate back into service to save lead time.
10. Teardown and burn-down inventory at close, as one list the operator can execute: hosts and destruction state; domains and renewals including auto-renew off; DNS records; certificates and ACME accounts; security-group and firewall rules; monitoring agents; mail authentication entries; cloud snapshots and backups; provider accounts; credentials and keys resident on those pieces. Each row names who destroys it, how, and how destruction is verified. Hand the data-class side of the same list to `secure-operator-comms`.

Gate: do not go to 6 until 4 produced registered, facted assets; do not go to 10 until every piece is either verified burned-down or explicitly handed to the operator as still standing.

## Evidence

- Positive: dated per-piece state — provisioning date, age at first use, health-check output (command plus the decisive line), reputation-check results with the sources queried, burn events with evidence, and the teardown row with its verification method. A hole in our own infrastructure record is itself a finding for the operator.
- Negative, honest: "reputation check on <domain/ip>: whois age <N> days, no listing across the three zones queried on <date>, reverse DNS absent — assessed clean for this purpose; absence of listing is not proof of good standing." Where a check could not run from this environment, NOT ASSESSED with the reason.
- Do not assert a third party's ownership or reputation beyond what you queried, and never present a previous domain owner's content as evidence about the target.
- Keys, tokens and provider credentials go to `save_task_credential` and are cited by ID; no live value in a fact, asset summary or report.

## Report shape

Title: `Engagement infrastructure: <piece/cell> <aged / burned / burned-down> — <consequence for this engagement>`.
Fields: piece reference (asset id), cell and purpose, provider and jurisdiction if the operator disclosed them, provisioning and teardown dates, age at first use, health and reputation checks with dates and sources, burn events with evidence, destroy method plus verifier, and credential IDs held on it.
