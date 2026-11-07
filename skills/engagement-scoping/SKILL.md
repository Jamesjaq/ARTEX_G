---
name: engagement-scoping
description: "Invoke this skill when an engagement is starting or changing and the rules of engagement, constraints and named in-scope asset set must be written down before a single packet leaves the platform: authorisation instrument, target list, windows, exclusions, third-party and cloud-provider rules, test-account provisioning, deconfliction path, and how each becomes a platform constraint."
---

# Engagement Scoping (RoE to constraint set)

Scope is the artefact every other seat is measured against. Done means: a written authorisation
reference, a registered asset set, a constraint list the platform renders on every prompt, a
deconfliction path with a name and a number, and a NOT ASSESSED list for everything the letter did
not buy. Anything unstated is not permission; it is a gap someone will guess at later.

## Boundaries

- Authorized use: nothing here runs against a host until it appears in the registered scope and the
  authorisation reference exists in writing. No verbal, assumed, or "they probably meant it" consent.
- You do not choose targets. You transcribe what the authorisation names and register it. If the
  letter is ambiguous, the ambiguity is a question for the operator, not a licence.
- This skill is not testing: no probing, no banner reads, no DNS requests, no path checks to
  "confirm" an asset is alive before it is scoped. Registration comes first; if it cannot be
  registered, it is out of scope and becomes a question.
- Owner of other subjects: tripwire/SOC-response procedure is `engagement-deconfliction`; captured
  secret material is `credential-handling`; our own indicators and infrastructure are
  `opsec-tradecraft` and `campaign-infrastructure`; the deliverable write-up is `engagement-report`.
- Stop condition: no written authorisation instrument = stop and report. Do not task a single
  specialist, do not queue a worker, do not "just resolve the domain".

## Procedure

Do not go to step 4 until 2 and 3 have produced registered rows and a written exclusion list.

1. **Locate the authorisation instrument** and record its reference verbatim as a fact: statement of
   work number, engagement letter date and signatories, bug-bounty program page plus the policy
   version date, or the lab ownership statement. Quote the clause you are relying on, with its
   section name. If several instruments exist (a contract plus a program page), record both and
   note which is narrower — the narrower one wins.
2. **Register the legal entity and the positive scope.** `list_companies` to see whether the client
   already exists; `add_company_scope` for the entity (it returns the `company_id` used for asset
   attribution). Then `add_task_scope` with the exact entries the letter names — domains,
   wildcard domains, CIDRs, named hosts, application URLs, tenant IDs. Transcribe; do not widen a
   wildcard into a sibling domain you assume is theirs, and do not narrow one either.
3. **Name the in-scope asset set explicitly.** For every host or application the letter lists by name
   rather than by range, `insert_assets` it inside the registered scope with the reason in the
   asset's attributes. Run `list_assets` and `list_untested_assets` afterwards and read the two back
   to the operator: that pair is the denominator every later coverage claim uses. A registration the
   scope gate refuses is a fact to record, not an obstacle to rephrase until it passes.
4. **Write the exclusion list as deny constraints** (`set_constraints`, batch, one clause each, each
   text quoting its RoE origin): named out-of-scope hosts and near neighbours that resolve nearby;
   third-party SaaS and payment providers in the chain; CDN/WAF fronts whose origin belongs to
   someone else; shared or multi-tenant infrastructure; OT, process-control and life/safety systems;
   employee-owned devices; any asset class the letter is silent about. Add matching **allow**
   constraints only where the RoE grants something narrower than default ("read-only queries against
   the staging replica", "passive reconnaissance only in phase 1").
5. **Windows and rate as constraints**: `deny` every action outside the agreed test window (state the
   timezone once, at the top of the text); `deny` volume classes the letter caps or forbids
   ("no load or DoS testing", "no more than N requests/second", "no credential stuffing against
   production identities"). Record the window with start and end dates as a fact so wrap-up seats
   can date-check without re-reading the letter.
6. **Third parties and cloud providers.** For each provider in the architecture, record whether the
   engagement needs advance notice or a filed test request before touching anything, and who files
   it — the client, not this seat. Fetch the vendor's current penetration-testing policy page to
   date the claim rather than quoting remembered terms, and write the result as a fact with the
   retrieval date. Convert the outcome into constraints: `deny` actions on provider infrastructure
   the client does not control, `deny` any action that would affect other tenants, and mark any
   provider whose notice is still unfiled as NOT ASSESSED with the reason.
7. **Test-account provisioning.** Ask the client, through `ask_operator`, for the identities the
   testing needs, one role at a time (unprivileged user, two users at the same privilege for
   cross-entity proofs, an admin, the API service identity, the MDM/VPN pairing). When they arrive,
   store every secret with `save_task_credential` and record only the role, tenant, provisioning
   date and credential ID as a fact. A credential found by testing rather than issued for testing is
   not an authorised identity: it goes to `credential-handling` and needs an explicit operator
   decision before any worker uses it.
8. **Deconfliction path as recorded state**: one fact carrying the named defender contact and role,
   the 24-hour number, the notification channel agreed, the client's own "stop immediately"
   conditions, and who authorises re-testing after a fix. Hand the operating procedure for a fired
   tripwire to `engagement-deconfliction`; this seat only records who to call.
9. **Objective and acceptance bar**: `set_goals` with the objective as the letter states it, each goal
   carrying its measurable proof criterion ("read one record belonging to tenant B from a tenant A
   session", not "test isolation"). Goals without an acceptance bar come back as guesswork.
10. **Operator confirmation before the first active step.** Render the whole set — entity, asset
    entries, exclusions, windows, identities, contacts, objective — as one compact list and put it
    to the operator with `ask_operator` for confirmation. Log the reply as a dated fact. A timeout
    or a partial answer is not confirmation: task only the confirmed lines and mark the rest.
11. **Drift control.** Whenever scope changes — an added asset, a lifted exclusion, a new window —
    re-run the relevant step and record a dated fact naming what changed and which instrument
    authorised the change. A scope change that is not written down did not happen, and every later
    seat will read the old boundary.

## Evidence

A positive scoping result is: the authorisation reference with the clause quoted, the registered
asset rows (`list_assets` IDs plus the `list_untested_assets` denominator), the constraint list as
rendered back by the platform, the contact fact, and the dated operator confirmation. Every line
traces to an instrument clause; no line traces to an assumption.

A negative is a named gap, not silence. For each class the engagement did not cover, write: the
asset or technique class, the method that would have tested it, the RoE clause or missing permission
that stopped it, and the verdict **NOT ASSESSED**. "Excluded by RoE clause 4.2 — third-party payment
provider, no testing permitted; NOT ASSESSED" is reportable. "We did not test payments" is not.
Where an entry could not be registered, state the entry, the refusal, and that nothing was sent to
it. Never record a scope you did not obtain as if it were clean.

## Report shape

Section "Scope and authorisation": fields = authorisation reference and clause, entities
(`company_id`), registered asset entries and count, in-scope denominator, exclusion list with the
clause behind each, test window with timezone, provisioned identities by credential ID, defender
contact and channel, operator confirmation date, and the NOT ASSESSED table.
