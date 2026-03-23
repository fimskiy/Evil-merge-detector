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
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#0d1117;color:#e6edf3;line-height:1.6}
a{color:inherit;text-decoration:none}

/* NAV */
nav{display:flex;align-items:center;justify-content:space-between;padding:0 48px;height:64px;border-bottom:1px solid #21262d;position:sticky;top:0;background:#0d1117;z-index:10}
.nav-logo{font-weight:700;font-size:15px}
.nav-logo span{color:#f85149}
.nav-links{display:flex;gap:32px;font-size:14px;color:#8b949e}
.nav-links a:hover{color:#e6edf3}
.nav-cta{background:#238636;color:#fff;padding:7px 16px;border-radius:6px;font-size:14px;font-weight:500}
.nav-cta:hover{background:#2ea043}

/* HERO */
.hero{max-width:780px;margin:0 auto;padding:100px 24px 80px;text-align:center}
.hero-badge{display:inline-block;background:#161b22;border:1px solid #30363d;border-radius:20px;padding:4px 14px;font-size:12px;color:#8b949e;margin-bottom:24px}
.hero-badge span{color:#f85149}
h1{font-size:52px;font-weight:700;line-height:1.15;letter-spacing:-.5px;margin-bottom:20px}
h1 em{color:#f85149;font-style:normal}
.hero-sub{font-size:18px;color:#8b949e;max-width:560px;margin:0 auto 40px}
.hero-actions{display:flex;gap:12px;justify-content:center;flex-wrap:wrap}
.btn-primary{background:#238636;color:#fff;padding:11px 24px;border-radius:6px;font-size:15px;font-weight:500}
.btn-primary:hover{background:#2ea043}
.btn-secondary{background:#21262d;color:#e6edf3;padding:11px 24px;border-radius:6px;font-size:15px;font-weight:500;border:1px solid #30363d}
.btn-secondary:hover{background:#30363d}

/* CODE BLOCK */
.hero-code{margin-top:56px;background:#161b22;border:1px solid #30363d;border-radius:8px;padding:20px 24px;text-align:left;font-family:"SFMono-Regular",Consolas,monospace;font-size:13px;line-height:1.7;max-width:600px;margin-left:auto;margin-right:auto}
.hero-code .c{color:#8b949e}
.hero-code .cmd{color:#79c0ff}
.hero-code .ok{color:#3fb950}
.hero-code .bad{color:#f85149}

/* SECTION */
section{padding:80px 24px}
.section-inner{max-width:960px;margin:0 auto}
.section-label{font-size:12px;font-weight:600;letter-spacing:1px;text-transform:uppercase;color:#f85149;margin-bottom:12px}
h2{font-size:32px;font-weight:700;margin-bottom:16px}
.section-sub{font-size:16px;color:#8b949e;max-width:560px;margin-bottom:48px}

/* HOW IT WORKS */
.steps{display:grid;grid-template-columns:repeat(3,1fr);gap:24px}
.step{background:#161b22;border:1px solid #21262d;border-radius:8px;padding:28px}
.step-num{font-size:12px;font-weight:600;color:#8b949e;margin-bottom:12px}
.step h3{font-size:16px;font-weight:600;margin-bottom:8px}
.step p{font-size:14px;color:#8b949e;line-height:1.6}
.step code{background:#21262d;padding:1px 6px;border-radius:4px;font-size:12px;font-family:monospace;color:#e6edf3}

/* INTEGRATIONS */
.integrations{display:grid;grid-template-columns:repeat(3,1fr);gap:16px;margin-top:48px}
.integration{background:#161b22;border:1px solid #21262d;border-radius:8px;padding:24px}
.integration h3{font-size:15px;font-weight:600;margin-bottom:8px}
.integration p{font-size:13px;color:#8b949e;margin-bottom:16px}
.integration pre{background:#0d1117;border-radius:6px;padding:12px;font-size:12px;font-family:monospace;color:#79c0ff;overflow-x:auto;line-height:1.6}

/* PRICING */
.pricing{display:grid;grid-template-columns:1fr 1fr;gap:24px;max-width:680px}
.plan{background:#161b22;border:1px solid #21262d;border-radius:8px;padding:32px}
.plan.featured{border-color:#238636}
.plan-name{font-size:14px;font-weight:600;color:#8b949e;margin-bottom:8px}
.plan-price{font-size:36px;font-weight:700;margin-bottom:4px}
.plan-price sup{font-size:18px;vertical-align:top;margin-top:6px;display:inline-block}
.plan-price sub{font-size:14px;font-weight:400;color:#8b949e}
.plan-desc{font-size:13px;color:#8b949e;margin-bottom:24px;padding-bottom:24px;border-bottom:1px solid #21262d}
.plan ul{list-style:none;display:flex;flex-direction:column;gap:10px;margin-bottom:28px}
.plan li{font-size:14px;display:flex;gap:8px;align-items:baseline}
.plan li::before{content:"✓";color:#3fb950;font-size:12px;flex-shrink:0}
.plan li.no::before{content:"–";color:#8b949e}
.plan li.no{color:#8b949e}
.plan-btn{display:block;text-align:center;padding:10px;border-radius:6px;font-size:14px;font-weight:500;background:#21262d;border:1px solid #30363d}
.plan-btn:hover{background:#30363d}
.plan.featured .plan-btn{background:#238636;border-color:#238636;color:#fff}
.plan.featured .plan-btn:hover{background:#2ea043}

/* PROBLEM */
.problem-grid{display:grid;grid-template-columns:1fr 1fr;gap:48px;align-items:center}
.problem-text p{color:#8b949e;font-size:15px;margin-bottom:16px}
.problem-text p strong{color:#e6edf3}
.diff-block{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:20px;font-family:monospace;font-size:13px;line-height:1.8}
.diff-block .fn{color:#8b949e;margin-bottom:8px;font-size:12px}
.diff-hash{color:#8b949e}
.diff-clean{color:#3fb950}
.diff-bad{color:#f85149}

/* DIVIDER */
hr{border:none;border-top:1px solid #21262d}

/* FOOTER */
footer{padding:40px 48px;display:flex;align-items:center;justify-content:space-between;border-top:1px solid #21262d;font-size:13px;color:#8b949e}
footer a:hover{color:#e6edf3}

@media(max-width:768px){
  nav{padding:0 20px}.nav-links{display:none}
  h1{font-size:34px}.steps,.integrations,.pricing{grid-template-columns:1fr}
  .problem-grid{grid-template-columns:1fr}
  footer{flex-direction:column;gap:16px;text-align:center}
}
</style>
</head>
<body>

<nav>
  <div class="nav-logo">Evil <span>Merge</span> Detector</div>
  <div class="nav-links">
    <a href="#how-it-works">How it works</a>
    <a href="#pricing">Pricing</a>
    <a href="https://github.com/fimskiy/Evil-merge-detector">GitHub</a>
    <a href="/dashboard">Dashboard</a>
  </div>
  <a class="nav-cta" href="https://github.com/apps/evil-merge-detector">Install App</a>
</nav>

<div class="hero">
  <div class="hero-badge"><span>New</span> — Now available on GitHub Marketplace</div>
  <h1>The merge commit<br>that <em>wasn't</em> reviewed</h1>
  <p class="hero-sub">Evil Merge Detector finds merge commits that introduce changes not present in either parent branch — the attack vector your code review misses.</p>
  <div class="hero-actions">
    <a class="btn-primary" href="https://github.com/apps/evil-merge-detector">Install GitHub App</a>
    <a class="btn-secondary" href="https://github.com/fimskiy/Evil-merge-detector">View on GitHub</a>
  </div>

  <div class="hero-code">
    <div class="c"># Scan your repository</div>
    <div><span class="cmd">$</span> evilmerge scan .</div>
    <div>&nbsp;</div>
    <div class="bad">CRITICAL  ab90bd7  vite.config.js</div>
    <div class="c">          Both parents had identical content.</div>
    <div class="c">          Merge result differs — manual edit detected.</div>
    <div>&nbsp;</div>
    <div class="ok">✓  23 merge commits checked, 1 issue found</div>
  </div>
</div>

<hr>

<!-- PROBLEM -->
<section>
  <div class="section-inner">
    <div class="problem-grid">
      <div class="problem-text">
        <div class="section-label">The problem</div>
        <h2>Hidden in the merge, invisible in the PR</h2>
        <p>When both parent branches contain <strong>identical files</strong>, Git's three-way merge algorithm outputs them unchanged. The only way to get a different result is to manually edit files during the merge.</p>
        <p>GitHub's PR diff doesn't show merge commits. <code>git log</code> doesn't surface the change. SAST tools scan files, not merge history. The injection is invisible.</p>
        <p>This is how malicious code ran undetected in a production repository for <strong>3.5 months</strong> — on every developer machine and every CI build.</p>
      </div>
      <div>
        <div class="diff-block">
          <div class="fn">vite.config.js — merge commit ab90bd7</div>
          <div class="diff-hash">Parent 1 MD5: aa82acb0c335…  ← clean</div>
          <div class="diff-hash">Parent 2 MD5: aa82acb0c335…  ← clean</div>
          <div class="diff-bad">Merge MD5:    2a54754defae…  ← DIFFERENT</div>
          <br>
          <div class="diff-hash">When both parents are identical,</div>
          <div class="diff-hash">Git cannot produce a different</div>
          <div class="diff-hash">output on its own.</div>
        </div>
      </div>
    </div>
  </div>
</section>

<hr>

<!-- HOW IT WORKS -->
<section id="how-it-works">
  <div class="section-inner">
    <div class="section-label">How it works</div>
    <h2>Simple detection, no false positives</h2>
    <p class="section-sub">For each merge commit, we reconstruct what Git should have produced and compare it to what the commit actually contains.</p>

    <div class="steps">
      <div class="step">
        <div class="step-num">Step 01</div>
        <h3>Find the merge base</h3>
        <p>Identify the common ancestor of the two parent commits — the starting point for the three-way merge algorithm.</p>
      </div>
      <div class="step">
        <div class="step-num">Step 02</div>
        <h3>Reconstruct the expected tree</h3>
        <p>Run a clean three-way merge of the parent trees. This is what Git would produce with no manual intervention.</p>
      </div>
      <div class="step">
        <div class="step-num">Step 03</div>
        <h3>Compare against reality</h3>
        <p>Diff the expected tree against the actual merge commit. Any difference is a file that was manually edited during the merge.</p>
      </div>
    </div>
  </div>
</section>

<hr>

<!-- INTEGRATIONS -->
<section>
  <div class="section-inner">
    <div class="section-label">Integrations</div>
    <h2>Works where you already work</h2>
    <p class="section-sub">Three ways to add evil merge detection — pick what fits your workflow.</p>

    <div class="integrations">
      <div class="integration">
        <h3>CLI</h3>
        <p>Scan any repository from the command line. Supports JSON and SARIF output for Code Scanning.</p>
        <pre>brew install fimskiy/tap/evilmerge
evilmerge scan .</pre>
      </div>
      <div class="integration">
        <h3>GitHub Action</h3>
        <p>Add to your workflow and get annotations directly on pull requests.</p>
        <pre>- uses: fimskiy/Evil-merge-detector@v1
  with:
    fail-on: warning</pre>
      </div>
      <div class="integration">
        <h3>GitHub App</h3>
        <p>Install once, get automatic checks on every PR. No workflow changes needed.</p>
        <pre>Install from GitHub Marketplace
→ automatic on every pull request
→ results in GitHub Checks</pre>
      </div>
    </div>
  </div>
</section>

<hr>

<!-- PRICING -->
<section id="pricing">
  <div class="section-inner">
    <div class="section-label">Pricing</div>
    <h2>Simple, per-organization pricing</h2>
    <p class="section-sub">The CLI and GitHub Action are always free and open source.</p>

    <div class="pricing">
      <div class="plan">
        <div class="plan-name">Free</div>
        <div class="plan-price">$0</div>
        <div class="plan-desc">For open source and personal projects</div>
        <ul>
          <li>Public repositories</li>
          <li>50 PR scans / month</li>
          <li>GitHub Checks integration</li>
          <li class="no">Private repositories</li>
          <li class="no">Scan history dashboard</li>
          <li class="no">Unlimited scans</li>
        </ul>
        <a class="plan-btn" href="https://github.com/apps/evil-merge-detector">Install for free</a>
      </div>
      <div class="plan featured">
        <div class="plan-name">Pro</div>
        <div class="plan-price"><sup>$</sup>7<sub>/month</sub></div>
        <div class="plan-desc">For teams and private repositories</div>
        <ul>
          <li>Public &amp; private repositories</li>
          <li>Unlimited PR scans</li>
          <li>GitHub Checks integration</li>
          <li>Private repositories</li>
          <li>Scan history dashboard</li>
          <li>Email alerts</li>
        </ul>
        <a class="plan-btn" href="https://github.com/marketplace/evil-merge-detector">Upgrade to Pro</a>
      </div>
    </div>
  </div>
</section>

<hr>

<footer>
  <div>Evil <span style="color:#f85149">Merge</span> Detector — open source on <a href="https://github.com/fimskiy/Evil-merge-detector" style="color:#58a6ff">GitHub</a></div>
  <div style="display:flex;gap:24px">
    <a href="/dashboard">Dashboard</a>
    <a href="https://github.com/fimskiy/Evil-merge-detector">Docs</a>
    <a href="https://github.com/apps/evil-merge-detector">Install</a>
  </div>
</footer>

</body>
</html>`
