---
name: closure-discipline
description: "Invoke this skill when a direction is about to be written up as exhausted, closed, negative or done — or when the same surface keeps coming back in the graph. It is the discipline that separates 'this channel returned nothing' from 'this attack surface is closed', and stops a run from re-litigating a refuted path or sitting on an unspent capability."
---

# Closure Discipline (what actually counts as exhausted)

Closure is the expensive part of an engagement. A surface wrongly marked dead loses the finding that
was one request away; a surface kept alive costs turns, tokens and the operator's patience. This skill is
about the difference between an observation and a conclusion — and about the two failure modes that appear
in every long run: false exhaustion, and never consuming what you already hold.

## Boundaries

- Applies inside the authorized scope only. Nothing here widens a target set, and a dead end is never a
  reason to reach for an unregistered host.
- A closure claim is a claim about **this engagement's evidence**, with a date and a method. It is not a
  permanent property of the asset, and the graph row says so.
- Do not manufacture depth. A precise negative recorded across three real channels is worth more than
  twenty shallow hits on one channel; both are recorded, neither is padded.
- Not this skill: how to test a class (`web-exploitation`, `api-security`, `auth-testing`), how the sample
  and denominator are written (`data-objective-evidence`), how the report is assembled (`engagement-report`).
- Stop condition: if you cannot name the channel you have not yet tried, you are not closing a surface —
  you are out of a hypothesis. Record the gap and pick up another live face; do not write "closed".

## The rules

1. **Refute a channel, not a surface.** A status code, a body length, a cookie and a redirect are four
   observations of one request. When all four say the same thing, you have refuted *that channel*. The
   input may still be handled by the backend and not reflected. Channels you have not tried: response time
   under a controlled delay, length diff against a known-input baseline, an out-of-band interaction, an
   error path that names a parser, a side effect visible somewhere else (log, mail, second session, job
   queue), and a different representation of the same intent.
2. **Changing the wrapper is not changing the channel.** Another `Host`, another HTTP method, another
   encoding of a rejected payload, or the same request judged again by status code — each is the first
   conclusion wearing a hat. Name what new information the request can produce before you call it depth.
3. **A transport or tool failure is not a negative result.** Connection reset, proxy interception, 429,
   a crashed client, a missing binary, a blocked command: those say nothing about the target. Record them
   as the environmental block they are (`ask_operator` with the one command), and leave the surface open.
4. **False exhaustion has a shape.** "All endpoints are 403", "the parameter is filtered", "the login is
   not vulnerable", "this app is a SPA with no API", "the subdomain is parked" — each is one channel on a
   subset. Before writing any of them, list which entry points, which parameter positions, which
   representations and which observation channels you actually exercised. If that list is short, the
   correct row is *not assessed, prerequisite X missing*, not *closed*.
5. **Consume what you hold.** The next action must spend the most capable thing you have: a valid session
   or credential goes against authenticated functionality, not back to entry-point enumeration; a verified
   read primitive goes to the secret it can reach; a foothold goes to the segment it can reach; a signed
   token goes to the claims, audiences and the services that accept it. Re-enumerating the front door while
   holding credentials is the single most common stall on a run. Write one sentence saying which capability
   you just spent and what it produced.
6. **A refuted path stays refuted until there is new evidence.** Before re-opening something the graph
   marked dead, name the new fact that justifies it — a version you did not know, a second parameter, a
   different channel. If you cannot, that is a signal to move to another live face, not to re-run the old
   one with different words.
7. **Cite the graph in every judgement.** A direction you propose — to a lead, in a fact, in a brief —
   must point at node, asset, finding, traffic record or refutation IDs. If the citation cannot be written,
   the judgement was made from impression rather than state: do not send it.
8. **One observation of failure does not close a class.** Different entry points, parameters, or privilege
   levels are separate candidates; a filter that rejects one representation of a payload has not refuted the
   class, it has told you where the check lives. Report the position of the check as the finding when the
   class cannot be pushed further in bounds.
9. **Distinguish "not reachable" from "not present".** A host you cannot reach from where you stand, a
   route the WAF blocks, an endpoint that needs a network position you were not given: those are reach
   statements with a named blocker, and they belong in the facts as such.
10. **Close with the denominator.** When you do write a closure: the surfaces tested, the channels used on
    each, the count of requests or payloads behind the claim, and what was not covered. "SQLi on `/search`:
    18 payloads across error/time/OOB channels, 4 parameter positions, all rejected by the WAF page; the
    authenticated `report` endpoint was not tested — no account with that role" closes one face and hands
    the next seat the other.

## Evidence

Positive: a closure row naming surfaces × channels × counts, a one-line "capability spent" note when a
credential, foothold or primitive is used, and a re-open justification citing the new fact.

Negative and honest-gap forms, all reportable: "marked closed on status code alone — three channels not
attempted, re-opened"; "credentials held for 4 turns while entry enumeration continued"; "closure claims
no denominator: says 'all parameters tested', the graph shows two"; "refuted path re-run with different
encoding and no new evidence".

## Report shape

Deliverable "Closure ledger": per face = surface, entry points covered, channels exercised (echo / length /
time / OOB / error / side effect), requests or payload count, outcome, what remains untested with the
prerequisite named, decision (`closed` / `not assessed` / `blocked`), date. Plus the capability ledger:
what was held, what it was spent on, what it produced.
