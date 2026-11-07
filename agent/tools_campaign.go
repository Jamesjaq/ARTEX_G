package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Autumn-27/artex/config"
	"github.com/Autumn-27/artex/db"
	actool "github.com/Autumn-27/norma/tool"
)

// Campaign graph (fact-intent engagement memory) — vendored, standalone, no-CTF.
// Integrated as native ARTEX agent tools wrapping campaign/lib/graph.py (stdlib).
// (1) store auto-scopes per task; (2) facts/intents bridge into artex assets/findings/exploration.

func campaignEnginePath() (string, error) {
	cands := []string{}
	if b := config.BaseDir(); b != "" && b != "." {
		cands = append(cands, filepath.Join(b, "campaign", "lib", "graph.py"))
		cands = append(cands, filepath.Join(b, "data", "campaign", "_engine", "graph.py")) // placeholder no-op
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		cands = append(cands, filepath.Join(d, "campaign", "lib", "graph.py"))
		cands = append(cands, filepath.Join(d, "..", "campaign", "lib", "graph.py"))
	}
	if cwd, err := os.Getwd(); err == nil {
		cands = append(cands, filepath.Join(cwd, "campaign", "lib", "graph.py"))
		cands = append(cands, filepath.Join(cwd, "ARTEX", "campaign", "lib", "graph.py"))
	}
	cands = append(cands, "campaign/lib/graph.py", "../campaign/lib/graph.py", "/app/campaign/lib/graph.py")
	seen := map[string]bool{}
	for _, c := range cands {
		c = filepath.Clean(c)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("campaign engine not found (expected campaign/lib/graph.py relative to artex)")
}

// campaignStoreResolved: explicit store wins; else per-task scoped; else .cg.
func (t *ToolSet) campaignStoreResolved(m map[string]json.RawMessage) string {
	if s := strArg(m, "store"); s != "" {
		return s
	}
	if t != nil && t.taskID > 0 {
		dir := filepath.Join(config.BaseDir(), "data", "campaign", fmt.Sprintf("task-%d", t.taskID))
		_ = os.MkdirAll(dir, 0o755)
		return dir
	}
	return ".cg"
}

func runCampaign(ctx context.Context, args ...string) (string, error) {
	engine, err := campaignEnginePath()
	if err != nil {
		return "", err
	}
	py := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		py = "python"
	}
	cmd := exec.CommandContext(ctx, py, append([]string{engine}, args...)...)
	var out, er bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &er
	runErr := cmd.Run()
	res := strings.TrimSpace(out.String())
	if es := strings.TrimSpace(er.String()); es != "" {
		if res != "" {
			res += "\n"
		}
		res += es
	}
	if runErr != nil {
		if res == "" {
			res = runErr.Error()
		}
		return res, fmt.Errorf("campaign: %w", runErr)
	}
	return res, nil
}

func campaignOK(out string) (actool.Result, error) { return actool.Text(out), nil }
func campaignErr(err error, out string) (actool.Result, error) {
	if out == "" {
		out = err.Error()
	}
	return actool.Errorf(out), nil
}

// CampaignTools returns the native campaign fact-intent graph tools (vendored, no-CTF).
func (t *ToolSet) CampaignTools() []actool.CoreTool {
	return []actool.CoreTool{
		t.campaignInit(), t.campaignStatus(), t.campaignFrontier(), t.campaignGraph(),
		t.campaignValidate(), t.campaignFact(), t.campaignIntent(), t.campaignDone(),
		t.campaignHint(), t.campaignRecall(), t.campaignKBSearch(), t.campaignDoc(),
		t.campaignClaim(), t.campaignComplete(),
	}
}

func (t *ToolSet) campaignInit() actool.CoreTool {
	return writeTool("campaign_init", "Initialize an engagement fact-intent graph (shared working memory). No-CTF, vendored in artex. Auto-scoped per task when no store is given.", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
		"goal":  str("engagement objective / goal"),
	}, "goal"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		var a struct{ Store, Goal string }
		_ = json.Unmarshal(in, &a)
		if strings.TrimSpace(a.Goal) == "" {
			return campaignErr(fmt.Errorf("goal required"), "")
		}
		store := t.campaignStoreResolved(toMap(in))
		out, err := runCampaign(ctx, "--store", store, "init", "--goal", a.Goal)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out + fmt.Sprintf("\n[store: %s]", store))
	})
}

func (t *ToolSet) campaignStatus() actool.CoreTool {
	return readTool("campaign_status", "Show engagement graph status (counts, kinds, open intents).", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(toMap(in)), "status")
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignFrontier() actool.CoreTool {
	return readTool("campaign_frontier", "List open, ready-to-work intents (frontier).", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(toMap(in)), "frontier")
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignGraph() actool.CoreTool {
	return readTool("campaign_graph", "Dump the full engagement graph (facts, intents, edges).", obj(map[string]any{
		"store":  str("graph dir (auto: per-task when omitted)"),
		"format": str("text or json (default text)"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		args := []string{"--store", t.campaignStoreResolved(m), "graph"}
		if f := strArg(m, "format"); f != "" {
			args = append(args, "--format", f)
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignValidate() actool.CoreTool {
	return readTool("campaign_validate", "Audit graph integrity (schemas, deps, status) without changing files.", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(toMap(in)), "validate")
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignFact() actool.CoreTool {
	return writeTool("campaign_fact", "Record a confirmed, typed fact (observation/evidence). Kinds: host,service,cred,vuln,loot,artifact,note. Bridges to artex assets/findings/exploration (best-effort).", obj(map[string]any{
		"store":       str("graph dir (auto: per-task when omitted)"),
		"description": str("confirmed observation (one-line gist with key evidence)"),
		"kind":        str("fact kind: host|service|cred|vuln|loot|artifact|note"),
		"data":        str("structured data as key=value pairs comma-separated, e.g. host=10.0.0.1,port=80"),
		"title":       str("short label (2-10 chars)"),
		"doc":         str("detailed markdown (commands, output, evidence)"),
		"from_intent": str("intent id this fact concludes (links+completes it)"),
	}, "description"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		desc := strArg(m, "description")
		if desc == "" {
			return campaignErr(fmt.Errorf("description required"), "")
		}
		store := t.campaignStoreResolved(m)
		args := []string{"--store", store, "fact", desc}
		kind := strArg(m, "kind")
		if kind != "" {
			args = append(args, "-k", kind)
		}
		if v := strArg(m, "title"); v != "" {
			args = append(args, "--title", v)
		}
		if v := strArg(m, "doc"); v != "" {
			args = append(args, "--doc", v)
		}
		if v := strArg(m, "from_intent"); v != "" {
			args = append(args, "--from-intent", v)
		}
		dataKV := parseDataKV(strArg(m, "data"))
		for k, v := range dataKV {
			args = append(args, "-d", k+"="+v)
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		if note := t.bridgeFact(ctx, kind, desc, strArg(m, "title"), dataKV); note != "" {
			out += "\n" + note
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignIntent() actool.CoreTool {
	return writeTool("campaign_intent", "Add a unit of work (intent) depending on fact ids. Bridges to artex exploration intent (best-effort).", obj(map[string]any{
		"store":       str("graph dir (auto: per-task when omitted)"),
		"description": str("what to do (one sentence)"),
		"from":        str("dependency fact ids, comma-separated (e.g. f001,f002 or 'goal')"),
		"title":       str("short label"),
		"doc":         str("detailed markdown: exact commands, evidence expected, timebox"),
		"confirm":     str("set 'true' to require confirmation before execution"),
	}, "description", "from"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		desc, from := strArg(m, "description"), strArg(m, "from")
		if desc == "" || from == "" {
			return campaignErr(fmt.Errorf("description and from required"), "")
		}
		store := t.campaignStoreResolved(m)
		args := []string{"--store", store, "intent", "--from", from, desc}
		if v := strArg(m, "title"); v != "" {
			args = append(args, "--title", v)
		}
		if v := strArg(m, "doc"); v != "" {
			args = append(args, "--doc", v)
		}
		if strings.EqualFold(strArg(m, "confirm"), "true") {
			args = append(args, "--confirm")
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		if note := t.bridgeIntent(desc, strArg(m, "title")); note != "" {
			out += "\n" + note
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignDone() actool.CoreTool {
	return writeTool("campaign_done", "Conclude an intent with the produced fact. Bridges result to artex (best-effort).", obj(map[string]any{
		"store":  str("graph dir (auto: per-task when omitted)"),
		"intent": str("intent id (e.g. i001)"),
		"fact":   str("confirmed result description"),
		"kind":   str("result fact kind"),
		"data":   str("key=value pairs comma-separated"),
		"title":  str("short label"),
		"doc":    str("detailed markdown evidence"),
	}, "intent", "fact"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		iid, f := strArg(m, "intent"), strArg(m, "fact")
		if iid == "" || f == "" {
			return campaignErr(fmt.Errorf("intent and fact required"), "")
		}
		store := t.campaignStoreResolved(m)
		args := []string{"--store", store, "done", iid, "--fact", f}
		kind := strArg(m, "kind")
		if kind != "" {
			args = append(args, "-k", kind)
		}
		if v := strArg(m, "title"); v != "" {
			args = append(args, "--title", v)
		}
		if v := strArg(m, "doc"); v != "" {
			args = append(args, "--doc", v)
		}
		dataKV := parseDataKV(strArg(m, "data"))
		for k, v := range dataKV {
			args = append(args, "-d", k+"="+v)
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		if note := t.bridgeFact(ctx, kind, f, strArg(m, "title"), dataKV); note != "" {
			out += "\n" + note
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignHint() actool.CoreTool {
	return writeTool("campaign_hint", "Add a commander/human hint (or list hints). Persists across restarts.", obj(map[string]any{
		"store":   str("graph dir (auto: per-task when omitted)"),
		"content": str("hint text (omit to list hints)"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		args := []string{"--store", t.campaignStoreResolved(m), "hint"}
		if c := strArg(m, "content"); c != "" {
			args = append(args, c)
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignRecall() actool.CoreTool {
	return readTool("campaign_recall", "Find prior facts/intents similar to a query (reuse before repeating).", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
		"query": str("task gist / keywords"),
		"limit": intp("max matches (default 5)"),
	}, "query"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		q := strArg(m, "query")
		if q == "" {
			return campaignErr(fmt.Errorf("query required"), "")
		}
		args := []string{"--store", t.campaignStoreResolved(m), "recall", q}
		if r, ok := m["limit"]; ok {
			var n int
			if json.Unmarshal(r, &n) == nil && n > 0 {
				args = append(args, "--limit", fmt.Sprintf("%d", n))
			}
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignKBSearch() actool.CoreTool {
	return readTool("campaign_kb_search", "Search the offline knowledge base (techniques/CVEs/writeups).", obj(map[string]any{
		"query": str("keywords (technique, product, CVE)"),
		"tag":   str("optional tag filter"),
		"limit": intp("max hits (default 5)"),
		"show":  str("entry id (kb001) to show full detail instead of searching"),
	}), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		if id := strArg(m, "show"); id != "" {
			out, err := runCampaign(ctx, "kb", "show", id)
			if err != nil {
				return campaignErr(err, out)
			}
			return campaignOK(out)
		}
		q := strArg(m, "query")
		if q == "" {
			return campaignErr(fmt.Errorf("query (or show) required"), "")
		}
		args := []string{"kb", "search", q}
		if v := strArg(m, "tag"); v != "" {
			args = append(args, "-t", v)
		}
		if r, ok := m["limit"]; ok {
			var n int
			if json.Unmarshal(r, &n) == nil && n > 0 {
				args = append(args, "--limit", fmt.Sprintf("%d", n))
			}
		}
		out, err := runCampaign(ctx, args...)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignDoc() actool.CoreTool {
	return readTool("campaign_doc", "Read the detailed doc (third layer) of a fact/intent node.", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
		"id":    str("node id (e.g. f001, i002)"),
	}, "id"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		id := strArg(m, "id")
		if id == "" {
			return campaignErr(fmt.Errorf("id required"), "")
		}
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(m), "doc", id)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignClaim() actool.CoreTool {
	return writeTool("campaign_claim", "Claim an open intent for a worker (reserve it).", obj(map[string]any{
		"store":  str("graph dir (auto: per-task when omitted)"),
		"intent": str("intent id"),
		"worker": str("worker name"),
	}, "intent", "worker"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		iid, w := strArg(m, "intent"), strArg(m, "worker")
		if iid == "" || w == "" {
			return campaignErr(fmt.Errorf("intent and worker required"), "")
		}
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(m), "claim", iid, "--worker", w)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

func (t *ToolSet) campaignComplete() actool.CoreTool {
	return writeTool("campaign_complete", "Mark the engagement goal satisfied (links a fact to goal).", obj(map[string]any{
		"store": str("graph dir (auto: per-task when omitted)"),
		"from":  str("fact id(s) satisfying the goal, comma-separated"),
		"note":  str("why the objective is met"),
	}, "from", "note"), func(ctx context.Context, in json.RawMessage) (actool.Result, error) {
		m := toMap(in)
		from, note := strArg(m, "from"), strArg(m, "note")
		if from == "" || note == "" {
			return campaignErr(fmt.Errorf("from and note required"), "")
		}
		out, err := runCampaign(ctx, "--store", t.campaignStoreResolved(m), "complete", "--from", from, "--note", note)
		if err != nil {
			return campaignErr(err, out)
		}
		return campaignOK(out)
	})
}

// ---- bridging (best-effort, never fatal) ----

func (t *ToolSet) bridgeFact(ctx context.Context, kind, desc, title string, data map[string]string) (note string) {
	defer func() { _ = recover() }()
	kind = strings.ToLower(strings.TrimSpace(kind))
	var bridged []string
	// asset bridge
	if t.as != nil {
		switch kind {
		case "host":
			if ip := firstNonEmpty(data["ip"], data["host"], data["addr"]); ip != "" && strings.Contains(ip, ".") {
				if id, err := t.as.UpsertIP(db.UpsertIPReq{IP: ip, TaskID: t.taskID}); err == nil && id > 0 {
					t.anchorOwner(id)
					bridged = append(bridged, fmt.Sprintf("asset:ip#%d", id))
				}
			} else if dom := firstNonEmpty(data["domain"], data["hostname"], data["host"]); dom != "" {
				if id, err := t.as.UpsertSubdomain(db.UpsertSubdomainReq{Domain: dom, TaskID: t.taskID}); err == nil && id > 0 {
					t.anchorOwner(id)
					bridged = append(bridged, fmt.Sprintf("asset:subdomain#%d", id))
				}
			}
		case "service":
			if u := firstNonEmpty(data["url"]); u != "" {
				if id, err := t.as.UpsertHTTPService(db.UpsertHTTPServiceReq{URL: u, TaskID: t.taskID}); err == nil && id > 0 {
					t.anchorOwner(id)
					bridged = append(bridged, fmt.Sprintf("asset:service#%d", id))
				}
			} else if port, _ := strconv.Atoi(firstNonEmpty(data["port"])); port > 0 {
				name := firstNonEmpty(data["product"], data["service"], "service")
				host := firstNonEmpty(data["host"], data["ip"])
				if host != "" {
					if id, err := t.as.UpsertOtherService(func() db.UpsertOtherServiceReq { r:=db.UpsertOtherServiceReq{ServiceName: name, Port: port, TaskID: t.taskID}; if netIsIP(host){r.IP=host}else{r.Domain=host}; return r }()); err == nil && id > 0 {
						t.anchorOwner(id)
						bridged = append(bridged, fmt.Sprintf("asset:service#%d", id))
					}
				}
			}
		}
	}
	// finding / exploration bridge
	if t.ts != nil && t.ts.ID() > 0 {
		if kind == "vuln" {
			sev := strings.ToLower(firstNonEmpty(data["severity"], "medium"))
			vname := firstNonEmpty(data["name"], title, "campaign finding")
			evid := firstNonEmpty(data["evidence"], data["poc"], desc)
			in := db.RecordFindingInput{TaskID: t.taskID, ExplorationID: t.ts.ID(), IntentID: t.ownerNode, VulnClass: firstNonEmpty(data["vulnclass"], data["cve"], "vuln"), Name: vname, Severity: sev, Summary: desc, Evidence: evid, Worker: t.worker}
			if rec, err := t.ts.RecordFinding(ctx, in); err == nil && rec != nil {
				bridged = append(bridged, fmt.Sprintf("finding#%d", rec.NodeID))
				if t.notify != nil {
					t.notify()
				}
			} else {
				// fallback to exploration fact
				bridged = append(bridged, t.bridgeExplorationFact(desc, kind))
			}
		} else {
			bridged = append(bridged, t.bridgeExplorationFact(desc, kind))
		}
	}
	// compact
	out := make([]string, 0, len(bridged))
	for _, b := range bridged {
		if b != "" {
			out = append(out, b)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return "[bridged -> artex: " + strings.Join(out, ", ") + "]"
}

func (t *ToolSet) bridgeExplorationFact(desc, kind string) string {
	if t.ts == nil || t.ts.ID() == 0 || strings.TrimSpace(desc) == "" {
		return ""
	}
	sum := desc
	if kind != "" {
		sum = "[" + kind + "] " + sum
	}
	payload := map[string]any{"summary": sum, "confidence": "observed"}
	id, err := t.ts.AddNode(db.KindFact, payload, 5, "confirmed", firstNonEmpty(t.worker, "campaign"), nil)
	if err != nil || id <= 0 {
		return ""
	}
	return fmt.Sprintf("fact#%d", id)
}

func (t *ToolSet) bridgeIntent(desc, title string) string {
	if t.ts == nil || t.ts.ID() == 0 || strings.TrimSpace(desc) == "" {
		return ""
	}
	payload := map[string]any{"summary": desc}
	if title != "" {
		payload["title"] = title
	}
	id, err := t.ts.AddNode(db.KindIntent, payload, 5, "open", firstNonEmpty(t.worker, "campaign"), nil)
	if err != nil || id <= 0 {
		return ""
	}
	if of, _ := t.ts.OriginFactID(); of > 0 {
		_ = t.ts.Link(of, db.RelSpawns, id)
	}
	return fmt.Sprintf("[bridged -> artex: intent#%d]", id)
}

func toMap(in json.RawMessage) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	_ = json.Unmarshal(in, &m)
	return m
}
func strArg(m map[string]json.RawMessage, k string) string {
	r, ok := m[k]
	if !ok {
		return ""
	}
	var s string
	if json.Unmarshal(r, &s) == nil {
		return strings.TrimSpace(s)
	}
	var n json.Number
	if json.Unmarshal(r, &n) == nil {
		return n.String()
	}
	return ""
}
func parseDataKV(s string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(s) == "" {
		return out
	}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if i := strings.Index(p, "="); i >= 0 {
			k := strings.TrimSpace(p[:i])
			v := strings.TrimSpace(p[i+1:])
			if k != "" {
				out[k] = v
			}
		}
	}
	return out
}
func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func netIsIP(s string) bool {
	for _, c := range s { if (c < '0' || c > '9') && c != '.' && c != ':' { return false } }
	return strings.Contains(s, ".") || strings.Contains(s, ":")
}
