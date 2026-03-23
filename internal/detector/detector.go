package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// sensitivePatterns are file path patterns that elevate severity to CRITICAL.
var sensitivePatterns = []string{
	".env", "credentials", "secret", "password", "token",
	"auth", "crypto", "private", "key", ".pem", ".p12",
}

// Detector analyzes merge commits for evil changes.
type Detector struct{}

func New() *Detector {
	return &Detector{}
}

// commitTrees holds the four trees needed for evil merge analysis.
type commitTrees struct {
	merge *object.Tree
	p1    *object.Tree
	p2    *object.Tree
	base  *object.Tree
	// hashes for display
	p1Hash   string
	p2Hash   string
	baseHash string
}

// extractTrees resolves all trees from a merge commit.
func extractTrees(mergeCommit *object.Commit) (*commitTrees, error) {
	if mergeCommit.NumParents() != 2 {
		return nil, fmt.Errorf("commit %s has %d parents (only 2-parent merges supported)",
			mergeCommit.Hash.String(), mergeCommit.NumParents())
	}

	parent1, err := mergeCommit.Parent(0)
	if err != nil {
		return nil, fmt.Errorf("getting parent 1: %w", err)
	}

	parent2, err := mergeCommit.Parent(1)
	if err != nil {
		return nil, fmt.Errorf("getting parent 2: %w", err)
	}

	bases, err := parent1.MergeBase(parent2)
	if err != nil {
		return nil, fmt.Errorf("finding merge-base: %w", err)
	}

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

	ct := &commitTrees{
		merge:  mergeTree,
		p1:     p1Tree,
		p2:     p2Tree,
		p1Hash: parent1.Hash.String()[:7],
		p2Hash: parent2.Hash.String()[:7],
	}

	if len(bases) > 0 {
		ct.base, err = bases[0].Tree()
		if err != nil {
			return nil, fmt.Errorf("getting base tree: %w", err)
		}
		ct.baseHash = bases[0].Hash.String()[:7]
	}

	return ct, nil
}

// AnalyzeMerge checks a merge commit for evil changes.
func (d *Detector) AnalyzeMerge(ctx context.Context, mergeCommit *object.Commit) (*models.MergeReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ct, err := extractTrees(mergeCommit)
	if err != nil {
		return nil, err
	}
	return d.buildReport(mergeCommit, ct, false)
}

// AnalyzeMergeDetailed is like AnalyzeMerge but also populates EvilChange.Diff
// for each finding. Used by the --commit flag for detailed inspection.
func (d *Detector) AnalyzeMergeDetailed(ctx context.Context, mergeCommit *object.Commit) (*models.MergeReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ct, err := extractTrees(mergeCommit)
	if err != nil {
		return nil, err
	}
	return d.buildReport(mergeCommit, ct, true)
}

func (d *Detector) buildReport(mergeCommit *object.Commit, ct *commitTrees, withDiff bool) (*models.MergeReport, error) {
	parent1, err := mergeCommit.Parent(0)
	if err != nil {
		return nil, fmt.Errorf("getting parent 1: %w", err)
	}
	parent2, err := mergeCommit.Parent(1)
	if err != nil {
		return nil, fmt.Errorf("getting parent 2: %w", err)
	}

	evilChanges, err := d.findEvilChanges(ct, withDiff)
	if err != nil {
		return nil, fmt.Errorf("finding evil changes: %w", err)
	}

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

	for _, ec := range evilChanges {
		if ec.Severity > report.MaxSeverity {
			report.MaxSeverity = ec.Severity
		}
	}

	return report, nil
}

// findEvilChanges compares the merge tree against parents and base.
func (d *Detector) findEvilChanges(ct *commitTrees, withDiff bool) ([]models.EvilChange, error) {
	var changes []models.EvilChange

	mergeFiles := make(map[string]string)
	if err := ct.merge.Files().ForEach(func(f *object.File) error {
		mergeFiles[f.Name] = f.Hash.String()
		return nil
	}); err != nil {
		return nil, fmt.Errorf("iterating merge tree: %w", err)
	}

	p1Files, err := treeFileHashes(ct.p1)
	if err != nil {
		return nil, fmt.Errorf("iterating parent1 tree: %w", err)
	}
	p2Files, err := treeFileHashes(ct.p2)
	if err != nil {
		return nil, fmt.Errorf("iterating parent2 tree: %w", err)
	}
	baseFiles, err := treeFileHashes(ct.base)
	if err != nil {
		return nil, fmt.Errorf("iterating base tree: %w", err)
	}

	for path, mergeHash := range mergeFiles {
		p1Hash, inP1 := p1Files[path]
		p2Hash, inP2 := p2Files[path]
		baseHash, inBase := baseFiles[path]

		if !inP1 && !inP2 && !inBase {
			ec := models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeAdded,
				Severity:   d.classifySeverity(path, true),
				Detail:     "File added only in merge commit, not present in any parent or base",
			}
			if withDiff {
				content, _ := fileContent(ct.merge, path)
				ec.Diff = formatAddedFile(content, ct.merge.Hash.String()[:7])
			}
			changes = append(changes, ec)
			continue
		}

		if mergeHash == p1Hash || mergeHash == p2Hash {
			continue
		}

		if inBase && baseHash == p1Hash && baseHash == p2Hash && mergeHash != baseHash {
			ec := models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   models.SeverityCritical,
				Detail:     "File unchanged in both branches but modified in merge commit",
			}
			if withDiff {
				old, _ := fileContent(ct.p1, path)
				new, _ := fileContent(ct.merge, path)
				ec.Diff = computeDiff(old, new, "P1 "+ct.p1Hash, "M  "+ct.merge.Hash.String()[:7])
			}
			changes = append(changes, ec)
			continue
		}

		changedInP1 := !inBase || baseHash != p1Hash
		changedInP2 := !inBase || baseHash != p2Hash

		if changedInP1 && !changedInP2 {
			ec := models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   d.classifySeverity(path, false),
				Detail:     "File changed only in first parent branch, but merge result differs from both parents",
			}
			if withDiff {
				old, _ := fileContent(ct.p1, path)
				new, _ := fileContent(ct.merge, path)
				ec.Diff = computeDiff(old, new, "P1 "+ct.p1Hash, "M  "+ct.merge.Hash.String()[:7])
			}
			changes = append(changes, ec)
			continue
		}

		if !changedInP1 && changedInP2 {
			ec := models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   d.classifySeverity(path, false),
				Detail:     "File changed only in second parent branch, but merge result differs from both parents",
			}
			if withDiff {
				old, _ := fileContent(ct.p2, path)
				new, _ := fileContent(ct.merge, path)
				ec.Diff = computeDiff(old, new, "P2 "+ct.p2Hash, "M  "+ct.merge.Hash.String()[:7])
			}
			changes = append(changes, ec)
			continue
		}

		if changedInP1 && changedInP2 {
			ec := models.EvilChange{
				FilePath:   path,
				ChangeType: models.ChangeModified,
				Severity:   models.SeverityInfo,
				Detail:     "File changed in both branches (potential conflict resolution), merge differs from both parents",
			}
			if withDiff {
				old, _ := fileContent(ct.p1, path)
				new, _ := fileContent(ct.merge, path)
				ec.Diff = computeDiff(old, new, "P1 "+ct.p1Hash, "M  "+ct.merge.Hash.String()[:7])
			}
			changes = append(changes, ec)
		}
	}

	for path := range p1Files {
		if _, inMerge := mergeFiles[path]; !inMerge {
			if _, inP2 := p2Files[path]; inP2 {
				ec := models.EvilChange{
					FilePath:   path,
					ChangeType: models.ChangeDeleted,
					Severity:   d.classifySeverity(path, false),
					Detail:     "File present in both parents but deleted in merge commit",
				}
				if withDiff {
					content, _ := fileContent(ct.p1, path)
					ec.Diff = formatDeletedFile(content, ct.p1Hash)
				}
				changes = append(changes, ec)
			}
		}
	}

	return changes, nil
}

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

// fileContent returns the text content of a file in the given tree.
// Returns empty string if the file doesn't exist or can't be read.
func fileContent(tree *object.Tree, path string) (string, error) {
	if tree == nil {
		return "", nil
	}
	f, err := tree.File(path)
	if err != nil {
		return "", nil //nolint:nilerr // file absent in this tree is not an error
	}
	return f.Contents()
}

// computeDiff returns a unified-style diff between oldContent and newContent.
func computeDiff(oldContent, newContent, oldLabel, newLabel string) string {
	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(oldContent, newContent)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)

	const maxContext = 3

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n", oldLabel)
	fmt.Fprintf(&sb, "+++ %s\n", newLabel)

	for _, d := range diffs {
		lines := strings.Split(strings.TrimSuffix(d.Text, "\n"), "\n")
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			for _, l := range lines {
				fmt.Fprintf(&sb, "- %s\n", l)
			}
		case diffmatchpatch.DiffInsert:
			for _, l := range lines {
				fmt.Fprintf(&sb, "+ %s\n", l)
			}
		case diffmatchpatch.DiffEqual:
			// Show up to maxContext lines; collapse the rest
			if len(lines) > maxContext*2 {
				for _, l := range lines[:maxContext] {
					fmt.Fprintf(&sb, "  %s\n", l)
				}
				fmt.Fprintf(&sb, "  ... (%d unchanged lines)\n", len(lines)-maxContext*2)
				for _, l := range lines[len(lines)-maxContext:] {
					fmt.Fprintf(&sb, "  %s\n", l)
				}
			} else {
				for _, l := range lines {
					fmt.Fprintf(&sb, "  %s\n", l)
				}
			}
		}
	}

	return sb.String()
}

// formatAddedFile shows a new file added only in the merge commit.
func formatAddedFile(content, mergeHash string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- /dev/null\n")
	fmt.Fprintf(&sb, "+++ M  %s (new file)\n", mergeHash)
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	for _, l := range lines {
		fmt.Fprintf(&sb, "+ %s\n", l)
	}
	return sb.String()
}

// formatDeletedFile shows a file deleted only in the merge commit.
func formatDeletedFile(content, p1Hash string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- P1 %s\n", p1Hash)
	fmt.Fprintf(&sb, "+++ /dev/null (deleted in merge)\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	for _, l := range lines {
		fmt.Fprintf(&sb, "- %s\n", l)
	}
	return sb.String()
}
