---
name: messaging-collab
description: "Invoke this skill when a task involves mail, chat or collaboration stores as a source of credentials, tokens, links and context: mailbox and channel search, inbox rules and delegated access, app-consent and token exposure, and attachment triage."
---

# Mail and collaboration stores as identity surface

Message stores hold other people's secrets in plain sight: pasted tokens, reset links,
file shares, app grants, and rules that quietly forward. Done means a bounded, citable
answer to "what could this access reach, and how much of it is secret material" — proved
with counts and message IDs, never with an exported mailbox.

## Boundaries

- Only accounts the engagement gave you (`list_task_credentials`) or that you own. Reading a
  real employee's mail is a scope and legality event, not a technical achievement; if the
  objective needs it, that is an approval question first.
- Never bulk-export or quote real message bodies into a fact, finding or report: cite by
  message ID + sender + date + folder, and describe content class rather than content.
- No sending, forwarding, replying, deleting, marking read, or rule creation on a real
  user's mailbox. Where the test requires a mutation, do it in your own test mailbox and say
  so.
- Not this skill: the credential itself and proof of its reach → `credential-handling`;
  directory paths that the mail intel opens → `ad-attack`; SaaS/OAuth token abuse across
  cloud APIs → `cloud-attack`; exfiltration channel measurement → `exfil-egress`,
  `collection-exfil`; rule/consent teardown → `persistence-cleanup`.

## Procedure

1. **Map the stores and their auth.** For the domain:
   `dig +short SRV _autodiscover._tcp.<domain>` and `SRV _imaps._tcp`, `nmap -Pn -p 25,110,143,443,587,993,995 -sV <mail hosts>`,
   `curl -sSI https://<host>/autodiscover/autodiscover.xml` plus the EWS
   (`/EWS/Exchange.asmx`) and MAPI-over-HTTP (`/mapi/emsmdb`) probes, and
   `nmap --script smtp-commands,imap-capabilities,pop3-capabilities`.
   For cloud: `curl -sk https://login.microsoftonline.com/<tenant>/v2.0/.well-known/openid-configuration`
   and read the endpoints out of it. `insert_assets` each endpoint with its auth schemes.
   Record whether NTLM/modern-auth/Basic is offered — a Basic-over-HTTP mail endpoint is
   itself a fact.
2. **Authenticate as yourself, once.** Use the given account; confirm which protocol
   actually works (EWS/MAPI/Graph/IMAP/SMTP/Teams/Slack API). Keep the session cookie or
   bearer token in the run dir by reference and register it with `save_task_credential`; the
   value never goes into a fact.
3. **Enumerate the access-grant surface before the content.** In your own mailbox:
   inbox rules (`GET /me/mailFolders/inbox/rules`, or `Get-InboxRule` on-prem), folder
   permissions, and any delegation the store reports. For on-prem Exchange, read the
   mailbox ACLs the directory publishes (`FullAccess` / `SendAs` / `SendOnBehalf` on your own
   and on engagement-owned mailboxes). Tenant-wide app consent
   (`oauth2PermissionGrants`, `servicePrincipals` with `appRoles`, `allowedTenants`) needs
   directory-read rights: check whether the engagement account actually has them, and if not
   record the consent surface as NOT ASSESSED rather than guessing. A rule forwarding to an
   external address, or a consent granting `Mail.Read` cross-tenant, is the finding; the
   message content is not.
4. **Search terms, bounded.** Query the search endpoint with a term list and a small page
   size (`$top=5`, or IMAP `SEARCH SUBJECT`/`TEXT` with a result cap) for
   `password|passwd|token|secret|api[-_ ]?key|bearer|invited|shared with you|reset link`,
   plus the URL schemes `http://` and `://` to surface internal links and hosts. Record
   hits-per-term as counts; open only the 2-3 items that carry the finding.
5. **Read the metadata of what you opened.** Item ID, sender, recipient list size, date,
   folder, whether it contains a secret or an internal URL. Put the *class and count* in the
   fact: "3 messages in Inbox carry long-lived bearer tokens for internal services", and cite
   the item IDs, never the token.
6. **Triage attachments without exfiltrating them.** Enumerate names, sizes and sha256 from
   the attachment endpoint; download only a file that is the objective itself and only from
   your own account, into the run dir, and register it by hash. Password-protected archives
   and macros are a note about the control (DLP/scanning verdict), not a challenge.
7. **Chat and wiki surfaces.** For Teams/Slack/Mattermost/Confluence/Jira: list the
   workspaces and channels the token can read, run the same bounded term search per store
   (`/api/search?query=`, capped), and note the sharing posture — publicly readable spaces,
   guest access, externally shared files/boards. A token's read scope is the measurement
   ("token reads 27 of 31 channels it was not a member of"), not the messages.
8. **Chain outward, then hand off.** Internal URLs, hostnames, service accounts and tenants
   found in step 5 become facts for the lead and assets for the graph (`ad-attack` for
   directory rights, `device-trust-mdm` for CA/tenant trust findings, `pivot-tunneling` for
   reachable hosts). Rule and consent artefacts you created in your own mailbox go to
   `persistence-cleanup` with the removal command.

## Evidence

- Positive: the endpoint and auth scheme used, the query or collection call, hits per term
  with the store total as denominator, item IDs and dates for the 2-3 cited items, an explicit
  "no message body reproduced" statement, and the class of secret found rather than the value.
- Negatives: "Searched Inbox, Sent Items and archive (1,203 items total) for the eight terms
  above via the Graph search endpoint over <window>: 0 items carrying credential material."
  Where you could not look: "Real-user mailboxes NOT ASSESSED — no delegated read was
  authorised; findings cover the two engagement-owned test mailboxes (41 items)."

## Report shape

`<Store> access as <account> exposes <class of secret/context> — <n> of <total> <items> matched; reachable via <endpoint/auth>`
Fields: store + protocol, principal, grant type (rule, delegation, app consent, token scope),
counts with denominator, cited item IDs, data class, what was not touched, remediation.
