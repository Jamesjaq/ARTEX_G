---
name: persistence-cleanup
description: "Invoke this skill when a task requires durable access on an authorised host or tenant, or when any engagement step has created a state change: the mechanism choice, the artefact ledger kept as each item is created, and the verified-removal inventory that must accompany the finding."
---

# Authorised persistence and the teardown ledger

Durability is demonstrated, never accumulated. Done means: one mechanism, on one named host,
registered in a ledger the instant it existed, its detectability observed, and its removal
independently verified before wrap-up — with any residue reported rather than quietly left.

## Boundaries

- Persistence is authorised, documented and reversible by construction: the test for any
  mechanism is whether the client can evict it with the tools and commands they already have.
  Boot, firmware/UEFI and hypervisor-level classes are lab-only territory and are never
  offered as an engagement action.
- One mechanism per demonstration, per host the engagement names. Writes to shared directory
  objects (ACLs, GPOs, group membership, service accounts, OAuth consent, inbox rules on a
  real mailbox) are approval-first: they outlive your session and are felt by other people.
- A ledger row that cannot be verified as removed is a reportable residual, not a footnote.
- Not this skill: how the foothold was obtained or escalated → `post-exploitation`,
  `initial-access`; whether the sensor stack saw the artefact → `telemetry-sources`,
  `detection-coverage`; implant/listener mechanics → `c2-ops`; secret material used by a
  persistence mechanism → `credential-handling`; directory rights themselves → `ad-attack`;
  retention and custody of the artefact files → `evidence-custody`.

## Procedure

1. **Open the ledger before the first write.** In the run dir, `teardown_ledger.tsv` with:
   `host | mechanism | exact artefact (path / key / task name / rule ID / cert serial) | creation command | removal command | independent verification command | created (UTC) | state | fact ID`.
   Every later step fills a row as it acts; nothing is reconstructed at the end.
2. **Get the pre-state snapshot.** Before creating anything:
   `schtasks /query /fo csv | head`, `reg query HKLM\Software\Microsoft\Windows\CurrentVersion\Run`,
   `Get-CimInstance -ClassName Win32_StartupCommand`, WMI consumer list
   (`Get-WMIObject -Namespace root\Subscription -Class __FilterToConsumerBinding`), services
   (`sc query type= all` / `systemctl list-unit-files --state=enabled`), `crontab -l` for the
   account, `authorized_keys`, `Get-LocalGroupMember Administrators`. Save to the run dir —
   this is the diff you will prove removal against.
3. **Pick the mechanism the objective actually needs**, cheapest and most evictable: scheduled
   task or cron/user unit (`schtasks /delete`, `crontab -r`), service
   (`sc delete`, unit file removal + daemon-reload), Run key
   (`reg delete ... /v <name> /f`), a file in a startup path (delete + confirm gone),
   `authorized_keys` line added with a comment identifying the engagement, WMI subscription
   (`Remove-WMIObject` on filter, consumer and binding — all three), AD-side (a right granted
   to your own object, reversed with the same ACL edit), mailbox-side rule or app consent
   (`Remove-InboxRule`, revoke grant in Entra), certificate minted for auth (revoke by request
   ID/serial — see `ad-pki-certificate`).
4. **Create, then register in the same turn.** After the creation command, immediately:
   `record_fact` (artefact path/ID, host, mechanism, removal command, ledger row),
   `insert_assets` if a new object appeared, `save_task_credential` if you created a
   service account or key pair (value by reference, never in the fact text). The ledger row
   gets the fact ID back.
5. **Observe detectability once.** Re-run the step-2 inventory and note whether the artefact
   shows up in the client's own tools, and whether anything logged or alerted — this is
   usually the actual finding: durable and visible, or durable and invisible. Do not tamper
   with the agent to make the test pass.
6. **Use it minimally.** One reconnection proving durability is enough (a fresh session at the
   same privilege, or one command). Do not build a second channel while the first is up.
7. **Teardown, then verify independently.** Run the removal command from the ledger, then the
   *verification* command that you did not use to create it (query the task list, read the key,
   list consumers, diff against the step-2 snapshot). Mark the row `verified` with the command
   output file. Remove engagement-created accounts/keys by revoking or deleting, and rotate
   any credential the mechanism consumed if it is still live.
8. **Close the ledger.** One consolidated fact: mechanisms created (n), verified removed (n),
   residues outstanding (list with host, artefact, why not removable, who must act). Every
   residual is also an explicit message to the operator — leftovers are never shipped silently
   inside a report body.

## Evidence

- Positive: ledger rows quoted with the artefact identifier, the creation and removal commands,
  both verification outputs (file path + the one proving line), the before/after snapshot diff,
  and the detectability statement (what saw it, what did not).
- Negative: "Durability NOT ASSESSED — no host in scope was authorised for a write; posture
  conclusions come from the read-only autorun inventory of 3 hosts." Or: "4 artefact classes
  tested (scheduled task, Run key, WMI subscription, cron): 4 created, 4 removed and verified
  against snapshots; 0 residues. WMI class not observed by the client's inventory tool."
- A finding that persistence is easy must still show removal, or it is an open incident.

## Report shape

`Authorised persistence on <host>: <mechanism> survived <reconnection test>, removed and verified <timestamp>`
Fields: host, mechanism, artefact identifier, creation/removal commands, detectability verdict
with sensor source, residues outstanding, and the durable-risk statement the client needs
(e.g. "an attacker could keep this foothold for 30 days and their only visible signal is <X>").
