---
name: engagement-modes
description: "Invoke this skill when the engagement's posture has to be stated: how much breadth, depth and time the bench is allowed to spend, and whether the target is a whole estate, a running service or a code diff. It sets the phase budget that every seat then works inside, so 'go deeper' and 'cover everything' stop meaning the same run."
---

# Engagement modes

A syndicate without a posture is a bench that does everything at medium depth forever. The
objective says *what*; the mode says *how hard, how wide, and against what kind of artefact*.
Pick one, write it into the task, and let the budget follow.

This exists because ARTEX already tiers each **seat** (run seconds and turns by rank) but had no
word for the shape of the **engagement** - so a mission to "check this host" and a mission to
"map the estate" consumed the same attention, and the planner had nothing to trade against.

## The four modes

### quick — answer one question
A named target, a named question, a short fuse. Recon to first meaningful response, one or two
technique families, and a result that says *confirmed* or *not present* - not a survey.

- Budget: one or two surfaces, no backlog growth, no chaining.
- Deliverable: the answer, the request and response that prove it, and what was not tested.
- Wrong choice when: the scope has hosts nobody has fingerprinted, or the question depends on
  credentials that do not exist yet.

### standard — work the surface honestly
The default. Fingerprint, enumerate, test the classes the technology actually exposes, prove or
rule out, and keep a NOT-ASSESSED list for what was skipped and why.

- Budget: every in-scope asset gets a pass; depth goes where the response says something.
- Deliverable: coverage number, findings with evidence, gaps filed against benches.
- Wrong choice when: the objective is an attack path rather than a surface list - then go deep.

### deep — find what others miss
Chaining and consequence. Full parameter and state coverage on the surfaces that answer,
cross-service edges followed, low-and-slow classes (race, cache, deserialisation, logic) worked
rather than noted, and findings combined into what they get an attacker.

- Budget: few surfaces, many passes; expect to spend the whole engagement on one host.
- Deliverable: a path with named preconditions, and the impact chain in numbers.
- Wrong choice when: the scope is wide and nothing has been fingerprinted - depth on an unknown
  surface is a false negative with a nicer name.

### diff — assess a change, not a host
The target is a code delta: a pull request, a release, a patch window. What is new, what did the
change un-guard, which sink became reachable. Structural passes first (`sg`/Tree-sitter for the
shape, `semgrep` for the sink classes, `gitleaks` for what got committed, `trivy fs` for what the
dependencies now pull in), then prove only the classes the diff actually introduced.

- Budget: the diff and its blast radius, not the repository.
- Deliverable: findings tied to changed lines, plus what the change *would* have enabled had it
  landed differently.
- Wrong choice when: there is no source, or the question is "is this deployed thing exploitable".

## Choosing, and saying so

1. Read the objective for a **noun**: an estate (standard/deep), a host (quick/standard), a path
   or a crown jewel (deep), a change (diff).
2. Read the scope for **width**: more than a dozen live assets with nothing fingerprinted caps you
   at standard until the recon bench has been paid.
3. Read the RoE for **permission**: an axis left `unspecified` removes whole classes from every
   mode - say which, rather than quietly narrowing.
4. Write the chosen mode into the task hints, name the phase budget in the brief you send down
   the chain, and re-state it when you escalate: a seat told "deep" that spends its turns on
   enumeration has been mis-tasked, not misbehaved.
5. Switching mid-engagement is a decision, not a drift. Say why the previous mode stopped being
   the right one.

## Boundaries

- A mode changes **effort**, never authorisation. Scope grants and the RoE still decide what may
  be contacted, written, or left the machine for - `deep` does not unlock anything.
- `diff` mode reads source that may be the client's crown jewels: work in the engagement's own
  checkout, keep artefacts in the session directory, and never paste secrets found in code into a
  finding - reference the file, the line and the class.
- Chaining findings in `deep` mode must not create state on a target that cannot be undone:
  derived impact is stated as derived, and the act you declined is named.

## Done means

For any mode: the mode is written down, the budget it implies is visible in the brief, the
deliverable matches what was promised, and what was deliberately not done is listed instead of
silently dropped.
