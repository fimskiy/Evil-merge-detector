package landing

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl = template.Must(template.New("landing").Parse(page))

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("landing template: %v", err)
	}
}

var page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Evil Merge Detector — Find hidden code in Git merge commits</title>
<meta name="description" content="Automatically detect evil merges — merge commits that introduce changes not present in either parent branch. CLI, GitHub Action, and GitHub App.">
<style>
:root {
  --bg: #07090d;
  --bg2: #0d1117;
  --bg3: #111820;
  --border: rgba(255,255,255,0.07);
  --border-bright: rgba(255,255,255,0.12);
  --red: #e63946;
  --red-dim: rgba(230,57,70,0.12);
  --red-glow: rgba(230,57,70,0.25);
  --green: #2dce89;
  --amber: #f4a261;
  --text: #c9d1d9;
  --text-dim: #586069;
  --text-mid: #8b949e;
  --mono: 'Courier New', Courier, monospace;
  --sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html { scroll-behavior: smooth; }

body {
  font-family: var(--sans);
  background: var(--bg);
  color: var(--text);
  line-height: 1.6;
  overflow-x: hidden;
}

a { color: inherit; text-decoration: none; }
code { font-family: var(--mono); }

/* GRID BACKGROUND */
.grid-bg {
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px);
  background-size: 48px 48px;
  pointer-events: none;
  z-index: 0;
}
.grid-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse 80% 60% at 50% 0%, rgba(230,57,70,0.06) 0%, transparent 70%);
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
  background: rgba(7,9,13,0.85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}

.nav-logo {
  font-family: var(--mono);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.03em;
  color: var(--text);
}
.nav-logo .accent { color: var(--red); }

.nav-links {
  display: flex;
  gap: 32px;
  font-size: 13px;
  color: var(--text-dim);
}
.nav-links a { transition: color 0.15s; }
.nav-links a:hover { color: var(--text); }

.nav-cta {
  font-size: 13px;
  font-weight: 600;
  font-family: var(--mono);
  letter-spacing: 0.04em;
  padding: 8px 18px;
  background: var(--red);
  color: #fff;
  border-radius: 4px;
  transition: opacity 0.15s, box-shadow 0.15s;
}
.nav-cta:hover {
  opacity: 0.88;
  box-shadow: 0 0 20px var(--red-glow);
}

/* HERO */
.hero {
  position: relative;
  z-index: 1;
  max-width: 860px;
  margin: 0 auto;
  padding: 110px 24px 90px;
  text-align: center;
}

.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-dim);
  border: 1px solid var(--border);
  padding: 5px 14px;
  border-radius: 2px;
  margin-bottom: 36px;
}
.hero-eyebrow .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--red);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 0 0 var(--red-glow); }
  50% { opacity: 0.6; box-shadow: 0 0 0 6px transparent; }
}

h1 {
  font-family: var(--mono);
  font-size: 54px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.02em;
  text-transform: uppercase;
  margin-bottom: 24px;
  color: #e6edf3;
}
h1 .threat { color: var(--red); }
h1 .cursor {
  display: inline-block;
  width: 3px;
  height: 0.85em;
  background: var(--red);
  vertical-align: middle;
  margin-left: 4px;
  animation: blink 1.1s step-end infinite;
}
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }

.hero-sub {
  font-size: 18px;
  color: var(--text-mid);
  max-width: 520px;
  margin: 0 auto 40px;
  line-height: 1.7;
}

.hero-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 64px;
}

.btn-primary {
  font-family: var(--mono);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 12px 28px;
  background: var(--red);
  color: #fff;
  border-radius: 4px;
  transition: opacity 0.15s, box-shadow 0.2s;
}
.btn-primary:hover {
  opacity: 0.9;
  box-shadow: 0 0 30px var(--red-glow);
}

.btn-secondary {
  font-family: var(--mono);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 12px 28px;
  background: transparent;
  color: var(--text-mid);
  border: 1px solid var(--border-bright);
  border-radius: 4px;
  transition: color 0.15s, border-color 0.15s;
}
.btn-secondary:hover {
  color: var(--text);
  border-color: rgba(255,255,255,0.25);
}

/* TERMINAL */
.terminal {
  background: var(--bg2);
  border: 1px solid var(--border-bright);
  border-radius: 6px;
  overflow: hidden;
  max-width: 620px;
  margin: 0 auto;
  text-align: left;
  box-shadow: 0 24px 80px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.04);
}
.terminal-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: rgba(255,255,255,0.03);
  border-bottom: 1px solid var(--border);
}
.dot-r { width:10px;height:10px;border-radius:50%;background:#ff5f57; }
.dot-y { width:10px;height:10px;border-radius:50%;background:#ffbd2e; }
.dot-g { width:10px;height:10px;border-radius:50%;background:#28c840; }
.terminal-title {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text-dim);
  margin-left: auto;
  margin-right: auto;
}
.terminal-body {
  padding: 20px 24px;
  font-family: var(--mono);
  font-size: 13px;
  line-height: 1.9;
}
.t-dim { color: var(--text-dim); }
.t-prompt { color: var(--text-mid); }
.t-cmd { color: #79c0ff; }
.t-bad { color: var(--red); font-weight: 700; }
.t-warn { color: var(--amber); }
.t-ok { color: var(--green); }

/* SECTIONS */
section {
  position: relative;
  z-index: 1;
  padding: 96px 24px;
}

.section-inner { max-width: 1000px; margin: 0 auto; }

.label {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--red);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.label::before {
  content: '';
  display: block;
  width: 20px;
  height: 1px;
  background: var(--red);
}

h2 {
  font-family: var(--mono);
  font-size: 30px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: -0.01em;
  color: #e6edf3;
  margin-bottom: 12px;
  line-height: 1.2;
}

.section-sub {
  font-size: 17px;
  color: var(--text-mid);
  max-width: 500px;
  margin-bottom: 52px;
  line-height: 1.7;
}

/* DIVIDER */
.divider {
  position: relative;
  z-index: 1;
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
  font-size: 16px;
  color: var(--text-mid);
  margin-bottom: 18px;
  line-height: 1.75;
}
.problem-text p strong { color: var(--text); }
.problem-text code {
  background: rgba(255,255,255,0.07);
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 12px;
}

.diff-card {
  background: var(--bg2);
  border: 1px solid var(--border-bright);
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 16px 48px rgba(0,0,0,0.4);
}
.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: rgba(255,255,255,0.03);
  border-bottom: 1px solid var(--border);
}
.diff-filename {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text-dim);
}
.diff-badge {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 2px 8px;
  background: var(--red-dim);
  color: var(--red);
  border: 1px solid rgba(230,57,70,0.3);
  border-radius: 2px;
}
.diff-body {
  padding: 20px;
  font-family: var(--mono);
  font-size: 12.5px;
  line-height: 2;
}
.diff-line { display: flex; gap: 12px; }
.diff-label { color: var(--text-dim); min-width: 70px; flex-shrink: 0; }
.diff-value { color: var(--text-mid); }
.diff-value.clean { color: var(--green); }
.diff-value.bad { color: var(--red); font-weight: 700; }
.diff-sep { height: 1px; background: var(--border); margin: 14px 0; }
.diff-note {
  font-size: 12px;
  color: var(--text-dim);
  line-height: 1.7;
  font-style: italic;
}

/* HOW IT WORKS */
.steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2px;
  background: var(--border);
  border: 1px solid var(--border);
  border-radius: 6px;
  overflow: hidden;
}
.step {
  background: var(--bg2);
  padding: 36px 32px;
  position: relative;
  overflow: hidden;
  transition: background 0.2s;
}
.step:hover { background: var(--bg3); }
.step-num {
  font-family: var(--mono);
  font-size: 72px;
  font-weight: 700;
  color: rgba(255,255,255,0.03);
  position: absolute;
  top: -8px;
  right: 16px;
  line-height: 1;
  user-select: none;
}
.step-label {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--red);
  margin-bottom: 14px;
}
.step h3 {
  font-family: var(--mono);
  font-size: 14px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  color: #e6edf3;
  margin-bottom: 10px;
}
.step p {
  font-size: 15px;
  color: var(--text-mid);
  line-height: 1.7;
}

/* INTEGRATIONS */
.integrations {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.integration {
  background: var(--bg2);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 28px;
  transition: border-color 0.2s;
}
.integration:hover { border-color: var(--border-bright); }
.int-icon {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-dim);
  background: rgba(255,255,255,0.05);
  display: inline-block;
  padding: 3px 8px;
  border-radius: 2px;
  margin-bottom: 16px;
}
.integration h3 {
  font-family: var(--mono);
  font-size: 15px;
  font-weight: 700;
  color: #e6edf3;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}
.integration p {
  font-size: 15px;
  color: var(--text-mid);
  margin-bottom: 20px;
  line-height: 1.65;
}
.integration pre {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 14px 16px;
  font-family: var(--mono);
  font-size: 12px;
  color: #79c0ff;
  overflow-x: auto;
  line-height: 1.8;
}

/* PRICING */
.pricing-wrap {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  max-width: 700px;
}

.plan {
  background: var(--bg2);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 36px;
  position: relative;
}

.plan.featured {
  border-color: rgba(230,57,70,0.4);
  background: linear-gradient(135deg, rgba(230,57,70,0.05) 0%, var(--bg2) 60%);
}
.plan.featured::before {
  content: 'RECOMMENDED';
  position: absolute;
  top: -1px;
  right: 24px;
  font-family: var(--mono);
  font-size: 9px;
  letter-spacing: 0.12em;
  padding: 4px 10px;
  background: var(--red);
  color: #fff;
  border-radius: 0 0 4px 4px;
}

.plan-tier {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-dim);
  margin-bottom: 16px;
}

.plan-price {
  margin-bottom: 6px;
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.price-currency {
  font-family: var(--mono);
  font-size: 20px;
  color: var(--text-mid);
  font-weight: 700;
}
.price-amount {
  font-family: var(--mono);
  font-size: 48px;
  font-weight: 700;
  color: #e6edf3;
  line-height: 1;
}
.price-period {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text-dim);
}

.plan-desc {
  font-size: 15px;
  color: var(--text-dim);
  margin-bottom: 28px;
  padding-bottom: 28px;
  border-bottom: 1px solid var(--border);
}

.plan ul {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 32px;
}
.plan li {
  font-size: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-mid);
}
.plan li .check {
  width: 16px;
  height: 16px;
  border-radius: 2px;
  background: rgba(45,206,137,0.12);
  border: 1px solid rgba(45,206,137,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 10px;
  color: var(--green);
}
.plan li.no .check {
  background: rgba(255,255,255,0.04);
  border-color: var(--border);
  color: var(--text-dim);
}
.plan li.no { color: var(--text-dim); }

.plan-btn {
  display: block;
  text-align: center;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 12px;
  border-radius: 4px;
  border: 1px solid var(--border-bright);
  color: var(--text-mid);
  transition: all 0.15s;
}
.plan-btn:hover {
  color: var(--text);
  border-color: rgba(255,255,255,0.25);
}
.plan.featured .plan-btn {
  background: var(--red);
  border-color: var(--red);
  color: #fff;
}
.plan.featured .plan-btn:hover {
  opacity: 0.88;
  box-shadow: 0 0 24px var(--red-glow);
}

/* FOOTER */
footer {
  position: relative;
  z-index: 1;
  padding: 36px 48px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.footer-logo {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text-dim);
}
.footer-logo .accent { color: var(--red); }
.footer-links {
  display: flex;
  gap: 28px;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-dim);
}
.footer-links a { transition: color 0.15s; }
.footer-links a:hover { color: var(--text); }

/* RESPONSIVE */
@media (max-width: 768px) {
  nav { padding: 0 20px; }
  .nav-links { display: none; }
  h1 { font-size: 32px; }
  h2 { font-size: 22px; }
  .hero { padding: 72px 20px 60px; }
  .problem-grid,
  .steps,
  .integrations,
  .pricing-wrap { grid-template-columns: 1fr; }
  .steps { gap: 1px; }
  section { padding: 64px 20px; }
  .divider { margin: 0 20px; }
  footer { flex-direction: column; gap: 20px; text-align: center; padding: 32px 20px; }
  .footer-links { justify-content: center; }
}
</style>
</head>
<body>

<div class="grid-bg"></div>

<!-- NAV -->
<nav>
  <div class="nav-logo">EVIL<span class="accent">_</span>MERGE<span class="accent">.</span>DETECT</div>
  <div class="nav-links">
    <a href="#how-it-works">How it works</a>
    <a href="#pricing">Pricing</a>
    <a href="https://github.com/fimskiy/Evil-merge-detector">GitHub</a>
    <a href="/dashboard">Dashboard</a>
  </div>
  <a class="nav-cta" href="https://github.com/apps/evil-merge-detector">Install App</a>
</nav>

<!-- HERO -->
<div class="hero">
  <div class="hero-eyebrow">
    <span class="dot"></span>
    Now available on GitHub Marketplace
  </div>

  <h1>
    The merge commit<br>
    that <span class="threat">wasn't</span> reviewed<span class="cursor"></span>
  </h1>

  <p class="hero-sub">
    Evil Merge Detector finds merge commits that introduce changes not present in either parent branch —
    the attack vector your code review misses.
  </p>

  <div class="hero-actions">
    <a class="btn-primary" href="https://github.com/apps/evil-merge-detector">Install GitHub App</a>
    <a class="btn-secondary" href="https://github.com/fimskiy/Evil-merge-detector">View on GitHub</a>
  </div>

  <div class="terminal">
    <div class="terminal-bar">
      <span class="dot-r"></span>
      <span class="dot-y"></span>
      <span class="dot-g"></span>
      <span class="terminal-title">evilmerge — scan</span>
    </div>
    <div class="terminal-body">
      <div class="t-dim"># Scan your repository</div>
      <div><span class="t-prompt">$ </span><span class="t-cmd">evilmerge scan .</span></div>
      <div>&nbsp;</div>
      <div class="t-bad">CRITICAL  ab90bd7  vite.config.js</div>
      <div class="t-dim">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Both parents had identical content.</div>
      <div class="t-dim">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Merge result differs — manual edit detected.</div>
      <div>&nbsp;</div>
      <div class="t-ok">&#10003;  23 merge commits checked &nbsp;|&nbsp; 1 critical issue found</div>
    </div>
  </div>
</div>

<div class="divider"></div>

<!-- PROBLEM -->
<section>
  <div class="section-inner">
    <div class="problem-grid">
      <div class="problem-text">
        <div class="label">The problem</div>
        <h2>Hidden in the merge,<br>invisible in the PR</h2>
        <p>When both parent branches contain <strong>identical files</strong>, Git's three-way merge algorithm outputs them unchanged. The only way to get a different result is to manually edit files during the merge.</p>
        <p>GitHub's PR diff doesn't show merge commits. <code>git log</code> doesn't surface the change. SAST tools scan files, not merge history. The injection is invisible.</p>
        <p>This is how malicious code ran undetected in a production repository for <strong>3.5 months</strong> — on every developer machine and every CI build.</p>
      </div>

      <div class="diff-card">
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
            When both parents are identical, Git cannot<br>
            produce a different output on its own.
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
      <div class="step">
        <div class="step-num">01</div>
        <div class="step-label">Step 01</div>
        <h3>Find the merge base</h3>
        <p>Identify the common ancestor of the two parent commits — the starting point for the three-way merge algorithm.</p>
      </div>
      <div class="step">
        <div class="step-num">02</div>
        <div class="step-label">Step 02</div>
        <h3>Reconstruct expected tree</h3>
        <p>Run a clean three-way merge of the parent trees. This is what Git would produce with no manual intervention.</p>
      </div>
      <div class="step">
        <div class="step-num">03</div>
        <div class="step-label">Step 03</div>
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
    <p class="section-sub">Three ways to add evil merge detection — pick what fits your workflow.</p>

    <div class="integrations">
      <div class="integration">
        <div class="int-icon">CLI</div>
        <h3>Command Line</h3>
        <p>Scan any repository from the terminal. Supports JSON and SARIF output for GitHub Code Scanning.</p>
        <pre>brew install fimskiy/tap/evilmerge
evilmerge scan .</pre>
      </div>
      <div class="integration">
        <div class="int-icon">Action</div>
        <h3>GitHub Action</h3>
        <p>Add to your workflow and get annotations directly on pull requests. Zero configuration.</p>
        <pre>- uses: fimskiy/Evil-merge-detector@v1
  with:
    fail-on: warning</pre>
      </div>
      <div class="integration">
        <div class="int-icon">App</div>
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

    <div class="pricing-wrap">
      <div class="plan">
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
          <li class="no"><span class="check">&ndash;</span> Scan history dashboard</li>
          <li class="no"><span class="check">&ndash;</span> Unlimited scans</li>
        </ul>
        <a class="plan-btn" href="https://github.com/apps/evil-merge-detector">Install for free</a>
      </div>

      <div class="plan featured">
        <div class="plan-tier">Pro</div>
        <div class="plan-price">
          <span class="price-currency">$</span>
          <span class="price-amount">7</span>
          <span class="price-period">/month</span>
        </div>
        <div class="plan-desc">For teams and private repositories</div>
        <ul>
          <li><span class="check">&#10003;</span> Public &amp; private repositories</li>
          <li><span class="check">&#10003;</span> Unlimited PR scans</li>
          <li><span class="check">&#10003;</span> GitHub Checks integration</li>
          <li><span class="check">&#10003;</span> Scan history dashboard</li>
          <li><span class="check">&#10003;</span> Email alerts</li>
          <li><span class="check">&#10003;</span> Priority support</li>
        </ul>
        <a class="plan-btn" href="https://github.com/marketplace/evil-merge-detector">Upgrade to Pro</a>
      </div>
    </div>
  </div>
</section>

<!-- FOOTER -->
<footer>
  <div class="footer-logo">EVIL<span class="accent">_</span>MERGE<span class="accent">.</span>DETECT &mdash; open source on <a href="https://github.com/fimskiy/Evil-merge-detector" style="color:#58a6ff">GitHub</a></div>
  <div class="footer-links">
    <a href="/dashboard">Dashboard</a>
    <a href="https://github.com/fimskiy/Evil-merge-detector">Docs</a>
    <a href="https://github.com/apps/evil-merge-detector">Install</a>
  </div>
</footer>

</body>
</html>`
