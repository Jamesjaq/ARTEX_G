---
name: fuzzing
description: "Invoke this skill when the task is to find a previously unknown flaw by driving a parser, file format, protocol grammar or protocol state machine with generated input in a lab: harness setup, corpus, coverage signal, crash triage and dedup."
---

# Fuzzing (novel-flaw discovery via harness, grammar or state machine)

Fuzzing produces candidates, not findings. The deliverable is a harness that runs, a corpus that reaches real code, a coverage signal proving the harness is honest, and a de-duplicated crash set where each unique crash reproduces on command.

Done means: one signature per unique crash, each with a minimized input, and a stated budget for whatever came out clean.

## Boundaries

- Fuzz a component you can build and run locally, or a lab instance of the in-scope product. High-rate mutation against a live third-party service is a denial-of-service event and out of bounds.
- Protocol-level requests against an in-scope live service only at the rate the engagement allows, and that surface needs the operator's go-ahead before it starts.
- Seed corpora are synthetic or public samples shaped like production data. Never fuzz with a customer dump.
- Rate limits, WAF blocks and throttling during a live-session fuzz are observations to record, not obstacles to out-technique; no sensor probing is part of this procedure.
- Not this skill:
  - recovering a vulnerability from a vendor fix — `patch-diff`;
  - turning a crash into a reliable proof of impact — `exploit-development`;
  - reverse engineering the parser before writing the harness — `rea`;
  - whether a published advisory already covers this component — `vulnerability-intel`.

## Procedure

1. **Choose the input boundary** where malformed data is cheapest: a library entry function (parse, decode, decompress, deserialize), a file or archive loader, a network message constructor, or an RPC request encoder. Prefer component-level harnessing over whole-application driving — same code path, orders of magnitude more execs per second.
2. **Build the harness.**
   - Signature: takes a byte buffer plus length, calls exactly one entry point, loops nothing.
   - Determinism: seed or stub randomness, stub time and network, silence logging (one print per exec destroys throughput), make state setup idempotent so a crash cannot poison the next run.
   - Instrumentation: `-fsanitize=fuzzer,address,undefined` for libFuzzer, `-fsanitize=fuzzer-no-link` plus a separate sanitizer build, `go test -fuzz FuzzX` for Go, `atheris` for Python, `afl-clang-lto` when the source is not yours to modify.
   - Gate: do not go to step 3 until the harness runs a known-good seed and exits 0.
3. **Seed the corpus** from valid inputs only: the format's own fixtures, protocol captures from your own lab traffic, public sample documents, generators for the format. Then minimize so each seed contributes unique coverage (`afl-cmin`, `cargo fuzz cmin`, libFuzzer `-merge=1` over the split set). A 10k-file unminimized corpus is noise; a 300-file minimized one is a signal.
4. **Give the mutator the format's alphabet.** AFL++ dictionary (`-x`) or libFuzzer `-dict` with magic bytes, tags and keywords: container signatures, ASN.1 and protobuf tags, JPEG/PNG/PDF chunk markers, the field names the parser switches on.
5. **For structures mutation cannot reach** (checksums, length prefixes, framing), switch to grammar or structure-aware generation: a template per record type with length and checksum fields computed by the generator, a grammar in langfuzz/pyfuzz style, or a custom mutator hook. Coverage from step 7 is what tells you whether the switch was needed.
6. **Stateful protocol work needs a state model, not random messages.** Write the legal transition table (connect, authenticate, negotiate, transmit, teardown), then mutate one axis at a time against that model:
   - out-of-order, duplicated or replayed messages;
   - oversized length fields against the declared size, truncated tails;
   - version or cipher downgrade;
   - timeout and keep-alive interleaving;
   - concurrent sessions sharing parser state.
   Keep a per-state disposition record; untested transitions mean untested code.
7. **Establish the coverage signal before trusting a long run.** `llvm-cov` or `sancov` edge counts, `-exact_count` for libFuzzer, `afl-whatsup` for AFL++; record edges, functions, corpus size and execs per second at T0. A curve flat within 10 minutes means the harness is not reaching the parser: fix step 2 rather than adding hours.
8. **Black-box fallback** when source or build is unavailable: a response-divergence oracle over status code, body length, error-string class and connection-level failure. Two controls required before a divergence becomes a candidate — the same input replayed, and a known-good input in the same window — because service jitter and GC pauses look like anomalies.
9. **Triage each crash.**
   - Reproduce it, then minimize while the crash persists (`afl-tmin`, libFuzzer `-shrink`, manual chunk deletion for protocol sequences).
   - Capture the sanitizer report verbatim, the top frames and the faulting address or offset; symbolize with `llvm-symbolizer`.
   - Classify the bug class from the report header: heap-buffer-overflow read or write, stack overflow past the guard page, use-after-free, double free, integer-overflow UB, null dereference, assertion.
10. **Dedup by signature, not by count.** Key each crash as sanitizer type plus module plus the first three non-library frames (or the faulting offset for the same instruction). Group, record the population per signature, then re-run each unique crash 5-10 times: nondeterministic ones are environmental — report them as flaky and drop them from the finding set.
11. **Hand off.** Every unique, deterministic crash goes to `exploit-development` with the minimized input file, build identity (binary sha256 plus version string), the harness command, the sanitizer report, and whether the code path exists in the shipping configuration.

## Evidence

Positive, per unique crash: minimized input plus its sha256; the one-line command that reproduces; sanitizer report verbatim; bug class and frames; reproduction ratio out of repeated attempts; build hash and configuration; whether the path is reachable in production, and an explicit "unknown" when it is.

Negative: "harness `fuzz_png_chunk` ran 4h12m, 38.4M execs, corpus 412 seeds minimized from 3.1k, edge coverage 1842 rising to a plateau of 1844 by minute 33, 0 sanitizer reports. This component is therefore NOT ASSESSED for memory safety beyond 1844 covered edges — the ICCP and eXIf writers remained uncovered (list attached)."

## Report shape

`<component> <build hash>: <bug class> in <function> via <input class> (<repro n/m, minimized k bytes>)`

Fields: harness path, seeds and dictionary used, sanitizer flags, coverage at stop, execs and duration, unique-crash signature, reproduction command, production reachability note.
