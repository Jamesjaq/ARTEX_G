---
name: stored-value-custody
description: "Invoke this skill when the surface holds value in its own form rather than a bank ledger: cryptocurrency exchange and brokerage platforms, custody and wallet infrastructure, node and RPC interfaces, withdrawal and hot-wallet pipelines, blockchain interfaces, seed and key material handling, and the operator tooling around them."
---

# Stored Value & Custody

Crypto surfaces fail in a small number of spectacular ways, and every one of them is irreversible. There is no reversal request, no chargeback scheme and no bank to call; a withdrawn coin is gone. That changes the risk calculus of every test below, and it changes how a finding must be written.

This skill covers the platform side (an exchange, a broker, a custodial product), the self-custody side (wallets, node interfaces, key ceremonies) and the seam between them. It does not replace `secrets-keymanagement` for key material generally — it says what is different when the key *is* the money.

## Boundaries

- **No value moves.** Withdrawal, transfer, sweep and trade tests run against accounts and wallets created for this engagement, on an environment the RoE names. A live hot wallet is mapped and authenticated, never spent from. If you cannot distinguish "our test wallet" from a production one, you do not have permission to touch it.
- **Chain state is a production dependency.** Sending a real transaction on a mainnet, even from your own account, can strand fees, break reconciliation, and land on someone's mempool watcher. Use testnet, a local node, or a lab chain. Mainnet interactions are limited to public reads you are authorised to make.
- **Seed phrases, private keys and full keystore files never enter a finding, a fact, the graph or a transcript.** The finding names where the material is, what reaches it, and what an attacker would be able to sign. That is complete evidence.
- **Hashes and dumps stay in the credential store by reference** (`save_task_credential`), and only for material the engagement generated itself.
- **Do not deobfuscate, publish or reuse a real exploit against an operating exchange.** Proof-of-concept code for a withdrawal-authorisation flaw is written against your own lab deployment. If the target is live, the finding is the reach plus the derived consequence.
- **Denominators, not dumps.** An exposed internal API that returns customer balances is proved with one record you are authorised to see and the count of what it returns.

## The layers to map

1. **Customer channel** — web and mobile platform, API keys and signing, 2FA and device binding, support and back-office tooling, email and SMS flows.
2. **Trading and ledger core** — order matching, balances, internal transfer journal, fee accounting.
3. **Custody** — cold storage, hot wallet topology, signer quorum (MPC, multisig, HSM), key ceremonies, address book and allow-lists.
4. **Blockchain interface** — node RPC (geth/besu/bitcoind/solana), explorer services, mempool and indexer services, mempool watchers, bridge infrastructure.
5. **Integration surfaces** — banking rails, FIAT on/off ramps, KYC/AML vendors, liquidity providers, third-party bridges.

The findings live disproportionately in layer 4 and 5: node RPCs bound to the wrong interface, indexers with permissive CORS, admin interfaces to bridge infrastructure, test keys left in production config, and third-party sign-in that mints a session with more scope than it should.

## Method

### 1. Prove who can cause a signature

The whole security of a custody product collapses to one question: what does it take to get this system to sign? Map the path (request → authorisation checks → quorum → broadcast) before testing anything, and record the control at each hop. A finding that says "the withdrawal endpoint lacked one of its four checks" is worth more than one that says "the endpoint was reachable".

### 2. Test the seams between layers, not the layers

- API keys and signing: replay windows, nonce handling, missing request-body binding, key scope escalation, keys that leaked into public config, JS bundles or logs.
- Identity: 2FA bypass at the *action* layer (login is protected, withdrawal is not), device-unbind flows, support-channel social paths, email-change flows that keep sessions alive.
- Ledger: race between two withdrawals, negative or zero-value transfers, fee bypass, rounding into dust accounts, internal transfer endpoints that skip the authorisation the external one performs.
- Node RPC: methods exposed without auth (`admin_`, `personal_`, debug tracing), JSON-RPC on a public interface, websocket hosts not origin-checked, RPC credentials in mounted config.
- Monitoring and indexers: dashboards with no auth that reveal address balances and hot-wallet topology, and therefore the shape of the target.

### 3. Chain the small ones

Crypto incidents are almost never one flaw. The report should read as a path: exposed config → signer service credentials → address allow-list read → a signature request the platform would accept. Each hop needs its own captured evidence and the derived end state: what could be signed, for whom, and how much is reachable by that path.

### 4. Say what is unprovable

Self-custody products fail in the user's hands, not only in the platform's. Where the risk is a design consequence — a signer quorum that lives on one machine, a recovery path that depends on one vendor — label it derived and state the assumption it rests on. Do not dress a design critique as a demonstrated breach.

## Done means

For each custody-relevant finding: which layer it lives in, what an attacker gains at that hop, what checks stand between that hop and a signature, evidence captured by reference (paths, field names, counts, method names — never key material), and the consequence stated in value and population. A withdrawal finding that did not move value is written as reach plus derived impact, and the act you declined is named.
