package dashboard

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/fimskiy/evil-merge-detector/app/internal/session"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type Handler struct {
	secret         []byte
	db             *store.Store
	tmpl           *template.Template
	billingEnabled bool
}

func New(secret []byte, db *store.Store, billingEnabled bool) *Handler {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTmpl))
	return &Handler{secret: secret, db: db, tmpl: tmpl, billingEnabled: billingEnabled}
}

type pageData struct {
	Login          string
	Plan           string
	Scans          []store.ScanRecord
	AppInstalled   bool
	BillingEnabled bool
	Upgraded       bool
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Get(r, h.secret)
	if !ok {
		http.Redirect(w, r, "/auth/github", http.StatusFound)
		return
	}

	data := pageData{
		Login:          sess.GitHubLogin,
		AppInstalled:   sess.InstallationID != 0,
		BillingEnabled: h.billingEnabled,
		Upgraded:       r.URL.Query().Get("upgraded") == "1",
	}

	if h.db != nil && sess.InstallationID != 0 {
		inst, err := h.db.GetInstallation(r.Context(), sess.InstallationID)
		if err == nil {
			data.Plan = inst.Plan
		}
		scans, err := h.db.RecentScans(r.Context(), sess.InstallationID, 50)
		if err == nil {
			data.Scans = scans
		}
	}

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		log.Printf("dashboard template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

var dashboardTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Evil Merge Detector — Dashboard</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0d1117; color: #e6edf3; min-height: 100vh; }
header { background: #161b22; border-bottom: 1px solid #30363d; padding: 0 24px; height: 60px; display: flex; align-items: center; justify-content: space-between; }
header h1 { font-size: 16px; font-weight: 600; color: #e6edf3; }
header h1 span { color: #f85149; }
.user { display: flex; align-items: center; gap: 12px; font-size: 14px; color: #8b949e; }
.plan { background: #21262d; border: 1px solid #30363d; border-radius: 12px; padding: 2px 10px; font-size: 12px; color: #58a6ff; text-transform: uppercase; letter-spacing: .5px; }
.plan.pro { color: #3fb950; }
a.logout { color: #8b949e; text-decoration: none; font-size: 13px; }
a.logout:hover { color: #e6edf3; }
main { max-width: 960px; margin: 40px auto; padding: 0 24px; }
h2 { font-size: 20px; font-weight: 600; margin-bottom: 20px; }
.empty { color: #8b949e; font-size: 14px; padding: 40px 0; text-align: center; }
.install-notice { background: #1c2128; border: 1px solid #388bfd; border-radius: 8px; padding: 20px 24px; margin-bottom: 32px; font-size: 14px; color: #e6edf3; }
.install-notice a { color: #58a6ff; }
.upgrade-banner { background: #1c2128; border: 1px solid #30363d; border-radius: 8px; padding: 20px 24px; margin-bottom: 32px; display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.upgrade-banner p { font-size: 14px; color: #8b949e; margin: 0; }
.upgrade-banner strong { color: #e6edf3; }
.upgrade-actions { display: flex; gap: 8px; flex-shrink: 0; }
.btn-upgrade { background: #238636; color: #fff; border: none; border-radius: 6px; padding: 7px 16px; font-size: 13px; font-weight: 500; cursor: pointer; text-decoration: none; }
.btn-upgrade:hover { background: #2ea043; }
.btn-upgrade.secondary { background: transparent; border: 1px solid #30363d; color: #8b949e; }
.btn-upgrade.secondary:hover { color: #e6edf3; border-color: #8b949e; }
.success-banner { background: #0f2a1a; border: 1px solid #238636; border-radius: 8px; padding: 16px 24px; margin-bottom: 32px; font-size: 14px; color: #3fb950; }
.manage-link { font-size: 13px; color: #8b949e; }
table { width: 100%; border-collapse: collapse; font-size: 13px; }
th { text-align: left; padding: 8px 12px; color: #8b949e; font-weight: 500; border-bottom: 1px solid #21262d; }
td { padding: 10px 12px; border-bottom: 1px solid #21262d; vertical-align: middle; }
tr:hover td { background: #161b22; }
.repo { font-weight: 600; color: #58a6ff; }
.pr { color: #8b949e; }
.sha { font-family: monospace; font-size: 12px; color: #8b949e; }
.evil { font-weight: 700; }
.evil.bad { color: #f85149; }
.evil.ok { color: #3fb950; }
.dur { color: #8b949e; }
</style>
</head>
<body>
<header>
  <h1>Evil <span>Merge</span> Detector</h1>
  <div class="user">
    <span>{{.Login}}</span>
    {{if .Plan}}<span class="plan {{if eq .Plan "pro"}}pro{{end}}">{{.Plan}}</span>{{end}}
    <a class="logout" href="/auth/logout">Sign out</a>
  </div>
</header>
<main>
  {{if .Upgraded}}
  <div class="success-banner">&#10003; Upgraded to Pro — thank you! Your account now has unlimited scans.</div>
  {{end}}
  {{if not .AppInstalled}}
  <div class="install-notice">
    GitHub App is not installed for your account yet.
    <a href="https://github.com/apps/evil-merge-detector" target="_blank" rel="noopener">Install it</a>
    to start scanning pull requests automatically.
  </div>
  {{end}}
  {{if and .BillingEnabled .AppInstalled}}
  {{if eq .Plan "free"}}
  <div class="upgrade-banner">
    <p><strong>Free plan</strong> — 50 PR scans per month. Upgrade to Pro for unlimited scans on private repositories.</p>
    <div class="upgrade-actions">
      <form method="POST" action="/billing/checkout" style="display:inline">
        <input type="hidden" name="period" value="monthly">
        <button type="submit" class="btn-upgrade">Upgrade $7/mo</button>
      </form>
      <form method="POST" action="/billing/checkout" style="display:inline">
        <input type="hidden" name="period" value="yearly">
        <button type="submit" class="btn-upgrade secondary">$67/yr</button>
      </form>
    </div>
  </div>
  {{end}}
  {{if eq .Plan "pro"}}
  <div style="margin-bottom:24px">
    <form method="POST" action="/billing/portal" style="display:inline">
      <button type="submit" class="btn-upgrade secondary manage-link">Manage subscription</button>
    </form>
  </div>
  {{end}}
  {{end}}
  <h2>Recent scans</h2>
  {{if not .Scans}}
  <p class="empty">No scans yet. Open a pull request in a repository where the app is installed.</p>
  {{else}}
  <table>
    <thead>
      <tr>
        <th>Repository</th>
        <th>PR</th>
        <th>Commit</th>
        <th>Evil merges</th>
        <th>Total merges</th>
        <th>Duration</th>
        <th>Scanned at</th>
      </tr>
    </thead>
    <tbody>
      {{range .Scans}}
      <tr>
        <td class="repo">{{.Owner}}/{{.Repo}}</td>
        <td class="pr">#{{.PRNumber}}</td>
        <td class="sha">{{slice .HeadSHA 0 7}}</td>
        <td class="evil {{if gt .EvilMerges 0}}bad{{else}}ok{{end}}">{{.EvilMerges}}</td>
        <td>{{.TotalMerges}}</td>
        <td class="dur">{{.DurationMs}}ms</td>
        <td>{{.ScannedAt.UTC.Format "2006-01-02 15:04"}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
  {{end}}
</main>
</body>
</html>`
