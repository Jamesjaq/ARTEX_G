---
name: model-data-security
description: "Invoke this skill when the objective is the data and artifacts behind an AI system: training sets, model weights, embedding/vector stores, RAG corpora, fine-tune pipelines, and secrets recoverable by prompt extraction — with reach proved by planted records and counts, and membership/inversion/poisoning claims held to a measured baseline."
---

# Model & Data Security (corpus, store, artifact)

The assets behind the model are ordinary data stores wearing a novelty costume: a vector DB with no tenant filter is an access-control flaw, a writable RAG index is a poisoning path, a downloadable weight file is a secret-exposure surface. Done means measured reach — how much of what was recovered, and what the control blocked — never a bulk dump or an uncited "it can leak training data" claim.

## Boundaries

- Authorized use only: a store, corpus, model artifact or pipeline you own or are contracted to assess; confirm it against registered in-scope assets first, and stop if ownership is not established.
- Sample, never bulk-exfiltrate: prove reach with planted records plus a count and a couple of representative rows. Do not pull a real corpus or real customer/PII records to make a point.
- No secret material in prose: recovered keys, tokens or weight-file credentials go to the credential store by reference (`save_task_credential`), cited by ID.
- Access/storage/crypto mechanics of the underlying datastore belong to their own platform skills; this skill owns the AI-specific interpretation (embeddings, inversion, poisoning, extraction) on top. Prompt-extraction at the app interface starts in `ai-redteam`; here you prove the *data-at-rest* consequence.

## Procedure

1. Map the data plane first — enumerate what exists and who may touch it:
   - training sets and their storage location, model weight files/buckets, embedding/vector stores and collections, the RAG corpus and its ingest path, fine-tune pipelines and their write access;
   - for each: read-path, write-path, and the access control in force;
   - bind findings to registered in-scope stores via `list_assets` / `insert_assets` / `node_detail`.
2. Reachability & authorization (treat it like the datastore it is): test tenant/ACL isolation on the vector store and corpus — can one query or one tenant read rows it should not? Use planted records with known IDs to count reach ("3 of 20 planted cross-tenant queries returned other-tenant rows; 17 rejected"). Highest-value class; needs no exotic AI tooling.
3. Exposure of artifacts:
   - are weight files, embedding dumps, or training snapshots in an accessible/unauthenticated bucket, endpoint, or model hub under the org's name?
   - prove read on a small ranged request (headers/metadata, or a few KB), not by pulling the whole artifact;
   - a publicly reachable weights file with embedded credentials is the finding.
4. Prompt-extracted secrets: determine whether system prompts, few-shot fixtures, or corpus-injected context contain secrets/PII and whether the model emits them. Demonstrate with a canary you planted into the corpus/prompt, cite the value by reference, and stop at the smallest proof — do not enumerate every secret.
5. Embedding inversion / reconstruction — only claim it recovers content if you can measure it:
   - run a small reconstruction experiment on planted source strings;
   - quantify what came back (exact-match rate, cosine similarity of reconstructed vs original);
   - compare against a baseline (random or trivial round-trip) and state the confidence; if it does not beat baseline, the honest result is "not demonstrated here".
6. Membership inference — frame as a binary decision task with a measured baseline:
   - build known-member and known-non-member sets;
   - report the true-positive/accuracy rate against a chance baseline (~50%) and a random-probe control;
   - a membership claim with no denominator and no baseline is marketing — record it inconclusive, not as a finding.
7. Poisoning — establish the write path FIRST (can an attacker or low-privilege user add/edit a document in the RAG corpus, training set, or a fine-tune input?), then show the effect with exactly ONE retrieved result that reflects the injected content. Write-path-plus-one-impact is the bar; without the write path it is a hardening note.
8. Fine-tune / pipeline supply-chain: check who can push a dataset or a checkpoint into the training or fine-tune path and what signs it. An unsigned checkpoint, a dataset with no provenance, or a writable build input is a poisoning/integrity route — prove the write access, then defer impact to step 7's standard.
9. Model extraction (stealing the model itself, distinct from its training data): if a hosted model is reachable only through an API, measure how much behaviour can be cloned — log a small number of input/output pairs from your own queries and state what fraction of behaviour a copy would need; do NOT scrape the endpoint into a full imitation dataset (that is both a bulk-harvest and an availability cost — keep it to a bounded sample within the impact budget).
10. Do not go to 5–9 until 1–2 have established what the store holds and who can reach it; inversion/membership/poisoning/extraction claims are meaningless without that baseline.

## Evidence

- Positive: which store/artifact, the access path, the control in force, measured reach as a count with denominator, and (for inversion/membership) the experiment design, the baseline compared against, and the confidence. State proof basis (exec-live vs static) on every claim.
- Negative: "Cross-tenant isolation on vector collection X tested via 20 planted records across N query classes over window Z; all out-of-tenant reads rejected -> assessed effective for this class" — or NOT ASSESSED with method and reason. An inversion/membership attempt that did not beat its baseline is written as measured-and-refuted, not omitted.

## Report shape

Title: `<class> on <store/artifact>: <reached/extracted/poisoned> <what> (<count/denominator>)`. Fields: store name + path, read/write access control, planted-record design, count with denominator, baseline + confidence (for inversion/membership), secret references by ID only, one representative sample (never a dump), and the control that should have stopped it.
