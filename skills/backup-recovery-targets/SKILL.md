---
name: backup-recovery-targets
description: "Invoke this skill when the backup estate is in scope: locating backup servers, agents, snapshots and recovery points, deciding what one backup identity can read, and assessing restore rights — without ever deleting, overwriting or disabling anything."
---

# Backup and Recovery Targets

The backup plane is usually the widest data access in the building and the least-monitored credential store. Done means: the backup product and its servers registered as assets, what one backup identity can enumerate and restore stated as a coverage number, the separation question answered (does the backup plane share credentials with production), and immutability/RPO claims written as observed or as open questions — nothing destroyed to learn it.

## Boundaries

- Absolute stop: never delete, overwrite, expire, encrypt, disable, or reconfigure a job, snapshot, recovery point, retention policy, or agent. Backup destruction is the ransomware precursor and the most destructive action in this catalogue; this seat assesses its possibility and reports the control gap, it does not demonstrate it on a live system.
- Restores only onto an engagement-owned target the operator names, one small file, recorded and cleaned up. A restore into production storage is a change to the client's data, whatever the report calls it.
- Mounted backup images are read-only, always. Sampling inside them follows `collection-exfil` rules: count + denominator + representative rows.
- Not this skill: cloud snapshots/buckets in an authorised tenant (`cloud-storage-data`), keys and vaults (`secrets-keymanagement`), extracting and cracking material from a mounted image (`credential-handling`), the resulting blast radius (`impact-analysis`).

## Procedure

1. Find the plane from what you already hold, rather than hunting for consoles:
   - Processes/services: `ps aux | grep -Ei 'veeam|nbu|netbackup|baksvc|gsrl|r11|bakcsi|commvault|simpana|bacula|duplicity|restic|rdplexposer|windowsserverbackup'` / `Get-Service | ?{$_.Name -match 'Veeam|NB|Backup|Agent'}`
   - Live sessions name the media server better than a scan does: `ss -tn state established` / `Get-NetTCPConnection -State Established | select RemoteAddress,RemotePort | sort -u`
   - Agent config: `C:\ProgramData\Veeam\Backup\*.xml`, the registry uninstall keys, `/etc/*backup*`, scheduled jobs, and the Windows Server Backup recovery-point list (`wbadmin get versions` — read-only, and it prints every recovery point on that host)
   - Backup storage mounts and shares: `findmnt`, `net use`, DFS roots, NFS exports whose path says `backup`, `VeeamBackup$`, `repository`, `repo1`.
2. Register what you found as assets so the whole bench stops rediscovering it: `insert_assets` `{"type":"ip","ip":"10.0.7.5","open_ports":[{"port":9380,"service":"veeam-bnr"},{"port":445,"service":"smb"}]}` plus `{"type":"service","service_name":"netbackup","port":1556,"ip":"10.0.7.5"}`. Derive the port/service from the banner or the agent config you read — not from a remembered list — and say in the fact which evidence named it.
3. Establish the identity question first, because it sizes everything else: what does the agent run as, and what does your held credential reach?
   - `Get-CimInstance Win32_Service -Filter "Name like '%Veeam%' or Name like '%Backup%'" | select Name,StartName,State` / `systemctl cat <unit> | grep -i user`
   - Service accounts named `Bkup`, `SQLBACKUP`, `veeam svc` in a domain context are usually the highest-privileged non-tier-0 accounts in the estate; record the group membership you can read, and hand the privilege question to `lateral-movement` rather than exploiting it yourself.
   - Test enumeration with the rights you hold, one call: the console REST/UI reachable? job list readable? repository listing readable? Record each as allowed or denied with the error line.
4. Answer "what does it protect" as a coverage number from read-only job/policy data: protection groups, machine list, agent count, database/application items, and recovery point inventory (`wbadmin get versions`, Veeam restore points per VM, snapshot lists). The denominator the report needs is "N of M estate hosts covered, oldest recovery point <date>, newest <date>".
5. The immutable/RPO question — ask the config, do not test the delete:
   - Read: retention settings per job, object-lock/WORM/copy-guard flags on the repository, air-gap or "backup to offline media" steps, separation of duties (who can edit a job), job schedule vs the stated RPO/RTO, offsite/second-copy presence, whether the repository requires multi-factor/role approval for deletion.
   - Where the console will not show it, write the question into a fact for the operator (`ask_operator` for a single read-only config export). Proving immutability is off by deleting a recovery point is a line you do not cross.
6. Backup as a data and credential source, subject to `collection-exfil` sampling rules:
   - Configuration XMLs and job files commonly carry stored service accounts, encrypted credential blobs, and the full host list — record path + which class, save accounts to `save_task_credential` by reference.
   - VM/VHD/VHDX files and database backups on a readable share = every host and store inside them. Mount read-only only when the objective needs the content: `Mount-VHD -Path <file> -ReadOnly`, or `udisksctl loop-setup -f <file>` then read the partition read-only; from a mounted image, the interesting objects are the SAM/NTDS/registry hives and the app config files — take the *presence and path* and hand the extraction to `credential-handling`.
   - Sample 1-3 items per store, count the rest, and note in the fact which recovery point (name, timestamp, size, source host) the sample came from.
7. Restore-rights assessment without restoring: enumerate the ACLs/RBAC roles readable for your identity (backup console role list, `Get-Acl` on the repository path, write permission on the restore target, self-service restore that writes back into production-adjacent storage). The finding is "identity X holds right Y over Z recovery points covering W hosts", which is provable by reading.
8. Answer the separation question in one fact: is the backup plane's credential set distinct from production (same service account used on both, same local admin password pattern, shared management VLAN, no MFA on the console, console reachable from the user segment)? This is the edge `attack-path-mapping` needs, so write it as an edge with asset ids.
9. Snapshot enumeration in cloud tenants only where the tenant is in scope, read-only (`describe-db-snapshots`, `compute disks list --filter snapshot`, `az snapshot list`) with counts, then defer the depth to `cloud-storage-data`.

## Evidence

- A positive needs: how the product/server was identified (the process line, config path, service name, or banner — quoted), the identity that made the read work, counts of protection objects and recovery points, one recovery point's metadata, and the restore-right statement labelled observed (from an ACL/role read) or inferred (from product defaults).
- A negative must separate the three kinds: "no backup agent observed on this host: 2 service lists and 2 connection tables read successfully, none matching a known agent → agent absent (observed)"; "backup server named in `hosts` file but not reachable through the current route → coverage NOT ASSESSED"; "console reachable, login blocked by MFA → job configuration NOT ASSESSED". None of these is "no backups exist".
- Immutability verdicts state their basis: "object-lock reported enabled on bucket `<id>` by the read-only config call" vs "unknown — retention settings not readable as this identity".

## Report shape

Title: "<product> at <host>: <identity> can <enumerate|restore> <N> protection groups covering <M> hosts, <K> recovery points; separation from production: <yes|no|not assessed>". `vulnclass` = backup-estate-exposure / insufficient-access-control-on-backup, `severity` from coverage × credential separation × restore rights, `asset_ids` = backup server + repository rows, `evidence` = identification lines + counts + the role/ACL read, and the summary explicitly states that no destructive test was performed.
