---
name: lateral-movement
description: "Invoke this skill when a foothold must be turned into reach: enumerating host-to-host candidates from a compromised machine, proving that a credential authenticates without logging in everywhere, and stating the blast radius of one set of credentials."
---

# Lateral movement and pivot reach

The output is a reach matrix, not a list of boxes you opened. Done means: for each credential
class, how many in-scope hosts are reachable, how many accept it, what the widest proven chain
is, and which interesting neighbours were refused because they are not in scope.

## Boundaries

- Confirm each host against the asset store before it receives a packet; register discovered
  neighbours and let the gate decide. A rejected or unregistered host is a recorded dead end.
- Prove the *lowest sufficient* tier. A finding needs reach plus authentication, not a shell on
  six machines. Every host you log into is an artefact and an incident volume for the client.
- Attempts use credential material that is correct or engagement-owned; two failed
  authentications against the same account on the same host is a lockout risk — stop and
  record, do not retry with variants.
- Not this skill: how the foothold escalated locally → `post-exploitation`; segments reachable
  only through the foothold and the tunnel mechanics → `pivot-tunneling`; directory paths and
  group rights → `ad-attack`; the credential values and their proof calls →
  `credential-handling`; relay/coercion routes to a hop → `privilege-relay`; cloud/SaaS reach →
  `cloud-attack`; egress measurement → `exfil-egress`.

## Procedure

1. **Build the candidate list from the host's own state — locally only.** `net session`,
   `net use`, `qwinsta`, `ss -tnp` / `netstat -antp` for established internal peers,
   `Get-ItemProperty` on `MapNetworkDrive` entries, SSH client history and `known_hosts`,
   shell history, mapped printers, recently accessed UNC paths, the AD computer objects in the
   same subnet (`ad-attack` step 4 output). Write the candidate list to the run dir before
   contacting anything.
2. **Tier 1 — reachability, no authentication.** From the foothold only:
   `nmap -Pn -p 22,135,139,445,389,1433,3389,5985,5986 <candidates>` (or per-host
   `nc -zG 2 <host> <port>` where nmap is absent). Record open/closed/filtered per host.
3. **Scope gate between 2 and 4.** For each host that answered, `list_assets`; anything not
   already registered goes through `insert_assets` and is not touched until it comes back
   in-scope. Log the refusals — "8 reachable hosts, 3 out of scope, not probed".
4. **Sample the authentication question.** Pick one or two hosts per segment/role from the
   in-scope set (say the sampling rule in the fact), never the whole fleet.
5. **Tier 2 — does the credential authenticate, without a session.** Pull the credential by
   reference (`list_task_credentials`), then materialise it into the interactive session's
   environment rather than the command line (`shell_open`, then `read -rs LVAR < file` with the
   file mode 0600 in the run dir, removed after use) so the value never lands in process
   arguments or the recorded request chain. Then:
   - SMB: `netexec smb <host> -u <user>` (auth verdict only) or an `IPC$`-only connection —
     no shares written, no commands.
   - Kerberos, target-free: `kinit` then `kvno cifs/<host>` / `host/<host>` — a service ticket
     from the KDC proves that hop is authorisable without ever contacting the host.
   - WinRM: enumerate the listener, authenticate, run nothing.
   - SSH: `ssh -o BatchMode=yes -o ConnectTimeout=5 <user>@<host> true` — exit status 0 means
     the key/credential was accepted, and nothing ran but `true`.
   - MSSQL/Oracle/Postgres: connection-authenticated version query only; `linked servers`
     enumeration via catalogue views, no cross-server execution.
6. **Tier 3 — execution, only where the objective needs it.** One host, one benign
   identity-returning command (`whoami` / `id`), and the artefacts it creates (service, temp
   file, logon profile) get registered in the `persistence-cleanup` ledger as they appear.
   Remote service creation, scheduled-task execution and WMI/DCOM activation on a production
   host are approval-first actions.
7. **Chain and compute blast radius.** Per credential class: hosts reachable / authenticating
   / executing, whether any is a domain controller or Tier-0, chain depth from the foothold,
   and the shortest path to the named objective. Write the matrix to the run dir as a table,
   then one consolidated fact summarising it with the denominators.
8. **Stop at the objective.** Once the widest proven chain is recorded, further hops are
   damage rather than proof. Record what you could have reached and deliberately did not.

## Evidence

- Positive: the per-tier counts with the exact check used at each tier ("reachable 14,
  authenticate 5 via SMB with credential ID 12, execution proven on 1 host, `whoami` returned
  <identity>"), the sampling rule, hosts touched and artefacts left, plus the token/`kvno`
  results for target-free proofs.
- Negative: "SSH: 6 hosts reachable on 22, 0 authenticate with the deploy key (credential ID
  9) — `BatchMode` probes returned Permission denied on all 6; no further attempts. RDP: not
  assessed — 3389 filtered everywhere in scope (nmap, <window>)."
- Out-of-scope refusals state the host identifier and why the refusal happened, so the report
  can show restraint as work.

## Report shape

`Credential <ID/type> from <foothold> reaches <n> of <m> in-scope hosts via <protocol>; <k> authenticate, <j> execute, nearest Tier-0 host <hostname>`
Fields: source host, credential reference, protocol/port, per-tier counts and denominators,
sampled-vs-total, artefacts and teardown state, out-of-scope hosts refused, remediation.
