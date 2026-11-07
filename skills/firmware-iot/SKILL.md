---
name: firmware-iot
description: "Invoke this skill when the target is a firmware image, an embedded service, or the software side of an authorized IoT/embedded device: image extraction and filesystem carve-out, service and component inventory, hardcoded credentials by reference, and update-channel integrity — analysed offline, without touching a device in service."
---

# Firmware & Embedded Software Assessment

Firmware analysis is offline reverse engineering of an image and its services: what is inside, what it speaks, what secrets it ships, and whether its update path is trustworthy. Done means an inventory plus at least one integrity/secret finding proven against the *image*, with every result labelled by the environment that produced it (static carve-out vs emulation vs device) — not a claim that a live device is exploitable.

## Boundaries

- Authorized use only: a firmware image supplied by the client/you own, or the software surface of an in-scope device you have permission to analyse. Extraction from a device you do not own is NOT this skill — that needs physical possession and goes through `hardware-interface-lab`, and never against a device carrying real patients, service, or process (see `cyber-physical-verticals`, `industrial-safety-bounds`).
- Offline-first: analyse the supplied image; do not flash, reboot, or drive a live/in-service device to confirm a result. A version match to a known flaw is exposure you rank and hand up, not something you exploit on a running clinical/industrial/automotive unit.
- Hardcoded keys, tokens, default credentials and certs go to the credential store by reference (`save_task_credential`), cited by ID. Never paste a live secret into a fact, finding, or PDF.
- Radio/over-the-air posture is `radio-sdr`; board-level pin/JTAG work is `hardware-interface-lab`; network exposure of a running service is `network-pentest`. This skill owns the image and its embedded software. For deep static structure, the `rea` MCP server can open the carved binaries/packages.

## Procedure

1. Obtain and fingerprint the artifact:
   - record sha256 of the image, vendor/product/model, and the version/build string; register it as an in-scope asset;
   - confirm it is the image the client intends you to analyse — a wrong-build result misleads the whole report.
2. Extract and carve the filesystem:
   - identify the container/layout with `binwalk -Me <img>` or `unblob`; locate the squashfs/jffs2/ubifs/romfs rootfs and the partition table, then expand them;
   - if the image is encrypted or entropy is uniformly high, record that as a protection finding rather than forcing it — dumping from a live board (`hardware-interface-lab`) may be the only path, and it is gated there.
3. Inventory the software (SBOM):
   - enumerate OS/distro, busybox vs full userland, each service binary + version, the kernel version, and third-party libraries;
   - correlate every component+version against advisory/KEV data; component CVE hits are exposure inventory the lead ranks, not exploits you run.
4. Map services & boot configuration. From `/etc`, init scripts, `inetd`/`xinetd`, systemd/busybox `rc`, and network config, list what listens, on which interface, with what default credentials or open debug (telnet/serial console, dropbear keys, anonymous FTP, unauthenticated HTTP/CoAP/MQTT). Note the web/management UI and any API it exposes.
5. Hunt hardcoded secrets: private keys, CA/cert bundles, default usernames+passwords, API keys, cloud endpoints, backdoor accounts, recovery tokens. Verify the credential actually matches the login path before claiming it works. Store each value by reference (`save_task_credential`) and cite the ID + the in-image file path, never the value.
6. Persistent config and secure-boot posture: inspect shipped config/backup defaults (a factory config that re-enables debug, a plaintext admin hash in `/etc/shadow` or a `uhttpd`/`lighttpd` conf), and the boot chain — is there a signature check, locked fuses/JTAG disable, or verified boot in the image, or does it load any kernel/rootfs? Report weak or absent anti-tamper as a hardening finding; confirm fuse/secure-boot state only from the image, never by blowing a fuse on hardware.
7. Update-channel integrity. Inspect the OTA/firmware update mechanism in the image and test its failure modes statically:
   - does it verify a signature, pin a certificate, check a version/HMAC, fetch over HTTPS?
   - unsigned update accepted? trust-any cert / no pinning? downgrade allowed? cleartext fetch, or an update host that is takeover-able?
   - each is an integrity finding — prove it from the update code/config, not by flashing a device.
8. Emulate before claiming runtime reachability: attempt QEMU (e.g. `qemu-user` or full-system) to see which services bind and answer. Label which environment produced each result — "reachable in emulation", "present in image but not reachable in emulation", "confirmed on device (only if the client ran it)". "Not reachable in emulation" is NOT "not vulnerable" — state it as an emulation limitation, never as a negative on the device.

Gate: do not report a service as exposed until 3–4 establish it is present and 8 (or a client-run device test) shows whether it actually listens.

## Evidence

- Positive: image sha256 + product/version, the exact offset/path/service where the artifact lives, the proof basis (static carve-out vs emulation vs client-run on device) labelled per result, and a measured count for data-class claims ("3 default creds in 2 config files; 1 matched the telnet login path"). For update-integrity: the code/config line that skips the check.
- Negative: "Update signature verification present — image analysis of <updater binary/config> shows an HMAC check with pinned key at <path>; no unsigned-update path found in the analysed image -> assessed effective (static analysis, not flashed on device)." A negative names the method and its limitation, never a blanket "no vulns".
- Secrets cited by ID only; representative strings only; no bulk dump of the filesystem or every key in the image.

## Report shape

Title: `<class> in <product/version> firmware: <component/service/update path> (<proof basis>)`. Fields: image sha256, product/version, artifact location (partition/file/offset), component+CVE (for exposure), the credential reference ID (never the value), update-integrity check that is missing/present, emulation vs device label on each runtime claim, and the remediation (sign/verify updates, remove default creds, pin certs, patch component).
