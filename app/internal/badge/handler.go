package badge

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

// ScanStore is the subset of store.Store required by the badge handler.
type ScanStore interface {
	LastScan(ctx context.Context, owner, repo string) (*store.ScanRecord, error)
}

// Handler serves SVG status badges at /badge/{owner}/{repo}.
type Handler struct {
	db ScanStore
}

func New(db ScanStore) http.Handler {
	return &Handler{db: db}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/badge/")
	path = strings.TrimSuffix(path, ".svg")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "usage: /badge/{owner}/{repo}", http.StatusBadRequest)
		return
	}
	owner, repo := parts[0], parts[1]

	label := "evil merges"
	value := "unknown"
	color := "#9f9f9f"

	if h.db != nil {
		if rec, err := h.db.LastScan(r.Context(), owner, repo); err == nil && rec != nil {
			if rec.EvilMerges == 0 {
				value = "passing"
				color = "#4c1"
			} else {
				value = fmt.Sprintf("%d found", rec.EvilMerges)
				color = "#e05d44"
			}
		}
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, max-age=0")
	writeBadge(w, label, value, color)
}

func charWidth(s string) int {
	return len(s)*7 + 20
}

var badgeTmpl = template.Must(template.New("badge").Parse(`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="20">
  <linearGradient id="s" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="{{.W}}" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="{{.LW}}" height="20" fill="#555"/>
    <rect x="{{.LW}}" width="{{.VW}}" height="20" fill="{{.Color}}"/>
    <rect width="{{.W}}" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="{{.LX}}" y="15" fill="#010101" fill-opacity=".3">{{.Label}}</text>
    <text x="{{.LX}}" y="14">{{.Label}}</text>
    <text x="{{.VX}}" y="15" fill="#010101" fill-opacity=".3">{{.Value}}</text>
    <text x="{{.VX}}" y="14">{{.Value}}</text>
  </g>
</svg>`))

func writeBadge(w http.ResponseWriter, label, value, color string) {
	lw := charWidth(label)
	vw := charWidth(value)
	_ = badgeTmpl.Execute(w, map[string]any{
		"W":     lw + vw,
		"LW":    lw,
		"VW":    vw,
		"LX":    lw / 2,
		"VX":    lw + vw/2,
		"Label": label,
		"Value": value,
		"Color": color,
	})
}
