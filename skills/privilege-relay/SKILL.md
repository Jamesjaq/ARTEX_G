---
name: privilege-relay
description: "Invoke this skill when a task must answer whether one host's authentication can be spent on another host: coercion primitives, NTLM or Kerberos relay, the signing and channel-binding controls that stop them, and delegation-chain exercises."
---

# Authentication relay and coercion (lab-first)

The question is whether host A's credential can be cashed at host B. The deliverable is a
measurement — which transports accept a relayed authentication, on which hosts, and which
control blocked it — not a foothold on a production server. Everything here is run in a lab
segment where you control both ends, or against hosts the RoE names one by one, and you report
which of the two you used.

## Boundaries

- Relay and coercion redirect or break live sessions on real machines. One coercion attempt
  per primitive per source host, evidence captured, then stop.
- Never propose or perform a change to make a relay work: no disabling SMB signing, LDAP
  signing or channel binding, no relaxing LDAP policies, no re-enabling NTLM. The control
  holding is the finding.
- Signed / EPA-enforced targets are escalated rather than worked around (see step 6). A
  write enabled *by* a successful relay — ACL add, service creation, `secretsdump` — needs
  named approval before the attempt, because the attempt is then the write.
- Not this skill: directory path selection → `ad-attack`; the ACL rights a relayed LDAP
  session could add → `ad-attack` / `persistence-cleanup`; captured hashes and their reach →
  `credential-handling`; moving along once you hold a session → `lateral-movement`,
  `pivot-tunneling`; wired/802.1X and NAC posture → `nac-8021x`; ADCS template classes →
  `ad-pki-certificate`.

## Procedure

1. **Posture survey first, read-only, on every candidate target.**
   `nmap -Pn --script smb-security-mode,smb-protocols -p 445 <targets>` (signing:
   disabled / enabled / required); LDAP: `nmap --script ldap-rootdse -p 389 <dc>` then test
   whether a simple unsigned bind is accepted at all; WinRM:
   `netexec winrm <host>` for the Negotiate/EP verdict; SMB null session:
   `smbclient -L <host> -U '%'`. Write the verdict table to the run dir. If signing is
   `required` on every target, the relay question is answered before you send a packet —
   record that and stop.
2. **Coercion only where authorised.** Primitives with a blast radius: MS-RPRN printer
   (SpoolSV/PrinterBug), MS-DFSNM (DFSCoerce), MS-EFSRPC (PetitPotam class), WebDAV-backed
   UNC forced through `\\<your host>@80\`, DCOM activation. Gate: the source host is one you
   own or one the RoE names; one attempt each; capture the inbound authentication and stop.
   Record which primitive you used even when it produced nothing.
3. **Relay only to targets that step 1 marked signing-not-required.** Set the listener on
   the lab segment (`ntlmrelayx -s <target> --no-smb-server` style, or the equivalent
   netexec relay module), and stop the moment the relay authenticates: record the resulting
   session identity and the fact that it was accepted. Do not continue into an admin-share
   write, service creation, ACL add or hash dump without approval — those are the writes the
   relay exists to enable.
4. **Prefer the read-only relay sink.** Where you only need to prove that a relayed
   authentication is accepted, relay to an HTTP/SMB sink that returns a session you can
   observe (`--servers`, proxied GET of a page you own, or an authenticated null-command
   session) rather than to a sink that mutates state.
5. **Delegation chains.** Constrained delegation: with the delegating account's credential
   pulled by reference from the store, exercise S4U2Self/S4U2Proxy against the SPN it may
   delegate to (`getST -spn <spn> -impersonate <your own test user> -dc-ip <dc>`) and record
   which service ticket you obtained. Unconstrained delegation: the artefact is inbound
   tickets on the delegating host — hand them to `credential-handling`, do not crack here.
   Resource-based delegation requires writing `msDS-AllowedToActOnBehalfOfOtherIdentity` on a
   real computer object: that is an approval gate, not a step.
6. **Escalate rather than improvise on hard targets.** A target that enforces signing or
   extended protection cannot be relayed, so the only ways forward are a control change
   (owned by the defender, never by you) or forcing the attempt against a production host and
   accepting broken sessions and an incident. Both are operator decisions. Hand up the single
   command with the target named and wait. When the control holds, the reportable result is
   "signing required on N/N targets"; say whether that is enforced by policy or is default
   behaviour, because the client needs to know which one they are relying on.
7. **Cross-check the poisoning angle.** Name-based resolution poisoning (LLMNR/NBT-NS/mDNS)
   only on a segment the RoE names, and only after you have recorded that the same hosts
   refuse relay anyway — otherwise you are answering a question nobody asked while sitting on
   the victim's broadcast domain.

## Evidence

- Positive: source host, primitive, transport, target service, the accept verdict, the
  identity the relay authenticated as, and an explicit "no state changed" / list of what was
  touched. Denominators: "coercion attempts 3, inbound authentications observed 1, relay
  accepted 1 of 4 candidate targets, all four checked for signing in step 1".
- Control-effectiveness result: the per-target signing/EPA table with the command that
  produced it.
- Negative with method: "NTLM relay NOT ASSESSABLE — SMB signing required on 6/6 in-scope
  hosts (`smb-security-mode`, <window>), LDAP simple bind refused unsigned on both DCs; no
  relay attempt run, no poisoning attempted (no in-scope broadcast segment)."

## Report shape

`Relay: <coercion primitive> from <source> accepted at <target service> as <identity> because <signing not required / EPA off / delegation usable>`
Fields: primitive, both endpoints, transport + port, posture verdict per host, action taken
after the accept, control that would have stopped it, remediation. Or the inverse finding:
`Control effective: <signing/EPA> blocked relay on N/N hosts`.
