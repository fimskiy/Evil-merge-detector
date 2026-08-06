package admin

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/fimskiy/evil-merge-detector/app/internal/session"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type Handler struct {
	secret     []byte
	db         *store.Store
	adminLogin string
	tmpl       *template.Template
}

func New(secret []byte, db *store.Store, adminLogin string) *Handler {
	tmpl := template.Must(template.New("admin").Parse(adminTmpl))
	return &Handler{secret: secret, db: db, adminLogin: adminLogin, tmpl: tmpl}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Get(r, h.secret)
	if !ok {
		http.Redirect(w, r, "/auth/github", http.StatusFound)
		return
	}
	if h.adminLogin == "" || sess.GitHubLogin != h.adminLogin {
		http.NotFound(w, r)
		return
	}
	if h.db == nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	stats, err := h.db.PlatformStats(r.Context())
	if err != nil {
		log.Printf("admin stats: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, stats); err != nil {
		log.Printf("admin template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

var adminTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Evil Merge Detector — Admin</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0d1117; color: #e6edf3; min-height: 100vh; }
header { background: #161b22; border-bottom: 1px solid #30363d; padding: 0 24px; height: 60px; display: flex; align-items: center; justify-content: space-between; }
header h1 { font-size: 16px; font-weight: 600; color: #e6edf3; }
header h1 span { color: #f85149; }
a.back { color: #8b949e; text-decoration: none; font-size: 13px; }
a.back:hover { color: #e6edf3; }
main { max-width: 720px; margin: 40px auto; padding: 0 24px; }
h2 { font-size: 20px; font-weight: 600; margin-bottom: 20px; }
.stats { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.stat { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 24px; }
.stat-val { font-size: 32px; font-weight: 700; color: #58a6ff; }
.stat-label { font-size: 13px; color: #8b949e; margin-top: 6px; }
</style>
</head>
<body>
<header>
  <h1>Evil <span>Merge</span> Detector — Admin</h1>
  <a class="back" href="/dashboard">Back to dashboard</a>
</header>
<main>
  <h2>Platform usage</h2>
  <div class="stats">
    <div class="stat">
      <div class="stat-val">{{.TotalInstallations}}</div>
      <div class="stat-label">Installations</div>
    </div>
    <div class="stat">
      <div class="stat-val">{{.TotalScans}}</div>
      <div class="stat-label">Total scans</div>
    </div>
    <div class="stat">
      <div class="stat-val">{{.ScansLast30Days}}</div>
      <div class="stat-label">Scans (last 30 days)</div>
    </div>
    <div class="stat">
      <div class="stat-val">{{.TotalEvilMerges}}</div>
      <div class="stat-label">Evil merges found</div>
    </div>
  </div>
</main>
</body>
</html>`
