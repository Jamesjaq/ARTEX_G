---
name: initial-access
description: "Invoke this skill for consented foothold work against named recipients: campaign design, technical delivery paths, and the human-approval gate that every people-facing vector must pass before anything is sent."
---

# Initial Access

The step from "we can reach them" to "we are inside". Two families of work live here: the
technical surface an exposure turns into a session, and the campaign that uses the delivery
mechanisms the technical stack already trusts. Done means one foothold, proven and recorded,
with the path that produced it written into the graph — not five attempts that left indicators.

## Boundaries

- Authorized targets only: hosts and recipients named in the engagement scope. The scope
  gate and the operator's approval are not advisory here — an access attempt against anyone
  outside them is an incident, not a test.
- **People-facing vectors are prepared, never executed by this seat.** Any step where a real
  person is the final actor — mail they open, a call they answer, a page they visit — ends
  with `ask_operator` naming the recipient, the method, the window and the consent that
  authorises it. `human-terrain` owns that side.
- No impersonative audio or video generation, and no credential material belonging to a real
  third party in a lure. Template content uses engagement-owned identities and lab hosts.
- Delivery infrastructure is engagement-owned and aged for the campaign; `campaign-infrastructure`
  owns stand-up and reputation, `c2-ops` owns listener hygiene. This skill does not configure either.
- This skill is not the exploitation phase for a service exposure. A public CVE against a
  listening port is `infra-exploitation` or `vulnerability-intel`.

## Procedure

1. **Establish the authorized vector set.** Read the engagement constraints and the RoE before
   designing anything. If the RoE is silent on people-facing testing, the vector list is
   technical-only and the gap goes in the report as NOT ASSESSED.
2. **Rank the real entry surfaces from what the map already says.** From the graph: exposed
   management interfaces and VPN/edge appliances (`edge-appliance`), reachable services with
   known-exploitable flaws (`vulnerability-intel`), mail and collaboration exposure
   (`messaging-collab`), SSO/authentication federation weakness (`auth-testing`), and supply
   or third-party reach (`cicd-supply-chain`, `fourth-party`). A foothold obtained through an
   unpatched exposed service is cheaper and quieter than a campaign — check it first.
3. **Design the technical delivery path only against an asset you already hold or a lab
   stand-in.** Macro/OLE, LOBIN abuse, file-type handler chains and browser surfaces are
   documented and detonated in the lab (`payload-lab-development`, `client-side`). What is
   reported about a live endpoint is whether the control caught it — see the lab-only rule below.
4. **Build the lure against the assessment's purpose, not the person.** The objective is to
   measure a control: does the mail stack quarantine an unexpected attachment class, does the
   proxy block the newly registered domain, does the endpoint stop the handler chain. Content
   stays generic to the role, cites an engagement-owned identity, and never clones a real
   vendor's page with a real customer's credentials in it.
5. **Pre-flight check, every time, before a send is even proposed:** recipient is in scope and
   named; sender identity is engagement-owned and aged; infrastructure has been reputation-
   checked in the last window; the RoE window is open now; a deconfliction contact exists if
   the target's SOC calls; the payload has a stop condition and a teardown record. Any "no"
   stops the plan.
6. **Hand the approval request up, then wait.** One `ask_operator` per campaign step, with the
   five facts above in the message body. A lead approves to the operator, never to itself.
7. **On a foothold, switch skills immediately.** First-ten-minutes posture capture is
   `post-exploitation` and `endpoint-discovery`; the map is `internal-mapping`. Record the
   entry vector as an asset-plus-fact pair so the next seat does not repeat the attempt.
8. **Register indicators as you go.** The domain, sending IP, lure hash and message ID go into
   the engagement's indicator ledger (`opsec-tradecraft`) so a later phase does not reuse them
   by accident and so the blue contact can be given them at deconfliction time.

## Evidence

A positive access result must carry: the named in-scope recipient or asset, the vector and
why it was authorized, the timestamp window, the foothold identity (host, user, session type),
the first posture read that proves reach, and evidence pointers (traffic/blob IDs, node IDs) —
secret material by credential-store reference only.

A negative is worth reporting only with its method: vectors attempted, count, over what window,
which control stopped each class. "Sent 3 variants over 2 days; 2 quarantined at the gateway,
1 delivered, 0 executed — endpoint control effective" is a finding. "Phishing did not work" is not.

Live-endpoint detection evasion stays lab-only: report the *control's blind spot* as a finding,
never a recipe for defeating a production sensor.

## Report shape

`Initial access via <vector> against <in-scope asset/recipient>` — fields: authorization
reference, vector class, technical root cause (exposure, patch state, or missing verification
step), foothold obtained, indicator list handed to the blue contact, and the recommendation
owner.
