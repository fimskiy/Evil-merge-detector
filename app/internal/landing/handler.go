package landing

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed og-image.png
var ogImage []byte

var tmpl = template.Must(template.New("landing").Parse(page))

var euCountries = map[string]bool{
	"AT": true, "BE": true, "BG": true, "CY": true, "CZ": true,
	"DE": true, "DK": true, "EE": true, "ES": true, "FI": true,
	"FR": true, "GR": true, "HR": true, "HU": true, "IE": true,
	"IT": true, "LT": true, "LU": true, "LV": true, "MT": true,
	"NL": true, "PL": true, "PT": true, "RO": true, "SE": true,
	"SI": true, "SK": true,
	"IS": true, "LI": true, "NO": true,
	"GB": true, "CH": true,
}

type pageData struct {
	ShowCookieBanner bool
}

func Handler(w http.ResponseWriter, r *http.Request) {
	country := r.Header.Get("CF-IPCountry")
	data := pageData{
		ShowCookieBanner: country == "" || euCountries[country],
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("landing template: %v", err)
	}
}

func OGImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(ogImage)
}

var privacyTmpl = template.Must(template.New("privacy").Parse(privacyPage))

func PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := privacyTmpl.Execute(w, nil); err != nil {
		log.Printf("privacy template: %v", err)
	}
}

var page = `<!DOCTYPE html>
<html lang="en">
<head>
<!-- GTM loads after cookie consent (see JS at bottom) -->
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Detect Evil Merge Commits in Git — Evil Merge Detector</title>
<meta name="description" content="Detect code injection via Git merge commits — a supply chain attack vector invisible to standard git security tools. CLI, GitHub Action, and GitHub App. Free for open source.">
<meta name="robots" content="index, follow, max-snippet:-1, max-image-preview:large, max-video-preview:-1">
<link rel="canonical" href="https://evilmerge.dev/">
<meta property="og:type" content="website">
<meta property="og:title" content="Detect Evil Merge Commits in Git — Evil Merge Detector">
<meta property="og:description" content="Detect hidden code injected into Git merge commits. Free CLI, GitHub Action, and GitHub App. Zero false positives.">
<meta property="og:url" content="https://evilmerge.dev/">
<meta property="og:site_name" content="Evil Merge Detector">
<meta property="og:image" content="https://evilmerge.dev/og-image.png">
<meta property="og:image:width" content="1200">
<meta property="og:image:height" content="630">
<meta property="og:image:alt" content="Evil Merge Detector — terminal showing CRITICAL evil merge detected">
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Detect Evil Merge Commits in Git — Evil Merge Detector">
<meta name="twitter:description" content="Find merge commits that inject hidden code. CLI, GitHub Action, GitHub App. Free for open source.">
<meta name="twitter:image" content="https://evilmerge.dev/og-image.png">
<meta name="twitter:image:alt" content="Evil Merge Detector terminal output showing evil merge detection">
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "Evil Merge Detector",
  "url": "https://evilmerge.dev/",
  "description": "Detect evil merge commits — merge commits that introduce code changes not present in either parent branch. Available as CLI, GitHub Action, and GitHub App.",
  "applicationCategory": "DeveloperApplication",
  "operatingSystem": "Linux, macOS, Windows",
  "offers": [
    {
      "@type": "Offer",
      "price": "0",
      "priceCurrency": "USD",
      "name": "Free",
      "description": "Public repositories, 50 PR scans per month"
    },
    {
      "@type": "Offer",
      "price": "7",
      "priceCurrency": "USD",
      "name": "Pro",
      "description": "Public and private repositories, unlimited scans"
    }
  ],
  "featureList": [
    "Evil merge commit detection",
    "Supply chain attack prevention",
    "Code injection detection via git merge history",
    "Git security scanning",
    "Detects GitHub-acknowledged design vulnerability in PR merge workflow",
    "Zero false positives",
    "CLI tool",
    "GitHub Action",
    "GitHub App",
    "SARIF output for GitHub Code Scanning",
    "Works offline"
  ],
  "isAccessibleForFree": true,
  "downloadUrl": "https://github.com/evilmerge-dev/Evil-merge-detector",
  "installUrl": "https://github.com/apps/evil-merge-detector"
}
</script>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg viewBox='0 0 32 32' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M16 2L28 6L28 18C28 24 22 29 16 31C10 29 4 24 4 18L4 6Z' fill='%23dc2626'/%3E%3Cpath d='M16 5L26 8.5L26 18C26 23 21 27 16 29C11 27 6 23 6 18L6 8.5Z' fill='%23b91c1c'/%3E%3Ccircle cx='16' cy='12' r='2.5' fill='white'/%3E%3Ccircle cx='11' cy='17' r='2' fill='white'/%3E%3Ccircle cx='16' cy='22' r='2.2' fill='white'/%3E%3Ccircle cx='21' cy='17' r='2.5' fill='%237f1d1d'/%3E%3Cline x1='20' y1='16' x2='22' y2='18' stroke='white' stroke-width='1.2' stroke-linecap='round'/%3E%3Cline x1='22' y1='16' x2='20' y2='18' stroke='white' stroke-width='1.2' stroke-linecap='round'/%3E%3C/svg%3E">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
:root {
  --bg: #ffffff;
  --bg-soft: #f8fafc;
  --bg-code: #f1f5f9;
  --border: #e2e8f0;
  --border-strong: #cbd5e1;
  --red: #dc2626;
  --red-soft: #fef2f2;
  --red-border: #fecaca;
  --text: #0f172a;
  --text-mid: #334155;
  --text-muted: #64748b;
  --green: #16a34a;
  --sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --mono: ui-monospace, 'SF Mono', 'Cascadia Code', 'Fira Code', Consolas, monospace;
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.08), 0 1px 2px rgba(0,0,0,0.04);
  --shadow-md: 0 4px 16px rgba(0,0,0,0.07), 0 2px 6px rgba(0,0,0,0.04);
  --shadow-lg: 0 12px 40px rgba(0,0,0,0.08), 0 4px 12px rgba(0,0,0,0.04);
  --radius: 8px;
  --ease: cubic-bezier(0.16, 1, 0.3, 1);
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html {
  scroll-behavior: smooth;
  scroll-padding-top: 64px;
}

body {
  font-family: var(--sans);
  background: var(--bg);
  color: var(--text);
  line-height: 1.6;
  overflow-x: hidden;
  -webkit-font-smoothing: antialiased;
}

a { color: inherit; text-decoration: none; }

:focus-visible {
  outline: 2px solid var(--red);
  outline-offset: 3px;
  border-radius: 3px;
}

/* NAV */
nav {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 48px;
  height: 60px;
  background: rgba(255,255,255,0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}

.nav-logo {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 10px;
}
.nav-logo svg { width: 32px; height: 32px; flex-shrink: 0; }
.nav-logo .brand-name { color: var(--text); }
.nav-logo .accent { color: var(--red); }

.nav-links {
  display: flex;
  gap: 32px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-muted);
}
.nav-links a { transition: color 0.15s; }
.nav-links a:hover { color: var(--text); }

.nav-cta {
  font-size: 13px;
  font-weight: 600;
  padding: 8px 18px;
  background: var(--red);
  color: #fff;
  border-radius: 6px;
  transition: background 0.15s, transform 0.15s var(--ease);
}
.nav-cta:hover {
  background: #b91c1c;
  transform: translateY(-1px);
}

/* ANIMATIONS */
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* Aurora blob animations */
@keyframes blobFloat {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(20px, -20px) scale(1.05); }
  66% { transform: translate(-15px, 15px) scale(0.95); }
}

/* Terminal entrance animation */
@keyframes terminalEntrance {
  from { opacity: 0; transform: translateY(30px) scale(0.98); }
  to   { opacity: 1; transform: translateY(0) scale(1); }
}

.hero-eyebrow { animation: fadeUp 0.45s var(--ease) 0.05s both; }
.hero h1      { animation: fadeUp 0.5s var(--ease) 0.15s both; }
.hero-sub     { animation: fadeUp 0.5s var(--ease) 0.25s both; }
.hero-actions { animation: fadeUp 0.45s var(--ease) 0.32s both; }

/* HERO */
.hero-wrap {
  position: relative;
  background:
    radial-gradient(ellipse 55% 60% at 90% 10%, rgba(220,38,38,0.18) 0%, transparent 65%),
    radial-gradient(ellipse 50% 55% at 10% 90%, rgba(99,102,241,0.15) 0%, transparent 65%),
    radial-gradient(ellipse 40% 40% at 50% 50%, rgba(220,38,38,0.05) 0%, transparent 70%),
    #ffffff;
}

.blob-1 { display: none; }
.blob-2 { display: none; }

.hero {
  position: relative;
  z-index: 1;
  max-width: 860px;
  margin: 0 auto;
  padding: 100px 24px 80px;
  text-align: center;
}

.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--red);
  background: var(--red-soft);
  border: 1px solid var(--red-border);
  padding: 5px 14px;
  border-radius: 100px;
  margin-bottom: 32px;
}
.hero-eyebrow .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--red);
}

h1 {
  font-size: clamp(2.25rem, 5vw + 0.5rem, 3.5rem);
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.03em;
  color: var(--text);
  margin-bottom: 20px;
}
h1 .threat { color: var(--red); }
.cursor {
  display: inline-block;
  width: 3px;
  height: 0.85em;
  background: var(--red);
  margin-left: 4px;
  vertical-align: middle;
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.hero-sub {
  font-size: 18px;
  color: var(--text-muted);
  max-width: 520px;
  margin: 0 auto 36px;
  line-height: 1.7;
  font-weight: 400;
}

.hero-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
  align-items: flex-start;
  flex-wrap: wrap;
  margin-bottom: 56px;
}

.cta-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}
.cta-note {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 400;
}

/* Primary button with animated gradient */
.btn-primary {
  font-size: 14px;
  font-weight: 600;
  padding: 11px 24px;
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  color: #fff;
  border-radius: 7px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(220,38,38,0.25);
}
.btn-primary:hover {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 8px 24px rgba(220,38,38,0.35);
  transform: translateY(-2px);
}
.btn-primary:active { transform: translateY(0); }

.btn-secondary {
  font-size: 14px;
  font-weight: 500;
  padding: 11px 24px;
  background: var(--bg);
  color: var(--text-mid);
  border: 1px solid var(--border-strong);
  border-radius: 7px;
  transition: border-color 0.15s, color 0.15s, transform 0.15s var(--ease);
}
.btn-secondary:hover {
  border-color: #94a3b8;
  color: var(--text);
  transform: translateY(-1px);
}
.btn-secondary:active { transform: translateY(0); }

/* TERMINAL */
.terminal {
  background: #f8fafc;
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  max-width: 600px;
  margin: 0 auto;
  text-align: left;
  box-shadow: var(--shadow-lg);
  animation: terminalEntrance 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.5s both;
}
.terminal-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  background: #f1f5f9;
  border-bottom: 1px solid var(--border);
}
.dot-r { width: 10px; height: 10px; border-radius: 50%; background: #ff5f57; }
.dot-y { width: 10px; height: 10px; border-radius: 50%; background: #ffbd2e; }
.dot-g { width: 10px; height: 10px; border-radius: 50%; background: #28c840; }
.terminal-title {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text-muted);
  margin: 0 auto;
}
.terminal-body {
  padding: 20px 24px;
  font-family: var(--mono);
  font-size: 13.5px;
  line-height: 1.9;
  color: var(--text-mid);
  min-height: 260px;
}
.t-dim    { color: var(--text-muted); }
.t-prompt { color: var(--text-muted); }
.t-cmd    { color: #2563eb; }
.t-bad    { color: var(--red); font-weight: 600; }
.t-ok     { color: var(--green); }
.t-cursor {
  display: inline-block;
  color: var(--text-muted);
  animation: blink 1s step-end infinite;
}
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }

/* STATS BAR */
.stats-bar {
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
  padding: 28px 24px;
  background: var(--bg-soft);
}
.stats-inner {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 48px;
  flex: 1;
  max-width: 200px;
}
.stat + .stat { border-left: 1px solid var(--border); }
.stat-val {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text);
  line-height: 1.2;
}
.stat-val .accent { color: var(--red); }

/* "0" false positives highlight */
.stat-val.zero {
  font-size: 2.5rem;
  font-weight: 800;
  color: #15803d;
  letter-spacing: -0.03em;
}

.stat-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  margin-top: 4px;
  white-space: nowrap;
}

/* SECTIONS */
section {
  padding: 96px 24px;
}
.section-inner { max-width: 1000px; margin: 0 auto; }

/* Section label pills */
.label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(220,38,38,0.08);
  border: 1px solid rgba(220,38,38,0.2);
  color: var(--red);
  padding: 4px 12px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  margin-bottom: 20px;
}

h2 {
  font-size: clamp(1.5rem, 2.5vw + 0.5rem, 2rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text);
  margin-bottom: 12px;
  line-height: 1.2;
}

.section-sub {
  font-size: 17px;
  color: var(--text-muted);
  max-width: 500px;
  margin-bottom: 48px;
  line-height: 1.7;
}

/* DIVIDER */
.divider {
  height: 1px;
  background: var(--border);
  margin: 0 48px;
}

/* PROBLEM SECTION */
.problem-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
}

.problem-text p {
  font-size: 15.5px;
  color: var(--text-mid);
  margin-bottom: 16px;
  line-height: 1.75;
}
.problem-text p strong { color: var(--text); font-weight: 600; }
.problem-text code {
  font-family: var(--mono);
  font-size: 12.5px;
  background: var(--bg-code);
  border: 1px solid var(--border);
  padding: 1px 6px;
  border-radius: 4px;
  color: var(--text-mid);
}

.diff-card {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  box-shadow:
    0 1px 2px rgba(0,0,0,0.04),
    0 4px 8px rgba(0,0,0,0.04),
    0 8px 16px rgba(0,0,0,0.04);
  transition: box-shadow 0.25s var(--ease), transform 0.25s var(--ease);
}
.diff-card:hover {
  box-shadow:
    0 2px 4px rgba(0,0,0,0.06),
    0 8px 16px rgba(0,0,0,0.08),
    0 16px 32px rgba(0,0,0,0.08);
  transform: translateY(-4px);
}
.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--bg-soft);
  border-bottom: 1px solid var(--border);
}
.diff-filename {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-muted);
}
.diff-badge {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 2px 8px;
  background: var(--red-soft);
  color: var(--red);
  border: 1px solid var(--red-border);
  border-radius: 4px;
}
.diff-body {
  padding: 20px;
  font-family: var(--mono);
  font-size: 13px;
  line-height: 2;
}
.diff-line { display: flex; gap: 14px; }
.diff-label { color: var(--text-muted); min-width: 70px; flex-shrink: 0; font-size: 12px; }
.diff-value { color: var(--text-mid); }
.diff-value.clean { color: var(--green); }
.diff-value.bad   { color: var(--red); font-weight: 600; }
.diff-sep { height: 1px; background: var(--border); margin: 14px 0; }
.diff-note {
  font-size: 12.5px;
  color: var(--text-muted);
  line-height: 1.65;
  font-style: italic;
  font-family: var(--sans);
}

/* HOW IT WORKS */
.steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.step {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 32px 28px;
  box-shadow:
    0 1px 2px rgba(0,0,0,0.04),
    0 4px 8px rgba(0,0,0,0.04),
    0 8px 16px rgba(0,0,0,0.04);
  transition: box-shadow 0.25s var(--ease), border-color 0.25s, transform 0.25s var(--ease);
}
.step:hover {
  box-shadow:
    0 2px 4px rgba(0,0,0,0.06),
    0 8px 16px rgba(0,0,0,0.08),
    0 16px 32px rgba(0,0,0,0.08);
  border-color: var(--border-strong);
  transform: translateY(-4px);
}
.step-num {
  font-size: 13px;
  font-weight: 700;
  color: var(--red);
  margin-bottom: 16px;
  font-family: var(--mono);
  letter-spacing: 0.04em;
}
.step h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 10px;
  letter-spacing: -0.01em;
}
.step p {
  font-size: 14.5px;
  color: var(--text-muted);
  line-height: 1.65;
}

/* INTEGRATIONS */
.integrations {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.integration {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 28px;
  box-shadow:
    0 1px 2px rgba(0,0,0,0.04),
    0 4px 8px rgba(0,0,0,0.04),
    0 8px 16px rgba(0,0,0,0.04);
  transition: box-shadow 0.25s var(--ease), border-color 0.25s, transform 0.25s var(--ease);
}
.integration:hover {
  box-shadow:
    0 2px 4px rgba(0,0,0,0.06),
    0 8px 16px rgba(0,0,0,0.08),
    0 16px 32px rgba(0,0,0,0.08);
  border-color: var(--border-strong);
  transform: translateY(-4px);
}
.int-badge {
  display: inline-block;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-mid);
  background: var(--bg-code);
  border: 1px solid var(--border);
  padding: 3px 8px;
  border-radius: 4px;
  margin-bottom: 16px;
  font-family: var(--mono);
}
.integration .int-badge.cli    { background: #f0fdf4; border: 1px solid #86efac; color: #16a34a; }
.integration .int-badge.action { background: #f5f3ff; border: 1px solid #c4b5fd; color: #7c3aed; }
.integration .int-badge.app    { background: #eff6ff; border: 1px solid #93c5fd; color: #2563eb; }
.integration:has(.cli):hover   { border-color: #86efac; }
.integration:has(.action):hover { border-color: #c4b5fd; }
.integration:has(.app):hover   { border-color: #93c5fd; }
.integration h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 8px;
  letter-spacing: -0.01em;
}
.integration p {
  font-size: 14.5px;
  color: var(--text-muted);
  margin-bottom: 18px;
  line-height: 1.6;
}
.integration pre {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-left: 3px solid #dc2626;
  border-radius: 0 8px 8px 0;
  padding: 16px 20px;
  font-family: var(--mono);
  font-size: 13px;
  line-height: 1.7;
  color: #1e293b;
  overflow-x: auto;
}

/* PRICING TOGGLE */
.pricing-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 32px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-muted);
}
.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
  background: var(--border-strong);
  border-radius: 99px;
  cursor: pointer;
  transition: background 0.2s;
  border: none;
  outline: none;
  flex-shrink: 0;
}
.toggle-switch.annual { background: var(--red); }
.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s cubic-bezier(0.16,1,0.3,1);
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-switch.annual .toggle-knob { transform: translateX(20px); }
.save-badge {
  font-size: 11px;
  font-weight: 600;
  background: #f0fdf4;
  border: 1px solid #86efac;
  color: #16a34a;
  padding: 2px 8px;
  border-radius: 99px;
  letter-spacing: 0.04em;
}
.toggle-label { transition: color 0.15s; }
.toggle-label.active { color: var(--text); }

/* PRICING */
.pricing-wrap {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  max-width: 680px;
  margin: 0 auto;
  padding-top: 20px;
}
.plan {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 32px;
  display: flex;
  flex-direction: column;
  box-shadow:
    0 1px 2px rgba(0,0,0,0.04),
    0 4px 8px rgba(0,0,0,0.04),
    0 8px 16px rgba(0,0,0,0.04);
}

/* Pro card — elevated and glowing */
.plan.featured {
  border: 2px solid #dc2626;
  background: #fffafa;
  position: relative;
  z-index: 1;
  box-shadow:
    0 4px 8px rgba(220,38,38,0.1),
    0 12px 24px rgba(220,38,38,0.12),
    0 24px 48px rgba(220,38,38,0.08);
}
.plan.featured::before {
  content: 'MOST POPULAR';
  position: absolute;
  top: -14px;
  left: 50%;
  transform: translateX(-50%);
  background: #dc2626;
  color: white;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 4px 14px;
  border-radius: 99px;
  white-space: nowrap;
}

.plan-tier {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 14px;
}
.plan-price {
  display: flex;
  align-items: baseline;
  gap: 3px;
  margin-bottom: 6px;
}
.price-currency {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-mid);
}
.price-amount {
  font-size: 48px;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text);
  line-height: 1;
}
.price-period {
  font-size: 13px;
  color: var(--text-muted);
}
.plan-desc {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border);
}
.plan ul {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 11px;
  margin-bottom: 28px;
  flex: 1;
}
.plan li {
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-mid);
}
.plan li .check {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #dcfce7;
  border: 1px solid #86efac;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 9px;
  color: var(--green);
  font-weight: 700;
}
.plan li.no .check {
  background: var(--bg-soft);
  border-color: var(--border);
  color: var(--text-muted);
}
.plan li.no { color: var(--text-muted); }

.plan-btn {
  display: block;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  padding: 11px 20px;
  border-radius: 7px;
  border: 1px solid var(--border-strong);
  color: var(--text-mid);
  transition: border-color 0.15s, color 0.15s, background 0.15s, transform 0.15s var(--ease);
}
.plan-btn:hover {
  border-color: #94a3b8;
  color: var(--text);
  transform: translateY(-1px);
}
.plan.featured .plan-btn {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  border-color: var(--red);
  color: #fff;
  box-shadow: 0 4px 12px rgba(220,38,38,0.25);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.plan.featured .plan-btn:hover {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border-color: #ef4444;
  box-shadow: 0 8px 24px rgba(220,38,38,0.35);
  transform: translateY(-2px);
}

.plan-note {
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 10px;
}

/* BOTTOM CTA */
.cta-section {
  text-align: center;
  background: var(--bg-soft);
  border-top: 1px solid var(--border);
}
.cta-section h2 {
  max-width: 560px;
  margin: 0 auto 14px;
}
.cta-section .section-sub {
  margin: 0 auto 32px;
}
.cta-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
  flex-wrap: wrap;
}

/* FOOTER */
footer {
  padding: 32px 48px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg);
}
.footer-logo {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
}
.footer-logo .accent { color: var(--red); }
.footer-links {
  display: flex;
  gap: 28px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
}
.footer-links a { transition: color 0.15s; }
.footer-links a:hover { color: var(--text); }
.footer-gh-link { color: #2563eb; }
.footer-gh-link:hover { color: #1d4ed8; }

/* Link arrow micro-interaction */
.link-arrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: gap 0.2s ease;
}
.link-arrow::after {
  content: '\2192';
  transition: transform 0.2s ease;
  display: inline-block;
}
.link-arrow:hover {
  gap: 10px;
}
.link-arrow:hover::after {
  transform: translateX(3px);
}

/* Scroll-triggered reveal */
.reveal {
  opacity: 0;
  transform: translateY(20px);
  transition: opacity 0.6s cubic-bezier(0.16, 1, 0.3, 1), transform 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}
.reveal.visible {
  opacity: 1;
  transform: translateY(0);
}
.reveal:nth-child(2) { transition-delay: 0.1s; }
.reveal:nth-child(3) { transition-delay: 0.2s; }
.reveal:nth-child(4) { transition-delay: 0.3s; }

/* FAQ */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
}
.faq-item {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  transition: border-color 0.15s;
}
.faq-item[open] {
  border-color: var(--border-strong);
}
.faq-item + .faq-item { border-top: none; border-radius: 0; }
.faq-item:first-child { border-radius: var(--radius) var(--radius) 0 0; }
.faq-item:last-child  { border-radius: 0 0 var(--radius) var(--radius); }
.faq-q {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 20px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  cursor: pointer;
  list-style: none;
  user-select: none;
  gap: 16px;
}
.faq-q::-webkit-details-marker { display: none; }
.faq-q::after {
  content: '+';
  font-size: 20px;
  font-weight: 300;
  color: var(--text-muted);
  flex-shrink: 0;
  transition: transform 0.2s var(--ease);
}
.faq-item[open] .faq-q::after {
  transform: rotate(45deg);
}
.faq-a {
  padding: 0 20px 18px;
  font-size: 14.5px;
  color: var(--text-muted);
  line-height: 1.7;
}

/* Mobile nav toggle */
.nav-menu-btn {
  display: none;
  background: none;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 6px 10px;
  cursor: pointer;
  color: var(--text-mid);
  font-size: 18px;
  line-height: 1;
}
.mobile-nav {
  display: none;
  flex-direction: column;
  gap: 0;
  background: var(--bg);
  border-bottom: 1px solid var(--border);
  padding: 8px 0;
}
.mobile-nav a {
  padding: 12px 20px;
  font-size: 15px;
  font-weight: 500;
  color: var(--text-mid);
  border-bottom: 1px solid var(--border);
  transition: color 0.15s, background 0.15s;
}
.mobile-nav a:last-child { border-bottom: none; }
.mobile-nav a:hover { color: var(--text); background: var(--bg-soft); }
.mobile-nav.open { display: flex; }

/* REDUCED MOTION */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
  .reveal { opacity: 1; transform: none; transition: none; }
}

/* TABLET (1024px) — 3-col → 2-col */
@media (max-width: 1024px) {
  .steps,
  .integrations { grid-template-columns: repeat(2, 1fr); }
}

/* MOBILE (768px) */
@media (max-width: 768px) {
  nav { padding: 0 20px; }
  .nav-links { display: none; }
  .nav-menu-btn { display: block; }

  .hero { padding: 56px 20px 48px; }
  .hero h1 { font-size: clamp(1.75rem, 8vw, 2.5rem); }
  .hero-sub { font-size: 16px; }
  .hero-actions { flex-direction: column; align-items: center; }
  .btn-primary, .btn-secondary { width: 100%; max-width: 320px; text-align: center; justify-content: center; }

  .terminal { max-width: 100%; }
  .terminal-body { font-size: 12px; padding: 14px 16px; }
  .terminal-body pre, .integration pre { overflow-x: auto; white-space: pre; }

  .stats-inner { flex-wrap: wrap; }
  .stat { min-width: 50%; padding: 20px 0; border-left: none !important; }
  .stat:nth-child(even) { border-left: 1px solid var(--border) !important; }
  .stat:nth-child(1), .stat:nth-child(2) { border-bottom: 1px solid var(--border); }

  .problem-grid,
  .steps,
  .integrations,
  .pricing-wrap { grid-template-columns: 1fr; }

  .integration pre { overflow-x: auto; font-size: 12px; }

  .pricing-toggle { flex-wrap: wrap; gap: 8px; }
  .plan.featured { transform: none; }
  .plan.featured::before { font-size: 10px; }

  section { padding: 56px 20px; }
  .section-sub { font-size: 15px; }
  .divider { margin: 0 20px; }

  footer { flex-direction: column; gap: 16px; text-align: center; padding: 28px 20px; }
  .footer-links { justify-content: center; flex-wrap: wrap; gap: 16px; }
}

/* SMALL PHONES (390px) */
@media (max-width: 390px) {
  .hero h1 { font-size: 1.65rem; }
  .stat-val.zero { font-size: 1.75rem; }
  .price-amount { font-size: 36px; }
  nav { padding: 0 16px; }
  section { padding: 48px 16px; }
}

/* COOKIE BANNER */
#cookie-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 600;
  background: var(--text);
  color: #e2e8f0;
  padding: 14px 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  font-size: 13.5px;
  line-height: 1.5;
  box-shadow: 0 -4px 20px rgba(0,0,0,0.15);
}
#cookie-banner p { margin: 0; }
#cookie-banner a { color: #93c5fd; text-decoration: underline; text-underline-offset: 2px; }
.cookie-actions { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
.cookie-btn {
  font-size: 13px;
  font-weight: 600;
  padding: 7px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s;
}
.cookie-btn:hover { opacity: 0.85; }
.cookie-btn-accept { background: var(--red); color: #fff; }
.cookie-btn-decline { background: transparent; color: #94a3b8; border: 1px solid #475569; }
@media (max-width: 640px) {
  #cookie-banner { flex-direction: column; align-items: flex-start; padding: 16px 20px; gap: 12px; }
}

</style>
</head>
<body{{if .ShowCookieBanner}} data-eu="1"{{end}}>

<!-- NAV -->
<nav aria-label="Main navigation">
  <div class="nav-logo">
    <svg viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg" role="img" aria-labelledby="logo-title-v2"><title id="logo-title-v2">Evil Merge Detector logo</title><path d="M256 24L444 88L444 264C444 364 354 446 256 482C158 446 68 364 68 264L68 88Z" fill="#da3633"/><path d="M256 62L418 118L418 264C418 348 338 418 256 450C174 418 94 348 94 264L94 118Z" fill="#b91c1c"/><line x1="256" y1="168" x2="182" y2="256" stroke="white" stroke-width="26" stroke-linecap="round"/><line x1="256" y1="168" x2="330" y2="256" stroke="white" stroke-width="26" stroke-linecap="round"/><line x1="182" y1="256" x2="256" y2="364" stroke="white" stroke-width="26" stroke-linecap="round"/><line x1="330" y1="256" x2="256" y2="364" stroke="white" stroke-width="26" stroke-linecap="round"/><circle cx="256" cy="168" r="28" fill="white"/><circle cx="182" cy="256" r="24" fill="white"/><circle cx="256" cy="364" r="26" fill="white"/><circle cx="330" cy="256" r="32" fill="#7f1d1d"/><line x1="313" y1="239" x2="347" y2="273" stroke="white" stroke-width="12" stroke-linecap="round"/><line x1="347" y1="239" x2="313" y2="273" stroke="white" stroke-width="12" stroke-linecap="round"/></svg>
    <span class="brand-name">Evil Merge <span class="accent">Detector</span></span>
  </div>
  <div class="nav-links">
    <a href="#how-it-works">How it works</a>
    <a href="#pricing">Pricing</a>
    <a href="https://github.com/evilmerge-dev/Evil-merge-detector">GitHub</a>
  </div>
  <a class="nav-cta" id="btn-nav-install" href="https://github.com/apps/evil-merge-detector">Install App</a>
  <button class="nav-menu-btn" id="nav-menu-btn" aria-label="Open menu" aria-expanded="false">&#9776;</button>
</nav>
<div class="mobile-nav" id="mobile-nav" aria-label="Mobile navigation">
  <a href="#how-it-works">How it works</a>
  <a href="#pricing">Pricing</a>
  <a href="https://github.com/evilmerge-dev/Evil-merge-detector">GitHub</a>
  <a href="https://github.com/apps/evil-merge-detector">Install App</a>
</div>

<main>

<!-- HERO -->
<div class="hero-wrap">
  <div class="blob-1" aria-hidden="true"></div>
  <div class="blob-2" aria-hidden="true"></div>

  <div class="hero">
    <div class="hero-eyebrow">
      <span class="dot" aria-hidden="true"></span>
      Open source · CLI + GitHub Action + GitHub App
    </div>

    <h1>
      Detect <span class="threat">evil merge</span> commits<br>before they ship<span class="cursor" aria-hidden="true"></span>
    </h1>

    <p class="hero-sub">
      Evil Merge Detector finds merge commits that introduce changes not present in either parent branch &mdash;
      the attack vector your code review misses.
    </p>

    <div class="hero-actions">
      <div class="cta-group">
        <a class="btn-primary" id="btn-hero-install" href="https://github.com/apps/evil-merge-detector">Start Scanning — Free</a>
        <span class="cta-note">No credit card required</span>
      </div>
      <a class="btn-secondary link-arrow" id="btn-hero-github" href="https://github.com/evilmerge-dev/Evil-merge-detector">View on GitHub</a>
    </div>

    <div class="terminal">
      <div class="terminal-bar">
        <span class="dot-r" aria-hidden="true"></span>
        <span class="dot-y" aria-hidden="true"></span>
        <span class="dot-g" aria-hidden="true"></span>
        <span class="terminal-title">evilmerge &mdash; scan</span>
      </div>
      <div class="terminal-body">
        <div class="t-dim"># Scan your repository</div>
        <div><span class="t-prompt">$ </span><span class="t-cmd" id="typed-cmd"></span><span class="t-cursor">▋</span></div>
        <div id="term-output" style="display:none">
          <div>&nbsp;</div>
          <div class="t-bad">CRITICAL  ab90bd7  vite.config.js</div>
          <div class="t-dim">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Both parents had identical content.</div>
          <div class="t-dim">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Merge result differs &mdash; manual edit detected.</div>
          <div>&nbsp;</div>
          <div class="t-ok">&#10003;  23 merge commits checked &nbsp;|&nbsp; 1 critical issue found</div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- STATS BAR -->
<div class="stats-bar">
  <div class="stats-inner">
    <div class="stat">
      <span class="stat-val zero" data-target="0" data-prefix="" data-suffix="">0</span>
      <span class="stat-label">False positives</span>
    </div>
    <div class="stat">
      <span class="stat-val" data-target="5" data-prefix="" data-suffix="">0</span>
      <span class="stat-label">Integrations</span>
    </div>
    <div class="stat">
      <span class="stat-val">CLI</span>
      <span class="stat-label">Works offline</span>
    </div>
    <div class="stat">
      <span class="stat-val">Free</span>
      <span class="stat-label">For public repos</span>
    </div>
  </div>
</div>

<div class="divider"></div>

<!-- PROBLEM -->
<section>
  <div class="section-inner">
    <div class="problem-grid">
      <div class="problem-text reveal">
        <div class="label">The problem</div>
        <h2>Hidden in the merge,<br>invisible in the PR</h2>
        <p>When both parent branches contain <strong>identical files</strong>, Git&rsquo;s three-way merge algorithm outputs them unchanged. The only way to get a different result is to manually edit files during the merge.</p>
        <p>GitHub&rsquo;s PR diff doesn&rsquo;t show merge commits. <code>git log</code> doesn&rsquo;t surface the change. SAST tools scan files, not merge history. The injection is invisible.</p>
        <p>This is how malicious code ran undetected in a production repository for <strong>several months</strong> &mdash; on every developer machine and every CI build.</p>
        <p>It&rsquo;s a <strong>supply chain attack</strong> via code injection &mdash; and it bypasses every standard git security tool.</p>
        <p>GitHub has confirmed this is an <strong>intentional design decision</strong> and does not plan to fix it. The responsibility falls entirely on your team to detect it.</p>
      </div>

      <div class="diff-card reveal">
        <div class="diff-header">
          <span class="diff-filename">vite.config.js &mdash; merge commit ab90bd7</span>
          <span class="diff-badge">Evil Merge</span>
        </div>
        <div class="diff-body">
          <div class="diff-line">
            <span class="diff-label">Parent 1</span>
            <span class="diff-value clean">aa82acb0c335&hellip; &larr; clean</span>
          </div>
          <div class="diff-line">
            <span class="diff-label">Parent 2</span>
            <span class="diff-value clean">aa82acb0c335&hellip; &larr; clean</span>
          </div>
          <div class="diff-line">
            <span class="diff-label">Merge</span>
            <span class="diff-value bad">2a54754defae&hellip; &larr; DIFFERENT</span>
          </div>
          <div class="diff-sep"></div>
          <div class="diff-note">
            When both parents are identical, Git cannot produce a different output on its own.
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<div class="divider"></div>

<!-- HOW IT WORKS -->
<section id="how-it-works">
  <div class="section-inner">
    <div class="label">How it works</div>
    <h2>Simple detection,<br>no false positives</h2>
    <p class="section-sub">For each merge commit, we reconstruct what Git should have produced and compare it to what the commit actually contains.</p>

    <div class="steps">
      <div class="step reveal">
        <div class="step-num">01</div>
        <h3>Find the merge base</h3>
        <p>Identify the common ancestor of the two parent commits &mdash; the starting point for the three-way merge algorithm.</p>
      </div>
      <div class="step reveal">
        <div class="step-num">02</div>
        <h3>Reconstruct expected tree</h3>
        <p>Run a clean three-way merge of the parent trees. This is what Git would produce with no manual intervention.</p>
      </div>
      <div class="step reveal">
        <div class="step-num">03</div>
        <h3>Compare against reality</h3>
        <p>Diff the expected tree against the actual merge commit. Any difference is a file manually edited during the merge.</p>
      </div>
    </div>
  </div>
</section>

<div class="divider"></div>

<!-- INTEGRATIONS -->
<section>
  <div class="section-inner">
    <div class="label">Integrations</div>
    <h2>Works where you<br>already work</h2>
    <p class="section-sub">Multiple ways to add evil merge detection &mdash; pick what fits your workflow.</p>

    <div class="integrations">
      <div class="integration reveal">
        <div class="int-badge cli">CLI</div>
        <h3>Command Line</h3>
        <p>Scan any repository from the terminal. Supports JSON and SARIF output for GitHub Code Scanning.</p>
        <pre>brew install fimskiy/tap/evilmerge
evilmerge scan .</pre>
      </div>
      <div class="integration reveal">
        <div class="int-badge action">Action</div>
        <h3>GitHub Action</h3>
        <p>Add to your workflow and get annotations directly on pull requests. Zero configuration.</p>
        <pre>- uses: evilmerge-dev/Evil-merge-detector@v1
  with:
    fail-on: warning</pre>
      </div>
      <div class="integration reveal">
        <div class="int-badge app">App</div>
        <h3>GitHub App</h3>
        <p>Install once, get automatic checks on every PR. No workflow changes needed.</p>
        <pre>Install from GitHub Marketplace
&#8594; automatic on every pull request
&#8594; results in GitHub Checks</pre>
      </div>
    </div>
  </div>
</section>

<div class="divider"></div>

<!-- PRICING -->
<section id="pricing">
  <div class="section-inner">
    <div class="label">Pricing</div>
    <h2>Simple, per-organization<br>pricing</h2>
    <p class="section-sub">The CLI and GitHub Action are always free and open source.</p>

    <div class="pricing-toggle">
      <span class="toggle-label active" id="lbl-monthly">Monthly</span>
      <button class="toggle-switch" id="billing-toggle" aria-label="Switch billing period" aria-pressed="false">
        <span class="toggle-knob"></span>
      </button>
      <span class="toggle-label" id="lbl-annual">Annual <span class="save-badge">Save 20%</span></span>
    </div>

    <div class="pricing-wrap">
      <div class="plan reveal">
        <div class="plan-tier">Free</div>
        <div class="plan-price">
          <span class="price-amount">$0</span>
        </div>
        <div class="plan-desc">For open source and personal projects</div>
        <ul>
          <li><span class="check">&#10003;</span> Public repositories</li>
          <li><span class="check">&#10003;</span> 50 PR scans / month</li>
          <li><span class="check">&#10003;</span> GitHub Checks integration</li>
          <li class="no"><span class="check">&ndash;</span> Private repositories</li>
          <li><span class="check">&#10003;</span> Scan history dashboard</li>
          <li class="no"><span class="check">&ndash;</span> Unlimited scans</li>
        </ul>
        <a class="plan-btn" id="btn-plan-free" href="https://github.com/apps/evil-merge-detector">Install for free</a>
        <p class="plan-note">No credit card required</p>
      </div>

      <div class="plan featured reveal">
        <div class="plan-tier">Pro</div>
        <div class="plan-price">
          <span class="price-currency">$</span>
          <span class="price-amount" id="pro-price">7</span>
          <span class="price-period" id="pro-period">/month</span>
        </div>
        <div class="plan-desc">For teams and private repositories</div>
        <ul>
          <li><span class="check">&#10003;</span> Public &amp; private repositories</li>
          <li><span class="check">&#10003;</span> Unlimited PR scans</li>
          <li><span class="check">&#10003;</span> GitHub Checks integration</li>
          <li><span class="check">&#10003;</span> Scan history dashboard</li>
          <li><span class="check">&#10003;</span> Priority support</li>
        </ul>
        <a class="plan-btn" id="btn-plan-pro" href="https://github.com/marketplace/evil-merge-detector">Upgrade to Pro</a>
        <p class="plan-note">Cancel anytime · No lock-in</p>
      </div>
    </div>
  </div>
</section>

<div class="divider"></div>

<!-- FAQ -->
<section id="faq">
  <div class="section-inner">
    <div class="label">FAQ</div>
    <h2>Common questions</h2>

    <div class="faq-list">
      <details class="faq-item reveal">
        <summary class="faq-q">Does it need access to my source code?</summary>
        <p class="faq-a">The GitHub App requests read-only access to repository contents and checks — the minimum required to scan merge commits. The CLI runs entirely locally; nothing leaves your machine.</p>
      </details>
      <details class="faq-item reveal">
        <summary class="faq-q">What counts as a PR scan?</summary>
        <p class="faq-a">One scan = one pull request event (opened or synchronized). Scans that find no merge commits in the PR are not counted against your limit.</p>
      </details>
      <details class="faq-item reveal">
        <summary class="faq-q">Can it produce false positives?</summary>
        <p class="faq-a">No. The detection is deterministic: if both parent branches have identical file content, Git's algorithm cannot produce a different output. Any difference is a manual edit — there is no ambiguity.</p>
      </details>
      <details class="faq-item reveal">
        <summary class="faq-q">Does it work with GitLab or Bitbucket?</summary>
        <p class="faq-a">The CLI works with any Git repository regardless of host. Ready-to-use CI templates for GitLab CI and Bitbucket Pipelines are available in the repository. The GitHub App and GitHub Checks integration are GitHub-only.</p>
      </details>
      <details class="faq-item reveal">
        <summary class="faq-q">Doesn&rsquo;t enabling &ldquo;Dismiss stale reviews&rdquo; prevent this?</summary>
        <p class="faq-a">Partially &mdash; and only going forward. &ldquo;Dismiss stale reviews&rdquo; forces re-review when new commits are pushed, which makes the attack harder to execute. But it doesn&rsquo;t scan your existing history for past injections, requires manual setup on every repository, and can be changed by any admin at any time. GitHub themselves confirmed this attack vector is <strong>working as designed</strong> and has no plans to address it &mdash; making detection, not just prevention, essential.</p>
      </details>
      <details class="faq-item reveal">
        <summary class="faq-q">What happens when I exceed 50 scans on the Free plan?</summary>
        <p class="faq-a">Additional PRs will not be scanned and the check run will be skipped with a note explaining the limit. No errors, no failed checks — just a nudge to upgrade.</p>
      </details>
    </div>
  </div>
</section>

<!-- BOTTOM CTA -->
<section class="cta-section">
  <div class="section-inner">
    <div class="label">Protect your codebase</div>
    <h2>Your next merge could be<br>hiding something.</h2>
    <p class="section-sub">Install the GitHub App and start scanning automatically &mdash; no workflow changes needed.</p>
    <div class="cta-actions">
      <a class="btn-primary" id="btn-cta-install" href="https://github.com/apps/evil-merge-detector">Install GitHub App</a>
      <a class="btn-secondary link-arrow" id="btn-cta-github" href="https://github.com/evilmerge-dev/Evil-merge-detector">View on GitHub</a>
    </div>
  </div>
</section>

</main>

<!-- FOOTER -->
<footer>
  <div class="footer-logo">Evil Merge <span class="accent">Detector</span> &mdash; open source on <a href="https://github.com/evilmerge-dev/Evil-merge-detector" class="footer-gh-link link-arrow">GitHub</a></div>
  <div class="footer-links">
    <a href="https://github.com/evilmerge-dev/Evil-merge-detector">Docs</a>
    <a href="https://github.com/apps/evil-merge-detector">Install</a>
    <a href="/privacy">Privacy Policy</a>
  </div>
</footer>

<!-- COOKIE BANNER -->
<div id="cookie-banner" style="display:none" role="dialog" aria-modal="true" aria-label="Cookie consent">
  <p>We use Google Analytics to understand how visitors use this site. No personally identifiable information is collected. <a href="/privacy">Privacy Policy</a></p>
  <div class="cookie-actions">
    <button class="cookie-btn cookie-btn-decline" id="cookie-decline">Decline</button>
    <button class="cookie-btn cookie-btn-accept" id="cookie-accept">Accept</button>
  </div>
</div>

<script>
(function() {
  // Scroll reveal
  var observer = new IntersectionObserver(function(entries) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1 });

  document.querySelectorAll('.reveal').forEach(function(el) {
    observer.observe(el);
  });

  // Animate counters
  function animateCounters() {
    var bar = document.querySelector('.stats-bar');
    if (!bar || bar.dataset.counted) return;
    var rect = bar.getBoundingClientRect();
    if (rect.top >= window.innerHeight) return;
    bar.dataset.counted = '1';
    bar.querySelectorAll('[data-target]').forEach(function(el) {
      var target = parseInt(el.dataset.target, 10);
      if (target === 0) return;
      var duration = 1400;
      var startTime = null;
      function step(ts) {
        if (!startTime) startTime = ts;
        var progress = Math.min((ts - startTime) / duration, 1);
        el.textContent = Math.round((1 - Math.pow(1 - progress, 3)) * target);
        if (progress < 1) requestAnimationFrame(step);
      }
      requestAnimationFrame(step);
    });
  }
  animateCounters();
  window.addEventListener('scroll', animateCounters, { passive: true });

  // Terminal typing animation
  var cmd = 'evilmerge scan .';
  var el = document.getElementById('typed-cmd');
  var cursor = el ? el.nextElementSibling : null;
  var output = document.getElementById('term-output');
  if (el) {
    var i = 0;
    function type() {
      if (i <= cmd.length) {
        el.textContent = cmd.slice(0, i);
        i++;
        setTimeout(type, 60 + Math.random() * 40);
      } else {
        if (cursor) cursor.style.display = 'none';
        setTimeout(function() {
          if (output) output.style.display = '';
        }, 300);
      }
    }
    setTimeout(type, 900);
  }

  // Mobile nav
  var menuBtn = document.getElementById('nav-menu-btn');
  var mobileNav = document.getElementById('mobile-nav');
  if (menuBtn && mobileNav) {
    menuBtn.addEventListener('click', function() {
      var open = mobileNav.classList.toggle('open');
      menuBtn.setAttribute('aria-expanded', open ? 'true' : 'false');
      menuBtn.innerHTML = open ? '&#10005;' : '&#9776;';
    });
    mobileNav.querySelectorAll('a').forEach(function(a) {
      a.addEventListener('click', function() {
        mobileNav.classList.remove('open');
        menuBtn.setAttribute('aria-expanded', 'false');
        menuBtn.innerHTML = '&#9776;';
      });
    });
  }

  // Pricing toggle
  var toggle = document.getElementById('billing-toggle');
  var proPrice = document.getElementById('pro-price');
  var proPeriod = document.getElementById('pro-period');
  var lblMonthly = document.getElementById('lbl-monthly');
  var lblAnnual = document.getElementById('lbl-annual');
  var isAnnual = false;
  if (toggle) {
    toggle.addEventListener('click', function() {
      isAnnual = !isAnnual;
      toggle.classList.toggle('annual', isAnnual);
      toggle.setAttribute('aria-pressed', isAnnual ? 'true' : 'false');
      if (proPrice) proPrice.textContent = isAnnual ? '67' : '7';
      if (proPeriod) proPeriod.textContent = isAnnual ? '/year' : '/month';
      if (lblMonthly) lblMonthly.classList.toggle('active', !isAnnual);
      if (lblAnnual) lblAnnual.classList.toggle('active', isAnnual);
    });
  }
})();

// Cookie consent + GTM (EU only, detected server-side via CF-IPCountry)
(function() {
  if (!document.body.dataset.eu) { loadGTM(); return; }

  var CONSENT_KEY = 'emd_cookie_consent';

  function loadGTM() {
    window.dataLayer = window.dataLayer || [];
    (function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
    new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
    j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
    'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
    })(window,document,'script','dataLayer','GTM-K6K4PH7S');
  }

  var consent = localStorage.getItem(CONSENT_KEY);
  if (consent === 'accepted') { loadGTM(); return; }

  var banner = document.getElementById('cookie-banner');
  if (!banner) return;

  var acceptBtn = document.getElementById('cookie-accept');
  var declineBtn = document.getElementById('cookie-decline');
  if (!acceptBtn || !declineBtn) return;

  if (consent !== 'declined') {
    banner.style.display = '';
    // Read offsetHeight after display change forces reflow
    requestAnimationFrame(function() {
      document.body.style.paddingBottom = banner.offsetHeight + 'px';
      acceptBtn.focus();
    });
  }

  acceptBtn.addEventListener('click', function() {
    localStorage.setItem(CONSENT_KEY, 'accepted');
    banner.style.display = 'none';
    document.body.style.paddingBottom = '';
    loadGTM();
  });

  declineBtn.addEventListener('click', function() {
    localStorage.setItem(CONSENT_KEY, 'declined');
    banner.style.display = 'none';
    document.body.style.paddingBottom = '';
  });
})();
</script>

</body>
</html>
`

var privacyPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Privacy Policy — Evil Merge Detector</title>
<meta name="robots" content="index, follow">
<link rel="canonical" href="https://evilmerge.dev/privacy">
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg viewBox='0 0 32 32' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M16 2L28 6L28 18C28 24 22 29 16 31C10 29 4 24 4 18L4 6Z' fill='%23dc2626'/%3E%3Cpath d='M16 5L26 8.5L26 18C26 23 21 27 16 29C11 27 6 23 6 18L6 8.5Z' fill='%23b91c1c'/%3E%3Ccircle cx='16' cy='12' r='2.5' fill='white'/%3E%3Ccircle cx='11' cy='17' r='2' fill='white'/%3E%3Ccircle cx='16' cy='22' r='2.2' fill='white'/%3E%3Ccircle cx='21' cy='17' r='2.5' fill='%237f1d1d'/%3E%3Cline x1='20' y1='16' x2='22' y2='18' stroke='white' stroke-width='1.2' stroke-linecap='round'/%3E%3Cline x1='22' y1='16' x2='20' y2='18' stroke='white' stroke-width='1.2' stroke-linecap='round'/%3E%3C/svg%3E">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
:root {
  --bg: #ffffff;
  --border: #e2e8f0;
  --red: #dc2626;
  --text: #0f172a;
  --text-mid: #334155;
  --text-muted: #64748b;
  --sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: var(--sans); background: var(--bg); color: var(--text); line-height: 1.7; -webkit-font-smoothing: antialiased; }
a { color: var(--red); text-decoration: none; }
a:hover { text-decoration: underline; }
nav {
  position: sticky; top: 0; z-index: 100;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 48px; height: 60px;
  background: rgba(255,255,255,0.9);
  backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}
.nav-logo { font-size: 18px; font-weight: 700; letter-spacing: -0.01em; color: var(--text); display: flex; align-items: center; gap: 10px; }
.nav-logo svg { width: 32px; height: 32px; flex-shrink: 0; }
.accent { color: var(--red); }
main { max-width: 720px; margin: 0 auto; padding: 64px 24px 96px; }
h1 { font-size: 2rem; font-weight: 700; letter-spacing: -0.02em; margin-bottom: 8px; }
.updated { color: var(--text-muted); font-size: 14px; margin-bottom: 48px; }
h2 { font-size: 1.1rem; font-weight: 600; margin-top: 40px; margin-bottom: 12px; }
p { color: var(--text-mid); margin-bottom: 16px; }
ul { color: var(--text-mid); padding-left: 20px; margin-bottom: 16px; }
li { margin-bottom: 6px; }
footer { text-align: center; padding: 32px 24px; border-top: 1px solid var(--border); color: var(--text-muted); font-size: 13px; }
@media (max-width: 640px) { nav { padding: 0 20px; } }
</style>
</head>
<body>
<nav>
  <a class="nav-logo" href="/">
    <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M16 2L28 6L28 18C28 24 22 29 16 31C10 29 4 24 4 18L4 6Z" fill="#dc2626"/>
      <path d="M16 5L26 8.5L26 18C26 23 21 27 16 29C11 27 6 23 6 18L6 8.5Z" fill="#b91c1c"/>
      <circle cx="16" cy="12" r="2.5" fill="white"/>
      <circle cx="11" cy="17" r="2" fill="white"/>
      <circle cx="16" cy="22" r="2.2" fill="white"/>
      <circle cx="21" cy="17" r="2.5" fill="#7f1d1d"/>
      <line x1="20" y1="16" x2="22" y2="18" stroke="white" stroke-width="1.2" stroke-linecap="round"/>
      <line x1="22" y1="16" x2="20" y2="18" stroke="white" stroke-width="1.2" stroke-linecap="round"/>
    </svg>
    <span>Evil Merge <span class="accent">Detector</span></span>
  </a>
</nav>

<main>
  <h1>Privacy Policy</h1>
  <p class="updated">Last updated: March 28, 2026</p>

  <h2>1. Data Controller</h2>
  <p>Evil Merge Detector (evilmerge.dev), operated by Dmitrii Fimskiy. For any privacy-related requests, contact: <a href="mailto:legal@evilmerge.dev">legal@evilmerge.dev</a></p>

  <h2>2. What Data We Collect</h2>
  <p>We use Google Analytics (via Google Tag Manager) to understand how visitors use this site. Google Analytics collects:</p>
  <ul>
    <li>Pages visited and time spent on each page</li>
    <li>Browser type, operating system, and device type</li>
    <li>Approximate geographic location (country and city, derived from IP address)</li>
    <li>Referral source (how you found the site)</li>
    <li>Session duration and navigation path</li>
  </ul>
  <p>IP addresses are anonymized before storage. We do not collect your name, email address, or any other personally identifiable information through Google Analytics.</p>
  <p>We also store a single item in your browser's <code>localStorage</code> (<code>emd_cookie_consent</code>) to remember your cookie consent decision. This item never leaves your browser.</p>

  <h2>3. Purpose and Legal Basis</h2>
  <p>We use analytics data to understand which features are most useful, identify usability issues, and improve the product. The legal basis for this processing is your consent (Article 6(1)(a) GDPR), given via the cookie banner on our site.</p>

  <h2>4. Data Retention</h2>
  <p>Google Analytics retains aggregated usage data for 14 months. The <code>localStorage</code> consent flag is retained in your browser until you clear your browser storage.</p>

  <h2>5. Third Parties</h2>
  <p>We share analytics data only with <strong>Google LLC</strong> (Google Analytics / Google Tag Manager). Google processes this data under its own privacy policy: <a href="https://policies.google.com/privacy" target="_blank" rel="noopener">policies.google.com/privacy</a>. Google LLC participates in the EU–US Data Privacy Framework.</p>
  <p>No other third parties receive your data from this website.</p>

  <h2>6. Your Rights (GDPR)</h2>
  <p>Under the GDPR you have the right to:</p>
  <ul>
    <li><strong>Access</strong> — request a copy of data we hold about you</li>
    <li><strong>Rectification</strong> — request correction of inaccurate data</li>
    <li><strong>Erasure</strong> — request deletion of your data</li>
    <li><strong>Restriction</strong> — request that we limit processing of your data</li>
    <li><strong>Objection</strong> — object to processing based on legitimate interest</li>
    <li><strong>Portability</strong> — receive your data in a structured, machine-readable format</li>
    <li><strong>Withdraw consent</strong> — withdraw your cookie consent at any time (see below)</li>
  </ul>
  <p>To exercise any of these rights, contact us at <a href="mailto:legal@evilmerge.dev">legal@evilmerge.dev</a>. We will respond within 30 days.</p>
  <p>You also have the right to lodge a complaint with the Polish data protection authority (UODO): <a href="https://uodo.gov.pl" target="_blank" rel="noopener">uodo.gov.pl</a></p>

  <h2>7. Withdrawing Consent</h2>
  <p>You can withdraw your cookie consent at any time by clicking the button below. Google Analytics will stop loading on future visits.</p>
  <p style="margin-top:16px">
    <button onclick="localStorage.removeItem('emd_cookie_consent');this.textContent='Done — reload the page to see the banner again';this.disabled=true;" style="background:#dc2626;color:#fff;border:none;border-radius:6px;padding:9px 18px;font-size:14px;font-weight:600;cursor:pointer;">Revoke cookie consent</button>
  </p>
  <p style="margin-top:12px">You can also opt out of Google Analytics tracking globally using the <a href="https://tools.google.com/dlpage/gaoptout" target="_blank" rel="noopener">Google Analytics Opt-out Browser Add-on</a>.</p>

  <h2>8. Cookies</h2>
  <p>We do not set any cookies directly. Google Analytics sets cookies on our behalf (e.g. <code>_ga</code>, <code>_ga_*</code>) only after you accept via the cookie banner. These cookies identify unique sessions for analytics purposes and expire after 13–24 months.</p>

  <h2>9. Changes to This Policy</h2>
  <p>We may update this policy. The date at the top of this page reflects the latest revision. Continued use of the site after changes constitutes acceptance of the updated policy.</p>
</main>

<footer>
  &copy; 2026 Evil Merge Detector &mdash; <a href="/">Home</a>
</footer>
</body>
</html>
`
