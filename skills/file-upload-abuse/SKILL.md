---
name: file-upload-abuse
description: "Invoke this skill when an upload on an authorized target lands somewhere executable, parseable or trusted: extension and content-type validation edges, filename and path handling, storage and CDN exposure, presigned and resumable surfaces, archive extraction (zip-slip), server-side rendering of uploaded content, and the descent path from a stored file to code execution."
---

# File Upload Abuse

The gap between "the server accepted my file" and "something executed, rendered, parsed or served it". Done = the landing zone mapped, the trust that was placed in the stored object demonstrated with one marker, and every artifact you placed accounted for and removed.

## Boundaries

- Authorized use: test only the upload points on the asset the intent names — the application you are engaged to test, not a real team's shared document store and not another tenant's bucket.
- Test with files you created, in storage you can identify, and prove execution on **your own** object: a marker that prints a fixed token, then deleted. Do not leave a handler in a live document root, do not overwrite a real object, and do not write anything into a path you did not create.
- Out of bounds outright: archive bombs and other resource-exhaustion payloads, anything that corrupts or encrypts shared data, defacement, and payloads whose real purpose is dodging antivirus on a live host. Anti-malware and content-dispatcher evasion is characterised as a control observation and detonated in a lab only — describe the consequence, never ship the recipe.
- Every artifact you place gets a cleanup line: exact path or object key, the request that placed it, and the request that removed it.
- Not: the LFI-to-RCE race over PHP session files (php2shell), server-side fetch of a URL parameter (ssrf-metadata), guessable object IDs and cross-tenant reads as an authorisation finding (api-security), browser/Office handler behaviour on the endpoint (client-side), the parser bug in source (codeaudit).

## Procedure

1. Inventory the ingest points on the tasked asset: multipart forms, base64 fields in JSON APIs, `multipart/form-data` GraphQL operations, import endpoints (CSV/XLSX/ZIP/vCard), avatar/logo/attachment fields, chunked and resumable uploads (tus `UPLOAD_OFFSET`/`UPLOAD_LENGTH`), presign endpoints (`/presign`, S3 `PutObject` signed URLs, `getUploadUrl`), rich-text paste with image upload, webhook bodies, and "fetch this URL into a file" flows.
2. Map the contract before probing it, from one honest upload: field name, whether the filename is client-supplied or server-generated, extension/MIME/magic-byte checks, size limits, where the bytes land (document root, object store, temp dir, a queue), the public URL shape, whether the server transforms it (resize, transcode, thumbnail, flatten), whether a scanner or DAST looks at it, and which edge (CDN/WAF) sits in front of the stored URL. The transform step matters twice: an image processor is a parser sink, and a transform usually strips your payload's active part.
3. Validation ladder, one axis per attempt, in this order: extension rules (case variants, alternate handler extensions `.phtml`, `.php5`, `.phar`, `.pht`, `.shtml`, `.asp;.jpg`, `.aspx`), trailing dot and space, null byte on legacy stacks, double extension, name-vs-type mismatch (valid image bytes with a handler extension and the reverse), `Content-Type` swap versus magic bytes, then filename metadata: path traversal (`../`, encoded separators, absolute paths), the RFC 5987 `filename*=UTF-8''...` form, CRLF in the filename, and reserved/odd names (`.`, `.htaccess`, `web.config`, a name that collides with an existing template).
4. Landing-zone analysis — this is the phase gate: do not chase execution until step 3 has told you **where the bytes go and who reads them**.
   - Directly under a handler-enabled document root, served by a host that resolves to the origin, and reachable by URL: the classic execution case.
   - Included later by the application (templates, config, exports, report builders, `include`/`require` of stored data): descent path — that is the LFI chain, hand it to php2shell with your evidence.
   - Reachable by any file-operation on the path (any `file_exists`, `imagecreatefrom*`, `unlink`, extraction, or deserialisation that opens it): a `.phar` you can place becomes object injection through whatever touches it. Prove the touch, not a gadget chain.
   - Extracted by an archive handler: zip-slip (`../` in the entry name), absolute paths, symlink and hardlink entries, entry name versus stored name confusion, and extraction outside the intended directory. Prove it by writing a marker into a directory you created, then read it back and delete it.
   - Rendered or converted server-side (PDF, image pipeline, office converter, preview): the active content is an XML entity, an SSRF or `file:` reference in the source doc, or a template expression in the metadata — report the class, hand the fetch detail to ssrf-metadata.
   - Stored in an object store or behind a CDN: public-read ACL, listing enabled, guessable keys, `Content-Disposition`/`X-Content-Type-Options` absent so a HTML/SVG object renders in origin context, and a signed URL that any reader can re-fetch.
5. Execution proof, minimal and reversible: upload one marker file that emits a fixed token (a handler that echoes a constant, not a shell), request it once, capture the request/response pair, delete it, confirm it is gone with a second request. One marker per landing zone.
6. Renderable-content abuse: SVG and HTML objects served from the app's own origin (stored XSS reaching session material), `Content-Disposition: inline` on user content, filename-driven header injection in the download, and CSV/spreadsheet formula injection into an export an admin opens (prove with a formula that evaluates in your own exported file, and report the read-back path, not a working weaponised formula).
7. Presigned and resumable surfaces: whether the client chooses the key (write outside your prefix, overwrite someone else's object, or place `index.html`/`.html` at a root key), whether the presign endpoint validates content type or size, whether an uncompleted multipart upload leaves a billable or reachable orphan, and whether the resumable offset can be pushed past your own session's boundary. Prove by writing a marker object to a path you can delete.
8. Storage-side enumeration: after your own upload, learn the key scheme and check whether keys are guessable and whether a second account of yours can read another's object — that boundary claim belongs to api-security, so record it as a tasked hand-off with your evidence rather than reporting it twice.
9. Downstream trust: does anything read the uploads directory as configuration or template, is there a job that processes the stored file, does a signed URL get handed to a third party? One read-back is the proof; the rest is a fact for your lead.

## Evidence

Positive: the upload request plus the server's own answer (stored name, key, path, URL), the landing zone as observed, the fetch or render request that produced the effect, the marker token returned (proof of execution/parse, not of a pattern), the count of variants accepted versus rejected with the rejecting layer named (app validation, WAF, scanner, storage ACL), the object ID or path, and the cleanup confirmation with the deleting request.

Negative: "`<METHOD /path>` field `<x>`: 12 variants (extension case, alternate handler, double extension, MIME/magic mismatch, `../` in filename, `filename*`, CRLF) — all rejected at `<layer>` with `<response tell>`; stored as `<server-generated name>` under `<prefix>`; a handler-extension marker at `<url>` returned 404 on 3 attempts; object listing 403 — execution and traversal closed for this input". Where you could not tell who reads the file: "landing path `<dir>` identified, reader unknown — descent path NOT ASSESSED, handed up as a fact".

## Report shape

"`<upload point>` stores `<file class>` at `<landing path/URL>` where it is `<executed|rendered|parsed|served trusted>` producing `<attacker outcome>`" with: endpoint and field, validation observed, landing zone, proving request pair, counts, marker path, and cleanup state.
