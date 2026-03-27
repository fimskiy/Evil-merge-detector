package scanner

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fimskiy/evil-merge-detector/internal/detector"
	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var knownBotNames = []string{
	"dependabot", "renovate-bot", "snyk-bot", "semantic-release-bot",
	"greenkeeper", "imgbot", "allcontributors",
}

func isBot(name, email string) bool {
	combined := strings.ToLower(name + " " + email)
	if strings.Contains(combined, "[bot]") {
		return true
	}
	lower := strings.ToLower(name)
	for _, b := range knownBotNames {
		if lower == b {
			return true
		}
	}
	return false
}

func shouldSkipCommit(c *object.Commit, opts models.ScanOptions) bool {
	hash := c.Hash.String()
	for _, h := range opts.IgnoredHashes {
		if strings.HasPrefix(hash, h) {
			return true
		}
	}
	name := c.Author.Name
	email := c.Author.Email
	for _, a := range opts.IgnoredAuthors {
		if strings.EqualFold(a, name) || strings.EqualFold(a, email) {
			return true
		}
	}
	if opts.IgnoreBots && isBot(name, email) {
		return true
	}
	return false
}

// matchGlob matches a file path against a glob pattern, supporting ** wildcards.
// *.lock matches any file with that extension at any depth.
// dist/** matches everything under dist/.
// src/**/*.js matches .js files anywhere under src/.
func matchGlob(pattern, filePath string) bool {
	fp := filepath.ToSlash(filePath)

	if m, _ := path.Match(pattern, fp); m {
		return true
	}
	if m, _ := path.Match(pattern, path.Base(fp)); m {
		return true
	}
	if !strings.Contains(pattern, "**") {
		return false
	}

	idx := strings.Index(pattern, "**")
	prefix := strings.TrimSuffix(pattern[:idx], "/")
	suffix := strings.TrimPrefix(pattern[idx+2:], "/")

	if prefix != "" && !strings.HasPrefix(fp, prefix+"/") {
		return false
	}
	if suffix == "" {
		return true
	}

	rest := fp
	if prefix != "" {
		rest = fp[len(prefix)+1:]
	}
	if m, _ := path.Match(suffix, rest); m {
		return true
	}
	if m, _ := path.Match(suffix, path.Base(rest)); m {
		return true
	}
	return false
}

func filterEvilChanges(changes []models.EvilChange, exclude, include []string) []models.EvilChange {
	if len(exclude) == 0 && len(include) == 0 {
		return changes
	}
	var out []models.EvilChange
	for _, ec := range changes {
		excluded := false
		for _, pat := range exclude {
			if matchGlob(pat, ec.FilePath) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		if len(include) > 0 {
			matched := false
			for _, pat := range include {
				if matchGlob(pat, ec.FilePath) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, ec)
	}
	return out
}

// resolveTagTime returns the commit time for a tag name.
func resolveTagTime(repo *git.Repository, tagName string) (time.Time, error) {
	ref, err := repo.Tag(tagName)
	if err != nil {
		return time.Time{}, fmt.Errorf("tag %q not found: %w", tagName, err)
	}
	// Try annotated tag first
	tagObj, err := repo.TagObject(ref.Hash())
	if err == nil {
		commit, err := tagObj.Commit()
		if err != nil {
			return time.Time{}, fmt.Errorf("resolving tag %q commit: %w", tagName, err)
		}
		return commit.Committer.When, nil
	}
	// Lightweight tag — hash points directly to a commit
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return time.Time{}, fmt.Errorf("resolving tag %q commit: %w", tagName, err)
	}
	return commit.Committer.When, nil
}

var errLimitReached = errors.New("limit reached")

// Scanner orchestrates repository scanning.
type Scanner struct {
	detector *detector.Detector
}

// New creates a new Scanner.
func New() *Scanner {
	return &Scanner{
		detector: detector.New(),
	}
}

// InspectCommit performs a detailed analysis of a single merge commit by hash.
// It populates EvilChange.Diff for each finding.
func (s *Scanner) InspectCommit(ctx context.Context, repoPath, hash string) (*models.MergeReport, error) {
	if repoPath == "" {
		repoPath = "."
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("opening repository at %s: %w", repoPath, err)
	}

	h := plumbing.NewHash(hash)
	commit, err := repo.CommitObject(h)
	if err != nil {
		return nil, fmt.Errorf("commit %s not found: %w", hash, err)
	}

	return s.detector.AnalyzeMergeDetailed(ctx, commit)
}

// Scan analyzes a repository for evil merges according to the given options.
func (s *Scanner) Scan(ctx context.Context, opts models.ScanOptions) (*models.ScanResult, error) {
	start := time.Now()

	repoPath := opts.RepoPath
	if repoPath == "" {
		repoPath = "."
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("opening repository at %s: %w", repoPath, err)
	}

	// Resolve --since-tag / --until-tag to timestamps
	if opts.SinceTag != "" {
		t, err := resolveTagTime(repo, opts.SinceTag)
		if err != nil {
			return nil, err
		}
		// go-git LogOptions.Since is inclusive (>=), so advance by one second
		// to get commits strictly after the tagged commit.
		t = t.Add(time.Second)
		opts.Since = &t
	}
	if opts.UntilTag != "" {
		t, err := resolveTagTime(repo, opts.UntilTag)
		if err != nil {
			return nil, err
		}
		opts.Until = &t
	}

	logOpts := &git.LogOptions{
		Order: git.LogOrderCommitterTime,
	}

	if opts.Branch != "" {
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(opts.Branch), true)
		if err != nil {
			return nil, fmt.Errorf("resolving branch %s: %w", opts.Branch, err)
		}
		logOpts.From = ref.Hash()
	}

	if opts.Since != nil {
		logOpts.Since = opts.Since
	}
	if opts.Until != nil {
		logOpts.Until = opts.Until
	}

	commitIter, err := repo.Log(logOpts)
	if err != nil {
		return nil, fmt.Errorf("getting commit log: %w", err)
	}

	result := &models.ScanResult{
		RepoPath: repoPath,
		Branch:   opts.Branch,
	}

	// Phase 1: collect qualifying merge commits
	var merges []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if c.NumParents() != 2 || shouldSkipCommit(c, opts) {
			return nil
		}
		result.TotalMerges++
		if opts.Limit > 0 && len(merges) >= opts.Limit {
			return errLimitReached
		}
		merges = append(merges, c)
		return nil
	})
	if err != nil && !errors.Is(err, errLimitReached) {
		return nil, fmt.Errorf("iterating commits: %w", err)
	}

	// Phase 2: analyze commits (parallel if workers > 1)
	reports := make([]*models.MergeReport, len(merges))
	total := len(merges)

	logProgress := func(n int, c *object.Commit) {
		if opts.Progress == nil {
			return
		}
		msg := strings.SplitN(c.Message, "\n", 2)[0]
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		fmt.Fprintf(opts.Progress, "[%d/%d] %s %s\n", n, total, c.Hash.String()[:7], msg)
	}

	workers := opts.Workers
	if workers <= 1 {
		for i, c := range merges {
			if ctx.Err() != nil {
				break
			}
			logProgress(i+1, c)
			report, err := s.detector.AnalyzeMerge(ctx, c)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					break
				}
				continue
			}
			reports[i] = report
		}
	} else {
		type job struct {
			idx    int
			commit *object.Commit
		}
		type res struct {
			idx    int
			report *models.MergeReport
		}
		jobs := make(chan job, workers)
		resCh := make(chan res, workers)

		var (
			wg      sync.WaitGroup
			progMu  sync.Mutex
			progIdx int
		)
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobs {
					progMu.Lock()
					progIdx++
					n := progIdx
					progMu.Unlock()
					logProgress(n, j.commit)

					report, err := s.detector.AnalyzeMerge(ctx, j.commit)
					if err != nil {
						continue
					}
					resCh <- res{j.idx, report}
				}
			}()
		}

		go func() {
			for i, c := range merges {
				select {
				case jobs <- job{i, c}:
				case <-ctx.Done():
					break
				}
			}
			close(jobs)
			wg.Wait()
			close(resCh)
		}()

		for r := range resCh {
			reports[r.idx] = r.report
		}
	}

	// Phase 3: filter and collect findings
	for _, report := range reports {
		if report == nil {
			continue
		}
		if len(report.EvilChanges) == 0 || report.MaxSeverity < opts.MinSeverity {
			continue
		}
		filtered := filterEvilChanges(report.EvilChanges, opts.Exclude, opts.Include)
		var bySeverity []models.EvilChange
		for _, ec := range filtered {
			if ec.Severity >= opts.MinSeverity {
				bySeverity = append(bySeverity, ec)
			}
		}
		if len(bySeverity) == 0 {
			continue
		}
		report.EvilChanges = bySeverity
		report.MaxSeverity = 0
		for _, ec := range bySeverity {
			if ec.Severity > report.MaxSeverity {
				report.MaxSeverity = ec.Severity
			}
		}
		result.Reports = append(result.Reports, *report)
		result.EvilMerges++
	}

	result.ScanDuration = time.Since(start)
	return result, nil
}
