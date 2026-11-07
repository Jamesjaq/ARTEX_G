---
name: client-side
description: "Invoke this skill when the question is what executes on an endpoint after it opens something: Office and PDF document handlers, shell/protocol handler chains, mail and gateway inspection paths, browser-side execution chains on an in-scope web app, and the trust boundary an artifact actually crosses."
---

# Client-Side Execution Chains

What happens on the endpoint after a human or a process opens something. Two surfaces share this skill: document and handler chains (delivery through mail/file, executed by the endpoint), and browser-side execution chains on an in-scope web app. Done = a named handler and version chain, a marker proving the primitive in a lab, the control outcome with counts, and an explicit statement of whether the live path was ever taken.

## Boundaries

- Authorized use: run against lab endpoints you control and the assets this engagement names. Delivery to and execution on a real employee's workstation requires an explicit approval naming that host; absent it, demonstrate the chain in your own lab endpoint and state plainly that the live path was not taken — that is the deliverable, not a lesser result.
- You prepare artifacts; a human performs any send, call or visit. Nothing in this skill dispatches anything to a real person, and nothing here widens a recipient list (initial-access and human-terrain own the delivery step).
- AV/EDR, gateway and telemetry behaviour is characterised in a lab you control and reported as a blind spot in the control. No evasion recipes against a live sensor: no tampering, blinding, spoofing or disabling of a production endpoint's protection (edr-bypass-re and payload-lab-development own that side of the work).
- Lab endpoints and files only for anything that can execute, self-replicate, or damage data. No destructive or encrypting payload at any point, in the lab or elsewhere.
- Browser-side: only in-scope web apps, your own accounts and sessions, and self-contained XSS proofs. No token harvest from a real user, no payload staged on infrastructure the engagement does not own.
- Not: the binary internals of a client app (rea), document upload handling inside the target application (file-upload-abuse), the mail-gateway or M365 tenant as a store to attack (messaging-collab), delivery and pretext design (initial-access), the backend contract of a mobile app (mobile-testing).

## Procedure

### A. Establish the endpoint's posture (read-only)

1. Ask what stack actually terminates at the endpoint, from authorised intel, not from a probe: mail client and version, Office build and its protection policy state, browser and version, PDF reader, archive tools, endpoint protection presence and version, OS build. Where a control's policy state decides the outcome (Protected View, attachment previewers, file-type blocking, Mark-of-the-Web handling, add-in policy, AMI), you need it before you can interpret a result — and if you cannot obtain it, say which variable is missing rather than reporting a chain that depends on it.

### B. Artifact and handler chains — lab only

2. Pick the artifact class for the stack you identified: Office (DDE and embedded object, external template reference, custom UI form, add-in load), PDF (document-level JavaScript, launch/embedded-file action, XFA form submission, annotation-triggered viewer code), Windows shell and protocol handlers (`.search-ms`, `.library-ms`, `.settingcontent-ms`, `.url`/`.desktop` with a remote target, `.lnk` icon-loading, WebDAV/SCDA-triggered credential prompts, handler chains that reach a network path from a browser), archive extraction, and the download/protocol-handoff paths a browser hands to a native app. One class per run, driven to a conclusion before opening another.
3. Locate the trust boundary the artifact crosses, and name it in the report: the gateway or filter step that inspects (or does not inspect) the object, the OS/office protection that would warn or sandbox, the point at which a handler is chosen, and what the user actually sees before code runs. The finding is that boundary, not the artifact.
4. Build in the lab: create the artifact in a work-dir script or template of your own, keep the active step a benign marker — a file written to a fixed lab path, or a token printed — and never a durable implant, credential harvester or destructive step. Describe any class you deliberately did not build, and why.
5. Detonate on the lab endpoint. Record the host, OS build, app version and patch state, then each attempt's outcome per control: gateway delivered or quarantined, warning shown or suppressed, handler invoked or blocked, code reached the marker file or not. Use the browser under a controlled driver only where the chain is browser-mediated.
6. Negative telemetry results are results. "This class produced no event in the lab stack" is a detection-coverage gap worth reporting; "here is how to keep the client from seeing it" is not this engagement's deliverable.

### C. Browser-side chains on an in-scope web app

7. Confirm the execution primitive before chasing a payload: identify the sink context (`href`, `src`, `srcdoc`, postMessage handler, template/`v-html`/`innerHTML`, `javascript:`/`data:` navigation, DOM-based router path, `window.name`, hash-driven navigation), whether the value is attacker-controlled at the server or only at the client, and whether your own session executes.
8. Chain to a real consequence: read or write something only an authenticated principal can reach, complete a state change the UI guards, or reach the token/session storage the app keeps in a place script can read (`localStorage`, a non-HttpOnly cookie, a global). CSP is part of the finding only if it does not stop that step; report what it blocked as well.
9. Reproduce deterministically under a controlled browser rather than by hand-clicking, and record the exact browser build. A chain you cannot reproduce in two runs is a hypothesis.
10. Handler trust edge worth naming: app-provided `postMessage` listeners accepting any origin, `window.open` handles retained across origins, cross-origin OAuth callback pages, and desktop protocol handlers registered by the product (`ms-appx`, custom URI schemes) that a web page can invoke — the last of these is a client-side finding when it reaches native code without a state check (compare with mobile-testing).

## Evidence

Positive: artifact or sink class, the named handler or component with its exact version, lab host and OS, the marker proof (path written plus its expected content, or the authenticated read that succeeded), the control matrix with counts (`gateway quarantined 3 of 3 .search-ms; Protected View blocked 3 of 3 DDE; PDF JS opened 3 of 3 with no warning`), and an explicit "lab detonation; live endpoint path not taken" plus the approval reference when a live host was used. For browser chains: the exact request or URL, the executed primitive, and what it read or changed in your own session.

Negative: "`<class>` tested on lab host `<W-1>`, Windows `<build>`, Office `<build>`: gateway quarantined 3 of 3, Protected View blocked 3 of 3 — class closed under those conditions. `<Other class>` NOT ASSESSED: no lab build with `<handler>` installed; `<Mark-of-the-Web path>` NOT ASSESSED: needs an untrusted host the engagement does not own". A result without the named control, version, and window is an assertion, not evidence.

## Report shape

"`Client-side: <artifact class> opens via <handler> on <platform/version> and reaches <primitive>; <control> did not inspect it (<n> of <m>)`" with fields: artifact class, handler and version, lab or live host identity, per-control outcome matrix, marker proof, chain endpoint, and the delivery-status disclaimer.
