---
name: cloud-storage-data
description: "Invoke this skill when the authorized objective is data reachability inside one cloud tenant: object stores, block and database snapshots, queues and topics, and data-lake tables, where the answer must be a counted exposure built from samples rather than a bulk copy."
---

# Cloud Storage and Data Exposure (object stores, snapshots, queues, lakes)

Inside one authorized tenant, storage is where the impact number lives. Enumerate the containers, classify each exposure mechanism, prove read or write from a sample, report the denominator.

Done means every bucket, container, queue and table in the inventory carries a disposition.

## Boundaries

- The tenant you were granted is the working boundary. Another tenant's object reached through a misconfigured grant is a finding about the grant — enumerate it, do not consume it.
- Sampling only: object counts plus a handful of representative objects. No `--recursive` copies, no whole tables, no bulk exports; prefer metadata and ranged reads over downloads.
- Destructive stop conditions: never drain, purge or expire a production queue; never create a public ACL to "test" exposure; never restore snapshots or clone production databases; never modify lifecycle, replication or public-access-block settings.
- Writes that are the finding's own mechanism (queue `SendMessage`, object `PutObject`) go to an object you create, name yourself, and then remove.
- Not this skill:
  - how the grant was obtained — `cloud-attack`;
  - encryption, key policy and rotation — `secrets-keymanagement`;
  - the application code that builds these paths — `codeaudit`; API-level IDOR on the app's endpoints — `api-security`;
  - whether the provider's control truly blocks access — `mitigation-status`.

## Procedure

1. **Inventory before judgement.** Per provider:
   - AWS: `aws s3api list-buckets`, then per bucket `get-bucket-acl`, `get-public-access-block`, `get-bucket-policy`;
   - GCP: `gcloud storage buckets list` plus the bucket's IAM policy bindings;
   - Azure: container list plus the public-access level (`blob` vs `container` vs `none`);
   - S3-compatible stacks (MinIO, Ceph, provider mirrors): the same API surface, but check the ACL implementation, since some ignore canned ACLs.
2. **Emit the inventory as a work-dir artifact**: container, region, ACL class, policy present, public-access-block state, object count (`s3 ls --summarize` or bucket inventory). That count column is the denominator every later finding cites.
3. **Public-ACL classes, kept separate.** Each is its own finding at its own severity:
   - canned `public-read` vs `public-read-write`;
   - bucket policy `Principal: "*"` with and without a `Condition`;
   - GCP `allUsers` vs `allAuthenticatedUsers`;
   - Azure container-level vs blob-level public access;
   - public-access-block enabled over a pre-existing public ACL (the state that looks safe and is not).
   A bucket writable by anonymous principals is an escalation path: note it for `cloud-attack`, do not exploit it here.
4. **Presigned and shared URLs.** Sweep the app's own responses, JS bundles, CI logs and archived traffic for provider signatures. Check `X-Amz-Expires` (a multi-day URL on an otherwise-private bucket is the finding), deprecated SigV2 signatures, and whether URLs survive credential rotation. Verify with HEAD, never by downloading the object.
5. **Requester-pays and redirect traps.** `RequestPayerBucketRequired` confirms existence and reveals payer configuration without granting read; use `--request-payer requester` only where the engagement covers the cost. On `PermanentRedirect` with an `x-amz-bucket-region` mismatch, confirm the target bucket is in your tenant before following it — a cross-account redirect target is its own leak class.
6. **Prefix confusion.** Read the policy's resource ARNs literally:
   - `bucket/cust-1*` also authorises `cust-11/` and `cust-12/`;
   - `ListBucket` without an `s3:prefix` condition exposes the whole key space;
   - wildcards mid-key rather than at the tail.
   Prove with a listing of your own prefix plus one foreign prefix name derived from the policy text; then report the pattern, not the foreign content.
7. **Snapshots, backups, images.** `describe-snapshots` with its `restorableBy` field, RDS/Cloud SQL automated-snapshot and export settings, disk exports landing in object stores, AMI launch permissions, unencrypted snapshots, backup-vault access policies. A snapshot shared beyond the org equals a full-copy exposure: state volume count and total size, and list the filesystem rather than mounting customer data.
8. **Queues and topics.** Policy `Principal: "*"` on SQS/SNS, Service Bus, Pub/Sub subscriptions. `SendMessage` open is integrity (attacker-controlled data into a production consumer); `ReceiveMessage` or consume open is confidentiality, and message bodies routinely carry tokens and PII. Prove read by sending a marked message to a queue you created and pulling 2-3 back. On production queues prove only the permission — policy text plus one non-consuming call.
9. **Data lakes and tables.** Glue, BigQuery, metastore-backed lakes: list databases and tables, take row counts from partition metadata, then `SELECT` with an explicit column list and a small `LIMIT`; for columnar files read the footer rather than the object. Check whether the tenancy column appears in the accessor path at all; report reachable rows and two masked representative rows, and hand the missing-filter code question to `codeaudit`.
10. **Disposition the inventory.** Every row from step 1 ends as exposed (class plus proof), protected (which control, tested how), or not assessed (why). Rows with no disposition make the engagement incomplete, not merely unfinished.

## Evidence

Positive: inventory denominator, exposure class, exact call and response, object and row counts, sample size, which fields were masked, the region and tenant the call ran in.

Negative: "checked 41 of 41 buckets for public ACL and `Principal: "*"` policies: 0 public, public-access-block on for all 41. Snapshot sharing checked on 12 volumes, `restorableBy` empty on all. Queue surface NOT ASSESSED — `sqs:GetQueueUrl` is denied to the current principal."

## Report shape

`<tenant>: <exposure class> on <container|queue|table> — <n> objects / <m> rows reachable (sampled <k>)`

Fields: provider, account, region, resource ARN, ACL or policy citation, counts, sample provenance, test-object cleanup status.
