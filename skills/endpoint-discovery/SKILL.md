---
name: endpoint-discovery
description: "Invoke this skill when the subject is one foothold's own state: who is logged in, what is running, what is configured, what is installed and patched, and what guards the host — captured read-only and written back as posture facts."
---

# Endpoint Discovery (foothold posture capture)

The host you hold is a document: it records who uses it, what it talks to, what guards it, and how stale it is. Done means one consolidated posture fact per category, anchored to the host's asset id, so movement seats plan from your read instead of logging in again.

## Boundaries

- Only a foothold this engagement established and the scope registers as in-bounds; confirm its asset id first. A host you cannot tie to an authorised foothold is a stop, not a starting point.
- Read-only by contract: no service restarts, no account creation, no log clearing, no policy or driver changes. Telemetry blinding and sensor tampering are lab work owned by `edr-bypass-re` — here you only describe the guard you see.
- Other people's data is not enumeration material: capture the shape (paths, counts, classes), do not read through a colleague's files to prove you could. Deep credential harvesting is `credential-handling`; the backup estate reachable from this box is `backup-recovery-targets`.
- Captured secret material (hashes, keys, tokens, cached credentials) goes to the credential store by reference and is cited by id; it never enters a fact, finding, or report body.

## Procedure

1. Anchor identity first — every later observation is only worth what this identity can do:
   - Windows: `whoami /all`, `whoami /priv`, `whoami /groups /fo csv`, `net localgroup administrators`
   - Linux: `id`, `sudo -n -l`, `cat /etc/passwd | awk -F: '$3==0'`
   - Record: SID/groups, privilege list (note SeDebugPrivilege, SeImpersonatePrivilege as *present*, not as an exploit claim), elevation state.
2. Users and sessions:
   - Windows: `query user`, `Get-CimInstance Win32_LoggedOnUser`, `Get-ChildItem C:\Users | select LastWriteTime`, `cmdkey /list` (target names present, not contents)
   - Linux: `w`, `last -Fi -a | head -40`, `lastlog`, `ls -la /home`, `ss -tnp` for other users' sessions
   - Fact shape: "interactive users in the last 30 days: 6 of 21 profiles; admin-group members: 2; service accounts with logon rights: 4".
3. Network posture from inside (this is what a sweep cannot see):
   - `ip -4 addr; ip route; ip neigh` / `ipconfig /all; route print`
   - Listeners: `ss -tlnp` / `Get-NetTCPConnection -State Listen | select LocalAddress,LocalPort,OwningProcess` — note which bind to 0.0.0.0 vs loopback
   - Established peers: `ss -tn state established` / `Get-NetTCPConnection -State Established`
   - Local firewall as the host sees it: `nft list ruleset` / `netsh advfirewall show allprofiles state` and the inbound allow rules on management ports.
4. Software and services:
   - Windows: `Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\* | select DisplayName,DisplayVersion`, `Get-Service | ?{$_.Status -eq 'Running'} | select Name,DisplayName,StartName`
   - Linux: `dpkg -l` / `rpm -qa --queryformat '%{NAME} %{VERSION} %{VENDOR}\n'`, `systemctl list-units --type=service --state=running --no-pager`
   - Runtimes that change the blast radius question: `docker ps`, `podman ps`, `/etc/kubernetes/manifests`, kubelet certs by path, `crictl pods`.
   - Register loopback/internal HTTP faces you found as `insert_assets` rows (url + port + version string) so other seats can query them instead of re-reading the host.
5. Patch level as a dated statement, not an adjective:
   - Windows: `systeminfo | findstr /B /C:"OS Name" /C:"OS Version"`, `Get-HotFix | sort InstalledOn -Descending | select -First 10` (or `wmic qfe list brief`)
   - Linux: `cat /etc/os-release; uname -a; rpm -qa --last | head -20` / `apt list --upgradable`
   - Write: "latest applied update KB5039211, 2025-06-10; kernel 5.15.0-91; 214 packages reported upgradable". Mapping a version to a specific CVE is `vulnerability-intel`'s job, not yours.
6. Guards and policy state (the fact that decides what the bench may do on this box):
   - Windows: `Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntiVirusProduct`, `Get-MpComputerStatus`, `GetProcessMitigationPolicy`-level checks you can read, Sysmon presence (`wevtutil el | findstr /i sysmon`)
   - Linux: `ps aux | grep -Ei 'falcon|csagent|elastic-sensor|sensor-|wsds|sophos|mcafee|trellix|carbonblack|clamav'`, `fltmc filters` on Windows, kernel module lists
   - Also record what else watches: `auditctl -l`, `journalctl -u auditd --since "-30m" | tail`, rsyslog/SIEM forwarder targets, Windows event-forwarding config.
   - Sensor present + signature 411 days stale is one fact; sensor absent *and the query succeeded* is another, and a denied query is a third — do not merge them.
7. Reach described by configuration (names and paths, not values): `~/.ssh/authorized_keys` and `known_hosts` (a reach map the identity wrote itself), `/etc/hosts`, `findmnt`/`/etc/fstab`/`mount` (SMB/NFS mounts of other hosts), `net use`, DFS roots, VPN profiles, `http_proxy`/`HTTPS_PROXY` and the PAC path, service config files naming other systems, backup agent config paths. Every peer named here is a candidate row for `internal-mapping`.
8. Persisted state the next seat needs: `schtasks /query /fo LIST /v` (focus on RunAs users), `crontab -l`, `/etc/cron*`, `systemctl list-timers`, startup folders/run keys, saved run-as credentials by target name, browser profile presence (path only), cloud/kube CLI config presence (`~/.aws/credentials`, `~/.kube/config` — existence, never contents).
9. Write it consolidated: one `record_fact` per category (identity, sessions, network, software, patch, guards, reach, persistence) using the batch `facts` array — summary one line, detail the itemised observations, evidence one command + the proving line, confidence observed|inferred, asset_ids [host id]. Eight micro-facts from one host bloats the graph; eight categories with items inside reads like a host.
10. State the capture's own limits: "readable as this identity: 41 of 41 running services, hotfix list denied (Win32_QuickFixEngineering access denied) → patch level NOT ASSESSED".

## Evidence

- A positive observation needs: the command, the output line, the window, and the host asset id. Version and patch claims are quoted strings ("falcon-sensor 7.10, AID present", "KB5034441 installed 2025-05-13"), never "outdated".
- Observed absence vs unmeasured: "SecurityCenter2 queried successfully, 0 registered AV products → no endpoint AV (observed)" is a finding candidate. "Query returned access denied → guard posture NOT ASSESSED" is not a clean bill of health and must not read like one.
- Counts need denominators: "3 of 12 listening sockets on non-loopback addresses", "214 upgradable packages, of which the package manager classified 61 as security updates".
- What counts as a finding: posture alone is inventory; a finding is posture meeting reach or a policy failure with a consequence — an EDR present but its tamper-protection off *on a host reachable from a user segment*, a world-readable mount of a DC share, a service running as a domain admin with an unquoted path, a management plane listening on 0.0.0.0 unauthenticated.

## Report shape

Title: "<host>: <state> — <measured consequence>" (e.g. "SRV01: EDR sensor present, signatures 411 days stale, SMB-reachable from the user VLAN"). `vulnclass` = host-posture / patch-debt / defence-staleness, `severity` from what the posture enables with its number, `asset_ids` = the host, `evidence` = command + quoted output line + the identity used.
