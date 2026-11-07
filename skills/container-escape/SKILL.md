---
name: container-escape
description: "Invoke this skill when the authorized foothold is inside a container or pod and the question is which runtime, image or cluster-RBAC boundary reaches the host, the node, or a neighbouring workload — and what can be proven without touching the node at all."
---

# Container Escape (runtime, image and cluster RBAC boundaries)

Three boundaries, assessed in order of increasing impact: the container's own configuration, the node it runs on, the cluster API it is authenticated to. Most engagements are satisfied by proving the first and characterising the second; crossing to the third is a decision, not a reflex.

Done means each boundary carries a verdict with its mechanism named.

## Boundaries

- The namespace, cluster and node set named in the engagement are the whole field. A second cluster reachable with a token you found is a finding about the token, not a new target.
- **Stop before touching the node.** The first pass is read-only: mount table, capability set, token audiences, RBAC verbs. Cross only when an allow-constraint explicitly names host or node actions.
- Even then, prove the crossing with one reversible action under a name you chose (write one file into a host mount, read the host's `/etc/hostname`), not a root shell.
- Never: cordon, drain, reboot or delete a node; write to `/etc`, cron, kubelet configuration or containerd data; delete pods you did not create; mutate RBAC, NetworkPolicy or admission configuration; copy `/etc/kubernetes/pki`, kubelet client certificates or another tenant's pod volumes. Cite those paths; do not read them out.
- Not this skill:
  - the cloud identity the node or pod role carries — `cloud-attack`;
  - the secret values found in image layers — `secrets-keymanagement`;
  - the pipeline that built the image — `cicd-supply-chain`;
  - the runtime or kernel CVE behind a mechanism — `vulnerability-intel`;
  - endpoint-sensor behaviour on the node — `edr-bypass-re`;
  - whether an admission policy truly enforces the securityContext — `mitigation-status`.

## Procedure

1. **Fix your position.** `cat /proc/1/cgroup`, `/.dockerenv`, `KUBERNETES_SERVICE_HOST` in env, `id` for uid/gid, downward-API pod and container names, presence of `docker` or `crictl` sockets. Record runtime and version strings (`runc --version`, `containerd --version`, kubelet version); those feed `vulnerability-intel`, they are not findings by themselves.
2. **Read the confinement instead of assuming it.**
   - `CapEff` and `CapBnd` from `/proc/self/status`, decoded with `capsh --decode=<hex>`;
   - `Seccomp:` line, `NoNewPrivs`, AppArmor or SELinux label from `/proc/self/attr/current`, `aa-status` where readable;
   - the declared `securityContext` from the API, if the token permits it.
   The reportable class is the gap between declared and effective: a dropped capability restored by `CAP_SYS_ADMIN`, one pod in a Deployment running `Unconfined`, `allowPrivilegeEscalation` still true.
3. **Enumerate mounts before writing anything.** `/proc/self/mountinfo` is the map:
   - host paths (`/`, `/etc`, `/root`, `/var/lib/kubelet`, `/var/log`, `/host`);
   - runtime sockets (`/var/run/docker.sock`, containerd or CRI-O socket) — a socket needs no exploit; name it as a node-compromise capability;
   - writable `/proc/sys` and `/sys` entries (`core_pattern`, `modprobe`, cgroup `release_agent`) — identify and report, do not drive;
   - `emptyDir` volumes shared with other containers, and hostPath CSI volumes belonging to neighbouring pods.
4. **Image-layer secrets.** With registry or local image-store access: `docker history --no-trunc` for ENV-baked credentials and `--build-arg` leftovers, the image config's `Env` and `Labels`, and a layer walk (`dive`, or export layers into a scratch dir) for `.netrc`, `.ssh`, `id_rsa`, `tfstate`, `.npmrc` and build-time `COPY` of host keys. Cite layer digest, path and identifier only; value handling follows `secrets-keymanagement`.
5. **Cluster RBAC from the token you already hold**, in-cluster only:
   `kubectl --token="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" --server=https://kubernetes.default.svc --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt auth can-i --list`
6. Judge that verb set one path per class, not as "over-privileged SA": `pods/exec` (neighbouring workloads); `secrets get` beyond the pod's own namespace; `serviceaccounts/token create` (mint another identity — the classic escalation); `clusterrolebindings create` (self-grant); `nodes/proxy`; `persistentvolumes get` (hostPath exposure across namespaces); `/metrics` read (leaks other pods' environment and tokens).
7. **Neighbouring workloads, read-only.** With permissive network policy plus `pods get`, enumerate reachable peers and prove reachability with one non-mutating request per peer class. Cross-namespace reach is a confidentiality finding; you do not need to exec into a neighbour to report it.
8. **Node-adjacent surfaces: look without crossing.** Kubelet 10250 (anonymous auth enabled? `/run` or `/exec` handlers present?), 10255 read-only port (stats expose pod environment and secret material), etcd 2379 without auth, node `/metrics`. Record the endpoint and the fact that it answers; escalate rather than creating a hostPath pod on a production node.
9. **Gate to host-level actions.** Only after steps 1-8 named the mechanism and an allow-constraint covers host actions: execute the smallest escape, capture one proof (identity, hostname, count of visible host mounts), stop, restore. Any pod, file or token created on the node goes into a fact with its ID plus a teardown line.
10. **Disposition each boundary.**
    - Boundary 1: confined, or escape-capable with the mechanism named;
    - Boundary 2: node reachable, route stated, marked attempted or not;
    - Boundary 3: cluster RBAC verdict at verb level.
    Anything declined for safety is written as "available, not executed under this engagement's constraints" — a deliverable, not a failure.

## Evidence

Positive: the mechanism quoted verbatim (`CapEff` value, the mountinfo line, `auth can-i --list` output, layer digest, endpoint response); the crossing proof (one filename, one hostname, one forged token identifier stored by reference); and counts — `3 of 9 pods Unconfined`, `SA token holds 14 verbs, 4 escalation-class`.

Negative: "read `/proc/self/mountinfo`: 11 entries, all overlay or tmpfs, no host paths, no runtime socket. `CapEff 0xa80425fb` with no `SYS_ADMIN`/`SYS_PTRACE`. Seccomp `2 (filter)`. Token verbs via `auth can-i --list`: no `pods/exec`, no `secrets get`. Node boundary NOT ASSESSED — no host path or socket present and kubelet unreachable from this pod network."

## Report shape

`<cluster/namespace>: container boundary <n> crossed via <mechanism> — <reach: node | N pods | cluster verb>`

Fields: pod and container ID, node name, image digest, runtime version, quoted mechanism, counts, executed-or-not, cleanup status.
