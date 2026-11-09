---
name: api-recon
description: Invoke this skill when collecting API endpoints from a website.
---

# API Recon (Frontend API Reconnaissance)

Under an **authorized** scope, discover as completely as possible: **backend APIs** (paths, methods, parameters, response bodies), **frontend routes**, and **UI feature triggers** (tabs, modals, table actions, etc.).

---

## Boundaries and Prohibitions (Agent must read — crossing these is out of scope)

This skill **only performs API / parameter-surface reconnaissance**. It is not a vulnerability discovery or penetration exploitation phase.

### Task Boundaries

| Scope | Allowed | Prohibited |
|---|---|---|
| **Target** | Enumerate paths, methods, parameters, routes, and UI triggers | SQLi/XSS/authorization bypass/brute force/fuzzing for vulnerabilities, request tampering attacks, destructive operations |
| **Authentication** | Use hooks + stubs/mocks to bypass **client-side** login gates | Ask the user for or guess account credentials; attempt real login form submission |
| **Runtime** | Hook endpoints without credentials and use mock responses to let the SPA enter the post-login shell | Workflows that depend on a real backend session to continue |

### Dynamic Analysis Without Credentials (Phase 3 default)

1. Use `preload.js` / `runtime_harvest.js` to **intercept and stub** login, permission, menu, and other bootstrap endpoints;
2. Return **structurally correct, business-success, possibly empty** mock bodies for business query endpoints;
3. Allow the frontend to render post-login pages even with no backend or 401 responses, so it triggers more XHR/fetch/WebSocket requests;
4. **Empty data, blank tables, and placeholder UI are expected** — do not switch to real login or vulnerability testing because of them.

**In one sentence**: use mocks to expand frontend routes and component mounting, and **record only outbound requests**; what the backend returns does not matter — what matters is which endpoints the frontend **still sends**.

### Hard Process Prohibitions

| Prohibited | Alternative |
|---|---|
| Grep/curl/Read the main entry `index-*.js` to extract API paths before Phase 1 is complete | Run `OUTDIR/harvest_static.py` |
| Write ad hoc scripts such as `extract_apis.py` to replace harvest | Modify `OUTDIR/harvest_static.py` and rerun it |
| Repeat the same failed grep/command ≥2 times | Change strategy: read tool logs, modify harvest, or consult the reference |
| Skip gates A/B and run the original `scripts/` versions directly | Copy into OUTDIR and adapt them to the target |
| Real usernames/passwords, OTP, OAuth, or other authentication | Stub/mock (see above) |
| Skip stubbing to obtain real data and perform authorization bypass/injection testing | Record only outbound requests; this stays within recon boundaries |
| Irreversible operations such as delete, export sensitive data, or bulk writes | The same applies to coverage clicks |
| Claim all pages and endpoints are discovered without completing runtime + dynamic enumeration | See “Definition of Done” or document limitations |
| Claim all parameters are understood without completing the parameter trigger matrix + diff | Phase 3b matrix + Phase 5 diff |
| Infer required/optional parameters from a single runtime sample | Use multi-sample diff or infer from validation rules/errors |

---

## Two-Layer Model + Execution Modes

| Layer | Output | Ceiling |
|---|---|---|
| **Static** (JS bundles) | Full endpoint paths, draft routes, candidate fields from request construction points | No HTTP methods; parameters require Phase 1b; misses URLs assembled at runtime |
| **Runtime** (live session) | Methods + bodies + responses + dynamic URLs + WS/SSE; multi-sample diff supplements parameters | Pages must actually render to issue requests; a single sample is insufficient to determine required/optional |

| Execution mode | Engine | Use case |
|---|---|---|
| **depth** | `runtime_harvest.js` (Puppeteer) | API inventory, METHOD/params/response bodies, WS/SSE, reproducible batch runs |
| **coverage** | browser + `preload.js` | Click tabs/modals/tables for deeper feature coverage |
| **both** | Run depth first, then coverage | Most complete, longest runtime |

**Parameter methodology** (no universal script): find paths with harvest/regex; find parameters with **anchor expansion + UI binding chain + multi-sample diff + error inference** (grep recipes in [reference.md](reference.md), section J).

---

## Definition of Done

Recon may be claimed complete only when all of the following are satisfied:

- [ ] **Static**: Phase 1 harvest produced `api_static.txt`, `routes.txt`, and `js/`
- [ ] **Runtime**: at least one of depth or coverage; coverage/both must complete **hook validation + dynamic enumeration loop**
- [ ] **Entered shell**: accessing a business path does not redirect back to `/login` (note hash routing)
- [ ] **Parameters**: coverage/both completed the parameter trigger matrix + `param_samples.json`; Phase 5 merged `params_merged.json`
- [ ] **Depth** (if module pages are blank): Phase 4 permission-tree reconstruction ran and Phase 3 reran until **module-level APIs** appeared (not only locale/bootstrap)
- [ ] **Delivery**: Phase 5 outputs are complete (see the Phase 5 output table); `insert_assets` wrote all discovered service and endpoint assets

---

## Scripts and Gates

All files in `scripts/` are reference templates only; **do not** run the originals directly and treat them as final results.

**Rule**: read first → adapt to the target → write into `OUTDIR` (such as `recon/`) → record `CHANGES.md`; if they do not fit, rewrite based on the methodology and borrow only the structure.

| Gate | When | Reference script → OUTDIR copy | Common required changes |
|---|---|---|---|
| **A (static)** | After Phase 0, before the **first** harvest/spider run | `harvest_static.py` / `spider_mpa.py` | **Most sites can run with the default regex**; change endpoint regex, webpack/Vite `publicPath`, MPA exclude/cookies only when the manifest/dialect does not match |
| **B (runtime)** | After Phase 2, before running depth/coverage | `runtime_harvest.js` / `preload.js` + `config.json` | Cookie/localStorage keys, neutralize success values, stubs, login regex, API prefix, hash/history |

**Mandatory SPA order** (not interchangeable; phase numbers take precedence over “explore first, script later”):

| Step | Must do | Prohibited |
|---|---|---|
| After Phase 0 | Next Bash command = `python3 OUTDIR/harvest_static.py <URL> OUTDIR` | curl/grep/Read the main entry `index-*.js` (usually >500KB) |
| Gate A | Copy the script → make small changes if needed → **run it immediately** | Manually extract APIs first, then decide whether to harvest |
| Before Phase 1 completes | Validate output with `wc -l`; if 404s occur, fix harvest and retry | Write ad hoc extract scripts; repeatedly grep URLs that were not downloaded |
| From Phase 1b onward | Grep only `OUTDIR/js/*.js` | Use the main bundle instead of harvest |

- ✅ Copy `harvest_static.py` → (optional) adjust regex → **run immediately**
- ❌ curl the main bundle → grep multiple times → write a temporary extract script → harvest last
- **MPA**: after Phase 0, the next Bash command = `python3 OUTDIR/spider_mpa.py ...`

---

## Tools and Output Constraints

| Constraint | Explanation |
|---|---|
| Large files | `index-*.js` files >100KB must **not** be Read/grepped into context; use OUTDIR scripts for batch processing |
| Grep output | Must include `\| head -20` or `-m 5`; keep only path summaries in conversation, do not paste bundle fragments |
| Validation | Use `wc -l`, `ls \| wc -l`; do not Read entire directories |
| Regex probing | Optional, ≤1 time, only on small chunks ≤50KB or HTML; formal static analysis must use harvest |
| Reference | Recipes/templates/troubleshooting are in [reference.md](reference.md); do not repeat the full text inline |

---

## Execution Roadmap

```
Phase 0 classify + OUTDIR
  → Gate A → Phase 1 harvest (★ run immediately ★)
  → Phase 1b parameter reverse engineering
  → Phase 2 auth three gates → config.json
  → Gate B → Phase 3 runtime + parameter matrix
  → Phase 4 permission tree (if needed) → rerun Phase 3
  → Phase 5 merge report + insert_assets for all discovered service and endpoint assets; never omit discovered assets when inserting
```

Check items in order; **do not enter the next phase until the previous one is complete**.

1. [ ] **Phase 0**: classify SPA/MPA; create `OUTDIR` → [Phase 0](#phase-0--classification)
2. [ ] **Gate A + Phase 1**: copy scripts → **run** harvest immediately → validate with `wc -l` → [Phase 1](#phase-1--static)
3. [ ] **Phase 1b**: anchor expansion + binding layer → `param_candidates.json` → [Phase 1b](#phase-1b--parameter-reverse-engineering)
4. [ ] **Phase 2**: auth three gates → `config.json` → [Phase 2](#phase-2--auth-three-gates)
5. [ ] **Gate B**: adjust runtime scripts → [Phase 3](#phase-3--runtime)
6. [ ] **Phase 3**: depth / coverage / both; confirm shell entry; parameter trigger matrix → `param_samples.json`
7. [ ] **Phase 4** (if needed): permission tree → patch stubs → rerun Phase 3 → [Phase 4](#phase-4--permission-tree-reconstruction)
8. [ ] **Phase 5**: merge outputs + report + `insert_assets` → [Phase 5](#phase-5--merge-and-report)

---

## Phase 0 — Classification

Fetch the entry HTML and **create `OUTDIR`** (do not modify `scripts/` inside the skill):

- **SPA**: empty shell + `<div id=app>` + chunks → Phases 1–5
- **MPA**: SSR + `<form>`, no endpoint bundle → after Gate A:

```bash
python3 recon/spider_mpa.py <BASE_URL> <OUTDIR> [--cookie "session=..."] [--max 300] [--depth 5] [--exclude "logout|delete|destroy"]
```

Outputs: `forms.txt`, `links.txt`, `api_inline.txt`. If an SPA yields ~0 forms, switch to Phase 1.

---

## Phase 1 — Static

Follow [Scripts and Gates](#scripts-and-gates) and [Tools and Output Constraints](#tools-and-output-constraints).

```bash
python3 recon/harvest_static.py <BASE_URL> <OUTDIR>
```

Harvest: parse HTML scripts → webpack/Vite manifest → download all lazy chunks → produce `js/`, `api_static.txt`, `routes.txt`, and `chunkmap.txt`.

```bash
wc -l OUTDIR/api_static.txt OUTDIR/routes.txt
ls OUTDIR/js | wc -l
```

- Compare chunk count vs manifest: if 404s occur, fix harvest and retry; do not manually curl individual chunks
- If `api_static.txt` is too small → loosen the endpoint regex inside OUTDIR and rerun (see reference)

### Phase 1b — Parameter Reverse Engineering

Paths come from Phase 1; parameter fields require separate recon. Follow the grep rules in [Tools and Output Constraints](#tools-and-output-constraints).

**Completion criteria**: for important endpoints, answer — field names, transport location, type inference, required or not, sample values, and confidence.

#### 1b.0 — Transport Forms

| Form | Where parameters live | Prioritize in static analysis |
|---|---|---|
| REST JSON | body + query | `(params\|data\|body)\s*:\s*\{` near the path anchor |
| GraphQL | `variables` | gql templates, `$page: Int` |
| Traditional form | urlencoded | `<form>`, `FormData` |
| File upload | multipart | `FormData.append` |
| Path parameters | `/user/:id` | route table + `useParams` / `$route.params` |
| Encrypted/signed | wrapped in `sign`/`data` | Hook the encryption function input (reference section D) |

Output: label each endpoint with `transport: query|json|form|graphql|encrypted`.

#### 1b.1 — Anchor Expansion

Use a known path as the anchor and expand the window to find request-construction objects:

```bash
grep -n '"/api/user/list"' OUTDIR/js/*.js | head -20
grep -rhoaE '.{0,120}("/api[^"]+").{0,200}' OUTDIR/js/*.js | head -20
grep -rhoaE '(params|data|body|payload)\s*:\s*\{' OUTDIR/js/*.js | head -20
```

| Wrapper layer | Parameter clues |
|---|---|
| axios instance | `data` / `params` |
| unified request helper | global fields injected by interceptors |
| OpenAPI client | generated method signatures |
| React Query / SWR | hook second argument |
| Vue composable | composable input |

Type remnants: `yup`/`zod`/rules, `Form.Item name=`, embedded Swagger.

→ `param_candidates.json`: `{ path, fields[], source: "static-callsite", confidence }`

#### 1b.2 — Binding Layer

```
Form field → onFinish/handleSubmit → transform → API payload
```

| Binding source | Technique |
|---|---|
| form submit | follow submit → transform → API |
| table search | `getFieldsValue()` → `params` |
| route | `:id` / `?tab=` |
| interceptor | global `tenantId`, pagination, sign |
| enum select | `options` → API enum values |

In DevTools, trace the call stack upward from `fetch`/`XHR.send` to the request-construction function.

#### 1b.3 — Three Request-Construction Questions (≠ Phase 2 auth gates)

| Question | What to answer |
|---|---|
| **Assembly** | Where is the payload built, and what transform traces exist |
| **Validation** | Required, pattern, enum |
| **Transport** | path / query / body / multipart / headers |

While reading the interceptor gate (Phase 2), also note globally injected fields (Authorization, `X-Tenant-Id`, sign).

#### 1b.4 — Hand-off to Phase 3

Candidate fields come from static analysis/the binding layer; **required/optional/conditional dependencies** must be determined through the Phase 3 parameter matrix + diff + Phase 5 error inference.

---

## Phase 2 — Auth Three Gates

Grep in `OUTDIR/js/` (with `head`) and write `config.json` (recipes in the reference):

| Gate | Question | Keywords |
|---|---|---|
| **Render gate** | How is “logged in” determined? | `isLogin`, `getToken`, Cookie/localStorage |
| **Interceptor gate** | What triggers a redirect to `/login`? | `response_code`, `errno`, axios interceptor |
| **Content gate** | Where do menus/permissions come from? | `menu`, `permission`, `role`, `acl`, `routes` |

Do not treat localStorage key names as credentials — confirm them from chunks/request chains.

**Exit = Gate B**: land the conclusions in `config.json` and adapt `OUTDIR/runtime_harvest.js` / `preload.js`.

### Phase 2b — API Observation (optional)

Use `preload.js` in OUTDIR to confirm session key names, Authorization, and nested API URLs:

| Setting | Output |
|---|---|
| `recordDetail: true` | `__API_RECON_DETAIL__` |
| `observe.xhrHeaders: true` | header observation |
| `extractUrlsFromResponse: true` | child APIs inside responses |
| `observe.storageReads/cookieReads: true` | feed back into config |
| `neutralizeVueRouter: true` | `__API_RECON_ROUTES__` |

Export in each coverage round: `__API_RECON_LOG__`, `__API_RECON_DETAIL__`, `__API_RECON_ROUTES__`, `__API_RECON_OBSERVE__`.

---

## Phase 3 — Runtime

Gate B must be complete; follow [Boundaries and Prohibitions](#boundaries-and-prohibitions-agent-must-read--crossing-these-is-out-of-scope) and the no-credential mock strategy.

Set `"runtimeMode": "depth" | "coverage" | "both"` in `config.json` (template in the reference).

### Hooks and Stubs (shared by depth + coverage)

| Layer | Scope | Purpose |
|---|---|---|
| L1 exact | auth/permission/bootstrap stubs | pass the initial auth gate |
| L2 negative correction | all JSON responses | rewrite unauthenticated codes to success |
| L3 fallback | `/api` requests not matched by L1 | return an empty success body to expand the UI |

- **depth**: fake auth + rewrite business codes with `forward` + `stubs`; traverse `routes` (hash/history); output `runtime_api.json`
- **coverage**: inject `preload.js` at **document-start** (CDP `addScriptToEvaluateOnNewDocument` or userscript)

Verification: `window.__API_RECON_PRELOAD__` exists; business paths do not redirect back to `/login`.

```bash
cd recon && npm install
node runtime_harvest.js config.json
```

### 3b — Coverage Dynamic Enumeration (required)

1. Main navigation/sidebar — click each item and wait 1–3s for network activity
2. Tabs — `role=tab`, `.ant-tabs-tab`
3. Tables — first-row view/edit/details
4. Toolbars — export, filter, create (**avoid irreversible delete actions**)
5. After entering each module — merge APIs/routes
6. SPA — use controlled `pushState` for paths in `routes.txt` that were not covered (do not do this for MPAs)

**Parameter trigger matrix** (required): record once per operation type in each module and **diff multiple samples**:

| Operation | Parameters usually added |
|---|---|
| Initial list load | pagination + default filters |
| Click search | keyword, filter |
| Advanced filters | more optional fields |
| Create/edit | full entity |
| Bulk/export/sort | `ids[]`, `exportType`, `sortField` |

**Under stubs, outbound body/headers are still real** — rely on the request. Record into `scan_raw.json`, `param_samples.json`, and `api_detail.json`.

- **Vue**: `neutralizeVueRouter: true` + document-start preload
- **React**: `routes.txt` + sidebar clicks + `pushState`
- **both**: run 3a depth first, then 3b coverage

---

## Phase 4 — Permission Tree Reconstruction

**Trigger**: module pages are blank / each route shows only bootstrap requests (such as locale) → the content gate has not passed.

| Symptom | Meaning |
|---|---|
| Shell entry succeeded | Render gate + interceptor gate passed |
| Sidebar items missing/clicks are blank | stub shape or permission codes are incomplete |
| Same few APIs on every route | `v-if permission` checks did not pass |
| `routes.txt` has far fewer routes than the bundle | supplement from the auth module |

```bash
grep -rhoaE '"/api[^"]*(permission|perm|role|menu|acl)[^"]*"' OUTDIR/js/*.js | sort -u | head -30
grep -rhoaE 'userRouteAuth|getResultTree|routeMap|routeLink|menuList|authList' OUTDIR/js/*.js | head -20
```

Typical chain: `role_permissions` (flat codes) + `permissions/all` (tree) → `getResultTree` → `userRouteAuth[CODE].url`.

```bash
python3 recon/extract_route_map.py recon/js recon/
python3 recon/build_perm_tree.py recon/js recon/ --config recon/config.json
```

Intermediate outputs: `route_map.json`, `userRouteAuth.json`, `permissions_tree.json`, `*_stub.json`, `perm_codes_all.txt`.

Stub checks: outer `response_code` matches the interceptor gate; flat codes align with the tree; `routes` covers all links in `route_map`.

After updating `config.json`, **rerun Phase 3**. For large SPAs, tune `waitUntil`, `routeTimeout`, and `perRouteMs` (see reference sections A3/I).

---

## Phase 5 — Merge and Report

### Output Table

| File | Stage | Content |
|---|---|---|
| `js/`, `api_static.txt`, `routes.txt`, `chunkmap.txt` | 1 | Static bundles and paths |
| `param_candidates.json` | 1b | Static parameter field candidates |
| `config.json` | 2 | Three gates + runtime configuration |
| `runtime_api.json` | 3a | Detailed depth recording (including WS/SSE) |
| `param_samples.json`, `scan_raw.json`, `api_detail.json` | 3b | Multi-samples, click log, detail |
| `route_map.json` etc. | 4 | Permission-tree intermediate files (if executed) |
| `params_merged.json` | 5 | Merged parameter fields + confidence |
| `api_merged.txt` | 5 | `METHOD /path [params] [static\|runtime\|both]` |
| `site_map.json` | 5 | Routes, APIs, params, feature points, limitations |
| **insert_assets** | 5 | Write all discovered service and endpoint assets into the asset store |

### 5b — Parameter Merge

Diff from `param_samples.json`; there is **no universal merge script**. Confidence rules are in reference section J7 (high/medium/low/pending trigger).

### 5c — Error Inference

Within the authorized scope, you may send incomplete requests and read 400 responses (**this is parameter recon, not vulnerability testing**): `field 'x' is required`, enum errors, etc. Note `data` wrappers, `variables`, and `bizData` before encryption.

The report must state: runtimeMode, static/runtime API counts, parameter confidence, uncovered modules, and a summary of `CHANGES.md` relative to the reference scripts.

Suggested `site_map.json` structure:

```json
{
  "site": "https://example.com",
  "runtimeMode": "both",
  "appType": "vue-spa",
  "routeGuardStrategy": ["nav-neutralize", "L1-auth", "L2-patch", "forward"],
  "apisFromStatic": [],
  "apisFromRuntime": [],
  "apis": [],
  "params": [{ "method": "POST", "path": "/api/user/list", "transport": "json", "fields": [] }],
  "frontendRoutes": [],
  "routesVerifiedByClick": [],
  "featuresTriggered": [],
  "limitations": ""
}
```

For more fields and grep recipes, see [reference.md](reference.md).

---

## General Notes

- **Framework agnostic**: webpack/Vite/Angular lazy loading use the same method
- **Transport**: REST/JSON, GraphQL, WebSocket, SSE; gRPC-web is out of scope
- **SSR**: client-side fetches can be recorded; RSC/Server Actions cannot be fully enumerated
- **Blind spots**: JSVMP, WASM, strong HMAC/mTLS validation → use static analysis and document limitations
- **Parameter blind spots**: conditional linkage, hidden params, WASM request construction → mark as “pending trigger” / “unreachable”
- **Static analysis is the safety net**: when runtime is blocked, static analysis can still enumerate endpoints

---

## Additional Resources

- Grep recipes, `config.json` templates, troubleshooting, hooks, parameter reverse engineering section J, and site_map template: **[reference.md](reference.md)**
- Reference script paths are listed in the [Scripts and Gates](#scripts-and-gates) table
