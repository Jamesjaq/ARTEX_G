---
name: cicd-supply-chain
description: "Invoke this skill when the authorized scope includes build pipelines, artefact repositories, infrastructure-as-code repositories or code-hosting credentials, and the question is whether the delivery path can be injected, redirected, replayed or read by someone outside the build system."
---

# CI/CD and Supply Chain (pipelines, artefact repos, IaC, code hosting)

The delivery path turns source into running infrastructure, so a hole here converts a low-privilege position into deployment rights. Assess four surfaces: pipeline definition and triggers, runner reach, package resolution, and signature enforcement.

Done means each surface has a verdict with the specific file, claim, resolver order or verification policy named.

## Boundaries

- Work the repositories, organisations, projects, registries and runner pools named in the engagement. A second tenant's registry reachable through a token you found is a finding about the token.
- Injection proofs use a benign marker: an echoed constant, a file written in the runner's workspace, a request to an out-of-band endpoint you operate.
- No data staged out of the runner, no persistence, no modified artefact shipped downstream.
- Never publish anything into the target's own package namespace, never rotate or revoke pipeline credentials, never merge or approve a change to a production pipeline, never delete a build cache or artefact. Proof pipelines run in a fork or project you control.
- Not this skill:
  - application source review — `codeaudit`;
  - the cloud identity a runner assumes once you are through — `cloud-attack`;
  - secret values found in variables or artefacts — `secrets-keymanagement`;
  - the image's runtime boundary — `container-escape`;
  - endpoint telemetry on build hosts — `edr-bypass-re`.

## Procedure

1. **Inventory the delivery path** and write it as an artifact: SCM host and org, pipeline engine (GitLab CI, GitHub Actions, Jenkins, Azure DevOps, Tekton/Argo), runner or agent pool per engine, package registries per ecosystem plus their auth, artefact stores, IaC repos, and the deploy target each maps to. Record which are in scope; do not test the rest.
2. **Read the pipeline definitions as deployed configuration**: `.gitlab-ci.yml`, `.github/workflows/*`, `Jenkinsfile`, `azure-pipelines.yml`, Argo/Tekton task specs. Hunt untrusted values interpolated into an executed shell:
   - PR title, branch name, commit message, label text, `github.event.*` context fields, webhook payload values;
   - job or stage names, and `git log` output fed into `sh -c`;
   - Jenkins Groovy reaching `evaluate` or `GroovyShell` outside the script-security sandbox;
   - GitLab rules interpolating a merge-request field into a `script:` line.
3. **Check trigger topology before claiming reachability.** `pull_request_target`, `workflow_run`, `issue_comment` and fork-PR-with-secrets events run trusted code with privileged tokens — pair each with whether the body then checks out or executes PR code. A definition that looks unsafe but fires only on `push` from a protected branch is a hardening note, not a path.
4. **Prove injection in your own copy of the pipeline**, not the target's live one: fork or scratch project, same engine, same trigger event, one benign marker step. Record which variable expanded into which command line and the marker's observed value. Touch the target's pipeline only under an explicit allow-constraint covering it.
5. **Runner reach.** This decides the blast radius of a step-4 finding:
   - runtime socket or CRI path mounted into the build container, privileged or host-networked runners;
   - egress from the runner, including whether cloud metadata endpoints answer;
   - the role its OIDC federation assumes — read the trust policy's `sub` and `aud` conditions and report repo-org-only or environment-free claims;
   - whether one self-hosted pool is shared across projects (cross-project token and cache bleed);
   - where registration tokens live and how long retired runners stay registered.
6. **Pipeline credential surface.** Enumerate variable names and scopes only (masked or unmasked, protected or environment-gated, org versus repo level); value handling follows `secrets-keymanagement`. Check the standard leak paths: secrets echoed in setup logs before masking, written into job state or restored caches, embedded in uploaded artefacts and build outputs, repeated in PR comments or status messages, and per-build cloud tokens (`id_token`, installation access tokens) whose assumed role carries no expiry or audience binding.
7. **Dependency confusion, proved inside an authorized namespace.**
   - Enumerate the private namespace and its scopes: `.npmrc` and lockfile registry pins, `pip` index configuration and `--extra-index-url` precedence, NuGet and Maven `settings.xml` source order, and which internal names appear in manifests but are absent from the private registry.
   - Publish a replacement only where authorized: an explicitly reserved test name, or your own registry configured identically in a scratch project.
   - Prove resolution preference with a dry run against a copy of the real configuration: `npm install --dry-run`, `pip download -d scratch/ --no-deps`, `mvn dependency:resolve`, `nuget list`.
   - The evidence is the selected version, source URL and integrity hash in that output. No postinstall payload, no exfiltrating install hook, no code-execution claim beyond what the resolution proof supports.
8. **Signed artefacts: check that someone verifies.** Test the signature exists (`cosign verify`, `notation`, `gpg --verify`, `npm audit signatures`, OCI attestation labels, apt and NuGet package signatures) **and** test that the deployment path enforces it: an admission policy verifying signatures with a non-empty identity and issuer regex, Kyverno or Gatekeeper policy, Helm chart provenance checks, package-manager integrity pinning. Artefacts signed while nothing verifies is the finding; name the verifier you looked for and where.
9. **IaC repositories.** Read provider and module configuration for the assumptions they encode: roles broader than any consumer uses, hardcoded account IDs, wildcard principals in provider blocks, `terraform.tfstate` committed with secrets (hand to `secrets-keymanagement`), applies without a review gate, and module sources pinned to mutable refs (`?ref=main`, no version constraint) so any push to that repo enters the build.
10. **Disposition.** Each surface from step 1 ends as path-proven, configuration-weak with the line cited, or not assessed with the missing access named — and the source-to-production chain gets one verdict per hop.

## Evidence

Positive: file and line of the interpolated value; the trigger event plus the token it receives; the marker observed in runner output; the trust-policy claim quoted; `selected <name>@<version> from <registry>, expected <private registry>` with hash and lockfile diff; counts such as `12 of 40 pinned internal names absent from the private registry`.

Negative: "reviewed 4 workflow files and 2 GitLab CI specs: 0 untrusted interpolations into `run`/`script` blocks, 0 `pull_request_target` bodies executing PR code; hosted runner pool with no shared self-hosted reuse; `cosign verify` passes on 18 of 18 production images but no admission verifier exists, so signature enforcement NOT ASSESSED beyond presence."

## Report shape

`<project/registry>: <surface: pipeline|runner|resolution|signature> flaw via <mechanism> — <deployment rights or artefact count obtained>`

Fields: engine or registry, repo and file:line, trigger event, token or role assumed, proof marker, verifier checked, cleanup of fork and test package.
