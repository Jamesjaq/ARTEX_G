---
name: financial-rail-assessment
description: "Invoke this skill when the surface moves money: core banking platforms, branch and teller consoles, payment switches and card gateways, interbank messaging rails, real-time settlement interfaces, mobile-money and agent-banking APIs, USSD gateways, reconciliation and settlement jobs. It is the bench discipline for testing value-bearing flows without creating losses, fabricating payments, or touching real customer records."
---

# Financial Rail Assessment

A money surface is different from every other surface on this engagement in one way that changes the whole method: **every request you send can create an irrevocable state change somewhere downstream**. A duplicated payment that clears is not a test artefact, it is a reconciliation case, an incident, and possibly a regulatory notification. The techniques below are ordinary web, API and protocol testing; what makes this a bench of its own is the discipline around them.

The standing gates in your system prompt still apply. This skill adds what is specific to value movement.

## Boundaries — these are the point of the skill

- **Never move real value.** Testing a transfer, top-up, payout or refund is done against accounts created for this engagement, inside an environment the RoE names as authorised (a UAT/sandbox/staging instance, or an owned lab replica of the switch). A production rail that settles to a real counterparty is read-only: you map it, you do not send it.
- **A negative test is not a transaction.** Where the flow must be pushed past a limit — double spend, replay, amount tampering, approval bypass — do it against your own test account, capture the request and response, and reverse it if the environment permits. If reversing is impossible, the test does not run; you report the reach and mark the consequence derived, not demonstrated.
- **One record, with denominators.** An exposure that shows customer data (statement, KYC file, transaction history, card index) is proved with a single record you are authorised to read plus a count of the population. Never dump a table to show it dumps.
- **No PAN, no CVV, no keys, no full card data in the report.** Reference what you found: last four of a test PAN, the field name, the log path, the count. A finding that carries real card data as evidence has committed a breach of the very control it reports.
- **Credentials and keys go to the credential store by reference** (save_task_credential), never pasted into facts, findings or transcripts.
- **Destructive acts on financial systems are off the table without a named, written authorisation**: ledger deletion, database restore, job disable on a settlement window, key rotation, and ATM/POS cash dispense tests belong to a lab bench and not to a live estate.
- **Settlement and cut-over windows are real.** A finding about a reconciliation job does not require you to trigger it during close. Time-based tests run in the lab or against a clock you control.
- **Fraud and AML rules are the client's defence.** Proving a rule can be evicted is a finding; evicting it in a way that would have let a real scam through a live control is an act with consequences for people who are not on this engagement. Derive that impact, do not demonstrate it.

## What a money surface actually consists of

Map these layers before attacking any of them, because the weak one is usually not the one everyone looks at:

1. **Channel** — branch/teller console, internet and mobile banking, agent portal, USSD or short-code gateway, API partner integration, ATM/POS network, back-office admin.
2. **Switch or hub** — message routing, ISO 8583-style or FIN/ISO 20022 traffic, MAC/signature fields, sequence and Stan handling, response codes, timeouts and retries.
3. **Core ledger** — the platform that owns balances and journals: internal APIs, batch jobs, end-of-day, GL interfaces.
4. **Value rails** — instant payment, card acquiring and issuing, wallets and mobile money, custody and blockchain interfaces, correspondent and settlement links.
5. **Trust edges** — partner and merchant integrations, webhook/IPN endpoints, sandbox keys that leaked into production config, service accounts shared across environments.
6. **Money-adjacent control planes** — limits and approval matrices, dual control and maker-checker, refund/chargeback tools, float and cash-management consoles, reconciliation exports.

## Method

### 1. Establish which layer is in scope, in writing

The scope grants must name the environment (production / UAT / lab replica) per channel, not just a hostname. A "bank app" that is production means read-only mapping plus config findings; the same app in UAT means you can run value tests. If the RoE does not distinguish them, ask before sending anything that could change a balance; `request_operator` with the exact test you want to run.

### 2. Find the value-bearing parameters first

On every mapped endpoint, identify which fields carry **amount, currency, account, beneficiary, reference id, status, fee, direction**. This is the target list. Business-logic findings in finance cluster here:

- amount tampering after validation (decimals, negatives, scientific notation, integer overflow on fee math);
- currency confusion between validation and posting;
- reference/Stan reuse → duplicate posting;
- status transitions that a client may set (pending → settled);
- fee fields that a caller supplies;
- IDOR on another customer's instruction by id.

### 3. Test authorisation with real identities

A financial flow is almost always role-shaped: maker vs checker, teller vs supervisor, branch vs head office, agent vs customer, merchant vs acquiring partner. Two accounts at the right levels, created for this engagement, are the proof. Without them the finding is a suspicion; log a gap (`report_gap`) naming the accounts the engagement needs.

### 4. Replay and race with measurement, not volume

Duplicate-submission and balance-check races are proved with a handful of concurrent requests against a test account, watching the ledger result, not with a load tool. Record: attempts, postings, balance delta, what the reconciliation job said afterwards.

### 5. Switch and messaging protocols: lab only for construction tests

Message-level tampering (MAC recompute, bitmap abuse, unexpected response codes, long messages, unsupported type 0/1/2 traffic) belongs to an owned switch simulator or a lab bench with your own counterparty. Against a real partner endpoint, the honest test is the **reachability, authentication and authorisation** of the interface and the security of its transport and config, not the wire format itself. If the engagement includes an owned switch, capture the exact message, the field you altered, and the accept/reject behaviour.

### 6. Follow the secrets, and stop at the reference

Config files, keystores, HSM client settings, `.env`, mounted secrets, partner certificates, webhook signing secrets and sandbox keys leaking into production are findings of the highest class on this bench. Report where they are, what they authenticate, and what an attacker would reach — the value never goes in the report.

### 7. Impact in money terms

When the reach is proved, state the consequence the way a finance stakeholder reads it: transaction types affected, value per instruction, population of accounts, whether the exposure is reversible by the client within one settlement cycle, and the control that should have stopped it. Amounts at risk are derived unless you actually moved them (which you did not).

## Done means

For every value-bearing surface you claimed: the layer named, the environment named, the request and response captured, the account or instruction that proves it (test data, referenced not dumped), the authorisation context, and the consequence expressed in money and denominators. Anything you could not safely demonstrate is written as a reach with a derived impact and the act you declined to perform.
