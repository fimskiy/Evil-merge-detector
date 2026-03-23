package dashboard

import (
	"html/template"
	"net/http"
	"time"

	"github.com/fimskiy/evil-merge-detector/app/internal/session"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type Handler struct {
	secret []byte
	db     *store.Store
	tmpl   *template.Template
}

func New(secret []byte, db *store.Store) *Handler {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTmpl))
	return &Handler{secret: secret, db: db, tmpl: tmpl}
}

type pageData struct {
	Login string
	Plan  string
	Scans []store.ScanRecord
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Get(r, h.secret)
	if !ok {
		http.Redirect(w, r, "/auth/github", http.StatusFound)
		return
	}

	data := pageData{Login: sess.GitHubLogin}

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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.Execute(w, data)
}

func fmtTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04")
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
