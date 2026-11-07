---
name: credential-handling
description: "Invoke this skill whenever captured secret material appears — passwords, hashes, tickets, API keys, tokens, private keys, certificates, cookies: registering it by reference, proving what it reaches with the smallest action, and keeping the value out of facts, findings and reports."
---

# Credential material: store by reference, prove reach, never export

A credential is a question about reach, not a trophy. Done means: the material is registered
with its provenance, one minimal action shows what it can do, and everything leaving the
platform identifies it by store ID and type — no live value in any fact, finding, report,
filename or command line.

## Boundaries

- Live values are typed into exactly two places: the `save_task_credential` call, and the one
  command that consumes them (via an environment variable, never argv). Everything else refers
  to the credential by its returned ID.
- Nothing leaves the platform. No pasting into an external service, no committing, no
  emailing, no sending a captured secret to a host the target suggested.
- Offline cracking is out of bounds unless the RoE grants it in writing. Log-in demonstrations
  use accounts the engagement owns or gives you.
- Not this skill: which rights a directory credential buys → `ad-attack`; how a session is
  carried to another host → `lateral-movement`; relay/coercion of live authentication →
  `privilege-relay`; where a foothold's secrets physically live and how to read them →
  `credential-access` work under `post-exploitation`; vaults/KMS/HSM design →
  `secrets-keymanagement`; retention and destruction of the artefact files →
  `evidence-custody`; teardown of an account you created → `persistence-cleanup`.

## Procedure

1. **Intake before inventory.** The moment material appears, `save_task_credential`
   (`domain`, `username`, `password` = the value or its canonical form, `note` = provenance:
   source host, file or process path, capture method, UTC timestamp, engagement principal who
   found it). Same domain+username is updated in place — re-record rather than duplicate, and
   update the row if the password rotates. Note whether it is your own test secret or a real
   user's: that changes what you may do with it.
2. **Keep the bulk out of prose.** Long objects (ccache, pfx, private key, `cookies.txt`,
   `NTDS.dit` slice) stay as files in the task's work dir, mode 0600, named by hash not by
   content (`<sha256-8>.ccache`). Register the *path + sha256 + type* in the credential note;
   never name a file after the secret. Delete the file when the proof is done and record the
   deletion in the same ledger row.
3. **Use a value without exposing it.** Read the row with `list_task_credentials`, then in an
   interactive session (`shell_open`) load it into an environment variable from a file:
   `read -rs SECRET < ./creds/<hash>.txt` — commands then reference `"$SECRET"`. Argument lists
   are visible in `ps`, shell traces and endpoint logs; heredocs and echoed values are captured
   in the traffic record. If a tool insists on argv, run it once and say so in the fact.
4. **Prove reach with the smallest action per class.**

| Material | Registration | Minimal reach proof |
|---|---|---|
| Password | username + value | one authenticated identity call (`netexec smb <host> -u <user>`, `kinit` + `klist`) |
| NTLM hash | value in store, `note` marks `ntlm` | pass-the-hash auth verdict, no execution |
| Kerberos ticket / TGT | converted ccache file by hash | `klist` + `kvno cifs/<host>` (KDC-side proof, target untouched) |
| Bearer / refresh token | value + issuer in `note` | one `/userinfo`-style call returning the identity and scope list |
| API key | service + key | one read-scope call, record the permission set it grants |
| SSH private key | key file by hash | `ssh -o BatchMode=yes <host> true` |
| Certificate + pfx | password in store, serial in fact | authenticate as your own account, record identity |
| Cookie / session | value in store | replay once against the owning endpoint, record returned identity |

   The finding is the identity, the permission set and the count of hosts or endpoints that
   accept it — not the value.
5. **Quantify instead of dumping.** For a class of material (e.g. "credentials in this config
   tree"), report counts by type and what each reaches: "9 secrets of types 4; 3 authenticate
   to in-scope services; 0 are reused outside the host." Never a table of plaintext.
6. **Cracking, where the RoE allows it.** Only against engagement-owned or your own captured
   material, in the work dir, with the algorithm/class reported rather than the recovered
   string; `save_task_credential` updates the row with the recovered value so it is tracked.
   Where it is not allowed, report the *strength statement* the capture implies (algorithm,
   key length, online-verifiability) and move on.
7. **Close the loop.** Every credential gets a disposition before wrap-up: rotated, disabled,
   or left with an explicit "client must rotate" line naming the account (never the secret).
   Anything you created — service account, key pair, SPN, certificate — is also a
   `persistence-cleanup` ledger row with its removal verified.

## Evidence

- Positive: credential ID + type + provenance line, the one command that consumed it, the
  identity/permission set returned, denominators ("tested against 5 in-scope hosts: 3 accept,
  2 refuse; 1 of 3 reaches Tier-0"), and the statement that no value was reproduced in graph
  text or output files.
- Negative: "Credential ID 12 (service account, captured from <path>): 0 authentications across
  14 in-scope SMB endpoints (`netexec`, <window>) — no reach, closed." Non-testing is also an
  evidence line: "Offline cracking NOT PERFORMED — RoE does not grant it; reported as NTLM
  verifiability against a live DC only."
- A redaction note belongs on every finding that names an account: what was withheld and why.

## Report shape

`Captured <credential type> for <account/service> (registered, ID <n>) authenticates to <n> of <m> in-scope <systems>, granting <capability>`
Fields: type, provenance (host/path/method/time), proof call used, accept counts, worst-case
principal reached, disposition (rotated/disabled/client-must-rotate), explicit
"secret value withheld — available in the credential store" line.
