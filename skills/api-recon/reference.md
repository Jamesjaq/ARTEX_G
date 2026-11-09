# api-recon — Reference Manual

Grep recipes, `config.json` templates, and troubleshooting. Run all grep commands against the `js/` directory. If bundles are single-line, you can first run `js-beautify` or `sed 's/}/}\n/g'`, but raw grep with a context window is usually sufficient.

## Script Notes

All files in `scripts/` are **reference templates** and must be adapted to the target site before execution. Typical change points:

| Script | Common items to adjust |
|---|---|
| `harvest_static.py` | endpoint regex, webpack/Vite manifest parsing, micro-frontend publicPath, retry/concurrency |
| `runtime_harvest.js` | neutralize field names and success values, stub match rules and body structure, route sources, WS recording, `waitUntil`/`routeTimeout`/`proxy` |
| `preload.js` | `loginPathRe`, L1 stubs, `neutralize.fields`, `apiPattern`, whether to enable L3, `recordDetail`, `observe.*`, `neutralizeVueRouter` |
| `spider_mpa.py` | `--exclude` destructive links, cookies, depth/max, same-origin filtering |
| `extract_route_map.py` | `routeMap` / `routeLink` regex, KEY naming patterns |
| `build_perm_tree.py` | `userRouteAuth` parsing, `ROOTS`/`PREFIX_PARENT` hierarchy heuristics, stub outer field names |
| `config.json` | unified entry point for all site-specific parameters above |

Place adjusted files in the task working directory (such as `recon/`) and document the specific changes relative to the reference scripts in the report.

---

## A. Three Reverse-Engineering Gates

### A1. Render Gate — “How is logged-in state determined?”

```bash
grep -rhoaE '.{0,40}(isLogin|isAuthenticated|loggedIn|hasLogin|requireAuth)\b.{0,80}' js | head
grep -rhoaE 'function (getUser|getToken|getAuth)[0-9]?\([^)]*\)\{.{0,200}' js | head
grep -rhoaE '(localStorage|sessionStorage)\.getItem\("[^"]+"\)' js | sort -u
grep -rhoaE '(Cookies?|cookie)\.(get|load)\("[^"]+"\)' js | sort -u
grep -rhoaE '\batob\(|JSON\.parse\(|jwt|decode' js | head
```

Find the chain `isLogin = f(getUser())` → `getUser = decode(storage.read(KEY))`, and determine the **storage key**, **container** (Cookie vs localStorage), and **encoding**:

| Encoding | How to forge in config |
|---|---|
| Plain string / `"1"` / token | `"value": "anything-truthy"` |
| `JSON.parse(x)` | `"value": "json:{\"id\":1,\"username\":\"admin\"}"` |
| `JSON.parse(atob(x))` | `"value": "b64json:{\"id\":1,\"username\":\"admin\"}"` |
| JWT | unsigned/`alg:none` JWT, or sign with a key found in the bundle |
| Encrypted (SM2/AES/RSA) | find hardcoded keys; if the render gate only needs a decodable blob you can forge it; otherwise fall back to static analysis |

→ Write into `cookies` / `localStorage`.

### A2. Interceptor Gate — “What triggers a redirect to /login?”

```bash
grep -rhoaE '.{0,60}(interceptors\.response|axios|request\.use).{0,120}' js | head
grep -rhoaE '.{0,40}(response_code|errcode|errno|\bcode\b|\bret\b|\bstatus\b)\s*[=!]==?\s*[\-0-9]{1,4}.{0,60}' js | head -20
grep -rhoaE '.{0,40}(not\s*logged\s*in|please\s*log\s*in|login\s*expired|session\s*expired|unauthorized|forbidden|token.{0,10}invalid).{0,40}' js | head
grep -rhoaE '.{0,30}(location\.href|router\.(push|replace)|navigate)\([^)]*login[^)]*\)' js | head
```

Determine the **field name**, **success value** (usually `0` or `200`), and the **failure value that triggers a redirect**. Verify with a junk session:

```bash
curl -sk -X POST -H 'Cookie: <fakekey>=junk' https://target/api/<protected> -d '{}' -H 'Content-Type: application/json'
```

→ Write into `neutralize.fields` + `neutralize.success`.

### A3. Content Gate — “Where do menus/permissions come from?”

```bash
grep -rhoaE '"/api[^"]*(permission|perm|role|menu|acl|resource|nav)[^"]*"' js | sort -u
grep -rhoaE '.{0,30}(menus|permissions|menuList|routeList|authList|role_permissions)\b.{0,120}' js | head
grep -rhoaE 'userRouteAuth|getResultTree|routeMap|routeLink|hasPermission|checkAuth' js | head
grep -rhoaE '([A-Z_][A-Z0-9_]*):\{name:"[^"]*",link:"/[^"]+"\}' js | head
```

**Two-layer data** (common in enterprise admin panels):

| API | Typical payload | Consumer |
|---|---|---|
| `.../role_permissions` | `{ permissions: string[], role_type }` | route guards, button-level ACL |
| `.../permissions/all` | `tree[{ code, position, children }]` | sidebar menu rendering |
| bundle `userRouteAuth` | `{ CODE: { url, name? } }` | code → frontend path |
| bundle `routeMap` | `{ KEY: { name, link } }` | alias resolution (webpack `o.DASHBOARD`) |

Read the consumer code to confirm how `getResultTree(tree, permissions)` filters and which field is checked by `v-if` / `hasAuth(code)`.

**Manual forge** (small sites): build a permissive payload → `stubs`.

**Full permission-tree reconstruction** (large sites where the sidebar/submodules remain blank): see **section I**.

---

## B. config.json Template

```json
{
  "baseUrl": "https://target/",
  "runtimeMode": "both",
  "chromium": "/usr/bin/chromium",

  "cookies": [
    { "name": "auth", "value": "b64json:{\"id\":1,\"username\":\"admin\",\"role\":\"admin\",\"func\":{},\"permissions\":[\"*\"]}" }
  ],
  "localStorage": { "token": "faketoken", "isLogin": "1" },

  "neutralize": {
    "fields": ["response_code", "code", "errno", "ret", "status"],
    "success": 0,
    "flags": { "success": true, "message": "ok" }
  },
  "forward": true,
  "loginUrlPattern": "/login",
  "apiPattern": "/api/|/rest/|/graphql",

  "mockTier": "L1+L2",
  "recordDetail": true,
  "observe": {
    "storageReads": false,
    "cookieReads": false,
    "xhrHeaders": true
  },
  "neutralizeVueRouter": true,
  "stubs": [
    {
      "match": "permissions/all|/menu|role_permissions",
      "body": {
        "response_code": 0, "code": 0,
        "data": {
          "permissions": ["*"],
          "menus": [
            { "name": "dashboard", "path": "/dashboard", "show": true, "children": [] },
            { "name": "alert", "path": "/alert", "show": true, "children": [] }
          ]
        }
      }
    }
  ],

  "explore": {
    "clickTabs": true,
    "clickTables": true,
    "pushStateFallback": true,
    "maxMenuItems": 50
  },

  "routes": ["/dashboard", "/alert", "/asset", "/device", "/report", "/config", "/system"],
  "waitMs": 1500, "perRouteMs": 900, "headless": true,
  "waitUntil": "domcontentloaded",
  "routeTimeout": 12000,
  "proxy": "",

  "captureResponses": true, "recordWs": true, "respMax": 600
}
```

Field notes:
- `runtimeMode`: `depth` (Puppeteer), `coverage` (browser MCP), `both`
- `cookies[].value` prefixes: `b64json:` → base64(JSON); `json:` → raw JSON; no prefix → literal
- `forward: true` forwards real requests and rewrites code fields; `false` uses fully offline stubs
- `mockTier`: preload layer levels enabled in coverage mode, e.g. `L1+L2`, `L1+L2+L3`
- `routes` comes from `routes.txt`; after forging menus, the harness automatically appends `<a href>`
- `captureResponses` / `recordWs` apply only to depth mode
- `waitUntil`: for large SPAs use `domcontentloaded` to avoid hangs with `networkidle2`
- `routeTimeout`: per-route `page.goto` timeout (milliseconds)
- `proxy`: Puppeteer `--proxy-server`; you can also set `HTTP_PROXY` / `HTTPS_PROXY`

### B1. Dual Stub Template (role_permissions + permissions/all)

```json
"stubs": [
  {
    "match": "role_permissions",
    "body": {
      "response_code": 0,
      "data": {
        "permissions": ["MONITOR", "MONITOR_ALERT", "THREAT", "ASSETS_RISK"],
        "role_type": "SUPER_ADMIN"
      }
    }
  },
  {
    "match": "permissions/all",
    "body": {
      "response_code": 0,
      "data": [
        {
          "code": "MONITOR",
          "position": 1,
          "children": [
            { "code": "MONITOR_ALERT", "position": 1, "children": [] }
          ]
        }
      ]
    }
  }
]
```

The outer field names (`response_code` / `code` / `data`) must match the A2 interceptor gate; `permissions` must cover all leaf codes in the tree.

---

## C. Coverage Mode: Preload Configuration

Edit the `CONFIG` object at the top of `scripts/preload.js`, or replace it before CDP injection:

```javascript
const CONFIG = {
  loginPathRe: /\/(login|signin)(\/|$|\?)/i,
  mockTier: 'L1+L2',
  forward: true,
  recordDetail: true,
  extractUrlsFromResponse: true,
  neutralizeVueRouter: true,
  observe: { storageReads: false, cookieReads: false, xhrHeaders: true },
  neutralize: { fields: ['response_code', 'code'], success: 0 },
  stubs: [ /* same as config.json stubs */ ],
  apiPattern: /\/(api|apis|v\d+|dev|internal|graphql)\//i,
};
```

Verification: `window.__API_RECON_PRELOAD__ === true` and the pathname is stable.

Export recorded results:

```javascript
JSON.stringify({
  apis: [...window.__API_RECON_LOG__],
  detail: window.__API_RECON_DETAIL__,
  routes: [...(window.__API_RECON_ROUTES__ || [])],
  observe: window.__API_RECON_OBSERVE__,
}, null, 2)
```

---

## D. Preload / Runtime Hook Capabilities

Browser hook capabilities built into preload (coverage) and runtime_harvest (depth), and their coverage:

| Hook capability | Value for API discovery | Coverage |
|---|---|---|
| Hook fetch / XHR.open | record request URL/method | ✅ `recordDetail` + `__API_RECON_LOG__` |
| Hook XHR.setRequestHeader | discover Authorization and other headers | ✅ `observe.xhrHeaders` |
| Hook localStorage/cookie reads | confirm session key names | ⚠️ optional `observe.storageReads/cookieReads` |
| Vue route extraction | supplement frontendRoutes | ✅ `__API_RECON_ROUTES__` (loaded routes) |
| Vue route guard neutralization / login redirect blocking | expand modules to trigger APIs | ✅ `neutralizeVueRouter` + native navigation neutralization |
| React route extraction | supplement routes | ⚠️ static analysis + clicks; no dedicated hook |
| Page navigation blocking (login path) | keep the page in place for analysis | ⚠️ blocks only login paths to avoid breaking business navigation |
| Hook crypto libraries (CryptoJS/SM, etc.) | encrypted parameters → plaintext API body | ❌ must manually hook the encryption function input; record conclusions in config |
| Anti-debug bypass | otherwise runtime cannot record APIs | ❌ must be handled manually; static analysis still works |

---

## E. Endpoint Extraction Regex (when static results are too few)

Loosen `extract_endpoints` in `harvest_static.py`, or do it manually:

```bash
grep -rhoaE '"/[a-z][A-Za-z0-9_/\-]{3,}"' js | sort -u
grep -rhoaE '/api/[a-zA-Z0-9_./-]+' js | sort -u
```

---

## F. Troubleshooting

| Symptom | Cause → Fix |
|---|---|
| Very few static APIs | endpoint dialect mismatch → loosen regex (section D/E) |
| chunk count ≪ manifest | CSS-only or undeployed chunks; 404s were retried |
| runtime still shows login page | render gate is wrong → recheck A1: key name, container, encoding, domain |
| entered shell but modules are blank | content gate → forge menus (A3); `routes` paths may be wrong |
| only bootstrap/locale on each route | permission codes incomplete → permission-tree reconstruction (section I); check `role_permissions` + `permissions/all` dual stubs |
| sidebar items exist but child pages are blank | tree missing intermediate nodes or codes inconsistent with `userRouteAuth` |
| every API redirects to login | interceptor gate → confirm `neutralize`; extend walk logic for nested fields |
| 0 WebSocket frames | subscription happens only after user interaction; increase `perRouteMs` |
| empty response bodies | real responses exist only when `forward: true` |
| Chromium missing | install chromium or set `config.chromium` / `CHROMIUM` |
| many mocks still redirect to login | hook too late or missing `location.href` setter → document-start + preload |
| all lists are empty | L3 empty arrays are normal; continue clicking tabs/settings/details |
| Redux actions mistaken for routes | filter internal paths containing get/set/change/clear/toggle/upload |
| Vue still redirects to login | preload not injected at document-start → change injection timing; or manually clear guards when `neutralizeVueRouter: false` |
| URLs in responses not in log | enable `extractUrlsFromResponse`; or extract manually from `__API_RECON_DETAIL__` |
| unknown Authorization header name | enable `observe.xhrHeaders` or inspect request headers in DevTools |
| runtime is very slow / times out | set `waitUntil: domcontentloaded`; reduce `routeTimeout`; do not use `networkidle2` |
| proxy connection fails | check `proxy` / environment variables; ensure Puppeteer and curl use the same proxy port |

---

## G. Hardened Targets

When the server progressively validates sessions (unforgeable signed cookies, server-rendered menus that cannot be stubbed), runtime will get stuck at the shell. Expected behavior:

- **Static analysis is sufficient for endpoint enumeration** — module paths are in the code
- If authorized, run the same harness with a **real session**: `forward: true`, no neutralize needed, capture real methods/params/responses

---

## H. Single-Task Checklist

1. Confirm authorization scope
2. **Read** `scripts/harvest_static.py` → adapt to the target → run → review `api_static.txt`, `routes.txt`
3. **Phase 1b**: path anchor expansion + binding layer → `param_candidates.json` (section J)
4. Reverse-engineer A1/A2/A3 → write a site-specific `config.json`
5. **Read and adapt** `runtime_harvest.js` / `preload.js` before execution
6. `runtimeMode=depth`: `npm install` → run the adjusted harvest script
7. `runtimeMode=coverage/both`: inject the adjusted preload at document-start → browser MCP dynamic enumeration + **parameter trigger matrix**
8. Modules do not render → **section I permission-tree reconstruction** → patch stubs → rerun
9. Parameter multi-sample diff + error inference → `params_merged.json`
10. Merge → `site_map.json` + `api_merged.txt`, honestly documenting coverage, gaps, and script changes

---

## I. Permission-Tree Reconstruction (Phase 4 Deepening)

Use this when forging simple `menus: [{ path, show: true }]` is ineffective and submodules still do not mount.

### I1. Locate the Auth Module

```bash
grep -l 'userRouteAuth' js/*.js
grep -l 'routeMap\|routeLink' js/*.js
grep -rhoaE 'getResultTree|role_permissions|permissions/all' js | head
```

Record: **permission API paths**, **response field names**, and **consumer chunk filenames**.

### I2. Extract routeMap

```bash
python3 scripts/extract_route_map.py recon/js recon/
# output: recon/route_map.json
```

If `[!] no routeMap pattern found`: loosen the regex in `extract_route_map.py`, or grep manually:

```bash
grep -rhoaE '([A-Z_][A-Z0-9_]*):\{name:"[^"]*",link:"/[^"]+"\}' js | head -20
```

### I3. Build the Permission Tree + Stubs

```bash
python3 scripts/build_perm_tree.py recon/js recon/ --config recon/config.json
```

Script logic:
1. Parse `userRouteAuth={MONITOR:{url:...},...}` (including webpack aliases `He=o.DASHBOARD`)
2. Resolve aliases with `route_map.json` → real paths
3. Infer parents by code prefix (`MONITOR_ALERT` → `MONITOR`)
4. Output `permissions_tree.json`, `permissions_all_stub.json`, `role_permissions_stub.json`
5. With `--config`, automatically write `stubs` and extend `routes` in `config.json`

**Adjust per target** (at the top of the script):
- `DEFAULT_ROOTS`: top-level module code list
- `DEFAULT_PREFIX_PARENT`: `PREFIX_` → parent mapping
- `DEFAULT_EXTRA_PARENT`: orphan nodes without prefix relationships

### I4. Validate Stub Consistency

```bash
# permissions count should be ≈ userRouteAuth entry count
wc -l recon/perm_codes_all.txt
# routes should cover all links in route_map
python3 -c "import json; m=json.load(open('recon/route_map.json')); r=set(json.load(open('recon/config.json'))['routes']); print('missing', [v['link'] for v in m.values() if v['link'] not in r])"
```

### I5. Rerun Runtime and Compare

```bash
node recon/runtime_harvest.js recon/config.json
# compare runtime_api.json counts before/after forging; check whether module APIs appear for /attack, /asset, etc.
```

| Before forge | After forge (success) |
|---|---|
| same 3–5 bootstrap requests on every route | different routes trigger different module APIs |
| only `/api/locale/language` | `/api/web/...` module endpoints appear |
| single-digit routes in `routes.txt` | `routes` expands to 80–110+ from route_map |

### I6. If It Still Fails

- **coverage mode**: click the sidebar + tabs; permission gating may request data only after interaction
- **stub fields**: compare a real API (curl + real session) with the stub nesting
- **extra guards**: grep for button-level checks such as `hasPermission|checkRole|func.` and extend `role_permissions.permissions`
- **static fallback**: module API paths still exist in `api_static.txt`; runtime only supplements METHOD/body; keep parameters in `param_candidates.json` + recorded samples

---

## J. Parameter Reverse Engineering (Phase 1b / 5b / 5c)

**Methodology, not a universal script.** Find paths with regex; find parameters with anchor expansion + UI binding chain + multi-sample diff + error inference.

### J1. Anchor Expansion — Find Request Construction from a Path

```bash
# use a known path from Phase 1 as the anchor
grep -n '"/api/user/list"' js/*.js
grep -rhoaE '.{0,120}("/api[^"]+").{0,200}' js | head
grep -rhoaE '(params|data|body|payload)\s*:\s*\{' js | head
grep -rhoaE '(get|post|put|delete|patch)\([^,]+,\s*\{' js | head
```

### J2. Wrapper Layers and Transport Forms

```bash
# axios / unified request helper
grep -rhoaE '(axios|request)\.(get|post|put|delete|patch)\(' js | head
grep -rhoaE 'interceptors\.(request|response)' js | head

# GraphQL
grep -rhoaE '(query|mutation)\s+\w+|gql`|graphql\(' js | head
grep -rhoaE '\$[a-zA-Z_]+\s*:\s*(Int|String|Boolean|\[)' js | head

# FormData / multipart
grep -rhoaE 'FormData|\.append\(' js | head

# path parameters
grep -rhoaE 'path:\s*"/[^"]*:[^"]+"' js | head
grep -rhoaE 'useParams|route\.params|\$route\.params' js | head
```

### J3. Validation Gate — Required / Format / Enum

```bash
grep -rhoaE '(required|message|pattern|enum|validator)\s*:' js | head
grep -rhoaE 'yup\.|zod\.|async-validator|Form\.Item|a-form-item|el-form-item' js | head
grep -rhoaE 'rules\s*:\s*\[|name:\s*["\'][a-zA-Z_]+["\']' js | head
grep -rhoaE 'label.*value|options\s*:\s*\[' js | head
```

### J4. Binding Layer — Form → API

```bash
grep -rhoaE 'onFinish|handleSubmit|getFieldsValue|validateFields' js | head
grep -rhoaE '(pick|omit|transform|dayjs|moment)\(' js | head
```

Runtime supplement: DevTools → Network → request → **Initiator** (call stack); trace upward from `fetch`/`send` to the request-construction function.

### J5. Encrypted Parameters

```bash
grep -rhoaE 'encrypt|decrypt|sign|CryptoJS|sm2|sm3|sm4|RSA|AES' js | head
```

**Do not guess fields from ciphertext** — hook the **input** to the encryption function and record the plaintext payload before encryption; write conclusions into `config.json` / `param_candidates.json`.

### J6. Parameter Trigger Matrix (required in Phase 3)

Record once per operation in each module and diff request bodies/queries:

| Operation | Focus |
|---|---|
| Initial list load | pagination defaults |
| Search | keyword, filters |
| Advanced filters | optional fields |
| Create/edit | full entity |
| Bulk/export | `ids[]`, `exportType` |
| Sort/pagination | `sortField`, `order` |

Output `param_samples.json`: `[{ "path", "method", "action": "search", "body", "query", "headers" }]`

### J7. Confidence Rules

| Confidence | Condition |
|---|---|
| **High** | static callsite + ≥2 consistent runtime samples |
| **Medium** | static only, or only 1 runtime sample |
| **Low** | inferred from responses/errors, not revalidated |
| **Pending trigger** | field known statically, UI/permissions not reached |

### J8. Quick Scenario Playbooks

| Scenario | Order |
|---|---|
| REST list page | J1 request-construction object → J6 four diffs → J3 rules |
| Create/edit form | J3 Form name → J4 submit chain → runtime submit + intentionally leave blank to observe 400 |
| GraphQL | J2 variable declarations → record variables for each operation at runtime |
| Encrypted body | J5 hook the input → fields before encryption are the real params |

### J9. Mapping to api-recon Phases

| api-recon | Parameter recon |
|---|---|
| Phase 1 static | J1 anchor expansion |
| Phase 2 A2 interceptor | globally injected fields (tenantId, sign) |
| Phase 3 runtime | J6 trigger matrix + `param_samples.json` |
| Phase 4 permission tree | different modules/forms differ → full fields trigger only with sufficient permissions |
| Phase 5 merge | `params_merged.json` + confidence; do not determine required fields from a single sample |

### J10. Troubleshooting

| Symptom | Fix |
|---|---|
| Field name appears statically but never at runtime | mark as “pending trigger”; supplement permission tree / click advanced filters / try each linked select option |
| Same path has different body shapes | normal — record separate entries by `action`; do not force-merge into one schema |
| Stub responses are fake but you want params | **inspect outbound request** bodies/headers; do not infer from stub responses |
| 400 reports a nested field | watch for outer wrappers such as `data`/`bizData`/`variables` |
| Only operation names are visible in GraphQL | expand the `variables` JSON; find `$var: Type` in static analysis |

---