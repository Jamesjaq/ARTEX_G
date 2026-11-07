---
name: secure-operator-comms
description: "Invoke this skill when operator identities, engagement channels or client data could cross into the target's sight: per-engagement identity separation, attribution-leak review of outbound artifacts, out-of-band communication where a surface may be monitored, and destruction of engagement material on the contract's clock."
---

# Secure Operator Communications (identity separation and data containment)

Two failures live here: an operator's personal identity surfacing in target-visible material, and client data travelling over a surface the target can read.

Done means every engagement runs on its own separated identities, every outbound artifact was leak-reviewed before it shipped, sensitive discussion has an out-of-band path where a surface may be monitored, and material is destroyed on schedule with an attestation.

This protects the operator, the client's data and the firm's defensibility — not the stealth of an operation.

## Boundaries

- Authorised use only, and note the direction: this governs our channels and our data handling, never the target's. Nothing here intercepts, accesses or monitors a target's communications.
- Who outside the engagement may see material is not an agent's call. Any transfer to a third party, regulator, vendor or insurer is the operator's own act with a named recipient and a recorded reason: prepare and check the channel, do not open it.
- This skill is NOT the target-facing indicator ledger (`opsec-tradecraft`), NOT custody and PII classification of artefacts (`evidence-custody`), NOT captured secret material (`credential-handling`), NOT the target's own mail stores as a data source (`messaging-collab`), NOT infrastructure burn-down (`campaign-infrastructure`).
- Stop conditions: a suspected compromise of an operator account or channel; a client-data leak into a public repository or to a wrong recipient; any request to move engagement material onto a surface the target controls.
- On any of those, pause the affected front and hand it up as an incident before any cleanup that would destroy the evidence of the leak itself.
- No engagement material — client data, findings, evidence files, captured credentials, hostnames from a report — is ever held in a personal account, on a personal machine, in a private cloud bucket, or in a public or unauthenticated repository, including "temporarily".

## Procedure

1. Build the identity register at engagement start, one row per identity the engagement uses: mail address, phone, VPS or provider account, code-hosting account, chat workspace, certificate contact, and the human who owns it.
2. `record_fact` the register, so a later auditor can tell engagement identity from personal identity from target identity.
3. Never reuse a personal address or personal account for any engagement function.
4. One identity set per engagement, no shared aliases. A mailbox, workspace or provider account used by two clients is the leak that surfaces in the second client's report.
5. Where the contract permits an organisation-level identity (a firm abuse contact, the firm's mail domain), use that — and mark which category each identity is in, because the reuse rule differs: engagement-specific never crosses, firm-level may.
6. Pre-flight every outbound artifact for attribution leaks before it can reach the target or a third party. The classes that actually leak:
   - usernames and home paths in scripts, comments and stack traces (`/home/<name>`, `C:\Users\<name>`);
   - a personal or firm mail address as a git commit author, or in a `From:` header;
   - a home or office IP in a curl invocation or in recorded traffic;
   - absolute work directories and personal filenames;
   - hostnames from `hostname` / `uname -n`, local timezone, locale and language settings.
7. Sweep with `search_all_worker_traces` and `traffic_search` for what was sent, and `Grep` over the work dir for what is sitting on disk.
8. Fix by restriping or rotating, and record what changed. Never silently edit another seat's evidence.
9. Nothing target-visible carries client data: files uploaded to a target, listener certificate subject fields, redirector page content, headers on a consented send, filenames inside an artifact — all of it can be read by the defender.
10. Use placeholders (`{client}`, `{target_ip}`) in anything that leaves our channel, and check the artifact with `Read`/`Grep` for real client names before release.
11. Decide per surface whether it may be monitored before using it for anything sensitive. Treat as monitored by default: the client's VPN and network, client-supplied mail or chat tenants, client webmail, any workspace the client controls, client-managed endpoints, and conference bridges on their estate.
12. A client-owned surface is fine for client business the contract expects; it is never the place for internal deliberation about that same client.
13. Route sensitive discussion out-of-band and name the channel in the fact: an internal chat or mail tenant the client cannot read, for anything about severity, a named person, a contract dispute, or our own exposure.
14. Where a monitored surface must be used with the client, raise it in the meeting rather than in writing on their system.
15. Never paste finding detail, credentials or client PII into a channel whose membership you have not checked.
16. Platform notification channels are an outbound path. Before pointing a task digest at Telegram, Slack/Teams, WeCom/Feishu/DingTalk, a generic webhook or SMTP: confirm every recipient is inside the engagement.
17. Review what the digest actually renders — finding titles, severities, hostnames, evidence excerpts — and keep credentials masked as the platform masks them.
18. Record the destination and the approver. Enabling a channel to a broad group is a disclosure.
19. Held material follows the contract's retention clock, not convenience. Classify by source: our artifacts, client data, captured credentials, personal data.
20. Confirm the retention period and destruction method with the operator, and attach a dated attestation naming what was destroyed and where.
21. Custody formality and PII scrub rules are `evidence-custody`'s. What is yours is that destruction must also reach the copies in operator mail, chat, local drives and provider snapshots — not just the platform database.
22. Access discipline while working: no client material cached on personal devices; no engagement code in a personal repository; no third-party AI, paste or file-transfer services for client data or captured content.
23. Keep screen-share and clipboard care when reviewing the credential store, and lock or log out shared operator sessions when stepping away.
24. Leak and compromise playbook: on any suspected leak, stop the affected front (`pause_task`), then record what leaked, where, when, who can now read it, and the last action that could have caused it.
25. Hand the operator one `ask_operator` with a recommended action — rotate, restrip, notify, destroy — and wait.
26. Identity rotation follows `opsec-tradecraft`; a leak that reaches the client's defenders is also a deconfliction event, so it goes to `engagement-deconfliction` in parallel.

Gate: do not go to 13 until 1-5 produced a register with no personal identities in engagement rows; do not go to 19 until 6-10 are clean for every artifact that left the engagement.

## Evidence

- Positive: a dated observation with its method — "identity register lists 4 engagement identities, none personal"; "pre-flight of <artifact class> found 2 personal-home paths in <location>, restriped before release, hash recorded"; "notification channel <destination> approved by <role> on <date>, digest content reviewed for finding titles and hostnames".
- A leak finding states blast radius honestly: what class of material, how many items, who could read it, over what window, and what rotation closed it. Never claim containment because it was "probably not noticed".
- Negative: "outbound artifact set for this engagement (N files, <window>) swept for personal-identity patterns via <method>; zero matches — assessed clean for the checked classes; unchecked: binary metadata and embedded git history."
- Name what you did not check; an unqualified "no leaks" is not a result. Where a store was unreachable, write NOT ASSESSED with the reason.
- Secrets, tokens and channel credentials stay in `save_task_credential`, cited by ID; webhook URLs, SMTP passwords and bot keys never appear in a fact, finding or report.

## Report shape

Title: `Operator channel / data handling: <issue class> on <surface> — <exposure or assessed-clean>`.

Fields: engagement and identity-register reference; surface and its monitoring assumption; artifact or channel inspected; method and window; findings by class with counts; rotation or restripe actions with dates; retention class and destruction-attestation reference; credential IDs involved; the operator decision authorising any external transfer.
