package scanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fimskiy/evil-merge-detector/internal/detector"
	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

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

	count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		// Respect context cancellation (Ctrl+C or timeout)
		if err := ctx.Err(); err != nil {
			return err
		}

		if c.NumParents() != 2 {
			return nil
		}

		result.TotalMerges++

		if opts.Limit > 0 && count >= opts.Limit {
			return errLimitReached
		}

		count++

		report, err := s.detector.AnalyzeMerge(ctx, c)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			// Skip problematic merges (e.g., shallow clones)
			return nil
		}

		if len(report.EvilChanges) > 0 && report.MaxSeverity >= opts.MinSeverity {
			var filtered []models.EvilChange
			for _, ec := range report.EvilChanges {
				if ec.Severity >= opts.MinSeverity {
					filtered = append(filtered, ec)
				}
			}
			if len(filtered) > 0 {
				report.EvilChanges = filtered
				result.Reports = append(result.Reports, *report)
				result.EvilMerges++
			}
		}

		return nil
	})

	if err != nil && !errors.Is(err, errLimitReached) {
		return nil, fmt.Errorf("iterating commits: %w", err)
	}

	result.ScanDuration = time.Since(start)
	return result, nil
}
