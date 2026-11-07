---
name: secrets-keymanagement
description: "Invoke this skill when the objective is the credential and cryptographic material itself: vaults, KMS/HSM keys, secret managers, SSH keys and TLS or signing certificates, plus the rotation and blast-radius questions that decide whether a leaked key matters."
---

# Secrets and Key Management (vaults, KMS/HSM, keys, certificates)

Secrets are the material that decrypts the rest of the engagement, so the deliverable is a capability statement, not a blob: which principal can use which key, how many consumers that reaches, when it was last rotated.

Done means each discovered secret is stored by reference, identified by public attributes only, and dispositioned by a use test that never required the value to appear in prose.

## Boundaries

- Work only the tenants, hosts and vault namespaces named in the engagement. A key policy naming another account is a finding about the policy, not a licence to act in that account.
- **No key material in prose**, ever: not in a fact, finding, report, artifact filename or command echo. Save with `save_task_credential` and cite the returned ID.
- Identify material by public attributes only: key ID or ARN, certificate serial plus SHA-256 fingerprint, SSH public-key fingerprint, secret path plus version number, credential last-4.
- Non-extractive defaults: no `CreateKey`, `ScheduleKeyDeletion`, key-policy edit, vault unseal, root-token generation, or destructive rotation against production. Rotation and breakage questions get answered by reading configuration, or by asking the operator.
- Not this skill:
  - the identity that reads the secret — `cloud-attack`;
  - buckets and volumes where credential files sit — `cloud-storage-data`;
  - credentials baked into images — `container-escape`; pipeline variables and runner credentials — `cicd-supply-chain`;
  - the application's own crypto implementation — `codeaudit`.

## Procedure

1. **Enumerate hiding places before testing any of them.**
   - Provider stores: Secrets Manager secret list and version IDs, SSM Parameter Store `SecureString` paths, GCP Secret Manager including old and destroyed versions, Azure Key Vault secrets/keys/certificates.
   - Host: process environments (`/proc/1/environ`, `/proc/<pid>/environ`, values redirected into a file rather than the transcript), mounted secret files and `emptyDir` leftovers, shell histories, systemd `EnvironmentFile`.
   - User configuration: `~/.aws/credentials`, `~/.config/gcloud/`, `~/.kube/config`, `~/.docker/config.json`, `~/.npmrc`, `~/.git-credentials`, PKCS#12 and JKS bundles.
   - Build and infra artifacts: `terraform.tfstate` and `*.tfvars`, cloud-init user-data, application `.env`, key material inside backup archives and `/tmp`.
   - Record per hit: location, identifier, owner, file mode. Do not print the value.
2. **Classify by capability type.** The class picks the proof in step 4 and the radius in step 5:
   - bearer — authenticates as whoever holds it;
   - decryptable — opens data;
   - signing — forges tokens, artefacts or TLS;
   - indirect — a KMS or Vault permission that lets the service use the key on your behalf.
3. **Answer the rotation question per credential.**
   - Access keys: `aws iam get-access-key-last-used`, plus the credentials report for age and services touched.
   - Secret managers: rotation schedule or rotation function attached, or absent.
   - Vault: token TTL versus max TTL, lease type, dynamic versus static backend, an unbounded TTL on a KV mount.
   - Certificates: `openssl x509 -noout -dates -serial -subject -issuer -fingerprint -sha256`, plus wildcard reuse across tenants and an internal CA issuing without name constraints.
   State age, rotation mechanism and reachability separately; "long-lived, never rotated, broadly readable" is the severity multiplier.
4. **Prove capability with the smallest self-contained test**, in this preference order:
   - Generate your own ciphertext, then call `kms:Decrypt` / `Sign` / `Encrypt` on it: proves the permission, touches no tenant data.
   - For a signing key, stand up a verification service in the lab and validate a token you minted against it. Do not submit forged tokens to the live application unless the engagement names that action.
   - For a decryptable data key, decrypt exactly one sample where authorized, take the population count from metadata (denominators come from `cloud-storage-data`), and stop.
   - For a bearer token, call one read-only authenticated endpoint and record the status code.
   - Resolve key/cert pairing by modulus-hash comparison (`openssl rsa -noout -modulus` versus `openssl x509 -noout -modulus`, both piped to `sha256sum`) — never by printing either artifact.
5. **Compute blast radius from policy text, not intuition.** KMS key-policy principals and grants, secret resource policies, Vault ACL policies covering the path (`vault policy list` plus a capabilities probe on the path you hold), wildcard aliases, `kms:Decrypt` without `kms:Encrypt` asymmetry, cross-account grant lists. Then count consumers: grep config and source for the secret's name or ARN to see how many services embed it. High consumer count plus impossible rotation is itself reportable.
6. **Non-extractable designs change the win condition.** With an HSM or Vault Transit you test *use*, not export: auth requirements on the sign and encrypt endpoints, key version and `exportable` flag, min-encryption-threshold state, whether unseal shares or the root credential live on the same host, in the same bucket, or in the same CI variable set as the thing they protect (that colocation is the finding), and unbounded decrypt or sign rates.
7. **Disposition every secret** from step 1: not reachable (permission denied, cited), reachable but not exploitable (reason), capability proven (test from step 4, identifiers recorded), or operator-material-only (handed up).
8. **Teardown.** Any key, token, AppRole or lease you created gets an ID in a fact and one removal line for the operator. Never leave a created credential behind "for the next phase".

## Evidence

Positive: location; identifier (ID/ARN/path/version/fingerprint only); capability class; the exact proving call and response code; reader and consumer counts; rotation age; the credential-store reference ID.

Negative: "enumerated 3 provider secret stores and 2 hosts' process environments: 14 secret identifiers found, 0 readable by the current principal (`AccessDeniedException` on 14 of 14 `GetSecretValue`; host files mode 0600, root-owned on both). Rotation age NOT ASSESSED — `iam:GenerateCredentialsReport` denied."

## Report shape

`<system>: <secret class> <identifier by ID/fingerprint> usable by <principal> — <n> consumers, last rotated <date | never>`

Fields: store or host, path or ARN, capability class, proving call, reader and consumer counts, credential reference ID, rotation recommendation.
