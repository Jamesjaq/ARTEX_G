---
name: ot-ics-assessment
description: "Invoke this skill when a control network, PLC/RTU/HMI, historian or process-data path is in authorized scope: active discovery, protocol identification (Siemens S7, Schneider Modicon/UMAS, Modbus, EtherNet/IP/CIP, DNP3, IEC-104, OPC UA, BACnet, GE SRTP, OMRON FINS, Mitsubishi MELSEC, Triconex, Codesys, PROFINET, etc.), Purdue-layer mapping and exposure reasoning. Active assessment is allowed within operator-defined scope; STRICTLY NO write/control/actuation (no writes, no setpoints/mode/force, no start/stop, no program up/download, no process commands)."
---

# OT / ICS Assessment (control networks)

Operational-technology targets answer differently from IT, but you **may actively assess in-scope assets** (scanning, fingerprinting, protocol discovery, read-only queries, auth/config assessment where safe) — the hard boundary is **NO WRITE/CONTROL/ACTUATION**. Done means an exposure-and-consequence picture: what an attacker *with this path* could reach, without ever changing process state.

## Boundaries

- **Authorized, in-scope only** (operator defines scope). If scope/ownership is unclear, confirm before acting.
- **Active allowed, NO write/control/actuation (absolute):** you may do active enumeration, service/protocol identification, read-only queries/reads, and non-destructive assessment. You must **NEVER** write coils/registers/objects/tags, change setpoints/mode/force, start/stop a CPU/process, upload/download programs/firmware, issue operate/direct-operate/select-before-operate/control commands, call actuating methods, or reinitialize/control devices. `industrial-safety-bounds` hard-stops any state-changing action and interaction with SIF/interlocks/actuators.
- **Be mindful of fragility:** active probes can stress sensitive devices — stay within scope/rate/intent; prefer identify/read-only NSE/scripts and verified read-only behavior.
- Consequence reasoning that would require *causing* an event belongs to `industrial-safety-bounds`; physical board/firmware work belongs to `firmware-iot` / `hardware-interface-lab`. No RF/SDR/expensive hardware.

## Procedure

1. **HARD BOUNDARIES (always):** Active assessment is allowed on **in-scope** assets (operator-scoped). **STRICTLY NO WRITE/CONTROL/ACTUATION** — never send: register/coil/object/tag writes, setpoint/mode/force changes, CPU start/stop, program/firmware upload/download, operate/direct-operate/select-before-operate/control commands, or any actuating method/call. Never interact with a SIF/interlock/actuator in a state-changing way. Stay in scope; be mindful of fragile devices (rate/intent). "It answers a read" is never a reason to send a write.
2. **Enumerate & identify (active + passive):** build the asset picture. Use passive sources where available (SPAN/flows/logs/inventory) and **active discovery** in scope (host/port discovery, `-sV` fingerprinting). Record controllers, RTUs, PLCs, HMIs, historians, EWS/gateways.
3. **Protocol identification & read-only discovery (active, no control):** identify what speaks on the segment and fingerprint it with **read/identify-only** methods. Common ICS/OT protocols & vendors (network-level, no RF/hardware): **Modbus/TCP 502** (incl. **Schneider Modicon/UMAS** 0x5A) — modbus-discover/read/ID only, no FC writes; **Siemens S7comm 102** (S7-300/400/1200/1500) — s7-info/s7-enumerate read-only, no CPU control/transfer/writes; **EtherNet/IP CIP 44818** (Rockwell/Allen-Bradley etc.) — enip-info read-only, no CIP writes; **DNP3 20000** — dnp3-info, **no operate/direct-operate/select/control**; **IEC-60870-5-104 2404** — identify, no command/actuation (no C_SC/C_DC/C_RC/C_SE); **OPC UA 4840** — enumerate endpoints/security, browse/read attributes, **no actuating method calls, no process-variable writes**; **BACnet/UDP 47808** — bacnet-info/read properties, no WriteProperty/reinitialize/control; **GE SRTP 18245**, **OMRON FINS 9600 (udp/tcp)**, **Mitsubishi MELSEC 5007/5006**, **Codesys 2455**, **PCWorx 1962**, **Triconex/Tristation 1502** (SIS — identify only, highly sensitive), **Red Lion Crimson 789**, **PROFINET** (discovery). Use nmap NSE info scripts and safe read-only clients; verify behavior is read/identify-only.
4. **Map every asset** to its Purdue layer and IT-side reachability (L0–L4; reachable from enterprise/IT by what path). Record via `insert_assets` / `record_fact`.
5. **Trace IT/OT seams:** boundary devices/crossing paths (data diodes, jump hosts/OEM tunnels, DMZ historians, VPN/remote access into control layer, vendor maintenance). Internet/IT-reachable HMIs/EWS/controllers are findings.
6. **Correlate exposure** to advisories (registered facts/published advisories, not live exploitation): matching vendor/CISA ICS advisory, reachable control path, reconcile observed vs inventory. Advisory-to-banner is inventory; a finding only when the path exists.
7. **Confirm with minimal, non-destructive checks (still no write/control):** active, scoped verification is fine; never escalate a read into a write/control to “prove impact”. The demonstration stops at identification/exposure — never inside process state change.

Gate: do not go to step 5 until 3–4 have produced a protocol-identified, layer-mapped inventory; reachability claims without that map are guesses.

## Evidence

- Positive: the path as measured — source observed (mirror/flow/read-only query), protocol identified, Purdue layer, and reachability. A finding is *exposure*, e.g. "HMI at L2 reachable from L4 without authentication (observed in mirrored traffic over window Z; single read-only banner query confirmed the listener)". Impact is reasoned downstream to the process, never produced.
- Negative (out of scope / not assessed): "Segment X NOT ASSESSED — out of operator-defined scope; inventory derived from provided data/mirror, saw N controllers, protocol posture recorded for M." Absence of testing is recorded honestly when out of scope or not assessed. (Active, read/identify assessment is performed on in-scope assets; no write/control/actuation is ever performed.)
- No live credentials, no vendor proprietary configs, no process setpoint values in prose — cite by reference; a captured secret goes to `save_task_credential`.

## Report shape

Title: `<asset> at <Purdue layer> <exposure/reachability> via <path>`. Fields: observation source (passive/read-only, with window), identified protocol, layer, IT-side reachability, matched advisory/version, consequence statement for the process (routed through `industrial-safety-bounds`), the segmentation/monitoring/patch the client must apply, and an explicit note of anything deliberately NOT tested for safety.
