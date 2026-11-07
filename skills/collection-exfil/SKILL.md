---
name: collection-exfil
description: "Invoke this skill when the objective is data: proving which records are reachable from the identity you hold and sampling them minimally, with a count and a denominator, without copying a table out."
---

# Collection and Minimal Exfiltration

Proving data reachability is a measurement, not a haul. Done means: for each store the objective cares about, the reach mechanism, a row/object count, the denominator of what was denied, and two or three representative rows described (class and identifier), with the exact capped query recorded so nobody has to read the store twice.

## Boundaries

- Sample, never bulk. Every data claim is a count + denominator + a couple of representative rows; a dumped table, a `.csv`/`.sql` export, or a multi-MB archive of target data is out of bounds no matter how easy the query is.
- Prefer rows you created: seed a record through the normal path first and read that back. Real records get touched only as the last 1-3 samples needed to prove the class, and their values get described, not stored.
- Not this skill: bytes crossing the perimeter (`exfil-egress`), secret material and its handling (`credential-handling`), the backup estate as a store (`backup-recovery-targets`), turning the numbers into impact language (`data-objective-evidence`), object storage/cloud specifics (`cloud-storage-data`).
- Stop conditions: no writes, no deletes, no schema changes, no re-encryption; if the objective can only be proven by altering or exporting production data, that is a blocked direction you hand up.

## Procedure

1. Write the objective as a testable sentence before touching anything: "role R can read N rows of class C from store S" / "the API returns other tenants' records for K of M identifier variants". Take the wording from the goal node (`graph_overview`) and put your testable form in the first fact, so the seat that follows you measures the same thing.
2. Establish reach without data first — a listing is usually enough:
   - SQL: `SELECT count(*) FROM information_schema.tables WHERE table_schema=<db>` then `SHOW TABLES` / `\dt` / `SELECT relname, reltuples FROM pg_class ORDER BY reltuples DESC LIMIT 20`
   - NoSQL: `show dbs`, `db.getCollectionNames().length`, `db.stats()`
   - File/share: `smbclient -L //<host> -U ...`, `showmount -e <host>`, `findmnt`, `ls <mount> | head` (names, not contents)
   - Object storage: bucket listing count (`aws s3 ls s3://<bucket> --recursive | wc -l` only if the tenant is in scope, else `list_objects_v2` with `MaxKeys=3`)
   - Cache/queue: `redis-cli --scan --pattern 'cust*' | head -20` (piped through `wc -l` for the count), `redis-cli dbsize`
3. Count with aggregate-only queries and record them; counts cost no rows and they are what the report needs:
   - `SELECT count(*)` per candidate table, `reltuples` estimates, `collStats`/`dbStats`, `du -sh` on file stores, index/table sizes
   - Always pair the number with the identity that produced it ("as app_user@%, read via the CRM connection string in <path>").
4. Sample under a hard cap and mask on the way in:
   - `SELECT <3 columns> FROM <table> LIMIT 3` / `head -3` of a file / one object per bucket / at most 3 distinct entities, none of them real customers when a seeded row exists
   - Record: column names, data classes present ("email, national-id-format, order total"), the object id or primary key of each sampled row, and a redaction note. Values themselves do not go into the fact, the finding, or the work dir.
   - If the row class is high-sensitivity (payment, medical, authentication material), the sample is the *shape*: "column `pan` holds 16-digit values passing Luhn in 3 of 3 sampled rows" beats pasting a card number.
5. Build the denominator honestly: how many candidate stores/tables/buckets did you attempt, how many answered, how many refused and with what error ("attempted 12 tables; 4 readable as this role, 7 denied by GRANT (quoted permission errors), 1 not attempted because it held a class the RoE excludes"). The 7 denials are the client's control working and belong in the report.
6. Capture the query once and cite it forever: put the exact capped statement in the fact's evidence line so the reporter and the retest can replay it. If you need the same store again, cite the earlier fact id instead of re-reading rows — a second pass over real data is waste with a privacy tail.
7. Config-embedded access is a reach path, not a trophy: connection strings, `.env`, `application.yml`, kube secrets mounted in a pod, service keyfiles. Record path + line reference + the engine/host it reaches, save the account to `save_task_credential` by reference, and move on to step 2 with those rights.
8. Know which store class hands the work to another seat: a backup file or VHD on disk → `backup-recovery-targets`; a cloud bucket/snapshot/queue → `cloud-storage-data`; keys that decrypt the rest → `secrets-keymanagement`; mail/collab stores → `messaging-collab`.
9. Close with one fact per store reached and a roll-up fact for the objective: {summary: "<store>: <N> rows of <class> readable as <role>; <M> of <K> tables attempted; 3 sampled, masked", detail: counts per table, denial list, sampled object ids, the identity/path used, evidence: the capped query + the count output line, confidence: observed}.
10. Then, and only then, `report_finding` with the measurement in the summary — not with the sample text in it.

## Evidence

- A positive needs all five: the reach mechanism (exact query/command, capped), the count, the denominator with the denial reasons, the identity/role that made it work, and the sample description without values. Missing any one, the claim is a hypothesis.
- A negative states what was attempted and what remains open: "4 read paths attempted against the CRM (direct SQL as app role, ORM endpoint with manipulated id, SMB share, on-disk backup file): 3 denied — errors quoted — 1 reached only the seeded test tenant. Production rows: 0 of 3 attempts. Payment-card class NOT ASSESSED (column encrypted at rest, key not reached)".
- "Reachable" is never written as "dumpable": the claim is about permission, and the number proves the permission.
- Time- and state-bound each count: "3.4M rows as of 14:02 UTC, from `reltuples` estimate (not an exact scan)" — an estimate and a count are different evidence.

## Report shape

Title: "<store>: <M> of <N> <data class> records readable as <role> (<K> sampled, masked)". `vulnclass` = sensitive-data-exposure / broken-access-control, `severity` from count × data class × the identity that reached it, `asset_ids` = the store's rows, `evidence` = the capped query + counts + denominator, `traffic_refs` where an HTTP path produced the read.
