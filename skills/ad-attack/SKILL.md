---
name: ad-attack
description: "Invoke this skill when a task involves the directory itself: LDAP/SMB enumeration of an in-scope domain, ACL and group topology reading, Kerberos account classes, delegation, forest trust edges, and the ordered recon-then-proof route from a foothold to privileged domain control."
---

# Active Directory enumeration and attack paths

The directory is a graph of rights, not a loot list. Done means: every edge on one
candidate path is backed by the query that showed it, the path is proved once against an
object you own, and everything you chose not to do is recorded. Enumeration quality is
judged per calling principal — the same domain answers differently to an anonymous bind, a
standard domain user, and a machine account.

## Boundaries

- Work only the domain/forest named in the engagement and present in the asset store; a
  trust edge that leads to a forest nobody tasked you with is a dead end to record, not
  a door to try.
- Read-only by default. No ACL grants, no group membership changes, no password resets, no
  GPO edits, no computer/service account creation except objects the engagement already
  owns for you. Any step that needs a write to a production object goes up before it runs.
- Not this skill: CA templates and enrollment rights → `ad-pki-certificate`; coercion and
  relay → `privilege-relay`; host-to-host reach from the foothold → `lateral-movement`;
  local escalation on the foothold → `post-exploitation`; hashes, tickets and their
  reach → `credential-handling`; graphing reachable paths → `attack-path-mapping`; Entra
  / cloud identity → `cloud-attack`, `device-trust-mdm`.
- Bound every sweep (paging, timeout, attribute list, one OU at a time). An unbounded
  query storm against a production DC is an availability event and a detection event.
- The ARTEX container ships nmap/dig/curl/jq/python3 only — no ldapsearch, Impacket,
  netexec, certipy or BloodHound collectors. Check `command -v ldapsearch netexec
  bloodhound-python certipy` before planning a step and hand the single missing-binary
  install up; do not hand-roll BER/SMB bytes in Bash to compensate.

## Procedure

1. **Fix the principal.** `list_task_credentials`, then record which account each
   enumeration runs as and how it holds auth (`klist` for a ccache, `whoami /groups` plus
   `whoami /priv` on a Windows foothold). Every later fact names this principal; an
   enumeration result without it is uninterpretable.
2. **Find the infrastructure.** `dig +short SRV _ldap._tcp.<domain>`,
   `_kerberos._tcp.<domain>`, `_gc._tcp.<domain>`; then
   `nmap -Pn -p 53,88,135,139,389,445,464,636,3268,3269 -sV <dc-ips>`. Read rootDSE
   (`ldapsearch -x -H ldap://<dc> -b '' -s base defaultNamingContext configurationNamingContext
   schemaNamingContext dnsHostName`), and functional level from `msDS-Behavior-Version` on
   `CN=Partitions,<configurationNamingContext>`. `insert_assets` every DC with its role.
3. **Anonymous vs authenticated delta.** Try `ldapsearch -x` (no bind) against a naming
   context; if it returns, capture what an unauthenticated caller can read (`objectClass`
   counts, `-E paging=500`, capped). Repeat as the step-1 principal. The *difference* is
   the finding-relevant fact, not either snapshot alone.
4. **Bounded object inventory.** One attribute-limited query per class, output to a file in
   the run dir, counts into the fact: computers, users, groups + membership
   (`member`/`memberTransitive`), OUs, GPO links (`gPCMachineExtensionNames`),
   `msDS-GroupMSAMembership` (gMSA). State which OUs you sampled and the domain total.
5. **ACL / ACE reading.** Pull `nTSecurityDescriptor` (SDDL) on Tier-0 objects: Domain
   Admins, other adminSDHolder-protected groups, Domain Controllers OU, AdminSDHolder,
   tier-0 computers. Resolve SDDL strings to trustees with
   `impacket-lookupsid <domain>/<own account>@<dc>` (bounded) or the ACL collector. Record
   trustee, right, inherited-or-not, object. AD CS enrollment rights (`0e10c9c8-82a8-4a8b-bbf6-33a92f682f1f`)
   belong to `ad-pki-certificate`.
6. **Kerberos account classes** — counts and identities, never the material itself:
   SPN-bearing non-server accounts (Kerberoastable), `userAccountControl` with
   `UF_DONT_REQUIRE_PREAUTH` (AS-REP-rollable), `TRUSTED_FOR_DELEGATION` /
   `TRUSTED_TO_AUTH_FOR_DELEGATION`, `msDS-AllowedToDelegateTo`,
   `msDS-AllowedToActOnBehalfOfOtherIdentity` (RBCD), password-not-required and
   `UF_DONT_EXPIRE_PASSWD` service accounts. Harvested hashes/tickets go to
   `credential-handling`; nothing here cracks.
7. **Delegation analysis.** For each edge from step 6: who may act as whom, against which
   SPN, hosted where, and whether the target is Tier-0. Constrained delegation is provable
   read-only (see `privilege-relay` for the S4U exercise); RBCD needs a write to a real
   computer object and therefore stops here.
8. **Trust edges.** `Get-ADTrust -Filter *` (or the Trusts collector): trust type,
   `TrustDirection`, `TrustAttributes` (downlevel/quarantined/non-transitive/external),
   SID filtering. Enumerate the remote side's SIDs only within the named forest. For each
   edge state explicitly: direction, what membership/right would have to be granted before
   anything crosses, and whether you can traverse it today. A topology fact is not a
   compromise.
9. **Path selection gate.** Hand raw edges up as facts; do not attempt a step-10 proof
   until one candidate edge has principal + right + object recorded and the lead has
   picked the shortest chain.
10. **One proof, on your own object.** Exercise the chosen edge against an object the
    engagement owns (your test account's DACL, your own computer object, a group you
    created). Prove identity before/after, touch nothing else, and write down the
    mechanisms you deliberately skipped because they would have been destructive.

## Evidence

- A positive edge fact = query command + the attribute line it returned + trustee and
  object names, `confidence: observed`, principal and enumeration window named.
- A proof fact = the command, the resulting identity/token principal, and every object ID
  touched. One consolidated fact per intent, details in `detail`, evidence one line.
- Denominators: "ACL readable on 14 Tier-0 objects; 2 grant WriteDacl to Domain Users",
  "412 SPN-bearing accounts enumerated; 6 Kerberoastable".
- Negatives carry method and coverage: "Kerberoastable: none, via
  `(&(objectClass=user)(servicePrincipalName=*))` over the full domain partition at
  <time>; no Kerberoast path — closed." Where a collector could not run: "session
  enumeration NOT ASSESSED — BloodHound collector absent, no operator install before
  wrap-up", never a silent blank.

## Report shape

`Directory escalation path: <principal> → <right/ACE or delegation> on <object> → <Tier-0 target>`
Fields: principal and auth method, edge class, proof basis (executed-live on own object /
static-analysis-unexecuted), objects touched, enumeration coverage and what was skipped,
and the single setting that breaks the edge.
