package detector

import (
	"fmt"
	"strings"

	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// sensitivePatterns are file path patterns that elevate severity to CRITICAL.
var sensitivePatterns = []string{
	".env", "credentials", "secret", "password", "token",
	"auth", "crypto", "private", "key", ".pem", ".p12",
}

// Detector analyzes merge commits for evil changes.
type Detector struct{}

// New creates a new Detector.
func New() *Detector {
	return &Detector{}
}

// AnalyzeMerge checks a merge commit for evil changes by comparing its tree
// against the merge-base and both parents.
func (d *Detector) AnalyzeMerge(mergeCommit *object.Commit) (*models.MergeReport, error) {
	if mergeCommit.NumParents() != 2 {
		return nil, fmt.Errorf("commit %s has %d parents (only 2-parent merges supported)", mergeCommit.Hash.String(), mergeCommit.NumParents())
	}

	parent1, err := mergeCommit.Parent(0)
	if err != nil {
		return nil, fmt.Errorf("getting parent 1: %w", err)
	}

	parent2, err := mergeCommit.Parent(1)
	if err != nil {
		return nil, fmt.Errorf("getting parent 2: %w", err)
	}

	// Get merge-base(s)
	bases, err := parent1.MergeBase(parent2)
	if err != nil {
		return nil, fmt.Errorf("finding merge-base: %w", err)
	}

	// Get trees
	mergeTree, err := mergeCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting merge tree: %w", err)
	}

	p1Tree, err := parent1.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting parent1 tree: %w", err)
	}

	p2Tree, err := parent2.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting parent2 tree: %w", err)
	}

	var baseTree *object.Tree
	if len(bases) > 0 {
		baseTree, err = bases[0].Tree()
		if err != nil {
			return nil, fmt.Errorf("getting base tree: %w", err)
		}
	}

	// Find evil changes
	evilChanges, err := d.findEvilChanges(mergeTree, p1Tree, p2Tree, baseTree)
	if err != nil {
		return nil, fmt.Errorf("finding evil changes: %w", err)
	}

	// Build report
	report := &models.MergeReport{
		CommitHash:   mergeCommit.Hash.String(),
		ShortHash:    mergeCommit.Hash.String()[:7],
		Message:      strings.TrimSpace(mergeCommit.Message),
		Author:       mergeCommit.Author.Name,
		AuthorEmail:  mergeCommit.Author.Email,
		Date:         mergeCommit.Author.When,
		ParentHashes: []string{parent1.Hash.String(), parent2.Hash.String()},
		EvilChanges:  evilChanges,
	}

	// Compute max severity
	for _, ec := range evilChanges {
		if ec.Severity > report.MaxSeverity {
			report.MaxSeverity = ec.Severity
		}
	}

	return report, nil
}

// findEvilChanges compares the merge tree against parents and base to find
// changes that don't belong to either parent branch.
func (d *Detector) findEvilChanges(mergeTree, p1Tree, p2Tree, baseTree *object.Tree) ([]models.EvilChange, error) {
	var changes []models.EvilChange

	// Collect all file paths from merge tree
	mergeFiles := make(map[string]string) // path -> hash
	if err := mergeTree.Files().ForEach(func(f *object.File) error {
		mergeFiles[f.Name] = f.Hash.String()
		return nil
	}); err != nil {
		return nil, fmt.Errorf("iterating merge tree: %w", err)
	}

	// Collect file hashes from parents and base
	p1Files, err := treeFileHashes(p1Tree)
	if err != nil {
		return nil, fmt.Errorf("iterating parent1 tree: %w", err)
	}
	p2Files, err := treeFileHashes(p2Tree)
	if err != nil {
		return nil, fmt.Errorf("iterating parent2 tree: %w", err)
	}

	baseFiles, err := treeFileHashes(baseTree)
	if err != nil {
		return nil, fmt.Errorf("iterating base tree: %w", err)
	}

	// Check each file in the merge commit
	for path, mergeHash := range mergeFiles {
		p1Hash, inP1 := p1Files[path]
		p2Hash, inP2 := p2Files[path]
		baseHash, inBase := baseFiles[path]

		// File exists in merge but not in any parent or base → new file added in merge
		if !inP1 && !inP2 && !inBase {
			changes = append(changes, models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeAdded,
				Severity:   d.classifySeverity(path, true),
				Detail:     "File added only in merge commit, not present in any parent or base",
			})
			continue
		}

		// File matches at least one parent → no evil change for this file
		if mergeHash == p1Hash || mergeHash == p2Hash {
			continue
		}

		// File unchanged across both branches (base == p1 == p2) but different in merge
		if inBase && baseHash == p1Hash && baseHash == p2Hash && mergeHash != baseHash {
			changes = append(changes, models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   models.SeverityCritical,
				Detail:     "File unchanged in both branches but modified in merge commit",
			})
			continue
		}

		// File changed in only one branch, but merge doesn't match either parent
		changedInP1 := !inBase || baseHash != p1Hash
		changedInP2 := !inBase || baseHash != p2Hash

		if changedInP1 && !changedInP2 {
			// Changed only in P1 branch, merge should match P1
			changes = append(changes, models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   d.classifySeverity(path, false),
				Detail:     "File changed only in first parent branch, but merge result differs from both parents",
			})
			continue
		}

		if !changedInP1 && changedInP2 {
			// Changed only in P2 branch, merge should match P2
			changes = append(changes, models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   d.classifySeverity(path, false),
				Detail:     "File changed only in second parent branch, but merge result differs from both parents",
			})
			continue
		}

		// File changed in both branches (conflict zone) and merge differs from both
		if changedInP1 && changedInP2 {
			changes = append(changes, models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   models.SeverityInfo,
				Detail:     "File changed in both branches (potential conflict resolution), merge differs from both parents",
			})
		}
	}

	// Check for files deleted only in the merge (present in both parents but absent in merge)
	for path := range p1Files {
		if _, inMerge := mergeFiles[path]; !inMerge {
			if _, inP2 := p2Files[path]; inP2 {
				// File exists in both parents but was deleted in merge
				changes = append(changes, models.EvilChange{
					FilePath:   path,
					ChangeType: models.ChangeDeleted,
					Severity:   d.classifySeverity(path, false),
					Detail:     "File present in both parents but deleted in merge commit",
				})
			}
		}
	}

	return changes, nil
}

// classifySeverity determines severity based on file path patterns.
func (d *Detector) classifySeverity(path string, isNew bool) models.Severity {
	lower := strings.ToLower(path)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return models.SeverityCritical
		}
	}
	if isNew {
		return models.SeverityCritical
	}
	return models.SeverityWarning
}

// treeFileHashes extracts a map of file path → blob hash from a tree.
func treeFileHashes(tree *object.Tree) (map[string]string, error) {
	files := make(map[string]string)
	if tree == nil {
		return files, nil
	}
	err := tree.Files().ForEach(func(f *object.File) error {
		files[f.Name] = f.Hash.String()
		return nil
	})
	return files, err
}
