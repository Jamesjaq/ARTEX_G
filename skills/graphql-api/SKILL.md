---
name: graphql-api
description: "Invoke this skill when the interface exposes an introspectable schema, batched queries, or a GraphQL-over-REST contract: schema recovery when introspection is off, field and operation level authorisation, persisted-query and batching bypasses, depth/alias/complexity controls, subscriptions, and injection through query arguments."
---

# GraphQL & Schema-Contract Testing

The query language's own failure modes, which REST-shaped tests never surface: schema disclosure, field/operation authorisation gaps, batching and depth as control bypasses, persisted-query allowlist holes, multipart uploads, and subscriptions. Done = a recovered operation list, every tasked boundary proven or cleared, one replayable request per claim.

## Boundaries

- The endpoint the intent names is the target; internal types you discover in a schema are inventory, not authorisation to go and ask for them elsewhere.
- Introspection being **on** is inventory, not a finding. Neither is a schema that merely exists.
- Depth, alias-count and batching probes are compute attacks against a live service: one small request that shows the control is absent, then stop. Anything meant to measure cost runs against a lab copy or after an explicit operator yes.
- Mutations only against objects you created; record every object ID you touched and add a cleanup line.
- Not: JWT/OAuth mechanics and REST BOLA (api-security), login and SSO state machines (auth-testing), a URL argument that makes the server fetch (ssrf-metadata), upload fields and where they land (file-upload-abuse), harvesting operation names out of the frontend bundle (api-recon).

## Procedure

1. Pin the contract before testing anything: POST-only or also GET, `application/json` versus `multipart/form-data` (the graphql-multipart-request shape), single object versus array body (batching), whether `variables` is sent as an object or a JSON string, presence of `extensions`/APQ, and the engine layer from its error shapes (graphql-js `Cannot query field "x" on type "y"`, Apollo Server/Router codes, Hasura `unauthorized`, Yoga/DGS and the .NET runtimes each format differently). The engine and version tell drives steps 4-7.
2. Recover the schema in tiers, and record which tier produced it: full introspection → a trimmed query if the full one is filtered (`__schema { queryType { name } }`) → with introspection disabled, rebuild from did-you-mean field suggestions (query a wrong field per type and read the candidates), operation names and `variables` blocks in the frontend bundle, an exposed playground/docs page, directive names reflected in errors, and APQ-only setups where hashes harvested from the bundle replay against the allowlist.
3. Write the operation inventory as the completeness artifact: per type the id-bearing arguments (`id`, `ids`, `node`, `slug`, `email`, `tenant`), the identity fields (`me`, `viewer`, `currentUser`), directives (`@auth`, `@hasRole`, `@rest`, `@accessible`, Hasura permissions blocks), custom scalars, and whether the model is type-level or field-level authorisation. Anything you could not recover is a stated gap, not silence.
4. Object and field authorisation, always with a second principal of your own: `node(id:)` global-relay walk (decode the base64 ID — `VXNlcjo0Mg==` is `user:42` — then re-encode others), numeric and UUID enumeration in `ids`, list-versus-detail differential (the list filters by tenant, the by-id query often does not), nested traversal reaching a field the REST surface hides, aliasing the same field twice, and `@skip`/`@include` or alias tricks against an authoriser that keys on the literal field name.
5. Transport and parser bypasses: mutation over GET, array batching to slip an operation past a per-operation authoriser or a limiter keyed on the URL, `variables` as a string versus an object changing validation, `operationName` mismatch, persisted-query allowlist bypass (send a `query` body alongside a hash the allowlist knows), content-type swap (`application/graphql`), and a batched body where the parser validates only the first element.
6. Depth, complexity and alias controls: one request with two aliases, one request at depth four. If rejected, name the limit and its message; if accepted, report the missing control with that one request. No scaling.
7. Injection through arguments: `where`/`filter`/`orderBy`/`format` strings into SQL (raw-expression exposure on Hasura and Prisma-style backends is the usual hit), NoSQL operator objects in args (`{ "$ne": null }`), LDAP/`dn` arguments, template-interpolated fields, and the args that are really another class's door — a `url` argument goes to ssrf-metadata, an `Upload` scalar goes to file-upload-abuse.
8. Subscriptions: negotiate the protocol (`graphql-ws` versus `subscriptions-transport-ws`), then test whether authorisation happens only at handshake, whether a revoked or rotated token stops the stream, and whether a subscription can be pointed at another principal's id on a shared channel.
9. Error intelligence: `extensions.exception` stack traces, `path`/`locations` reflection, debug/trace modes, playgrounds in production. These are asset attributes; report one only when it hands something over (schema plus a live authorisation gap is a chain, a stack trace alone is a hardening note).
10. Persist every confirmed operation as a single replayable request in the finding, and mark which ones are one-shot because they consume a nonce or mutate state.

## Evidence

Positive: the exact body (query plus `variables`, with real personal or secret values described rather than pasted), the second account or tenant that proves the boundary is real, which fields returned data they must not, how many such objects out of how many tested, the engine/version observed, and the rejected variants with their error strings.

Negative: "`<endpoint>`: introspection disabled, schema recovered from bundle plus field suggestions (partial: `<n>` fields, `<m>` types unrecovered); `node(id:)` with 3 foreign ids from account B → `unauthorized`; batching rejected (`message: batch limit`); alias count rejected; depth-4 accepted once, not scaled — so cost amplification NOT ASSESSED because no lab copy was available. Field-level auth NOT ASSESSED for `<type>`: no second tenant account." The recovery tier belongs in the negative, because a partial schema means partial coverage.

## Report shape

"`GraphQL <Type.field> on <endpoint> returns <another principal's data / bypasses <control>>`" with: endpoint, operation and variables, auth context (which account), fields reached with counts, engine/version, and the rejected-variant summary.
