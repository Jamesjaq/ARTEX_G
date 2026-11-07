---
name: database-attack
description: "Invoke this skill when the tasked surface is a data engine rather than the app in front of it: SQL or NoSQL listeners, their authentication posture, trust links between engines (linked servers, federation, replication, database links), and code execution reachable behind a query."
---

# Database Attack (engines, trust links, execution behind a query)

The engine as an attack surface: who can reach the listener, what the account you hold actually
grants, which other engines it trusts, and what a query can make the host do. Done means every
tasked instance has an exposure verdict, a data-reach measurement with its denominator, a trust
link inventory, and — where execution was proved — the statement pair that created and removed
your own object.

## Boundaries

- Authorized use: connect only to instances the task registered, using only the accounts the task
  issued. Following a link reaches another asset: check the registered list before you cross it,
  and if the far side is not listed, stop and record the link as a fact instead of traversing it.
- Read-only proof first. Execution through a database is a finding's mechanism when demonstrated on
  an object you created; enabling a server feature, writing outside the data directory, or loading
  code into a production engine is a server-side change needing the operator's explicit yes for that
  instance. Never `FLUSHALL`, `SHUTDOWN`, `DROP`, `TRUNCATE`, `KILL`, `CONFIG SET` or a bulk
  `UPDATE`/`DELETE` on a production store — any of those is an outage or data loss, not a PoC.
- Real rows are not evidence to paste. Sample the shape: a count, a denominator, and two or three
  representative values with the sensitive parts elided, and objects referenced by ID. Hashes,
  keys and connection-string secrets go to `save_task_credential` and are cited by ID only.
- Not this skill: finding the listener (`network-pentest`), non-DB service banners
  (`infra-exploitation`), SQL injection through an application (`web-exploitation`, `api-security`),
  quantifying blast radius (`data-objective-evidence`), sampling at volume (`collection-exfil`),
  storing captured material (`credential-handling`), domain paths reached through a service account
  (`ad-attack`), backup stores (`backup-recovery-targets`), object stores and lakes
  (`cloud-storage-data`).

## Procedure

1. Enumerate the tasked instances and their listeners from the registered assets: address, port,
   bind scope (loopback / internal / world), TLS optional or required, and reachability from each
   vantage you have — before any query runs.
2. Unauthenticated reach, read commands only, per engine:
   - Redis: `redis-cli -h <host> INFO`, `CLIENT LIST`, `CONFIG GET maxmemory`. Exposure is proved
     by `INFO`; `SET`, `CONFIG SET dir`, `MODULE LOAD` and the cron/SSH write chain are lab
     detonation — on a live instance report the exposure, do not run the chain.
   - MongoDB: `mongosh --quiet --eval 'db.adminCommand({ping:1})'`, then `listDatabases` only if an
     unauthenticated read is permitted; no `find()` across user collections.
   - Elasticsearch/OpenSearch: `_cluster/health`, `_cat/indices?format=json` (names and counts are
     the exposure); no `_delete_by_query`, no mapping writes.
   - ClickHouse / InfluxDB / Neo4j: `SHOW DATABASES` over HTTP, `/_all_dbs`, the unauthenticated
     Bolt or HTTP root; also check for a bundled UI (`/_utils`, Fauxton, the play console).
   - Memcached: `printf 'stats\r\n' | nc <host> 11211` — a stats reply means the store is open.
   - Cassandra: `cqlsh -e "DESCRIBE KEYSPACES"`; the banner's authenticator line says `AllowAll`.
3. Authenticated posture with the issued account, and nothing beyond it:
   - MySQL/MariaDB: `SELECT USER(),CURRENT_USER(),VERSION();` account counts per auth plugin
     (`SELECT plugin,COUNT(*) FROM mysql.user GROUP BY plugin`) — counts, never the authentication
     strings; then `@@secure_file_priv` and `@@require_secure_transport`.
   - PostgreSQL: `SELECT current_user,version();` plus `rolname,rolsuper,rolcanlogin,rolreplication
     FROM pg_roles`, `pg_has_role(current_user,'pg_write_server_files','usage')` and
     `SHOW password_encryption;`
   - MSSQL: `SELECT @@version,SUSER_SNAME(),IS_SRVROLEMEMBER('sysadmin');` plus the surface-area
     read `SELECT name,value_in_use FROM sys.configurations WHERE name IN ('xp_cmdshell',
     'show advanced options','OLE Automation Procedures');`
   - Oracle: `lsnrctl status <host>` for version, services and whether the listener password is
     unset; open-account reads with the issued account only.
4. Data reach, measured: for each sensitive store the objective cares about, one
   `SELECT COUNT(*)` (or `db.<coll>.estimatedDocumentCount()`, `_count`) as the numerator against
   the store's own total, then `LIMIT 3` of non-sensitive columns or elided values to prove the
   shape. Write "N of M rows reachable as <account>; sample elided" — that is the measurement, not
   a dump.
5. Trust links between engines — inventory first, traverse only in scope, one probe each:
   - MSSQL linked servers: `SELECT name,product,data_source FROM sys.servers;` then one
     `SELECT * FROM OPENQUERY(<link>,'SELECT @@version')` per in-scope link to prove the stored
     credential flows. `sys.linked_logins` and `sys.remote_logins` say whether any local principal
     can use the link or only a fixed mapped login — that distinction is the finding.
   - MySQL: `FEDERATED` tables, and `mysql.servers`, which stores remote credentials in cleartext —
     exposure you register by reference.
   - PostgreSQL: `foreign_servers`/`foreign_data_wrappers` for `postgres_fdw` and `dblink`, then
     `pg_user_mappings` — a mapping holding a password is material for the credential store.
   - Oracle database links: `SELECT owner,username,host FROM all_db_links;` — following one needs
     the operator's yes unless the far side is registered and tasked.
   - Replication and CDC: `SHOW REPLICA STATUS`, `pg_stat_replication`, MSSQL distribution DBs,
     `rs.conf()` (reveals internal hosts), Kafka/Debezium topics carrying whole tables. A replica
     target is a second asset: register it.
6. Code execution behind a query. Prove the grant exists first — usually that is the whole finding —
   and demonstrate the effect only where the RoE allows, only on an object you created:
   - MSSQL: the `xp_cmdshell` configuration row plus `IS_SRVROLEMEMBER('sysadmin')`. Enabling the
     feature is a server change: ask. Where approved, run one command and prove it by the state it
     changed, not by a shell banner.
   - PostgreSQL: `COPY FROM PROGRAM` needs superuser or `pg_execute_server_program`; prove the
     membership, then where approved read one known file via `pg_read_file`. Untrusted languages
     (`plpython3u`, `c`) need a `CREATE FUNCTION`: task-named, called once, dropped, both recorded.
   - MySQL/MariaDB: `FILE` privilege plus `@@secure_file_priv`; an empty `secure_file_priv` with
     `FILE` held is read-any-path and write-to-webroot exposure. UDF loading is lab detonation,
     reported as exposure, never run on a live instance.
   - Oracle: `UTL_INADDR.get_host_address`/`UTL_HTTP` for out-of-band reach of a host you control;
     `DBMS_SCHEDULER` job creation is execution — own object, drop after, job name recorded.
   - NoSQL: MongoDB `$where` and stored JS, Elasticsearch painless execution
     (`{"script":{"source":"Math.max(1,2)"}}` is a harmless proof), Redis `MODULE LOAD` and `EVAL`
     as exposure classes.
7. Auth reflection out of a DB engine (`xp_dirtree`, `LOAD_FILE` of a UNC path, UTL_TCP to your host)
   coerces a service account to authenticate to you. Blast radius: one attempt against the named
   asset, evidence captured, stop, hand the relay consequence to `privilege-relay`.
8. Revert and record: every object you created is dropped in the same session, every touched account
   or table ID is listed in the evidence, and the closing fact says which instance is back to its
   original state and which was changed and needs the owner to know.

## Evidence

Positive: engine and exact version string, the account used, the statement pair (or command and
output), the measured reach with its denominator ("2 of 5 linked servers traversed in scope; 3
far sides unregistered, not crossed"), the created object and its removal, proof basis
(executed-live / executed-lab), and any material referenced by credential-store ID.

Negative: "Unauthenticated read tested via the engine's own read commands across 6/6 tasked
instances over window Z; 0 accepted an anonymous read; execution grants read from
`sys.configurations`/`pg_roles` and are all absent — feature enablement NOT ASSESSED, not
permitted by RoE". Never write "no SQL risk" from a version string.

## Report shape

"<Engine> at <host:port> lets <principal> <execute/read/traverse> into <scope>, affecting <N rows
of M in <store>>" — fields: engine/version, listener scope and vantage, account and its grants,
the statement pair, measured rows with denominator, linked/trusted instances crossed, objects
created and removed, and credential references for anything captured.
