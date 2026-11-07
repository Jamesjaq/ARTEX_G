---
name: ad-pki-certificate
description: "Invoke this skill when a task touches Active Directory certificate services: locating CAs and enrollment endpoints, enumerating template ACLs and enrollment rights, classifying ESC-style template misconfigurations, and proving a certificate-based escalation path without loading the CA."
---

# AD CS certificate templates and enrollment rights

A CA vouches for whoever the template lets it vouch for. Done means: the template class and
the exact attribute bytes that put it in that class, the principal who holds the right, what
the issued certificate authenticates as, and one issued-then-revocable proof against your own
principal — the CA left otherwise untouched.

## Boundaries

- In-scope domain and its own PKI only. Enumerate read-only first; a certificate request is
  a state change on a service every host in the forest trusts.
- One request per template under test, for a subject you own. Minting a certificate whose
  SAN is a real Tier-0 account, then using it to authenticate, is an incident — the lead
  decides whether that demonstration is warranted, and it is not this skill's call.
- Not this skill: relay against the WebSync/`/certsrv` endpoint (ESC8) → `privilege-relay`;
  ACL topology and the rest of the escalation graph → `ad-attack`; the pfx/key material and
  its reach → `credential-handling`; vaults, KMS/HSM and on-prem certificate stores →
  `secrets-keymanagement`; revocation/teardown bookkeeping → `persistence-cleanup`.
- Do not touch CA private keys, `certmap.rul`, CA database files, or CA policy to make a
  test work. Missing binary (certipy/OpenSSL tooling) is a provisioning gap, not a licence
  to hand-roll CRIM/MSCEP.

## Procedure

1. **Locate the PKI.** From the configuration partition:
   `ldapsearch -Y GSSAPI -b "CN=Configuration,<DC>" "(objectClass=certificationAuthority)" cn dNSHostName cNCName cATrustAuthorities cACertificateChain`.
   Then endpoints: `nmap -Pn -p 445,636,443,50443 -sV <ca>`, and
   `curl -sk -o /dev/null -w '%{http_code}\n' https://<ca>/certsrv/`. `insert_assets` the CA
   host and each enrollment endpoint you can actually reach.
2. **Enumerate templates read-only.**
   `ldapsearch -Y GSSAPI -b "CN=Certificate Templates,CN=Configuration,<DC>" "(objectClass=pKICertificateTemplate)" cn pKICertificateTemplateId pKIExtnKeyUsage pKIKeyUsage flags "msPKI-Certificate-Name-Flag" "msPKI-RA-Signature" "msPKI-Template-Schema-Version" nTSecurityDescriptor`
   → write the LDIF to the run dir, count the total. `certipy template list` gives the same
   picture decoded; either way the denominator is the number of templates in the forest, not
   the number you read.
3. **Decode what the attributes mean.** `pKIExtnKeyUsage` is DER — parse with
   `openssl asn1parse` or the collector's decoder; `flags` and
   `msPKI-Certificate-Name-Flag` are bitfields (ENROLLEE_SUPPLIES_SUBJECT = 0x1);
   `msPKI-RA-Signature` = manager approval required. For the ACL, resolve who holds the
   enrollment extended right `0e10c9c8-82a8-4a8b-bbf6-33a92f682f1f`, the auto-enrollment
   right `ce10c9c8-...` pair, and any Write/WriteDacl over the template object.
4. **Classify each template against the table below.** One row per class; the anchor column
   is what must be visible in output before you name the class.

| Class | Anchor attribute / setting | What it buys |
|---|---|---|
| Subject forgery (ESC1) | ENROLLEE_SUPPLIES_SUBJECT set + auth EKU + no manager approval | enroll as any UPN |
| Any/overlapping EKU (ESC2) | EKU `anyExtendedKeyUsage` or OK-as-auth EKU | usable for auth if CA policy allows |
| Enrollment agent (ESC3) | `1.3.6.1.4.1.311.20.2.1` ENROLLMENT_AGENT EKU + CA "allow enrollment agent" | request on behalf of others |
| CA / template object rights (ESC4) | WriteDacl/Full Control on the CA or template object | change the above yourself |
| PKI container rights (ESC5) | Rights over `CN=PKI` containers, CA host local admin, backup/restore on the CA | subvert issuance or the store |
| CA-level SAN override (ESC6) | CA `editflags` SAN-without-template-supply | forgery for every template |
| CA officer role (ESC7) | Manage CA / Issue / `CA Administrator` without separation of duties | approve and issue |
| Machine cert auto-enrollment | Domain Machine/Auth template, no manager approval | trusted machine session |
| Relay to Web enrollment (ESC8) | reachable `/certsrv` RPC-over-HTTP + signing-free source | handled by `privilege-relay` |

5. **Rank by Tier-0 relevance.** Does the EKU chain to NT Authentication
   (`1.3.6.1.4.1.311.20.2.1`) and/or Smart Card Logon (`1.3.6.1.5.2.3.1`)? Is PKINIT
   usable? Does the DC actually map certificate subject → account (smart-card logon /
   certificate mapping rules)? Record the mapping question as a fact when you cannot
   confirm it from the directory.
6. **Prove the issue once, against yourself.**
   `certipy req -u <own account> -p <by reference> -ca <ca name> -target <ca host> -template <ESC class template> -upn <your own test UPN>`.
   Verify the SAN you supplied landed: `openssl x509 -in <own>.pfx -inform DER -noout -text | grep -A2 'Subject Alternative Name'`.
   Then prove the authentication path with `certipy auth -pfx <own>.pfx -dc-ip <dc>` — the
   returned identity must be your own test account. That is the end-to-end proof; do not
   substitute a real user's UPN.
7. **Register material by reference.** The pfx stays in the run dir; its password and any
   derived secret go to `save_task_credential` (domain = CA or target domain, username = the
   subject). Facts and findings carry CA name, template common name + OID, request ID and
   certificate serial, SAN, EKU list — never key bytes.
8. **Leave a revocation handle.** Record request ID + serial so the client can find and
   revoke it, request revocation approval, and if the certificate is being held as durable
   access it goes into the `persistence-cleanup` ledger like any other artefact.

## Evidence

- Positive: template `cn` + `pKICertificateTemplateId`, the attribute bytes that put it in
  the class, the trustee holding the enrollment right, one successful request (request ID,
  serial, SAN as issued), the authentication result against your own account, and the
  denominator: "41 templates in the forest, 27 readable as <principal>, 4 with an auth EKU
  open to Domain Users, 1 class-ESC1, 1 issued successfully, 1 rejected by CA policy
  (`manager approval`)", plus what you did not attempt.
- Negative: "No template exposes an authentication EKU to unprivileged principals —
  enumerated all 41 via LDAP on the configuration partition at <time>, decoded EKU + ACL for
  each." ESC8: "not assessed — 443 filtered on both CAs (nmap, <time>), so the Web
  enrollment path was never reachable."

## Report shape

`ADCS template <cn> (<oid>) is <ESC class>: <principal> may enroll and supply the subject, issuing to <EKU> on <CA>`
Fields: CA + endpoint, template OID, class and anchor attribute, ACL trustee, proof basis,
request ID/serial for revocation, remediation (flag, EKU, approval, or ACL to fix).
