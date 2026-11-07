---
name: telemetry-sources
description: "Invoke this skill when the question is what the sensor estate can produce at all: which agent or provider emits which event, whether it is enabled, collected and ingested, what it costs to turn on the ones that were off, and the gap statement when a source cannot be turned on."
---

# Telemetry Sources (the sensor estate and its fidelity)

Before anyone asks "was it detected", the stack has to be capable of saying anything. This subject
maps which sensor produces which event class, per platform and per segment, and separates three verdicts
with different fixes and different owners: no sensor produces it, the sensor produces it but policy does
not collect it, it is collected but never reaches the SIEM. Done means a source-by-source statement with
coverage, lag, retention and field completeness — and a named gap for every class the estate cannot
produce.

## Boundaries

- Authorized use: the estate is mapped from client documentation, policy state, licence tier and
  permitted read access. Sensor presence is a fingerprinting question. Nothing here disables, uninstalls,
  pauses, reconfigures or tampers with an endpoint agent, log forwarder or cloud audit trail — that is an
  availability action against the client's own protection.
- Enabling a source is a recommendation the client performs, in their change process, on their window and
  their licences. This seat names the switch, the cost and the owner; it never flips it.
- Verdicts stay per platform and per segment; a partial sample is labelled as one. One host's policy state
  does not describe an estate.
- Not this skill: the paired action-by-action verdict (`detection-coverage`), analytic authoring
  (`detection-engineering`), endpoint telemetry RE (`edr-bypass-re`), OT/process-network constraints
  (`industrial-safety-bounds`), the client's own SOC response (`engagement-deconfliction`).

## Procedure

Do not go to step 4 until 2 and 3 have separated the three verdicts for every source.

1. **Inventory the estate.** From client documentation, the assets already registered (`list_assets`),
   the policy state they share, and permitted recon, enumerate every producer: endpoint agents,
   forwarders, cloud trails, identity logs, network sensors, and the SIEM or archive each lands in.
   Register each as an asset or fact with the platform and version it runs on. A licence purchased is
   not an agent deployed — verify presence before crediting coverage.
2. **Map each source to the event classes it emits, and to where each lands** (index or dataset) and the
   retention that applies. Working frame, per category — list only what the estate actually has:
   - **Endpoint (Windows)**: EDR process/network/file/image-load/registry events; Sysmon 1, 3, 7, 11,
     12-14, 17-18, 22, 25; Security 4688 (command-line text only if that policy is on), 4672,
     4624/4625, 4720/4726, 4728/4732/4756, 5140/5145, 4698/4702, 7045/4697, 1102, 4663; PowerShell
     script-block logging 4104.
   - **Endpoint (Linux/macOS)**: `auditd` syscall rules (`execve`, `openat`, privilege change), eBPF
     agent equivalents, journald/rsyslog auth and sudo events, cron and systemd-unit writes, macOS
     unified log plus endpoint-security-extension events.
   - **Network**: NetFlow/IPFIX (5-tuple, bytes, timestamps, sampling rate), resolver and forwarder DNS
     query logs, firewall/NAT session logs, proxy egress (URL, identity, action, bytes), IDS/IPS alerts
     with signature-set version, packet-capture spans and retention, VPN auth and split-tunnel detail,
     TLS-inspection presence and exemptions.
   - **Cloud**: control-plane audit (CloudTrail org trail, multi-region, S3/Lambda data events; Azure
     Activity; GCP Cloud Audit Logs with an org sink), object-store access logs, VPC/VNet flow logs with
     aggregation interval, WAF/CDN logs (rule ID, action, real client IP behind the CDN), container
     runtime and Kubernetes API audit, workload identity telemetry, IAM change logs.
   - **Identity and collaboration**: DC security logs (4768/4769 with pre-auth and encryption failure
     codes, 4776, 4771, 4624 type 3/10, 5136, 4662/4670, 4928-4932 ADCS), certificate-services audit
     (template, requester, SAN upn), Entra/AzureAD sign-in logs (Conditional Access result, MFA method,
     device state, risk detections), ADFS/SAML/OIDC gateway logs, LDAP query logging (off by default —
     name the diagnostics level needed), PAM/vault session and checkout logs, MDM compliance and
     enrolment events, mail transport tracking, admin audit, journaling, OAuth-grant audit.
3. **Sort every event class into exactly one of three verdicts and never blur them:**
   - **not covered by a sensor** — no product in the estate emits this class at all.
   - **covered but not collected** — the platform can produce it but policy or a default is off:
     command-line auditing, script-block logging, Sysmon absent, DNS server logging, LDAP diagnostics at
     None, Kerberos failure subcategories unaudited, flow sampling at 1:4096, S3 access logging off,
     CloudTrail not org-wide, Entra sign-in logs kept under 30 days without a Log Analytics workspace,
     unlicensed or container hosts with no agent.
   - **collected but not ingested** — the event exists in the platform but has no pipeline, no index,
     no retention long enough to search, or an ingest lag longer than the window a detection would
     need.
   Reporting the third as the first sends the client to buy something they already own; reporting the
   second as the first hides a free configuration fix.
4. **For each "off" source, write the enablement ask with its cost.** Name the exact switch and where it
   lives: the GPO "Include command line in process creation events", `auditpol /set` subcategory names,
   the Sysmon install and ruleset, PowerShell script-block logging, DNS server logging, AD CS audit
   level, CloudTrail org trail plus multi-region plus S3 data events, VPC Flow Logs, GCP org-level sink,
   Entra diagnostic settings into Log Analytics, WAF log delivery, MDM compliance export scheduling. Per
   switch state: licence implication, ingest volume added (with its basis), performance cost, whether a
   maintenance window or reboot is needed, the owner, and the rollback path. The client performs every
   change on this list — no agent output flips a policy.
5. **For each source that cannot be turned on, write the gap statement**: source, the event class the
   engagement needed, why it is unavailable (platform or OS version lacks the event, licence tier does
   not include it, data-residency or works-council restriction, the vendor's product does not emit it,
   a third-party SaaS the client cannot log, an OT or process-network segment where a probe is not
   permitted), who decides, what compensates (network-level, identity-level, or a control change rather
   than a log), and the residual blind spot expressed in the terms `detection-coverage` uses. A gap
   statement is a deliverable; a silent blank is not.
6. **Record fidelity facts per source, measured where possible**: ingest lag from event occurrence to
   queryable in the SIEM (ask the defender for the timestamp of a known benign event and compare),
   retention length, sampling rate, field completeness (which of the named fields are actually
   populated versus empty), timezone and clock source, and host or segment coverage as a fraction of
   the estate with the denominator and the sample basis stated.
7. **Close the loop**: every `not instrumented` verdict in `detection-coverage` cites a row here, and every
   analytic in `detection-engineering` that depends on a source listed as off ships with that dependency
   named. Where a whole tactic has no source, that is a platform-capability finding — a fact for the
   advisory front, not a detection task.

## Evidence

Positive: per source, a dated statement naming the producer and version, the event classes it emits,
the verdict from step 3 with the config, policy or documentation pointer that proves it, coverage
(count of hosts or segments reporting over the count that should), retention, sampling, and measured
ingest lag where obtained. Partial reads are labelled partial with the sample basis.

Negative and honest forms: "no DNS query telemetry exists for segment S: verified against the
documented resolver logging policy and the defender's confirmation on <date>; therefore DNS-based
detection is impossible in that segment and every technique needing it is recorded as not
instrumented"; "agent coverage unknown — deployment inventory not provided; 3 of 3 sampled hosts
reported the agent, estate unverified"; "event class X present in platform but no pipeline into SIEM:
gap reported to owner Y". Never report an absence of events as an absence of activity, and never
infer a source is enabled because the licence exists.

## Report shape

Row pattern: `<source> (<producer> <version>, <platform/segment>) — event classes <list> — verdict
<not covered | not collected | not ingested | enabled> — coverage <n>/<denominator> — lag <t> —
retention <d>`.
Fields per row: producer, event classes, verdict, proving pointer (policy/config/doc), enablement
switch and cost if off, owner, gap statement if it cannot be turned on, compensating source, and the
`detection-coverage` rows it explains.
