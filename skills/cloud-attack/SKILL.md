---
name: cloud-attack
description: "Invoke this skill when an authorized engagement hands you a foothold inside a cloud tenant (AWS, GCP, Azure, or a provider-managed workload) and the question is which identity or resource path actually extends what that principal can do, and how far the resulting blast radius reaches."
---

# Cloud Attack Path (IAM, workload identity, blast radius)

A cloud foothold is a set of grants, not a host. Work two path classes separately — **identity paths** (obtain a different or larger principal) and **resource paths** (use the grant you already hold) — and finish each confirmed path with a counted blast radius.

Done means: every hop proven by one named API call, every path class dispositioned, no foreign tenant entered.

## Boundaries

- Operate only against the account/project/subscription IDs named in the engagement; resolve them before the first provider call.
- A resource readable through a share belonging to another tenant is evidence of a boundary hole, **not** a new target.
- Stop at the policy statement that permits the cross-tenant read. Use the engagement's designated secondary tenant, or an object you created in your own, as the proof control.
- Read-only enumeration plus non-mutating evaluation is the default. Stop conditions:
  - no `CreateRole`, `PutRolePolicy`, `AttachRolePolicy`, `CreateAccessKey` or `PassRole` against live roles;
  - no KMS key creation, no snapshot export, no ACL or public-access change;
  - no logging, trail or sensor tampering of any kind.
- Mutating proof steps need the operator's explicit go-ahead. Record every object ID you create, for teardown.
- Not this skill:
  - application authorisation logic in source — `codeaudit`; API-layer BOLA/IDOR — `api-security`;
  - object-store contents and data sampling — `cloud-storage-data`;
  - key policy, secret material, rotation questions — `secrets-keymanagement`;
  - container-to-node and in-cluster RBAC — `container-escape`;
  - whether a provider control (SCP, IMDSv2, org policy) actually blocks the path — `mitigation-status`;
  - which published flaw opened the door — `vulnerability-intel`.

## Procedure

1. **Pin the identity.** `aws sts get-caller-identity`, `gcloud auth list`, `az ad signed-in-user show`. Record how it was obtained: access key, instance metadata, federated web identity, workload pool.
   Write account/project ID, region set and principal ARN as a fact; that anchors every later count.
2. **Collect the grant set.** Managed plus inline policies for the principal (`aws iam list-attached-role-policies`, `get-role`; GCP `get-iam-policy` bindings; Azure effective role assignments).
3. **Reduce it into three layers and keep them distinct in the report:**
   - identity policy (what the principal is granted);
   - permissions boundary, SCP, or org policy (what is silently capped);
   - session policy (what this particular credential narrowed to).
   A grant that appears but never fires is a privilege-reduction finding, not an attack path. Name the layer that killed it.
4. **Split the hunt by path class.** Do not start step 5 until this step produced either a usable hop or an explicit zero for the identity class.
   - Identity-path candidates: `sts:AssumeRole` on roles whose trust policy omits source-account conditions; `iam:PassRole` onto a service you control; `CreateAccessKey` on a human user; `iam:CreateRole` plus self-attach; GCP `iam.serviceAccounts.getAccessToken` or `roles/iam.serviceAccountTokenCreator`; Azure managed-identity assignment to a VM you hold.
   - Resource-path candidates: object stores, parameter and secret stores, snapshots and backups, shared images, queues, databases reachable with the current grant.
5. **Prove the hop, not the destination.** A successful `AssumeRole` returning credentials for the target role is the finding. Stop there; hand the new principal to a fresh intent or the operator. Store returned credentials with `save_task_credential` and cite the ID.
6. **Workload identity: check binding conditions, not the mechanism.**
   - Metadata paths: IMDSv2 token requirement and hop limit, GCE metadata required header, Azure IMDS identity header and `identity=false` handling.
   - Federation: `sts:AssumeRoleWithWebIdentity` via IRSA, GCP workload identity pools, Azure federated credentials — read the trust policy's `sub`, `aud`, `iss` conditions.
   - Reportable patterns: a role trustable by any pod in a namespace; a federated trust keyed on repo org but not environment or branch; a wildcard audience.
   - Pair each with `mitigation-status` before scoring: confirm the control's state on this tenant's actual configuration, with its check date.
7. **Blast radius arithmetic per path.** State principal → capability → object count in ratio form: `1 of 6 candidate role trusts succeeded; the assumed role lists 43 buckets, 2 carrying other-tenant prefixes`. A path with no numbers is a narrative.
8. **Shared-resource and cross-tenant surface.** Enumerate the share classes whose *policy* you can read: snapshot and image launch permissions, repository and bucket policies with `Principal: "*"` and no `Condition`, missing `aws:PrincipalOrgID` or tenant-ID guards, `allUsers` bindings. Report the grant and the object count under it; never consume another tenant's data to prove it.
9. **Teardown and visibility.** List everything you created, one copy-pastable removal line per object, confirmation recorded as a fact. Note which enumeration calls are provider-logged (CloudTrail, Activity Log, Cloud Audit Logs) — "the blue team can see this activity" is a legitimate report line, and never a licence to touch logging.

## Evidence

Positive: the exact API call and response code; the policy statement (identifiers only, values redacted); the hop proven; `succeeded/attempted`; reachable object counts; the account and region the call ran in; cleanup confirmation.

Negative, with method: "enumerated the grant set of 3 principals via `list-attached-*-policies` plus boundary documents; attempted 6 identity-crossing calls, 0 succeeded over 40 minutes; identity path closed for this principal. Resource path NOT ASSESSED: the SCP denies `s3:ListAllMyBuckets`, so no inventory was obtainable."

## Report shape

`<tenant>: <identity|resource> path from <principal> to <capability> (<blast radius: n accounts / n objects>)`

Fields: provider, account/project ID, region, principal ARN (redacted), hop chain, policy citation, counts, created-object IDs, cleanup status.
