---
name: iot-protocols
description: Invoke this skill when you encounter IoT/OT protocol ports or devices outside Web/HTTP — CoAP, Modbus/industrial control, BLE Bluetooth, Zigbee/802.15.4, UPnP/SSDP, mDNS/DNS-SD, AMQP, etc. Provides discovery, enumeration, authentication, and authorization testing methods for each protocol; write/control operations are read-only by default and require authorization.
---

# IoT / OT Protocol Testing Quick Reference

Outside the web, IoT/OT devices expose many non-HTTP protocols. This skill provides discovery and assessment methods for **authorized testing** by protocol. MQTT is covered separately in `mqtt-iov`; satellite communications in `satcom-recon`.

---

## Boundaries and Prohibitions (Agent must read)

| Allowed | Prohibited |
|---|---|
| Discovery, resource/service enumeration, reading, and **determination** of authentication/authorization issues | **Write registers/coils** to OT/industrial control devices, change device state, or perform any operation that could affect physical processes |
| Passive sniffing, reading broadcast/discovery responses | DoS against production devices, firmware flashing, factory reset, or provisioning hijacking |
| Small-dictionary weak-credential probing | Large-scale brute force or overloading low-power devices |

> ⚠️ **Industrial control/OT physical-safety red line**: **write operations** for Modbus/BACnet/DNP3 and similar protocols can directly drive valves, motors, or PLCs, endangering site safety and people. **Read-only by default (read coils/registers, discover)**; finding that “unauthenticated writes are possible” is itself a high-severity vulnerability — report it and **do not actually write**.

---

## Port Quick Reference

| Protocol | Port | Description |
|---|---|---|
| CoAP | 5683/udp (5684 DTLS) | REST-like protocol for constrained devices |
| Modbus | 502/tcp | Industrial control, most common |
| BACnet | 47808/udp | Building automation |
| DNP3 | 20000/tcp | Power/water utilities |
| UPnP/SSDP | 1900/udp | Discovery; amplification attack surface |
| mDNS/DNS-SD | 5353/udp | LAN service discovery |
| AMQP | 5672/tcp (RabbitMQ management 15672) | Message queue |
| BLE | — | Requires a Bluetooth adapter |
| Zigbee/802.15.4 | — | Requires KillerBee hardware |

Baseline round:

```bash
nmap -sU -sV -p 5683,1900,5353,47808 <HOST>
nmap -sV -p 502,20000,5672,15672 <HOST>
```

---

## CoAP (5683/udp)

```bash
coap-client -m get coap://<HOST>/.well-known/core      # resource discovery (core entry point)
nmap -sU -p 5683 --script coap-resources <HOST>
```

- Enumerate resources listed in `/.well-known/core` and GET each value.
- Test whether DTLS (5684) is enforced and whether unauthenticated requests are accepted; for PUT/POST/DELETE write operations, only make a **permission determination** (whether they are rejected) and do not actually change device state.

## Modbus (502/tcp, industrial control — read-only)

```bash
nmap -p 502 --script modbus-discover <HOST>            # enumerate unit IDs / device identification
# mbtget / pymodbus read-only: read coils and registers
mbtget -r3 -a 0 -n 10 <HOST>                           # read holding registers
```

- Modbus has no built-in authentication — if it is reachable, it can be read; this is often itself an exposure finding.
- **Never send write function codes** (05/06/0F/10) to production devices; if you need to verify “writability”, do so only on an authorized, isolated test PLC.

## BLE (requires a Bluetooth adapter)

```bash
bluetoothctl scan on                                   # discover devices
gatttool -b <MAC> --primary                            # enumerate GATT services
gatttool -b <MAC> --characteristics                    # enumerate characteristics
gatttool -b <MAC> --char-read -a <handle>              # read characteristic value
```

- Focus on: whether Just Works pairing is used (no MITM protection), whether characteristics expose sensitive data in plaintext for read/write, and whether replay is possible.
- Advanced: `bettercap` (BLE module), `nRF Connect`, `btlejack` (connection hijacking/sniffing; requires authorization).

## Zigbee / 802.15.4 (requires KillerBee hardware)

```bash
zbstumbler                                             # scan channels/networks
zbdump -f 15 -w capture.pcap                           # capture the specified channel
```

- Focus on default/leaked network keys (Trust Center Link Key, the classic `ZigBeeAlliance09`), plaintext joining, and replay of switch/lock commands (only on authorized test devices).

## UPnP / SSDP (1900/udp)

```bash
nmap -sU -p 1900 --script upnp-info <HOST>
upnpc -l                                               # list IGD port-mapping capabilities
```

- Focus on: unauthenticated port-mapping additions (internal-network exposure), device description XML exposing model/serial number, and SSDP reflection/amplification surface.

## mDNS / DNS-SD (5353/udp)

```bash
avahi-browse -a -t                                     # enumerate all broadcast services on the LAN
nmap -sU -p 5353 --script dns-service-discovery <HOST>
```

- Quickly map which devices and services exist on the LAN (printers, NAS, cameras, smart home devices, Chromecast, etc.) and their hostnames/ports, providing a map for subsequent targeted testing.

## AMQP (5672 / RabbitMQ management 15672)

```bash
nmap -p 5672 --script amqp-info <HOST>
# RabbitMQ management console default guest/guest (older versions allowed remote login)
curl -s -u guest:guest http://<HOST>:15672/api/overview | head
```

---

## Companion: Firmware and Hardware (when you have a device/firmware sample)

- Firmware extraction and analysis: `binwalk -eM firmware.bin`, `firmware-mod-kit`; look for hardcoded credentials, certificates, private keys, backdoors, update URLs, and broker/API addresses.
- Hardware debug interfaces (UART/JTAG/SPI Flash): require physical access to the device and fall under hardware testing; this skill only notes their existence.
- Broker addresses/certificates/topic patterns extracted from firmware feed back into `mqtt-iov`; extracted APIs feed back into `api-recon`.

---

## Delivery

- `report_finding` for each confirmed issue: unauthenticated reachability, default/leaked keys, plaintext transport, unauthenticated writes (OT), unauthenticated port-mapping changes (UPnP), default management credentials, and firmware hardcoded credentials.
- `insert_assets` to record discovered service/device assets (protocol, port, model, firmware version).
- For OT/physical findings, **explicitly state in the report that no write operations were performed and only permission determination was done**, and provide remediation recommendations (add authentication, network isolation, disable default accounts/keys).
