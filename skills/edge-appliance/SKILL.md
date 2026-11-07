---
name: edge-appliance
description: "Invoke this skill when the tasked asset is an edge appliance rather than an ordinary host: VPN concentrator, load balancer, gateway, SD-WAN or out-of-band controller and its management plane. Covers identifying the box, reading its patch posture, proving known-class exposure against this instance, and why every finding here is reported on its own risk page."
---

# Edge Appliance (VPN, load balancer, gateway and management planes)

The box in front of everything. An appliance carries the whole remote workforce, terminates the
credentials of every user, and is the one device that usually has no EDR, no backup restore, and
a change window nobody owns. This skill establishes what the appliance is, how stale it is, and
what it exposes — with read-only proof, because losing VPN is losing the business, not just the
target. Done means: vendor/build identified, patch posture stated, exposure demonstrated from
each vantage the engagement has, and an explicit list of what was deliberately not touched.

## Boundaries

- Authorized use: only the appliance addresses the task names, and only the interfaces the RoE
  allows you to reach. An appliance that answers on an unlisted management IP is a fact for the
  lead, not a target.
- Why this is separated from ordinary hosts, and why the report keeps it that way: one box equals
  every remote session, every load-balanced origin, and often the trust path to the directory;
  these flaws are mass-exploited within weeks of disclosure; and they are patched on a vendor
  window, so "current build" and "exposed build" are different risk statements than for a
  server. Findings here carry the shared-risk framing and go on their own page.
- Hard stops on a live appliance, no exceptions without the operator's explicit yes for that
  device: no reboot, no commit/save/config-change, no firmware upgrade or downgrade, no interface
  flap, no credential or PSK reset, no account creation, no session disconnect, no cluster
  failover. Do not loop the login endpoint (auth-lockout plus CPU on a device that also
  terminates traffic), and do not run exploit code against production — a crashed appliance is an
  incident, not a PoC.
- Not this skill: ordinary service attacks (`infra-exploitation`), which published flaw fits the
  build (`vulnerability-intel`), supplied firmware images (`firmware-iot`), board-level interfaces
  (`hardware-interface-lab`), certificate issuance and inspection roots (`ad-pki-certificate`).

## Procedure

1. Identify before asserting. Read only what the appliance already serves:
   `curl -ksD- -o /dev/null https://<host>/`, the login path, the cookie names, and the TLS
   certificate subject. Markers that actually discriminate: `BIGipServer*` cookie and `/tmui/`,
   `/xui/` (F5); `/sslvpn/` or `/dsEnterprise/` and `S_SESS*` cookies (Pulse/Pulse Secure);
   `/remote/login` and `/sslvpn/` portals plus `FORTINET*` cert OU (FortiGate);
   `/global-protect/login.esp` and `prelogin-cookie` (Palo GlobalProtect); `/portal/` and
   `checkCI`/`SESS_ID` (Check Point); `nfactor`/`CitrixGatewayLogin` (Citrix); `/screens/` and
   `sw.js` (Aruba); `/ui/` plus serial in banner (many controllers). Record vendor/model/build as
   a fact; state the evidence line you read it from, not the guess.
2. Patch posture, read-only: build string from the login page or cert, serial or hostname
   disclosure, and any version endpoint that answers without auth. Compare against the vendor
   advisory for that line — that comparison is `vulnerability-intel`'s job; here you hand it the
   build and take back the candidate class. Record how far behind current the build is as a
   number (one build, N releases behind, M months).
3. Map exposure from each vantage you legitimately have — this is the appliance finding. For the
   same interface, record: reachable from the internet side, reachable from an internal user
   VLAN, reachable only from the dedicated OOB network. `list_assets` for both vantages; a
   management console (`/tmui/`, `/ui/`, `/cgi-bin/`, vendor admin ports 8443/4443/8003) that
   answers on the WAN-facing interface is exposure you can report without any exploit.
4. Unauthenticated-surface pass, one request per endpoint, no enumeration loops: the login page
   and its static assets, the portal/agent download paths, a known config/status path, and the
   appliance's own API path. A page that returns data without a session is the finding; a page
   that returns 401 is the control working, and both go in the coverage table.
5. Config and key material exposure: if a backup/config endpoint or a leaked file
   (`config.xml`, a support bundle, an exported PSK) is reachable without auth, fetch it once,
   confirm the format, and register every secret inside via `save_task_credential` by reference.
   Never quote a PSK, hash, private key or token in a fact or the report; cite the credential ID
   and state which field type it was.
6. Crypto and tunnel posture: `openssl s_client -connect <host>:443` for chain, self-signed
   reuse across interfaces, and whether the same certificate is used on the user-facing and
   management plane; `nmap --script ssl-enum-ciphers -p443 <host>` for downgraded suites. For
   IKE/ESP: one `ike-scan <host>` pass to read the proposal set and the peer ID. Never dictionary
   a PSK against a live concentrator.
7. Session handling, only against the test identity the task issued: does the portal put the
   credential or a session token in a query string or `Referer`-leaking redirect, is the group or
   role held client-side where tampering changes reach, does logout invalidate the cookie. One
   login, one request pair, evidence captured, done — no session-hammering, no concurrent-user
   load.
8. Known-class exposure, the honest version: a build matching a mass-exploited class is
   inventory until you demonstrate the effect against this instance through a read-only
   primitive (an unauthenticated endpoint that returns appliance data, a reachability proof on an
   internal-only path). If the demonstration needs exploit traffic, state the class, the
   precondition you verified, and that verification is pending a lab or an operator-approved
   window; do not fire it blind.
9. Close-out facts: interfaces you did not touch, tests you skipped and why, and the exact device
   state you left it in ("no writes, no commits, no reboots"). The lead needs that sentence as
   much as the finding.

## Evidence

Positive: vendor/model/build plus the line you read it from, the interface and port, the vantage
it was reachable from, the request/response pair (or command and output), the exposure or effect,
and an explicit "device state unchanged" statement with the timestamp window. Counts over
endpoints tried: "1 of 6 unauthenticated paths returned appliance configuration data".

Negative: "Build <read from X> checked against the vendor advisory line; no unauthenticated
endpoint beyond the portal reachable from either vantage across the 14 paths attempted in
window Z; exploit-class behaviour NOT ASSESSED — no live-appliance exploit traffic permitted".
An untested control is never written as safe.

## Report shape

"Edge appliance <vendor model, build> at <host> exposes <interface/endpoint> from <vantage>,
without <control>" — fields: device role and who depends on it, build and how many releases
behind, vantage matrix, the proving request, credential-store references for any material in a
config, blast radius statement (all remote users / all balanced origins), device state left
unchanged, and the remediation window the owner actually controls.
