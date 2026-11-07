---
name: mitigation-status
description: "Invoke this skill when the question is whether a platform control actually stops a technique — OS hardening, a sensor, an admission policy, a provider SCP, or browser and runtime protection — and the answer must be a dated measurement on the build in front of you rather than a recollection."
---

# Mitigation Status (does the control stop the technique, on this build)

Mitigation knowledge expires: vendors flip defaults, revoke features, backport fixes and rename behaviour without publishing a CVE. This skill converts "I believe X blocks Y" into a measured verdict on a named build with a check date, plus the condition that would invalidate it.

Done means: the technique stated as a testable claim, the control identified and versioned, a lab measurement recorded, and a verdict class chosen.

## Boundaries

- Measure on a lab instance of the same build, or read configuration on the authorized asset. **Never probe, blind or disable a production sensor** — no countermeasure testing against the customer's EDR console, no stopping services, no tampering with policy delivery.
- Detection-evasion mechanics belong to `edr-bypass-re`: lab detonation only, and its mechanism mapping feeds this skill's verdict rather than the reverse.
- Nothing here proposes changing this platform's own guard, intercept or scope rules; those are the operator's surface.
- An unmeasured assumption is not a verdict. If the control's state cannot be verified on the build in front of you, the answer is "unknown on this build" plus what is needed to resolve it.
- Not this skill:
  - whether the underlying flaw exists — `fuzzing`, `patch-diff`, `vulnerability-intel`;
  - whether the technique works when the control is absent — `exploit-development`;
  - how the sensor is implemented — `edr-bypass-re`;
  - the container and cluster boundary itself — `container-escape`, `cloud-attack`.

## Procedure

1. **Write the claim as a testable proposition**: "<control> prevents <technique> on <product> <build> under <configuration>". "ASLR makes heap addresses unpredictable" is not testable; "on this allocator and build the chunk address varies by fewer than 6 bits across 200 launches" is.
2. **Fingerprint the control** and record the check date beside every field:
   - OS and build: `uname -a`, `/etc/os-release`, `cmd /c ver`, `sw_vers`;
   - package and patch level: `dpkg -l`, `rpm -qa --last`, the vendor patch bundle, distro errata status for the specific identifier;
   - sensor or policy version: agent version, signature or policy revision, policy last-applied timestamp;
   - configuration provenance: distribution default, GPO or MDM, kubelet or admission configuration, provider SCP, managed profile. Cite which one set the value.
3. **Apply the dated-knowledge rule.** Never assert mitigation behaviour from memory. Two acceptable bases: a dated vendor document naming the exact version, or a local measurement performed on this build. Record which one, with the date and version checked.
4. **Re-verify triggers** — any of these voids an earlier verdict, so re-run step 5 rather than reusing the number:
   - the build or version differs from the one the earlier check named;
   - the vendor shipped an update, or changed a default between releases;
   - the check is older than the engagement's intel TTL (the clock `vulnerability-intel` maintains), or predates a configuration change on the asset;
   - the verdict is about to move a severity, close a finding, or justify skipping a test;
   - two sources disagree about the control's behaviour.
5. **Measure the setting**, using the observable that answers the question: `sysctl kernel.randomize_va_space`, `/proc/sys/kernel/yama/ptrace_scope`, `getenforce` and `sestatus`, `aa-status`, `lscpu` flags for CET, shadow stack and IBRS classes, `Get-CimInstance -ClassName Win32_DeviceGuard`, `Get-MpComputerStatus`, the registry key behind the policy, `kubectl get pod -o jsonpath` for the effective securityContext, an admission webhook's `namespaceSelector` and `objectSelector` scope, the provider's effective SCP or org-policy list, a storage public-access state.
6. **Then measure that it is enforced.** The common gap is present but not applied, and it is the finding: a kernel parameter set globally but overridden for one namespace; a seccomp profile declared in a template the pod does not use; a GPO linked but never refreshed; an admission controller installed in audit mode; a policy scoped to one resource class while the technique uses another.
7. **Detonate in the lab against the same build.** Run the technique's minimal form for the planned number of attempts, record the outcome per variant, and capture both the control's reaction (blocked, errored, logged, alerted, allowed) and the failure text. Where the technique crosses into telemetry impairment, keep the measurement to detection coverage on your own lab sensor and write it as a blind-spot finding about the control.
8. **Choose exactly one verdict class per technique/control pair, each with its evidence:**
   - BLOCKS — mechanism fails; attempts and error text recorded;
   - PARTIALLY BLOCKS — blocks variant A, permits variant B; name the boundary condition;
   - DOES NOT BLOCK — succeeds here, with the scope gap identified (control exists, its coverage excludes this path);
   - NOT EFFECTIVE HERE — control absent, disabled, or wrong version on this build;
   - NOT ASSESSED — could not measure; state the missing access, host or permission.
9. **Keep mitigation distinct from fix, and lab state distinct from fleet state.** A hardened default is not a root-cause fix (same warning `patch-diff` carries); a control on in the lab can be off in production. Where the fingerprint scan allows it, give the denominator — "on for 31 of 44 hosts" — and say how you counted.
10. **Date and version-stamp every fact you record**, so a later reader knows what needs re-checking: control, product, build, configuration source, check date, measurement method, verdict.

## Evidence

Positive: the control's exact version and configuration quoted from the observable; the check date; the lab detonation command; `attempts/variants` with per-variant outcome; the enforcement-scope proof (which selector or profile actually applies); the fleet denominator where available.

Negative: "checked AppArmor on Ubuntu 24.04.1, kernel 6.8.0-49, profile `docker-default` enforced (measured 2026-05-14): technique X blocked 20/20 attempts with `Operation not permitted` from the profile's deny rule on the target path. Variant Y NOT ASSESSED — it needs the `ptrace` rule class this profile does not constrain."

## Report shape

`Mitigation status: <control> vs <technique> on <product> <build> (checked <date>) — <verdict class>`

Fields: control version and configuration source, measurement command, per-variant outcomes with counts, enforcement scope, fleet denominator, re-verify trigger, and the finding this verdict modifies.
