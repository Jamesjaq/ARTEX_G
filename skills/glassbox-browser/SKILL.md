---
name: glassbox-browser
description: "Invoke this skill when a web flow must be driven the way a person drives it — a real logged-in Chrome profile behind Glass Box: authenticated and session-bound flows, MFA/OTP walls, SPA surfaces a crawler cannot reach, watering-hole pages, and visual evidence of what the browser actually did."
mcps: glassbox
---

# Glass Box Browser Drive (a human's browser, watched by a human)

Glass Box is a byte-transparent MCP tee in front of a browser engine that drives **the
operator's real Chrome through an installed extension** — real profile, real cookies, real
history, real fingerprint — with a live dashboard at `http://127.0.0.1:7777` showing every call,
argument and screenshot as it happens. It is the difference between "an HTTP request that looks
like a browser" and "the browser". It is also the least anonymous thing in the toolbelt, so it is
used for the flows that need it and never for the ones that do not.

## Boundaries

- Authorized use: the site or application is in scope, and the RoE permits acting **with a real
  user's authenticated session**. Acting inside someone's account is not a stealthier crawl — it
  is a different authority, and it needs to be granted explicitly.
- **Identity leakage is the cost of this tool.** Traffic leaves from the operator's IP, their
  hostname, their browser fingerprint and their logged-in identities; the target sees a person,
  not a scanner. Before the first navigation state it in one line and get the operator's decision
  through `ask_operator`, offering the alternatives in order of preference: a dedicated
  engagement browser profile, a lab host, the headless `browser` (Playwright) MCP, or plain
  `shell_open`/curl for what needs no rendering.
- **Nothing is done in a real person's account that the plan did not list.** Read-only by default.
  Writing, sending, purchasing, deleting, changing settings or anything a helpdesk would call an
  incident is destructive-or-externally-visible: it goes to the operator first. The same rule as
  everywhere else in this platform — the gate is irreversibility, not difficulty.
- The dashboard is the control, so do not defeat it: `NEEDS YOU` on that screen is a real halt
  (an OTP, a captcha handed over, an explicit `browser_ask_user`). Back-off states
  (`RECONNECTING`) are infrastructure, not permission to stop.
- This is not a bot-net for scale. Rate, volume and repetition belong to `opsec-tradecraft`;
  a person-shaped browser driven at scanner rates is the worst of both tools.
- Not this skill: static crawl and API-surface mapping (`api-recon`, `playwright-cli` on the
  headless engine), auth-flow *design* (`auth-testing`), lure page delivery (`initial-access`,
  `human-terrain`), mobile app webviews (`mobile-testing`).
- Stop condition: if the glassbox MCP is disabled or the extension has never linked, that is an
  operator action — report the gap with `mcp_status` output and continue with the headless engine
  where the flow allows it.

## Procedure

Do not act in an authenticated account (step 5) before step 2 has the operator's decision and
step 4 has confirmed the session you are actually in.

1. **Check the bridge.** `mcp_status` for the glassbox server; if it is not enabled, ask the
   operator to enable it (and to load `engine/extension` in Chrome once — it self-dials via
   `alarms`, there is no Connect button). An unlinked extension gives ~3.4 s of honest
   `RECONNECTING`, so retry a scan cycle rather than abandoning the run.
2. **Declare the identity cost and get the call.** One `ask_operator`: which profile, which
   source IP, which logged-in identities will be visible to the target, and what the flow needs
   them for. Record the answer as a dated fact.
3. **Open the tab and take the baseline.** `browser_list_tabs` to see what is already open, then
   `browser_navigate`. Immediately `browser_snapshot` for structure and `browser_screenshot` for
   the visual baseline. The snapshot is the cheap read; spend the screenshot only where the image
   is the evidence or where you need coordinates.
4. **Establish who you are, honestly.** `browser_get_cookies` and `browser_get_local_storage` for
   the session class, and read the app's own identity display. Never assume you are anonymous or
   logged out; if the session is someone else's, stop and re-ask.
5. **Drive the flow like a person, in the minimum steps.** `browser_fill` and
   `browser_click` / `browser_set_combobox` / `browser_select_option` for controls,
   `browser_press_key` for keyboard-only ones, `browser_scroll` and `browser_wait_for_network`
   instead of fixed sleeps, `browser_dismiss_overlays` for cookie and consent layers. Read a
   region with `browser_get_page_content` (8k default) or `browser_extract_list` for repeating
   structures; `browser_execute_script` when the DOM question is genuinely a script question.
6. **On a wall, use its own door.** `browser_solve_captcha` runs the documented
   detect → checkbox → vision → grid ladder and only hands over with `action:"ask_human"`; an OTP
   or a hardware prompt is `browser_ask_user` — that is the run pausing for a human, which is
   exactly what the dashboard is for. Never guess a code.
7. **Keep the tabs straight.** `browser_get_new_tab` / `browser_switch_tab` after every popup or
   target=_blank action; `browser_list_frames` and `browser_select_frame` before touching
   anything inside an iframe, or the read is from the wrong document.
8. **Capture evidence as you go, not afterwards.** Each step that matters gets a screenshot and
   the page URL; `browser_console_logs` for client errors worth a finding; the request side still
   goes through the platform's traffic tools, so the browser action and the wire record line up.
9. **Extract what the flow legitimately produced** — tokens, URLs, document classes — by
   reference (`save_task_credential` for anything secret) and write the reachable state into the
   graph (`insert_assets`, `record_fact`) so a later seat does not need to re-drive the flow.
10. **Leave the browser as you found it.** Close what you opened (`browser_close_tab`), do not
    log out someone else's session, do not change their settings, and record the teardown. If a
    destructive-looking action was approved and performed, note the reversal step and whether it
    was verified.

## Evidence

Positive: the operator's identity-cost decision, the profile and source of traffic, the flow steps
as dated screenshots plus the URLs and DOM state that show the result, the session class you
acted in, and the wire-side record for the same actions. A claim about what an authenticated user
can reach needs the screenshot of the reach, not the inference.

Negative and honest-gap forms: "glassbox MCP disabled during the window — NOT ASSESSED via real
profile; headless result reported separately", "extension never linked (RECONNECTING for the
whole attempt)", "OTP wall: human handoff at <time>, no code available, flow stopped", "operator
declined real-profile use; only unauthenticated surface tested". Say which engine produced which
result — headless and real-browser findings are not interchangeable evidence.

## Report shape

`<flow> reachable by an authenticated browser session on <app>` — fields: authorization reference
and identity-cost decision, profile/source of traffic, session class, the steps with dated
evidence pointers, the control that did or did not stop the flow (MFA, device binding, step-up,
transaction approval), and the recommendation tied to the control that failed.
