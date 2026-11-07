---
name: device-trust-mdm
description: "Invoke this skill when a task turns on endpoint trust as an authorisation signal: MDM/Intune enrolment and compliance state, hybrid and workplace join, conditional-access policy gaps, passkey and device-bound credential flows."
---

# Device trust, compliance and conditional access

"Trusted device" is an assertion some policy makes, not a property of the machine. Done means
naming the signal a decision actually consumed (join state, compliance flag, certificate,
PRT, TPM/passkey), showing whether the platform *verified* it or merely accepted it, and
quantifying what a credential alone gets without any device at all.

## Boundaries

- Enrolment, compliance and trust tests run against your own account and your own or lab
  devices, and the report says which device class you used.
- Re-enrolling, wiping, quarantine-placing or revoking a real employee's managed device is an
  availability action against that person: approval first, or use a lab twin.
- MFA and phishing-resistant authentication outcomes are reported as control effectiveness
  with numbers, never as a how-to-defeat-it recipe. Do not change tenant or device policy to
  make a test pass.
- Not this skill: on-prem certificate templates and enrollment rights → `ad-pki-certificate`;
  directory rights and escalation edges → `ad-attack`; captured tokens and their reach →
  `credential-handling`; cloud IAM paths → `cloud-attack`; wired/802.1X posture →
  `nac-8021x`; endpoint telemetry and its blind spots → `edr-bypass-re`, `telemetry-sources`.

## Procedure

1. **Identify the trust model on the endpoint** (only on a host you own or the RoE names).
   Windows: `dsregcmd /status` → `AzureAdJoined` / `EnterpriseJoined` / `NgcSet` / device ID
   and certificate thumbprints; enrolment authority from
   `reg query HKLM\SOFTWARE\Microsoft\Enrollments /s` (provider, UPN, enrolment ID);
   `Get-ChildItem Cert:\LocalMachine\My` for the workplace-join and machine certificates;
   TPM/vBS presence from `dsregcmd`/`tpm.msc` output. Record which of these a remote attacker
   could even observe.
2. **Read the device objects the directory publishes.** In AD, `msDS-Device` objects carry
   `isManaged`, `isCompliant`, `msDS-DeviceOwner`, `msDS-DeviceId`, `msDS-KeyCredentialLink`;
   query bounded: `ldapsearch -Y GSSAPI -b "<defaultNamingContext>" "(objectClass=msDS-Device)" dn isManaged isCompliant`
   (`-E paging=200`, output to the run dir). This is how you tell whether compliance is
   *computed and stored* versus *checked at access time*.
3. **Map the enrolment path a normal user can drive.** Test, with your own account:
   workplace-join/self-service enrolment reachability (`enterpriseregistration.<domain>` over
   443/80, `curl -sSI`), Intune enrolment eligibility, Autopilot or bulk-token enrolment
   leftovers, and whether enrolling requires an approval, a platform SSO extension, or a
   compliance check that can be satisfied by a device you control. The finding here is the
   number of trust-gating signals an unprivileged user can mint themself.
4. **Prove what conditional access actually gates.** From a compliant-state device you
   control and from a non-enrolled device, run the same sign-in to the same application and
   compare: token issued or challenge presented, which claims landed in the token
   (`wam`/`msal` cache decode, or `jq` the JWT body for `sub`, `amr`, `xms_ce`/`unique_key`,
   `urn:microsoft:claims`). Never inline a token into a fact — claims only, and the raw token
   registered via `save_task_credential`.
5. **Enumerate the gap classes explicitly**, per application: legacy/protocol-excluded
   clients, workload identities and service principals, guest and external-user flows,
   apps not in the policy's coverage, terms-of-use-only policies, "mark compliant if
   enrolled" platform SSO shortcuts, filter rules that exclude a whole device class, and
   policies that gate on platform or location instead of compliance.
6. **Check device-bound credentials as a trust signal.** Passkey/FIDO2 and key-trust flows:
   whether attestation is enforced, whether a self-created authenticator with no attestation
   is accepted, and whether the second factor is bound to the device or merely registered
   from it. Test with your own account and your own key material.
7. **On-prem trust variants.** Hybrid join computed trust, certificate-based machine trust
   issued without compliance (that route belongs to `ad-pki-certificate` and `privilege-relay`
   for the auth exercise), and NAC/posture checks that only look at a certificate rather than
   at compliance state.
8. **Report the reach, not the drama.** Each conclusion is: signal consumed, verified-or-not,
   which app/protocol, and what an attacker holding only the credential gets.

## Evidence

- Positive: device class used (own/lab/which join state), the exact claims or attributes the
  decision accepted, the comparison attempts with counts ("4 sign-ins: 2 from enrolled
  compliant device, 2 from unenrolled; app X issued a token to all 4; policy Y challenged 2"),
  the attribute or policy ID that is the gap, and the remediation line.
- Negative: "Compliance verified at access time for the 2 applications in scope — unenrolled
  device denied with the enrolment-required error on 2/2 attempts." Absence of policy
  visibility is stated: "CA policy set NOT ASSESSED — the engagement account has no
  `Policy.Read.All`, so conclusions come from observed authentication behaviour only, over
  <window>."
- Anything destructive you declined to do (revoke, wipe, re-enrol a staff device) is recorded
  as a deliberate limitation.

## Report shape

`Conditional access / device trust: <app or resource> admits <credential class> from <device state> because <signal unverified or excluded>`
Fields: tenant/domain, policy or gap class, affected app(s) + protocol, device state tested,
verification verdict, attempt counts with denominator, artefacts needing approval, remediation.
