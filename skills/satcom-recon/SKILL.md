---
name: satcom-recon
description: Invoke this skill when assessing satellite-communications targets — management surfaces of VSAT/satellite modems and ground stations (IP-reachable Web/SNMP/SSH management interfaces), and passive capture and demodulation analysis of downlink signals using SDR (Iridium/Inmarsat/VSAT/AIS/ADS-B/ACARS, etc.). Default is passive reception + management-surface testing; do not transmit.
---

# Satellite Communications (SATCOM) Reconnaissance and Testing

Satellite communications assessments have two complementary paths: **① IP-reachable ground-device management surfaces** (closest to conventional penetration testing, high value) and **② passive capture and demodulation of RF downlinks** (requires SDR hardware, more intelligence/exposure analysis). This skill organizes both paths under the premise of an **authorized assessment**.

---

## Boundaries and Prohibitions (Agent must read — crossing these is out of scope)

| Scope | Allowed | Prohibited |
|---|---|---|
| **RF** | **Passive reception / IQ recording / downlink demodulation** (in permitted receive bands, within authorized scope) | **Transmit any signal** to satellites/uplink links; jamming or uplink replay |
| **GNSS** | Analyze and detect the impact of spoofing/jamming | Actually transmit GPS/GNSS spoofing signals (illegal in many jurisdictions and can endanger aviation/maritime safety) |
| **Content** | Analyze signaling, exposure, and plaintext configuration | Decrypt/disclose third-party private communications; violate local interception/privacy laws |
| **Management surfaces** | Perform conventional penetration testing on VSAT/modem management interfaces within the authorized scope | Perform any testing on unauthorized operator/third-party terminals |

> ⚠️ **Legal red line**: **transmission** in satellite bands generally requires a license. Unauthorized transmission (including GNSS spoofing and uplink jamming/replay) is illegal in most countries and may endanger real aviation/maritime/emergency systems. This skill **defaults to receive-only, analyze-only, and management-surface testing only**. Any TX/spoofing/replay must have explicit written authorization and be performed in a shielded/isolated controlled environment.

---

## Path A: Ground-Device Management Surfaces (priority, conventional penetration testing)

VSAT/satellite modems and ground-station equipment (iDirect, Hughes HN/HT, Gilat, Comtech, Newtec, BGAN/Inmarsat terminals, etc.) usually expose an **IP-reachable management surface**. This is the most realistic attack surface, and the methods are consistent with conventional internal-network/web penetration testing:

```bash
# Identify devices and management services (Web/SNMP/Telnet/SSH/proprietary ports)
nmap -sV -p 22,23,80,161,443,8080 <MODEM_IP>
# SNMP often uses public/private and can expose model, firmware, and configuration
snmpwalk -v2c -c public <MODEM_IP> system 2>/dev/null | head
```

Focus areas:
- **Default/weak credentials** (admin/admin, admin/iDirect, installer, etc.), unauthenticated status pages, and known CVEs for the firmware version.
- Use the **api-recon** skill for interface reconnaissance on management web surfaces, then assess conventional web vulnerabilities (authentication bypass, command injection, configuration download).
- Configuration backups often contain satellite parameters, routing, SNMP community strings, and next-hop internal networks — high-value information disclosure.
- Many terminals expose management interfaces directly to the public internet (they can be located in mapping data by favicon/banner fingerprints); this is often itself a finding.

> VSAT/BGAN **link plans and RF parameters** (frequency, symbol rate, modulation) can be read from the management surface and used to guide passive reception in Path B without trial-and-error frequency scanning.

---

## Path B: SDR Passive Capture and Demodulation (requires hardware)

Hardware: RTL-SDR (can receive Iridium/Inmarsat/AIS/ADS-B in L-band with an appropriate LNA/antenna), HackRF/USRP (wider bandwidth), suitable antennas, and a low-noise amplifier. Software can usually be installed on Kali/Linux via apt/pip.

### B1 — Spectrum and Recording

```bash
rtl_sdr -f 1626000000 -s 2048000 -g 40 iridium.iq      # record IQ (example: near Iridium uplink/L-band)
# Or use gqrx / SDR++ for visual tuning, and inspectrum for IQ constellations and burst structure
```

### B2 — Protocol Demodulation (downlink/broadcast, passive)

| System | Band/carrier | Common tools | What can be observed |
|---|---|---|---|
| **Iridium** | 1616–1626 MHz | `gr-iridium` + `iridium-toolkit` | Frames/signaling, some plaintext paging messages (exposure analysis) |
| **Inmarsat STD-C / AERO** | L-band ~1.5 GHz | `scytale-C`, `JAERO` | EGC safety-net/ship broadcasts, aviation ACARS-over-satellite |
| **VSAT (DVB-S/S2 downlink)** | Ku/C band (requires satellite dish + DVB card) | `TSDuck`, GSExtract-like tools | Encapsulated IP streams may be plaintext — a classic “satellite link leakage” surface |
| **AIS (maritime)** | 161.975/162.025 MHz (including satellite-borne) | `AIS-catcher`, `rtl-ais` | Vessel position/identity (exposure/tracking analysis) |
| **ADS-B (aviation)** | 1090 MHz | `dump1090` | Aircraft position/identity |
| **ACARS** | VHF 131 MHz / satellite | `acarsdec`, `JAERO` | Aviation text messages |
| **Weather satellites** | 137 MHz (APT) / L (HRPT) | `SatDump` | Imagery (a harmless target to verify the receive chain works) |

Method: first use SatDump to receive a NOAA/METEOR weather satellite image to confirm the antenna/gain/sampling chain works, then demodulate the target system. Use decoded signaling/plaintext for **exposure and data-leakage assessment**; do not decrypt private content.

### B3 — GNSS (analysis/detection only)

- Analyze the **impact surface of GPS/GNSS spoofing and jamming** (whether devices have anti-spoofing and whether they blindly trust NMEA).
- **Do not transmit** spoofing signals. If the authorized scope truly requires anti-spoofing verification, use a licensed signal simulator in a shielded room, performed by RF/hardware professionals — this skill does not provide guidance for actual transmission.

---

## Delivery

- Report management-surface findings (default credentials, unauthenticated pages, CVEs, configuration leaks) individually with `report_finding`, and use `insert_assets` to record modem/ground-station IPs, services, and management endpoints.
- RF-side output: list of received systems, demodulatable signaling types, and whether plaintext IP/sensitive broadcasts exist (exposure conclusion), noting receive conditions and limitations; retain raw IQ and decoded content as evidence without disclosing third-party private communications.
- Clearly distinguish “IP-reachable management-surface vulnerabilities” (conventional issues that can be remediated) from “RF exposure” (often design/encryption-absence issues whose remediation involves the link layer).
