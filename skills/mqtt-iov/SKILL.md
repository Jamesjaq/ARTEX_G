---
name: mqtt-iov
description: Invoke this skill when testing an MQTT broker or an Internet of Vehicles (IoV)/IoT backend (telemetry, TSP cloud, device shadow, command dispatch by VIN/device ID). Covers anonymous access, topic enumeration, ACL authorization bypass, weak credentials, sensitive data in retained messages, TLS configuration, and broker version vulnerabilities.
---

# MQTT & Internet of Vehicles (IoV) Backend Testing

MQTT is the most common message bus for IoT and Internet of Vehicles (TSP/telemetry) backends. Under an **authorized** scope, this skill systematically assesses the security of an MQTT broker and the business topics it carries.

---

## Boundaries and Prohibitions (Agent must read — crossing these is out of scope)

| Scope | Allowed | Prohibited |
|---|---|---|
| **Access** | Test anonymous connections, weak credentials, and client ID conflicts | Large-scale brute force against real credentials or overloading the broker |
| **Topics** | Subscribe for enumeration, read visible messages, inspect `$SYS` statistics | Publish to **control/command** topics messages that could affect physical devices |
| **Authorization** | Verify whether one device identity can read/write another device's topics (cross-VIN horizontal authorization bypass) | Actually trigger actions on another vehicle/device (unlock, start, OTA, steering, etc.) |
| **Availability** | Read and observe passively | DoS, message storms, clearing retained messages, or kicking production devices offline by taking over client IDs |

> ⚠️ **Physical-safety red line**: command topics in IoV/industrial MQTT (such as `.../command`, `.../rpc/request`) may directly drive real vehicles or devices. **By default, only determine "whether publishing is possible" (use a harmless probe message or verify on an isolated test device), and never send commands that produce physical actions to production targets**; if verification is truly necessary, it must have written authorization and be performed in a safe, controlled environment. Finding write-authorization bypass is itself a high-severity vulnerability — you do not need to cause real-world consequences to “prove” it.

---

## Tools

- `mosquitto_sub` / `mosquitto_pub` (`mosquitto-clients` package) — most universal.
- `mqttx` (MQTTX CLI), `mqtt` (HiveMQ mqtt-cli) — support MQTT 5.0 and WebSocket.
- `mqttsa`, `mqtt-pwn` — dedicated MQTT security assessment/enumeration tools.
- `nmap` NSE: `mqtt-subscribe`.
- Ports: `1883` plaintext / `8883` TLS / `9001` (common WebSocket, path often `/mqtt`).

---

## Execution Roadmap

```
Phase 0 identify ports and deployment form → Phase 1 anonymous access → Phase 2 topic enumeration + $SYS fingerprint
  → Phase 3 authentication / ACL authorization bypass → Phase 4 retained messages / sensitive data → Phase 5 TLS + version vulnerabilities
  → summarize + insert_assets + report_finding
```

### Phase 0 — Ports and Deployment Form

```bash
nmap -p 1883,8883,9001,61613,15675 -sV <HOST>
```

Distinguish: native MQTT (1883/8883), MQTT over WebSocket (9001 `/mqtt`, or `wss://host/mqtt` behind nginx), and broker management interfaces (EMQX Dashboard 18083, RabbitMQ-MQTT plugin 15672).

### Phase 1 — Anonymous Access

```bash
# If you can connect and receive anything, anonymous read access exists (a high-severity starting point)
mosquitto_sub -h <HOST> -p 1883 -t '$SYS/#' -v -W 5
mosquitto_sub -h <HOST> -p 1883 -t '#' -v -W 10        # wildcard subscription to all topics
```

- Connection refused / username and password required → proceed to Phase 3.
- If `$SYS/#` is readable, anonymous access is usually enabled (`allow_anonymous true`) — record it as a finding.

### Phase 2 — Topic Enumeration + Broker Fingerprint

```bash
mosquitto_sub -h <HOST> -t '$SYS/broker/version' -v -C 1    # broker and version
mosquitto_sub -h <HOST> -t '$SYS/broker/clients/#' -v -W 5   # online client count
```

- Subscribe to `#` for a few dozen seconds and summarize topic naming patterns: IoV examples often include `vehicle/{vin}/telemetry`, `device/{id}/up`, `tsp/{tenant}/...`, `$share/...` (shared subscriptions).
- Infer **tenant/device identifier positions** (VIN, deviceId, IMEI) from the topic structure — these are the anchors for Phase 3 authorization testing.
- Do not dump raw messages; summarize “which topic families exist and which segment contains the identifier”.

### Phase 3 — Authentication and ACL Authorization Bypass (core)

```bash
# Weak/default credentials (small dictionary, low rate; no large-scale brute force)
mosquitto_sub -h <HOST> -u admin -P public -t '#' -v -W 5
# Common defaults: admin/public, admin/admin, emqx/public, guest/guest, mqtt/mqtt
```

**Horizontal authorization bypass (most critical in IoV)**: use device A's identity to attempt subscribing to or writing device B's topics.

```bash
# With A's credentials, subscribe to B's VIN topic — receiving data means cross-device read authorization bypass
mosquitto_sub -h <HOST> -u <A_user> -P <A_pass> -t 'vehicle/<OTHER_VIN>/#' -v -W 10
# With A's credentials, attempt to publish to B's command topic — use a harmless placeholder message to determine whether the ACL allows it
#   Only verify write permission; do not send real commands. If the broker accepts it (no disconnect/no ACL rejection), it is authorization bypass
mosquitto_pub -h <HOST> -u <A_user> -P <A_pass> -t 'vehicle/<OTHER_VIN>/noop-probe' -m 'authz-test' -q 1
```

Decision points: whether the ACL restricts devices to their own topic subtrees by client-id/username; whether wildcards `#`/`+` are restricted; whether `$share` shared subscriptions bypass the ACL. **Finding cross-VIN telemetry reads or writable command topics for other devices is a critical authorization bypass** — `report_finding` based on that; do not trigger physical actions.

### Phase 4 — Retained Messages and Sensitive Data

- Retained messages are pushed immediately when a subscription starts and often contain credentials, tokens, locations, firmware URLs, or configuration.
- The first messages received from `mosquitto_sub -t '#' -v` are retained messages; summarize sensitive fields in them.
- LWT (Last Will and Testament) messages and IPs/hostnames in `$SYS` may also expose internal topology.

### Phase 5 — TLS and Version Vulnerabilities

```bash
# Check whether TLS is enforced, whether the certificate is self-signed/expired, and whether client certificates (mTLS) are required
openssl s_client -connect <HOST>:8883 -servername <HOST> </dev/null 2>/dev/null | openssl x509 -noout -subject -issuer -dates
```

- Business traffic over plaintext 1883 means transport is unencrypted (a man-in-the-middle can read and modify it).
- Check broker versions for CVEs: Mosquitto, EMQX, HiveMQ, VerneMQ, and NanoMQ all have historical authentication bypass/denial-of-service issues.
- Test management surfaces (EMQX Dashboard, RabbitMQ management console) for default credentials and unauthenticated APIs as well (this returns to web testing methods; api-recon can help).

---

## Internet of Vehicles (IoV / TSP) Focus Areas

- **Identifier-based authorization bypass** is the top IoV issue: VIN/IMEI/deviceId are often predictable or extractable from apps/firmware. If the ACL only checks “connected” rather than “connected as your own device”, it enables horizontal control across a fleet.
- MQTT is often only one link: combine it with TSP cloud HTTP APIs (use api-recon), vehicle app reverse engineering (extract broker addresses/certificates/topic patterns), and firmware analysis (extract embedded credentials with binwalk).
- In the command dispatch chain (App → Cloud → MQTT → Vehicle), focus on whether **each hop verifies device ownership**, not just the app layer.

---

## Delivery

- `report_finding` for each confirmed issue: anonymous access, weak/default credentials, cross-device read/write authorization bypass, plaintext transport, sensitive data in retained messages, unauthorized management surfaces, and known CVEs.
- `insert_assets` to record broker services (host:port, service_name=mqtt, version) and management endpoints.
- Preserve evidence: subscribe/publish commands and broker responses, `$SYS` fingerprint, and request/response evidence for authorization determinations; **do not retain real telemetry privacy content** — record only field names and the authorization fact.
